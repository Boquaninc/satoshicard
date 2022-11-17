package core

import (
	"encoding/hex"
	"errors"
	"math/big"
	"satoshicard/util"
	"sync"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/wire"
	"github.com/sCrypt-Inc/go-scryptlib"
)

const (
	DEFAULT_INDEX    = -1
	DEFAULT_TXID     = ""
	GAMBLING_CAPITAL = 100000000
)

type ParticipantContext struct {
	Id       string
	Hash     *big.Int
	Preimage *big.Int
	Pubkey   *btcec.PublicKey
	Txid     string
	Index    int
}

func NewParticipantContext(id string) *ParticipantContext {
	return &ParticipantContext{
		Id:    id,
		Index: DEFAULT_INDEX,
		Txid:  DEFAULT_TXID,
	}
}

type GameContextState int

const (
	GAME_CONTEXT_STATE_WAIT_OPEN       = 0
	GAME_CONTEXT_STATE_WAIT_PLAYER     = 1
	GAME_CONTEXT_STATE_WAIT_STEP1_INFO = 2
	GAME_CONTEXT_STATE_WAIT_STEP2_INFO = 3
	GAME_CONTEXT_STATE_WAIT_STEP3_INFO = 4
	GAME_CONTEXT_STATE_DONE            = 5
)

type GameContext struct {
	Id                  string
	State               GameContextState
	ParticipantContexts []*ParticipantContext
	WaitPlayerDone      func(string)
	L                   sync.Locker
	Step1ResultChannel  chan string
	Step3ResultChannel  chan map[string]string
	Contract            *scryptlib.Contract
}

func NewGameContext(uid string, contractPath string, WaitPlayerDone func(string)) *GameContext {
	desc, err := scryptlib.LoadDesc(contractPath)
	if err != nil {
		panic(err)
	}

	contract, err := scryptlib.NewContractFromDesc(desc)
	if err != nil {
		panic(err)
	}
	return &GameContext{
		Id:             util.RandStringBytesMaskImprSrcUnsafe(8),
		WaitPlayerDone: WaitPlayerDone,
		ParticipantContexts: []*ParticipantContext{
			NewParticipantContext(uid),
		},
		Contract:           &contract,
		Step1ResultChannel: make(chan string),
		Step3ResultChannel: make(chan map[string]string),
	}
}

func (gameContext *GameContext) SetState(state GameContextState) {
	gameContext.State = state
}

func (gameContext *GameContext) getState() GameContextState {
	return gameContext.State
}

func (gameContext *GameContext) CheckStateAndGame(gameId string, state GameContextState) bool {
	if gameContext.Id != gameId {
		return false
	}
	if gameContext.getState() != state {
		return false
	}
	return true
}

func (ctx *GameContext) Open() error {
	ctx.L.Lock()
	defer ctx.L.Unlock()
	panic("todo")
}

func (ctx *GameContext) GetParticipantContext(id string) (*ParticipantContext, error) {
	for _, ParticipantContext := range ctx.ParticipantContexts {
		if ParticipantContext.Id == id {
			return ParticipantContext, nil
		}
	}
	return nil, errors.New("user not found")
}

func (ctx *GameContext) AddParticipantLock(uid string) (string, error) {
	ctx.L.Lock()
	defer ctx.L.Unlock()
	if !ctx.CheckStateAndGame(ctx.Id, GAME_CONTEXT_STATE_WAIT_PLAYER) {
		return "", errors.New("CheckStateAndGame fail")
	}
	for _, participantContext := range ctx.ParticipantContexts {
		if participantContext.Id == uid {
			return "", errors.New("already in room")
		}
	}
	participantContext := NewParticipantContext(uid)
	ctx.ParticipantContexts = append(ctx.ParticipantContexts, participantContext)
	ctx.SetState(GAME_CONTEXT_STATE_WAIT_STEP1_INFO)
	ctx.WaitPlayerDone(uid)
	return ctx.Id, nil
}

func (ctx *GameContext) GetUnSignGenesisTx() (*wire.MsgTx, error) {
	readyCount := 0
	participantContexts := ctx.ParticipantContexts

	for _, participantContext := range participantContexts {
		if participantContext.Hash != nil &&
			participantContext.Pubkey != nil &&
			participantContext.Index != -1 &&
			participantContext.Txid != "" {
			readyCount++
		}
	}
	if readyCount < 2 {
		return nil, errors.New("not ready")
	}

	playerContexts := ctx.ParticipantContexts
	constructorParams := map[string]scryptlib.ScryptType{
		"hash1": scryptlib.NewIntFromBigInt(playerContexts[0].Hash),
		"hash2": scryptlib.NewIntFromBigInt(playerContexts[1].Hash),
		"user1": scryptlib.NewPubKey(util.ToBecPubkey(playerContexts[0].Pubkey)),
		"user2": scryptlib.NewPubKey(util.ToBecPubkey(playerContexts[1].Pubkey)),
	}
	contract := ctx.Contract
	err := contract.SetConstructorParams(constructorParams)
	if err != nil {
		panic(err)
	}

	script, err := contract.GetLockingScript()
	if err != nil {
		panic(err)
	}
	scriptHex := script.String()

	scriptByte, err := hex.DecodeString(scriptHex)
	if err != nil {
		panic(err)
	}

	msgTx := wire.NewMsgTx(2)
	util.AddVin(msgTx, playerContexts[0].Txid, playerContexts[0].Index)
	util.AddVin(msgTx, playerContexts[1].Txid, playerContexts[1].Index)
	util.AddVout(msgTx, scriptByte, GAMBLING_CAPITAL)
	//todo 找零
	return msgTx, nil
}

