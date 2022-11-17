package core

import (
	"encoding/hex"
	"errors"
	"math/big"
	"satoshicard/util"
	"sync"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/sCrypt-Inc/go-scryptlib"
)

const (
	DEFAULT_INDEX    = -1
	DEFAULT_TXID     = ""
	GAMBLING_CAPITAL = 100000000
)

type ParticipantContext struct {
	Id           string
	Hash         *big.Int
	Preimage     *big.Int
	Pubkey       *btcec.PublicKey
	Txid         string
	Index        int
	UnlockScript []byte
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
	ParticipantContexts []*ParticipantContext
	L                   sync.Locker
	Contract            *scryptlib.Contract
	ContractPath        string
	BitcoinCli          *rpcclient.Client
	OnAddParticipant    func(string)
}

func NewGameContext(uid string, contractPath string, OnAddParticipant func(string)) *GameContext {
	desc, err := scryptlib.LoadDesc(contractPath)
	if err != nil {
		panic(err)
	}

	contract, err := scryptlib.NewContractFromDesc(desc)
	if err != nil {
		panic(err)
	}
	return &GameContext{
		Id: util.RandStringBytesMaskImprSrcUnsafe(8),
		ParticipantContexts: []*ParticipantContext{
			NewParticipantContext(uid),
		},
		Contract:         &contract,
		ContractPath:     contractPath,
		L:                &sync.Mutex{},
		OnAddParticipant: OnAddParticipant,
	}
}

func (ctx *GameContext) Clear() {
	ctx = NewGameContext(ctx.Id, ctx.ContractPath, ctx.OnAddParticipant)
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
	if uid == "" {
		return "", errors.New("empty uid")
	}
	ctx.L.Lock()
	defer ctx.L.Unlock()
	if len(ctx.ParticipantContexts) > 2 {
		return "", errors.New("room already full")
	}
	rivalUid := ""
	for _, participantContext := range ctx.ParticipantContexts {
		if participantContext.Id == uid {
			return "", errors.New("already in room")
		}
		rivalUid = participantContext.Id
	}
	participantContext := NewParticipantContext(uid)
	ctx.ParticipantContexts = append(ctx.ParticipantContexts, participantContext)
	util.PrintJson(ctx.ParticipantContexts)
	ctx.OnAddParticipant(uid)
	return rivalUid, nil
}

type UtxoAndHash struct {
	Txid   string
	Index  int
	Hash   string
	Pubkey string
}

func (ctx *GameContext) SetUtxoAndHashLock(uid string, step1Info *UtxoAndHash) error {
	ctx.L.Lock()
	defer ctx.L.Unlock()
	participantContext, err := ctx.GetParticipantContext(uid)
	if err != nil {
		return err
	}
	if participantContext.Txid != "" {
		return errors.New("step1 info already set")
	}

	pubkeyByte, err := hex.DecodeString(step1Info.Pubkey)
	if err != nil {
		return err
	}
	pubkey, err := btcec.ParsePubKey(pubkeyByte, btcec.S256())
	if err != nil {
		return err
	}

	hash, ok := big.NewInt(0).SetString(step1Info.Hash, 10)
	if !ok {
		return errors.New("error hash")
	}

	participantContext.Txid = step1Info.Txid
	participantContext.Index = step1Info.Index
	participantContext.Pubkey = pubkey
	participantContext.Hash = hash
	return nil
}

func (ctx *GameContext) GetGenesisTxLock(sign bool) (*wire.MsgTx, error) {
	readyCount := 0
	ctx.L.Lock()
	defer ctx.L.Unlock()
	participantContexts := ctx.ParticipantContexts
	for _, participantContext := range participantContexts {
		if participantContext.Hash != nil &&
			participantContext.Pubkey != nil &&
			participantContext.Index != -1 &&
			participantContext.Txid != "" && (!sign || participantContext.UnlockScript != nil) {
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
	util.AddVin(msgTx, playerContexts[0].Txid, playerContexts[0].Index, playerContexts[0].UnlockScript)
	util.AddVin(msgTx, playerContexts[1].Txid, playerContexts[1].Index, playerContexts[1].UnlockScript)
	util.AddVout(msgTx, scriptByte, GAMBLING_CAPITAL)
	return msgTx, nil
}

func (ctx *GameContext) SetUnlockScriptLock(uid string, sig []byte) error {
	ctx.L.Lock()
	defer ctx.L.Unlock()
	participantContext, err := ctx.GetParticipantContext(uid)
	if err != nil {
		return err
	}
	if participantContext.UnlockScript != nil {
		return errors.New("unlockScript already set")
	}
	participantContext.UnlockScript = sig
	return nil
}

func (ctx *GameContext) SetPreimageLock(uid string, preimageStr string) error {
	preimage, ok := big.NewInt(0).SetString(preimageStr, 10)
	if !ok {
		return errors.New("error preimage")
	}
	hash := util.GetHash(preimage)
	ctx.L.Lock()
	defer ctx.L.Unlock()
	participantContext, err := ctx.GetParticipantContext(uid)
	if err != nil {
		return err
	}

	if participantContext.Preimage != nil {
		return errors.New("preimage already set")
	}

	if participantContext.Hash.Cmp(hash) != 0 {
		return errors.New("cant be right preimage")
	}
	participantContext.Preimage = preimage
	return nil
}

func (ctx *GameContext) GetRivalPreimage(uid string) (*big.Int, error) {
	ctx.L.Lock()
	defer ctx.L.Unlock()
	for _, participantContext := range ctx.ParticipantContexts {
		if participantContext.Id == uid {
			continue
		}
		if participantContext.Preimage == nil {
			return nil, errors.New("rival preimage not set")
		}
		return participantContext.Preimage, nil
	}
	return nil, errors.New("rival not found")
}

func (ctx *GameContext) SendGenesisLock() error {
	msgTx, err := ctx.GetGenesisTxLock(true)
	if err != nil {
		return err
	}
	_, err = ctx.BitcoinCli.SendRawTransaction(msgTx, true)
	if err != nil {
		return err
	}
	return nil
}
