package ui

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"os"
	"satoshicard/client"
	"satoshicard/conf"
	"satoshicard/server"
	"satoshicard/util"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
)

func NewRpcClient(RpcClientConfig *conf.RpcClientConfig) *rpcclient.Client {
	connCfg := &rpcclient.ConnConfig{
		User:         RpcClientConfig.Username,
		Pass:         RpcClientConfig.Password,
		Host:         RpcClientConfig.Host,
		HTTPPostMode: true,
		DisableTLS:   true,
	}
	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		panic(err)
	}
	return client
}

type Event string

type UIEvent struct {
	Event  Event
	Params string
}

const (
	EVENT_HOST     Event = "host"
	EVENT_JOIN     Event = "join"
	EVENT_JOINED   Event = "joined"
	EVENT_PREIMAGE Event = "preimage"
	EVENT_SIGN     Event = "sign"
	EVENT_PUBLISH  Event = "publish"
	EVENT_OPEN     Event = "open"
	EVENT_CHEKC    Event = "check"
	EVENT_WIN      Event = "win"
	EVENT_LOSE     Event = "lose"
)

type UIStateCode int

const (
	UI_STATE_WAIT_DECIDE_MODE     UIStateCode = 0
	UI_STATE_WAIT_PLAYER          UIStateCode = 1
	UI_STATE_WAIT_PREIMAGE_UTXO   UIStateCode = 2
	UI_STATE_WAIT_SIGN            UIStateCode = 3
	UI_STATE_WAIT_PUBLISH_OR_OPEN UIStateCode = 4
	UI_STATE_WAIT_OPEN            UIStateCode = 5
	UI_STATE_WAIT_CHECK           UIStateCode = 6
	UI_STATE_WAIT_WIN_OR_LOSE     UIStateCode = 7
	UI_STATE_WAIT_CLOSE           UIStateCode = 8
)

type UIState struct {
	Code   UIStateCode
	Params []string
}

type ClientGameContext struct {
	GensisPreLockingScript []byte
	GensisPreValue         int64
	Txid                   string
	Index                  int
	UnlockScript           []byte
	Hash                   *big.Int
	Preimage               *big.Int
}

type UIContext struct {
	Id           string
	State        *UIState
	EventChannel chan *UIEvent
	GameClient   client.Client
	RpcClient    *rpcclient.Client
	GameServer   *server.GameServer
	PrivateKey   *btcec.PrivateKey
	GameContext  *ClientGameContext
}

func NewUIContext(config *conf.Config) *UIContext {
	id := util.RandStringBytesMaskImprSrcUnsafe(8)

	privateKeyByte, err := hex.DecodeString(config.Key)
	if err != nil {
		panic(err)
	}
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyByte)

	ctx := &UIContext{
		Id:           id,
		EventChannel: make(chan *UIEvent),
		State:        &UIState{Code: UI_STATE_WAIT_DECIDE_MODE},
		GameClient:   nil,
		PrivateKey:   privateKey,
		GameContext:  &ClientGameContext{},
		RpcClient:    NewRpcClient(config.RpcClientConfig),
	}
	Server := server.NewGameServer(config.Listen, config.ContractPath, ctx.RpcClient, ctx.OnAddParticipant)
	ctx.GameServer = Server
	go ctx.ProcessEventLoop()
	// go ctx.ReadLoop()
	return ctx
}

func (uictx *UIContext) SetState(code UIStateCode, params []string) {
	uictx.State = &UIState{
		Code:   code,
		Params: params,
	}
}

func (uictx *UIContext) CheckStateIn(codes ...UIStateCode) bool {
	for _, code := range codes {
		if uictx.State.Code == code {
			return true
		}
	}
	return false
}

func (uictx *UIContext) OnAddParticipant(uid string) {
	event := &UIEvent{
		Event:  EVENT_JOINED,
		Params: uid,
	}
	uictx.EventChannel <- event
}

func (uictx *UIContext) ReadLoop() {
	buf := bufio.NewReader(os.Stdin)
	for {
		lineByte, isPrefix, err := buf.ReadLine()
		if err != nil {
			panic(err)
		}
		if isPrefix {
			panic("too long input")
		}
		line := string(lineByte)
		index := strings.Index(line, ":")
		if index == -1 {
			continue
		}
		UIEvent := &UIEvent{
			Event:  Event(line[0:index]),
			Params: line[index+1:],
		}
		uictx.EventChannel <- UIEvent
	}
}

