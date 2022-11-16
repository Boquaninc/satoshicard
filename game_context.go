package main

import (
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/sCrypt-Inc/go-scryptlib"
)

// type User struct {
// 	Key *btcec.PrivateKey
// }

// func (this *User) Address() btcutil.Address {
// 	key := this.Key.PubKey()
// 	pkHash := btcutil.Hash160(key.SerializeCompressed())
// 	addr, err := btcutil.NewAddressPubKeyHash(pkHash, gNet)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return addr
// }

// func (this *User) Pubkey() *btcec.PublicKey {
// 	return this.Key.PubKey()
// }

// func (this *User) BecPubkey() *bec.PublicKey {
// 	return (*bec.PublicKey)(this.Key.PubKey().ToECDSA())
// }

// func NewUser(key string) *User {
// 	keyByte, err := hex.DecodeString(key)
// 	if err != nil {
// 		panic(err)
// 	}
// 	privateKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), keyByte)
// 	return &User{
// 		Key: privateKey,
// 	}
// }

type PlayerContext struct {
	Id       string
	Txid     string
	Index    int
	Pubkey   *btcec.PublicKey
	Preimage *big.Int
	Hash     *big.Int
}

func NewPlayerContext(id string) *PlayerContext {
	return &PlayerContext{
		Id: id,
	}
}

type GameContextState int

const (
	WAIT_PLAYERS    GameContextState = 0
	WAIT_STEP1_INFO GameContextState = 1
)

type GameContext struct {
	Id                 string
	State              GameContextState
	PlayerContextSet   map[string]*PlayerContext
	Contract           *scryptlib.Contract
	Step1ResultChannel chan string
}

func NewGameContext(host string, contractPath string) *GameContext {
	desc, err := scryptlib.LoadDesc(contractPath)
	if err != nil {
		panic(err)
	}

	contract, err := scryptlib.NewContractFromDesc(desc)
	if err != nil {
		panic(err)
	}
	return &GameContext{
		Id:    RandStringBytesMaskImprSrcUnsafe(8),
		State: WAIT_PLAYERS,
		PlayerContextSet: map[string]*PlayerContext{
			host: NewPlayerContext(host),
		},
		Step1ResultChannel: make(chan string, 2),
		Contract:           &contract,
	}
}

func (gameContext *GameContext) AddPlayer(id string) error {
	if gameContext.GetState() != WAIT_PLAYERS {
		return errors.New("game already start or room is full")
	}
	playerContext := NewPlayerContext(id)
	if len(gameContext.PlayerContextSet) > 2 {
		panic("len(gameContext.PlayerContextSet) > 2")
	}
	gameContext.PlayerContextSet[playerContext.Id] = playerContext
	if len(gameContext.PlayerContextSet) != 2 {
		return nil
	}
	gameContext.SetState(WAIT_STEP1_INFO)
	return nil
}

type Step1Info struct {
	Txid   string
	Index  int
	Hash   string
	Pubkey string
}

func (gameContext *GameContext) SetStep1Info(id string, step1Info *Step1Info) (chan string, error) {
	if gameContext.GetState() != WAIT_STEP1_INFO {
		return nil, errors.New("gameContext.GetState() != WAIT_STEP1_INFO")
	}
	playerContext, ok := gameContext.PlayerContextSet[id]
	if !ok {
		panic("player not in the room")
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

	playerContext.Txid = step1Info.Txid
	playerContext.Index = step1Info.Index
	playerContext.Pubkey = pubkey
	playerContext.Hash = hash
	return gameContext.Step1ResultChannel, nil
}

func (gameContext *GameContext) ProcessStep1() {
	readyCount := 0
	for _, playerContext := range gameContext.PlayerContextSet {
		if playerContext.Hash != nil &&
			playerContext.Pubkey != nil &&
			playerContext.Index != -1 &&
			playerContext.Txid != "" {
			readyCount++
		}
	}
	if readyCount < 2 {
		return
	}
	//todo
	gameContext.Step1ResultChannel <- ""
	gameContext.Step1ResultChannel <- ""
}

func (gameContext *GameContext) SetState(state GameContextState) {
	gameContext.State = state
}

func (gameContext *GameContext) GetState() GameContextState {
	return gameContext.State
}
