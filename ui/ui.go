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
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/txscript"
	"github.com/btcsuite/btcd/wire"
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
	EVENT_HOST        Event = "host"
	EVENT_JOIN        Event = "join"
	EVENT_JOINED      Event = "joined"
	EVENT_PREIMAGE    Event = "preimage"
	EVENT_SIGN        Event = "sign"
	EVENT_PUBLISH     Event = "publish"
	EVENT_OPEN        Event = "open"
	EVENT_TAKEDEPOSIT Event = "takedeposit"
	EVENT_CHEKC       Event = "check"
	EVENT_WIN         Event = "win"
	EVENT_LOSE        Event = "lose"
)

type UIStateCode int

const (
	UI_STATE_WAIT_DECIDE_MODE           UIStateCode = 0
	UI_STATE_WAIT_PLAYER                UIStateCode = 1
	UI_STATE_WAIT_PREIMAGE_UTXO         UIStateCode = 2
	UI_STATE_WAIT_SIGN                  UIStateCode = 3
	UI_STATE_WAIT_PUBLISH_OR_OPEN       UIStateCode = 4
	UI_STATE_WAIT_OPEN                  UIStateCode = 5
	UI_STATE_WAIT_CHECK_OR_TAKE_DEPOSIT UIStateCode = 6
	UI_STATE_WAIT_WIN_OR_LOSE           UIStateCode = 7
	UI_STATE_WAIT_CLOSE_WIN             UIStateCode = 8
	UI_STATE_WAIT_CLOSE_LOSE            UIStateCode = 9
	UI_STATE_WAIT_CLOSE_WIN2            UIStateCode = 10
)

type UIState struct {
	Code   UIStateCode
	Params []string
}

type ClientGameContext struct {
	GensisPreLockingScript []byte
	GensisPreValue         int64
	GensisPreTxid          string
	GensisPreIndex         int
	PlayerIndex            int
	UnlockScript           []byte
	Hash                   *big.Int
	Preimage               *big.Int
	Number1                *big.Int
	Number2                *big.Int
}

type UIContext struct {
	Id                string
	State             *UIState
	EventChannel      chan *UIEvent
	GameClient        client.Client
	RpcClient         *rpcclient.Client
	GameServer        *server.GameServer
	PrivateKey        *btcec.PrivateKey
	GameContext       *ClientGameContext
	RivalPubkey       *btcec.PublicKey
	GameContractPath  string
	LockContractPath  string
	GenesisMsgTxCache *wire.MsgTx
}

func NewUIContext(config *conf.Config, mode int) *UIContext {
	id := util.RandStringBytesMaskImprSrcUnsafe(8)

	privateKeyByte, err := hex.DecodeString(config.Key)
	if err != nil {
		panic(err)
	}
	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), privateKeyByte)

	ctx := &UIContext{
		Id:               id,
		EventChannel:     make(chan *UIEvent),
		State:            &UIState{Code: UI_STATE_WAIT_DECIDE_MODE},
		GameClient:       nil,
		PrivateKey:       privateKey,
		GameContext:      &ClientGameContext{},
		RpcClient:        NewRpcClient(config.RpcClientConfig),
		LockContractPath: config.LockContractPath,
		GameContractPath: config.GameContractPath,
	}
	Server := server.NewGameServer(config.Listen, config.GameContractPath, config.LockContractPath, ctx.RpcClient, ctx.OnAddParticipant)
	ctx.GameServer = Server
	go ctx.ProcessEventLoop()
	if mode == 0 {
		go ctx.ReadLoop()
	}
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

func (uictx *UIContext) DoEventHost(*UIEvent) {
	if !uictx.CheckStateIn(UI_STATE_WAIT_DECIDE_MODE) {
		panic(errors.New("command not for now"))
	}
	uictx.GameServer.Open()
	internalClient := client.NewInternalClient(uictx.GameServer)
	joinResponse := internalClient.Join(uictx.Id)
	uictx.GameContext.PlayerIndex = joinResponse.Index
	uictx.GameClient = internalClient
	uictx.SetState(UI_STATE_WAIT_PLAYER, nil)
	return
}