func (uictx *UIContext) DoEventHost(*UIEvent) error {
	if !uictx.CheckStateIn(UI_STATE_WAIT_DECIDE_MODE) {
		return errors.New("command not for now")
	}
	uictx.GameServer.Open()
	internalClient := client.NewInternalClient(uictx.GameServer)
	internalClient.Join(uictx.Id)
	uictx.GameClient = internalClient
	uictx.SetState(UI_STATE_WAIT_PLAYER, nil)
	return nil
}

func (uictx *UIContext) DoEventJoin(event *UIEvent) error {
	if !uictx.CheckStateIn(UI_STATE_WAIT_DECIDE_MODE) {
		return errors.New("command not for now")
	}
	client := client.NewHttpClient(event.Params)
	response, err := client.Join(uictx.Id)
	if err != nil {
		return err
	}
	uictx.GameClient = client
	uictx.SetState(UI_STATE_WAIT_PREIMAGE_UTXO, []string{response.Rival})
	return nil
}

func (uictx *UIContext) DoEventJoined(event *UIEvent) error {
	if !uictx.CheckStateIn(UI_STATE_WAIT_PLAYER) {
		return errors.New("command not for now")
	}
	uictx.SetState(UI_STATE_WAIT_PREIMAGE_UTXO, []string{event.Params})
	return nil
}

func (uictx *UIContext) DoEventPreimage(event *UIEvent) error {
	if !uictx.CheckStateIn(UI_STATE_WAIT_PREIMAGE_UTXO) {
		return errors.New("command not for now")
	}
	preimage, ok := big.NewInt(0).SetString(event.Params, 10)
	if !ok {
		return errors.New("wrong number")
	}

	hash := util.GetHash(preimage)

	pubkey := uictx.PrivateKey.PubKey()

	addr := util.Pubkey2Address(pubkey)

	txid, err := uictx.RpcClient.SendToAddress(addr, server.GAMBLING_CAPITAL*server.MAX_FACTOR+server.EACH_FEE)
	if err != nil {
		return err
	}

	tx, err := uictx.RpcClient.GetRawTransaction(txid)
	if err != nil {
		return err
	}
	for _, vout := range tx.MsgTx().TxOut {
		script, err := txscript.ParsePkScript(vout.PkScript)
		if err != nil {
			continue
		}
		voutAddr, err := script.Address(util.GetNet())
		if err != nil {
			continue
		}
		if voutAddr.String() != addr.String() {
			continue
		}

		uictx.GameContext.GensisPreLockingScript = vout.PkScript
		uictx.GameContext.GensisPreValue = vout.Value
	}

	uictx.GameContext.Hash = hash
	uictx.GameContext.Preimage = preimage
	uictx.GameContext.Txid = txid.String()
	uictx.GameContext.Index = 0

	setUtxoAndHashRequest := &server.SetUtxoAndHashRequest{
		UserId:   uictx.Id,
		Hash:     hash.String(),
		Pubkey:   hex.EncodeToString(pubkey.SerializeCompressed()),
		Pretxid:  txid.String(),
		Preindex: 0,
	}

	err = uictx.GameClient.SetUtxoAndHash(setUtxoAndHashRequest)
	if err != nil {
		return err
	}
	uictx.SetState(UI_STATE_WAIT_SIGN, nil)
	return nil
}