func (ctx *GameContext) ProcessStep1() error {
	msgTx, err := ctx.GetUnSignGenesisTx()
	if err != nil {
	}
	rawTx := util.SeserializeMsgTx(msgTx)

	ctx.Step1ResultChannel <- rawTx
	ctx.Step1ResultChannel <- rawTx
	close(ctx.Step1ResultChannel)
	return nil
}

type Step1Info struct {
	Txid   string
	Index  int
	Hash   string
	Pubkey string
}

func (ctx *GameContext) SetStep1InfoLock(uid string, gameid string, step1Info *Step1Info) (chan string, error) {
	if ctx.Id != gameid {
		return nil, errors.New("not this game")
	}
	ctx.L.Lock()
	defer ctx.L.Unlock()
	if !ctx.CheckStateAndGame(gameid, GAME_CONTEXT_STATE_WAIT_STEP1_INFO) {
		return nil, errors.New("CheckStateAndGame fail")
	}
	participantContext, err := ctx.GetParticipantContext(uid)
	if err != nil {
		return nil, err
	}
	if participantContext.Txid != "" {
		return nil, errors.New("step1 info already set")
	}

	pubkeyByte, err := hex.DecodeString(step1Info.Pubkey)
	if err != nil {
		return nil, err
	}
	pubkey, err := btcec.ParsePubKey(pubkeyByte, btcec.S256())
	if err != nil {
		return nil, err
	}

	hash, ok := big.NewInt(0).SetString(step1Info.Hash, 10)
	if !ok {
		return nil, errors.New("error hash")
	}

	participantContext.Txid = step1Info.Txid
	participantContext.Index = step1Info.Index
	participantContext.Pubkey = pubkey
	participantContext.Hash = hash
	err = ctx.ProcessStep1()
	if err == nil {
		ctx.SetState(GAME_CONTEXT_STATE_WAIT_STEP2_INFO)
	}
	return ctx.Step1ResultChannel, nil
}
func (ctx *GameContext) SetStep1InfoAndWaitRawTxLock(uid string, gameid string, step1Info *Step1Info) (string, error) {
	receiveResultChan, err := ctx.SetStep1InfoLock(uid, gameid, step1Info)
	if err != nil {
		return "", err
	}
	rawTx, ok := <-receiveResultChan
	if !ok {
		return "", errors.New("never receive any result")
	}
	return rawTx, nil
}

func (ctx *GameContext) ProcessStep3() error {
	readyCount := 0
	for _, participantContext := range ctx.ParticipantContexts {
		if participantContext.Preimage != nil {
			readyCount++
		}
	}
	if readyCount < 2 {
		return errors.New("not ready")
	}
	uid2PreimageStr := map[string]string{}

	ctx.Step3ResultChannel <- uid2PreimageStr
	ctx.Step3ResultChannel <- uid2PreimageStr
	close(ctx.Step3ResultChannel)
	return nil
}

func (ctx *GameContext) SetStep3InfoLock(uid string, gameid string, preimageStr string) (chan map[string]string, error) {
	preimage, ok := big.NewInt(0).SetString(preimageStr, 10)
	if !ok {
		return nil, errors.New("error preimage")
	}
	hash := util.GetHash(preimage)
	ctx.L.Lock()
	defer ctx.L.Unlock()
	if !ctx.CheckStateAndGame(gameid, GAME_CONTEXT_STATE_WAIT_STEP2_INFO) {
		return nil, errors.New("CheckStateAndGame fail")
	}
	participantContext, err := ctx.GetParticipantContext(uid)
	if err != nil {
		return nil, err
	}

	if participantContext.Preimage != nil {
		return nil, errors.New("preimage already set")
	}

	if participantContext.Hash.Cmp(hash) != 0 {
		return nil, errors.New("cant be right preimage")
	}
	participantContext.Preimage = preimage
	err = ctx.ProcessStep3()
	if err == nil {
		ctx.SetState(GAME_CONTEXT_STATE_DONE)
	}
	return ctx.Step3ResultChannel, nil
}

func (ctx *GameContext) SetStep3InfoAndWaitPreimageLock(uid string, gameid string, preimageStr string) (string, error) {
	receiveResultChan, err := ctx.SetStep3InfoLock(uid, gameid, preimageStr)
	if err != nil {
		return "", err
	}
	result := <-receiveResultChan
	otherPreimage := ""
	for key, value := range result {
		if key == uid {
			continue
		}
		otherPreimage = value
		break
	}
	return otherPreimage, nil
}