func (uictx *UIContext) DoEventJoin(event *UIEvent) {
	if !uictx.CheckStateIn(UI_STATE_WAIT_DECIDE_MODE) {
		panic(errors.New("command not for now"))
	}
	client := client.NewHttpClient(event.Params)
	response := client.Join(uictx.Id)
	uictx.GameClient = client
	uictx.GameContext.PlayerIndex = response.Index
	uictx.SetState(UI_STATE_WAIT_PREIMAGE_UTXO, []string{response.Rival})
	return
}

func (uictx *UIContext) DoEventJoined(event *UIEvent) {
	if !uictx.CheckStateIn(UI_STATE_WAIT_PLAYER) {
		panic(errors.New("command not for now"))
	}
	uictx.SetState(UI_STATE_WAIT_PREIMAGE_UTXO, []string{event.Params})
	return
}

func (uictx *UIContext) DoEventPreimage(event *UIEvent) {
	if !uictx.CheckStateIn(UI_STATE_WAIT_PREIMAGE_UTXO) {
		panic(errors.New("command not for now"))
	}
	preimage, ok := big.NewInt(0).SetString(event.Params, 10)
	if !ok {
		panic(errors.New("wrong number"))
	}

	hash := util.GetHash(preimage)

	pubkey := uictx.PrivateKey.PubKey()

	addr := util.Pubkey2Address(pubkey)

	txid, err := uictx.RpcClient.SendToAddress(addr, server.GENESIS_FAUCET_AMOUNT)
	if err != nil {
		panic(err)
	}

	tx, err := uictx.RpcClient.GetRawTransaction(txid)
	if err != nil {
		panic(err)
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
	uictx.GameContext.GensisPreTxid = txid.String()
	uictx.GameContext.GensisPreIndex = 0

	setUtxoAndHashRequest := &server.SetUtxoAndHashRequest{
		UserId:   uictx.Id,
		Hash:     hash.String(),
		Pubkey:   hex.EncodeToString(pubkey.SerializeCompressed()),
		Pretxid:  txid.String(),
		Preindex: 0,
	}

	uictx.GameClient.SetUtxoAndHash(setUtxoAndHashRequest)
	uictx.SetState(UI_STATE_WAIT_SIGN, nil)
	return
}

func (uictx *UIContext) DoEventSign(event *UIEvent) {
	if !uictx.CheckStateIn(UI_STATE_WAIT_SIGN) {
		panic(errors.New("command not for now"))
	}
	getGenesisTxRequest := &server.GetGenesisTxRequest{
		Sign: false,
	}
	getGenesisTxResponse := uictx.GameClient.GetGenesisTx(getGenesisTxRequest)
	msgtx := util.DeserializeRawTx(getGenesisTxResponse.Rawtx)

	var unlockScript []byte = nil
	for index, vin := range msgtx.TxIn {
		if vin.PreviousOutPoint.Hash.String() != uictx.GameContext.GensisPreTxid ||
			vin.PreviousOutPoint.Index != uint32(uictx.GameContext.GensisPreIndex) {
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
		panic(errors.New("utxo not found"))
	}

	setGenesisTxUnlockScriptRequest := &server.SetGenesisTxUnlockScriptRequest{
		UserId:          uictx.Id,
		UnlockScriptHex: hex.EncodeToString(unlockScript),
	}

	uictx.GameClient.SetGenesisTxUnlockScript(setGenesisTxUnlockScriptRequest)
	uictx.SetState(UI_STATE_WAIT_PUBLISH_OR_OPEN, nil)
	return
}

func (uictx *UIContext) DoEventPublish(event *UIEvent) {
	if !uictx.CheckStateIn(UI_STATE_WAIT_PUBLISH_OR_OPEN) {
		panic(errors.New("command not for now"))
	}

	txid := uictx.GameClient.Publish()
	uictx.SetState(UI_STATE_WAIT_OPEN, []string{txid})
	return
}

func (uictx *UIContext) DoEventOpen(event *UIEvent) {
	if !uictx.CheckStateIn(UI_STATE_WAIT_PUBLISH_OR_OPEN, UI_STATE_WAIT_OPEN) {
		panic(errors.New("command not for now"))
	}

	getGenesisTxRequest := &server.GetGenesisTxRequest{
		Sign: true,
	}
	getGenesisTxResponse := uictx.GameClient.GetGenesisTx(getGenesisTxRequest)

	genesisMsgTx := util.DeserializeRawTx(getGenesisTxResponse.Rawtx)
	uictx.GenesisMsgTxCache = genesisMsgTx

	index := int64(-1)
	if uictx.GameContext.PlayerIndex == 0 {
		index = 1
	} else if uictx.GameContext.PlayerIndex == 1 {
		index = 2
	}

	vout := genesisMsgTx.TxOut[index]

	txInPoint := &util.TxInPoint{
		PreTxid:    genesisMsgTx.TxHash().String(),
		PreIndex:   index,
		Value:      vout.Value,
		LockScript: vout.PkScript,
		HashType:   txscript.SigHashAll | util.SigHashForkID,
	}
	txCtx := util.NewTxContext()

	hashTimeLockOpenUnlockContext := util.NewHashTimeLockOpenUnlockContext(uictx.LockContractPath, uictx.GameContext.Preimage, uictx.PrivateKey)
	txCtx.AddVin(txInPoint, hashTimeLockOpenUnlockContext)

	address := util.PrivateKey2Address(uictx.PrivateKey)

	pkScript, err := txscript.PayToAddrScript(address)
	if err != nil {
		panic(err)
	}

	txCtx.AddVout(server.OPEN_AMOUNT, pkScript)

	openMsgTx := txCtx.Build()

	openMsgTxid, err := uictx.RpcClient.SendRawTransaction(openMsgTx, false)
	if err != nil {
		panic(err)
	}

	setPreimageRequest := &server.SetPreimageRequest{
		UserId:   uictx.Id,
		Preimage: uictx.GameContext.Preimage.String(),
	}
	uictx.GameClient.SetPreimage(setPreimageRequest)
	uictx.SetState(UI_STATE_WAIT_CHECK_OR_TAKE_DEPOSIT, []string{openMsgTxid.String()})
	return
}

func (uictx *UIContext) DoEventTakeDeposit(event *UIEvent) {
	if !uictx.CheckStateIn(UI_STATE_WAIT_CHECK_OR_TAKE_DEPOSIT) {
		panic(errors.New("command not for now"))
	}
	genesisMsgTx := uictx.GenesisMsgTxCache

	index := int64(-1)
	if uictx.GameContext.PlayerIndex == 0 {
		index = 2
	} else if uictx.GameContext.PlayerIndex == 1 {
		index = 1
	}
	vout := genesisMsgTx.TxOut[index]
	txInPoint := &util.TxInPoint{
		PreTxid:    genesisMsgTx.TxHash().String(),
		PreIndex:   index,
		Value:      vout.Value,
		LockScript: vout.PkScript,
		HashType:   txscript.SigHashAll | util.SigHashForkID,
	}
	txCtx := util.NewTxContext()
	txCtx.RpcClient = uictx.RpcClient
	txCtx.LockTime = time.Now().Unix()

	hashTimeLockOverTimeContext := util.NewHashTimeLockOverTimeContext(uictx.LockContractPath, uictx.PrivateKey)
	txCtx.AddVin(txInPoint, hashTimeLockOverTimeContext)

	address := util.PrivateKey2Address(uictx.PrivateKey)

	pkScript, err := txscript.PayToAddrScript(address)
	if err != nil {
		panic(err)
	}

	txCtx.AddVout(server.OPEN_AMOUNT, pkScript)

	openMsgTx := txCtx.Build()

	openMsgTxid, err := uictx.RpcClient.SendRawTransaction(openMsgTx, false)
	if err != nil {
		panic(err)
	}
	uictx.SetState(UI_STATE_WAIT_CLOSE_WIN2, []string{openMsgTxid.String()})
	return
}

func (uictx *UIContext) DoEventCheck(event *UIEvent) {
	if !uictx.CheckStateIn(UI_STATE_WAIT_CHECK_OR_TAKE_DEPOSIT) {
		panic(errors.New("command not for now"))
	}

	request := &server.GetRivalPreimagePubkeyRequest{
		UserId: uictx.Id,
	}
	getRivalPreimageResponse := uictx.GameClient.GetRivalPreimage(request)

	pubkeyByte, err := hex.DecodeString(getRivalPreimageResponse.Pubkey)
	if err != nil {
		panic(err)
	}
	rivalPubkey, err := btcec.ParsePubKey(pubkeyByte, btcec.S256())
	if err != nil {
		panic(err)
	}

	selfPreimage := uictx.GameContext.Preimage
	rivalPreimage, ok := big.NewInt(0).SetString(getRivalPreimageResponse.Preimage, 10)
	if !ok {
		panic(errors.New("err rival preimage"))
	}

	var number1 *big.Int = nil
	var number2 *big.Int = nil
	if uictx.GameContext.PlayerIndex == 0 {
		number1 = selfPreimage
		number2 = rivalPreimage
	} else {
		number2 = selfPreimage
		number1 = rivalPreimage
	}

	cards := util.GetCardStrs(number1, number2)

	selfCards := ""
	rivalCards := ""
	if uictx.GameContext.PlayerIndex == 0 {
		selfCards = cards[0]
		rivalCards = cards[1]
	} else {
		selfCards = cards[1]
		rivalCards = cards[0]
	}

	uictx.GameContext.Number1 = number1
	uictx.GameContext.Number2 = number2

	uictx.RivalPubkey = rivalPubkey
	uictx.SetState(UI_STATE_WAIT_WIN_OR_LOSE, []string{
		selfPreimage.String(),
		selfCards,
		rivalPreimage.String(),
		rivalCards,
	})
	return
}

func (uictx *UIContext) DoEventWin(event *UIEvent) {
	if !uictx.CheckStateIn(UI_STATE_WAIT_WIN_OR_LOSE) {
		panic(errors.New("command not for now"))
	}

	genesisMsgTx := uictx.GenesisMsgTxCache

	factor, err := strconv.ParseInt(event.Params, 10, 64)
	if err != nil {
		panic(err)
	}

	txCtx := &util.TxContext{
		RpcClient:               uictx.RpcClient,
		LockTime:                0,
		Vins:                    make([]*util.VinContext, 0, 2),
		Vouts:                   make([]*wire.TxOut, 0, 1),
		SupplementFeePrivateKey: uictx.PrivateKey,
	}

	selfAddress := util.PrivateKey2Address(uictx.PrivateKey)
	selfScript, err := txscript.PayToAddrScript(selfAddress)
	if err != nil {
		panic(err)
	}
	seflAmount := server.GAMBLING_CAPITAL * (server.MAX_FACTOR + factor)
	txCtx.AddVout(seflAmount, selfScript)

	rivalAddress := util.Pubkey2Address(uictx.RivalPubkey)
	rivalScript, err := txscript.PayToAddrScript(rivalAddress)
	if err != nil {
		panic(err)
	}
	rivalAmount := server.GAMBLING_CAPITAL * (server.MAX_FACTOR - factor)
	txCtx.AddVout(rivalAmount, rivalScript)

	vin := &util.TxInPoint{
		PreTxid:    genesisMsgTx.TxHash().String(),
		PreIndex:   0,
		Value:      genesisMsgTx.TxOut[0].Value,
		LockScript: genesisMsgTx.TxOut[0].PkScript,
		HashType:   txscript.SigHashAll | util.SigHashForkID,
	}

	niuniuV1UnlockCtx := util.NewNiuNiuV1UnlockContext(uictx.GameContractPath, factor, uictx.GameContext.Number1, uictx.GameContext.Number2, uictx.GameContext.Hash)
	txCtx.AddVin(vin, niuniuV1UnlockCtx)

	msgTx := txCtx.SupplementFeeAndBuildByFaucet()

	hash, err := uictx.RpcClient.SendRawTransaction(msgTx, true)
	if err != nil {
		panic(err)
	}
	uictx.SetState(UI_STATE_WAIT_CLOSE_WIN, []string{hash.String()})
	return
}

func (uictx *UIContext) DoEventLose(event *UIEvent) {
	if !uictx.CheckStateIn(UI_STATE_WAIT_WIN_OR_LOSE) {
		panic(errors.New("command not for now"))
	}
	uictx.SetState(UI_STATE_WAIT_CLOSE_LOSE, nil)
	return
}

func (uictx *UIContext) DoEvent(event *UIEvent) {
	switch event.Event {
	case EVENT_HOST:
		uictx.DoEventHost(event)
	case EVENT_JOIN:
		uictx.DoEventJoin(event)
	case EVENT_JOINED:
		uictx.DoEventJoined(event)
	case EVENT_PREIMAGE:
		uictx.DoEventPreimage(event)
	case EVENT_SIGN:
		uictx.DoEventSign(event)
	case EVENT_PUBLISH:
		uictx.DoEventPublish(event)
	case EVENT_OPEN:
		uictx.DoEventOpen(event)
	case EVENT_TAKEDEPOSIT:
		uictx.DoEventTakeDeposit(event)
	case EVENT_CHEKC:
		uictx.DoEventCheck(event)
	case EVENT_WIN:
		uictx.DoEventWin(event)
	case EVENT_LOSE:
		uictx.DoEventLose(event)
	default:
		panic(errors.New("unknown event"))
	}
}

func (uictx *UIContext) HandleState() {
	state := uictx.State
	switch state.Code {
	case UI_STATE_WAIT_DECIDE_MODE:
		fmt.Println(`> command list:
	host 
		Command Description: Only two players are allowed to participate. Take player 1 creating a room as an example. 
		Player 1 enters "host:" and the default room number is "127.0.0.1:10001". At this point, let’s wait for player 2 to enter the room
	join
		Command Description: Player 2 enters "join:127.0.0.1:10001" to join the room
	preimage
		Command Description: Player 1 and Player 2 each enter a large value. For example: "preimage:14058023580860238450283". 
		Only numbers can be entered, entering other characters will result in an error
	sign
		Command Description: If the players have finished executing preimage, type "sign:" to sign
	publish
		Command Description: Only one player is required to enter "publish:" to publish
	open
		Command Description: Each player enters "open:" to show the card
	takedeposit
		Command Description: If one of the players does not open, 
		the other player can enter "takedeposit:" to take the bets of both players
	check
		Command Description: Each player enters "check:" to see the hand of the others. In this way, 
		they can know how much they won or lost and the corresponding multiple
	win
		Command Description: The winning player enters "win: multiple" (e.g. win: 3) to generate a proof and then gets the prize to end the game. 
		An error will be reported if the value entered is incorrect
	lose
		Command Description: There's no need for the losing player to operate anything to end the game, 
		or he or she can also enter "lose:" command to end the game.\n`)
	case UI_STATE_WAIT_PLAYER:
		fmt.Printf("> host a game successful,now please wait for another player\n")
	case UI_STATE_WAIT_PREIMAGE_UTXO:
		fmt.Printf("> we got a player %s,now game start,please input a really really big number\n", state.Params[0])
	case UI_STATE_WAIT_SIGN:
		fmt.Printf("> already set step 1 info ,wait other user set done and you may sign\n")
	case UI_STATE_WAIT_PUBLISH_OR_OPEN:
		fmt.Printf("> already set sign genesis ,wait other user set done and you may publish or open\n")
	case UI_STATE_WAIT_OPEN:
		fmt.Printf("> already publish %s,wait other user open their cards,you may check\n", state.Params[0])
	case UI_STATE_WAIT_CHECK_OR_TAKE_DEPOSIT:
		fmt.Printf("> already open the cards %s,wait other user open their cards,you may check\n", state.Params[0])
	case UI_STATE_WAIT_WIN_OR_LOSE:
		fmt.Printf("> your preimage is %s, got card %s,\n   other player preimage is %s got card %s\n", state.Params[0], state.Params[1], state.Params[2], state.Params[3])
	case UI_STATE_WAIT_CLOSE_WIN:
		fmt.Printf("> game over,you win,second txid is %s\n", state.Params[0])
	case UI_STATE_WAIT_CLOSE_WIN2:
		fmt.Printf("> game over,rival quit,second txid is %s\n", state.Params[0])
	case UI_STATE_WAIT_CLOSE_LOSE:
		fmt.Printf("> game over you lose\n")
	default:
		panic("unknown state")
	}
}

func (uictx *UIContext) ProcessEvent() {
	event := <-uictx.EventChannel
	util.Try(
		func() {
			uictx.DoEvent(event)
			uictx.HandleState()
		},
		func(i interface{}) {
			errMsg := util.GetErrInterfaceMsg(i)
			fmt.Printf("> ProcessEvent DoEvent %s %s\n", event.Event, errMsg)
		},
	)
}

func (uictx *UIContext) ProcessEventLoop() {
	uictx.HandleState()
	for {
		uictx.ProcessEvent()
	}
}