func (uictx *UIContext) DoEventSign(event *UIEvent) error {
	if !uictx.CheckStateIn(UI_STATE_WAIT_SIGN) {
		return errors.New("command not for now")
	}
	getGenesisTxRequest := &server.GetGenesisTxRequest{
		Sign: false,
	}
	getGenesisTxResponse, err := uictx.GameClient.GetGenesisTx(getGenesisTxRequest)
	if err != nil {
		return err
	}
	msgtx := util.DeserializeRawTx(getGenesisTxResponse.Rawtx)
	if err != nil {
		return err
	}

	var unlockScript []byte = nil
	for index, vin := range msgtx.TxIn {
		if vin.PreviousOutPoint.Hash.String() != uictx.GameContext.Txid ||
			vin.PreviousOutPoint.Index != uint32(uictx.GameContext.Index) {
			continue
		}
		unlockScript = util.GetP2PKHUnlockScript(
			msgtx,
			index,
			uictx.PrivateKey,
			uictx.GameContext.GensisPreLockingScript,
			uictx.GameContext.GensisPreValue)
		break
	}

	if unlockScript == nil {
		return errors.New("utxo not found")
	}

	setGenesisTxUnlockScriptRequest := &server.SetGenesisTxUnlockScriptRequest{
		UserId:          uictx.Id,
		UnlockScriptHex: hex.EncodeToString(unlockScript),
	}

	err = uictx.GameClient.SetGenesisTxUnlockScript(setGenesisTxUnlockScriptRequest)
	if err != nil {
		return err
	}
	uictx.SetState(UI_STATE_WAIT_PUBLISH_OR_OPEN, nil)
	return nil
}

func (uictx *UIContext) DoEventPublish(event *UIEvent) error {
	if !uictx.CheckStateIn(UI_STATE_WAIT_PUBLISH_OR_OPEN) {
		return errors.New("command not for now")
	}

	txid, err := uictx.GameClient.Publish()
	if err != nil {
		return err
	}
	fmt.Println(txid)
	uictx.SetState(UI_STATE_WAIT_OPEN, nil)
	return nil
}

func (uictx *UIContext) DoEventOpen() error {
	if !uictx.CheckStateIn(UI_STATE_WAIT_PUBLISH_OR_OPEN) {
		return errors.New("command not for now")
	}
	panic("todo")
}

func (uictx *UIContext) DoEventCheck() error {
	if !uictx.CheckStateIn(UI_STATE_WAIT_CHECK) {
		return errors.New("command not for now")
	}
	panic("todo")
}

func (uictx *UIContext) DoEvent(event *UIEvent) error {
	switch event.Event {
	case EVENT_HOST:
		return uictx.DoEventHost(event)
	case EVENT_JOIN:
		return uictx.DoEventJoin(event)
	case EVENT_JOINED:
		return uictx.DoEventJoined(event)
	case EVENT_PREIMAGE:
		return uictx.DoEventPreimage(event)
	case EVENT_SIGN:
		return uictx.DoEventSign(event)
	case EVENT_PUBLISH:
		return uictx.DoEventPublish(event)
	case EVENT_WIN:
		panic("todo")
	case EVENT_LOSE:
		panic("todo")
	default:
		return errors.New("unknown event")
	}
}

func (uictx *UIContext) HandleState() {
	state := uictx.State
	switch state.Code {
	case UI_STATE_WAIT_DECIDE_MODE:
		fmt.Printf("> join a game or host a game\n")
	case UI_STATE_WAIT_PLAYER:
		fmt.Printf("> host a game successful,now please wait for another player\n")
	case UI_STATE_WAIT_PREIMAGE_UTXO:
		fmt.Printf("> we got a player %s,now game start,please input a really really big number\n", state.Params[0])
	case UI_STATE_WAIT_SIGN:
		fmt.Printf("> already set step 1 info ,wait other user set done and you may sign\n")
	case UI_STATE_WAIT_PUBLISH_OR_OPEN:
		fmt.Printf("> already set sign genesis ,wait other user set done and you may publish or open\n")
	case UI_STATE_WAIT_OPEN:
		fmt.Printf("> already open,wait other user open their cards,you may check\n")
	case UI_STATE_WAIT_CHECK:
		fmt.Printf("> already open the cards,wait other user open their cards,you may check\n")
	case UI_STATE_WAIT_WIN_OR_LOSE:
		fmt.Printf("> you got card %s,other player got card %s\n", state.Params[0], state.Params[1])
	case UI_STATE_WAIT_CLOSE:
		fmt.Printf("> game over\n")
	default:
		panic("unknown state")
	}
}

func (uictx *UIContext) ProcessEvent() {
	event := <-uictx.EventChannel
	err := uictx.DoEvent(event)
	if err != nil {
		fmt.Printf("> ProcessEvent DoEvent %s %s\n", event.Event, err)
		return
	}
	uictx.HandleState()
}

func (uictx *UIContext) ProcessEventLoop() {
	uictx.HandleState()
	for {
		uictx.ProcessEvent()
	}
}
