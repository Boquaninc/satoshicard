package server

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/big"
	"net/http"
	"satoshicard/util"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/sCrypt-Inc/go-scryptlib"
)

const (
	GAMBLING_CAPITAL = 10000000
	MAX_FACTOR       = 10
	EACH_FEE         = 1000000

	EACH_GAME_AMOUNT = GAMBLING_CAPITAL * MAX_FACTOR
	EACH_LOCK_AMOUNT = GAMBLING_CAPITAL * MAX_FACTOR

	GAME_VOUT_AMOUNT = EACH_GAME_AMOUNT * 2

	GENESIS_FAUCET_AMOUNT = EACH_GAME_AMOUNT + EACH_LOCK_AMOUNT + EACH_FEE

	JOIN_URI                         = "/join"
	SET_UTXO_AND_HASH_URI            = "/set_utxo_and_hash"
	GET_GENESIS_TX_URI               = "/get_genesis_tx"
	SET_GENESIS_TX_UNLOCK_SCRIPT_URI = "/set_genesis_tx_unlock_script"
	PUBLISH_URI                      = "/publish"
	SET_PREIMAGE_URI                 = "/set_preimage"
	GET_RIVAL_PREIMAGE_PUBKEY_URI    = "/get_rival_preimage_pubkey"
)

type Request interface {
	JoinRequest |
		SetUtxoAndHashRequest |
		GetGenesisTxRequest |
		SetGenesisTxUnlockScriptRequest |
		PublishRequest |
		SetPreimageRequest |
		GetRivalPreimagePubkeyRequest
}

type Response interface {
	JoinResponse |
		SetUtxoAndHashResponse |
		GetGenesisTxResponse |
		SetGenesisTxUnlockScriptResponse |
		PulishResponse |
		SetPreimageResponse |
		GetRivalPreimagePubkeyResponse
}

func HttpAspect[T1 Request, T2 Response](
	handler func(request *T1) (*T2, error),
) func(rsp http.ResponseWriter, req *http.Request) {
	return func(rsp http.ResponseWriter, req *http.Request) {
		request := new(T1)
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		if len(body) == 0 {
			body = []byte("{}")
		}
		err = json.Unmarshal(body, request)
		if err != nil {
			return
		}
		response, err := handler(request)
		if err != nil {
			rsp.Write(util.MakeHttpJsonResponseByError(err, nil))
			return
		}
		rsp.Write(util.MakeHttpJsonResponseByInterface(response))
		return
	}
}

type ParticipantContext struct {
	Id           string
	Hash         *big.Int
	Preimage     *big.Int
	Pubkey       *btcec.PublicKey
	Txid         string
	Index        int
	UnlockScript []byte
}

const (
	DEFAULT_INDEX = -1
	DEFAULT_TXID  = ""
)

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

type GameServer struct {
	Id                      string
	ParticipantContexts     []*ParticipantContext
	L                       sync.Locker
	GameContract            *scryptlib.Contract
	GameContractPath        string
	LockContract            *scryptlib.Contract
	LockContractPath        string
	RpcClient               *rpcclient.Client
	OnAddParticipant        func(string)
	Server                  *http.Server
	GenesisTxid             string
	UnSignGenesisMsgTxCache *wire.MsgTx
	SignGenesisMsgTxCache   *wire.MsgTx
}

func NewGameServer(listen string, gameContractPath string, lockContractPath string, rpcClient *rpcclient.Client, OnAddParticipant func(string)) *GameServer {
	gameDesc, err := scryptlib.LoadDesc(gameContractPath)
	if err != nil {
		panic(err)
	}

	gameContract, err := scryptlib.NewContractFromDesc(gameDesc)
	if err != nil {
		panic(err)
	}

	lcokDesc, err := scryptlib.LoadDesc(lockContractPath)
	if err != nil {
		panic(err)
	}

	lockContract, err := scryptlib.NewContractFromDesc(lcokDesc)
	if err != nil {
		panic(err)
	}
	server := &GameServer{
		Id:                  util.RandStringBytesMaskImprSrcUnsafe(8),
		ParticipantContexts: []*ParticipantContext{},
		GameContract:        &gameContract,
		GameContractPath:    gameContractPath,
		LockContract:        &lockContract,
		LockContractPath:    lockContractPath,
		L:                   &sync.Mutex{},
		OnAddParticipant:    OnAddParticipant,
		Server:              &http.Server{Addr: listen},
		RpcClient:           rpcClient,
	}
	http.HandleFunc(JOIN_URI, HttpAspect(server.JoinLock))
	http.HandleFunc(SET_UTXO_AND_HASH_URI, HttpAspect(server.SetUtxoAndHashLock))
	http.HandleFunc(GET_GENESIS_TX_URI, HttpAspect(server.GetGenesisTxLock))
	http.HandleFunc(SET_GENESIS_TX_UNLOCK_SCRIPT_URI, HttpAspect(server.SetGenesisTxUnlockScriptLock))
	http.HandleFunc(PUBLISH_URI, HttpAspect(server.PublishLock))
	http.HandleFunc(SET_PREIMAGE_URI, HttpAspect(server.SetPreimageLock))
	http.HandleFunc(GET_RIVAL_PREIMAGE_PUBKEY_URI, HttpAspect(server.GetRivalPreimagePubkeyLock))
	return server
}

func (gameServer *GameServer) GetParticipantContext(id string) (*ParticipantContext, error) {
	for _, ParticipantContext := range gameServer.ParticipantContexts {
		if ParticipantContext.Id == id {
			return ParticipantContext, nil
		}
	}
	return nil, errors.New("user not found")
}

func (gameServer *GameServer) Close() {
	err := gameServer.Server.Close()
	if err != nil {
		panic(err)
	}
}

func (gameServer *GameServer) Open() {
	go gameServer.Server.ListenAndServe()
}

type JoinRequest struct {
	Id string `json:"id"`
}

type JoinResponse struct {
	Rival string `json:"rival"`
	Index int    `json:"index"`
}

func (gameServer *GameServer) JoinLock(request *JoinRequest) (*JoinResponse, error) {
	if request.Id == "" {
		return nil, errors.New("empty uid")
	}
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	if len(gameServer.ParticipantContexts) >= 2 {
		return nil, errors.New("room already full")
	}
	rivalUid := ""
	for _, participantContext := range gameServer.ParticipantContexts {
		if participantContext.Id == request.Id {
			return nil, errors.New("already in room")
		}
		rivalUid = participantContext.Id
	}
	participantContext := NewParticipantContext(request.Id)
	gameServer.ParticipantContexts = append(gameServer.ParticipantContexts, participantContext)
	if len(gameServer.ParticipantContexts) >= 2 {
		go gameServer.OnAddParticipant(request.Id)
	}
	return &JoinResponse{
		Rival: rivalUid,
		Index: len(gameServer.ParticipantContexts) - 1,
	}, nil
}

type SetUtxoAndHashRequest struct {
	UserId   string `json:"user_id"`
	Hash     string `json:"hash"`
	Pubkey   string `json:"pubkey"`
	Pretxid  string `json:"pretxid"`
	Preindex int    `json:"preindex"`
}

type SetUtxoAndHashResponse struct {
}

func (gameServer *GameServer) SetUtxoAndHashLock(request *SetUtxoAndHashRequest) (*SetUtxoAndHashResponse, error) {
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	participantContext, err := gameServer.GetParticipantContext(request.UserId)
	if err != nil {
		return nil, err
	}
	if participantContext.Txid != "" {
		return nil, errors.New("step1 info already set")
	}

	pubkeyByte, err := hex.DecodeString(request.Pubkey)
	if err != nil {
		return nil, err
	}
	pubkey, err := btcec.ParsePubKey(pubkeyByte, btcec.S256())
	if err != nil {
		return nil, err
	}

	hash, ok := big.NewInt(0).SetString(request.Hash, 10)
	if !ok {
		return nil, errors.New("error hash")
	}

	participantContext.Txid = request.Pretxid
	participantContext.Index = request.Preindex
	participantContext.Pubkey = pubkey
	participantContext.Hash = hash
	return &SetUtxoAndHashResponse{}, nil
}

type GetGenesisTxRequest struct {
	Sign bool `json:"sign"`
}

type GetGenesisTxResponse struct {
	Rawtx string `json:"rawtx"`
}

func GetConstructorLockScript(constructorParams map[string]scryptlib.ScryptType, contract *scryptlib.Contract) []byte {
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
	return scriptByte
}

func (gameServer *GameServer) GetGenesisMsgTx(sign bool) (*wire.MsgTx, error) {
	readyCount := 0
	participantContexts := gameServer.ParticipantContexts
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

	if sign && gameServer.SignGenesisMsgTxCache != nil {
		return gameServer.SignGenesisMsgTxCache, nil
	}

	if sign && gameServer.SignGenesisMsgTxCache == nil {
		msgTx := gameServer.UnSignGenesisMsgTxCache.Copy()
		msgTx.TxIn[0].SignatureScript = gameServer.ParticipantContexts[0].UnlockScript
		msgTx.TxIn[1].SignatureScript = gameServer.ParticipantContexts[1].UnlockScript
		gameServer.SignGenesisMsgTxCache = msgTx
		return gameServer.SignGenesisMsgTxCache, nil
	}

	if !sign && gameServer.UnSignGenesisMsgTxCache != nil {
		return gameServer.UnSignGenesisMsgTxCache, nil
	}

	playerContexts := gameServer.ParticipantContexts
	msgTx := wire.NewMsgTx(2)

	gameConstructorParams := map[string]scryptlib.ScryptType{
		"hash1":     scryptlib.NewIntFromBigInt(playerContexts[0].Hash),
		"hash2":     scryptlib.NewIntFromBigInt(playerContexts[1].Hash),
		"maxfactor": scryptlib.NewIntFromBigInt(big.NewInt(MAX_FACTOR)),
		"user1":     scryptlib.NewPubKey(util.ToBecPubkey(playerContexts[0].Pubkey)),
		"user2":     scryptlib.NewPubKey(util.ToBecPubkey(playerContexts[1].Pubkey)),
	}

	gameScriptByte := GetConstructorLockScript(gameConstructorParams, gameServer.GameContract)
	util.AddVout(msgTx, gameScriptByte, GAME_VOUT_AMOUNT)

	matureTime := time.Now().Unix() + 60*60
	lockConstructorParams1 := map[string]scryptlib.ScryptType{
		"matureTime":   scryptlib.NewInt(matureTime),
		"preimageHash": scryptlib.NewIntFromBigInt(playerContexts[0].Hash),
		"pubkey":       scryptlib.NewPubKey(util.ToBecPubkey(playerContexts[1].Pubkey)),
	}

	lockScriptByte1 := GetConstructorLockScript(lockConstructorParams1, gameServer.LockContract)
	util.AddVout(msgTx, lockScriptByte1, EACH_LOCK_AMOUNT)

	lockConstructorParams2 := map[string]scryptlib.ScryptType{
		"matureTime":   scryptlib.NewInt(matureTime),
		"preimageHash": scryptlib.NewIntFromBigInt(playerContexts[1].Hash),
		"pubkey":       scryptlib.NewPubKey(util.ToBecPubkey(playerContexts[0].Pubkey)),
	}

	lockScriptByte2 := GetConstructorLockScript(lockConstructorParams2, gameServer.LockContract)
	util.AddVout(msgTx, lockScriptByte2, EACH_LOCK_AMOUNT)

	util.AddVin(msgTx, playerContexts[0].Txid, playerContexts[0].Index, nil)
	util.AddVin(msgTx, playerContexts[1].Txid, playerContexts[1].Index, nil)
	gameServer.UnSignGenesisMsgTxCache = msgTx
	return gameServer.UnSignGenesisMsgTxCache, nil
}

func (gameServer *GameServer) GetGenesisMsgTxLock(sign bool) (*wire.MsgTx, error) {
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	return gameServer.GetGenesisMsgTx(sign)
}

func (gameServer *GameServer) GetGenesisTxLock(request *GetGenesisTxRequest) (*GetGenesisTxResponse, error) {
	msgTx, err := gameServer.GetGenesisMsgTxLock(request.Sign)
	if err != nil {
		return nil, err
	}
	return &GetGenesisTxResponse{
		Rawtx: util.SeserializeMsgTx(msgTx),
	}, nil
}

type SetGenesisTxUnlockScriptRequest struct {
	UserId          string `json:"user_id"`
	UnlockScriptHex string `json:"unlock_script_hex"`
}

type SetGenesisTxUnlockScriptResponse struct {
}

func (gameServer *GameServer) SetGenesisTxUnlockScriptLock(request *SetGenesisTxUnlockScriptRequest) (*SetGenesisTxUnlockScriptResponse, error) {
	unlockScrip, err := hex.DecodeString(request.UnlockScriptHex)
	if err != nil {
		return nil, err
	}
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	participantContext, err := gameServer.GetParticipantContext(request.UserId)
	if err != nil {
		return nil, err
	}
	if participantContext.UnlockScript != nil {
		return nil, errors.New("unlockScript already set")
	}
	participantContext.UnlockScript = unlockScrip
	return &SetGenesisTxUnlockScriptResponse{}, nil
}

type PublishRequest struct {
}

type PulishResponse struct {
	Txid string `json:"txid"`
}

func (gameServer *GameServer) PublishLock(request *PublishRequest) (*PulishResponse, error) {
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	msgTx, err := gameServer.GetGenesisMsgTx(true)
	if err != nil {
		return nil, err
	}
	hash, err := gameServer.RpcClient.SendRawTransaction(msgTx, true)
	if err != nil {
		return nil, err
	}
	gameServer.GenesisTxid = hash.String()
	return &PulishResponse{
		Txid: hash.String(),
	}, nil
}

type SetPreimageRequest struct {
	UserId   string `json:"user_id"`
	Preimage string `json:"preimage"`
}

type SetPreimageResponse struct {
}

func (gameServer *GameServer) SetPreimageLock(request *SetPreimageRequest) (*SetPreimageResponse, error) {
	preimage, ok := big.NewInt(0).SetString(request.Preimage, 10)
	if !ok {
		return nil, errors.New("error preimage")
	}
	hash := util.GetHash(preimage)
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	participantContext, err := gameServer.GetParticipantContext(request.UserId)
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
	return &SetPreimageResponse{}, nil
}

type GetRivalPreimagePubkeyRequest struct {
	UserId string `json:"user_id"`
}

type GetRivalPreimagePubkeyResponse struct {
	Preimage string `json:"preimage"`
	Pubkey   string `json:"pubkey"`
}

func (gameServer *GameServer) GetRivalPreimagePubkeyLock(request *GetRivalPreimagePubkeyRequest) (*GetRivalPreimagePubkeyResponse, error) {
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	for _, participantContext := range gameServer.ParticipantContexts {
		if participantContext.Id == request.UserId {
			continue
		}
		if participantContext.Preimage == nil {
			return nil, errors.New("rival preimage not set")
		}
		return &GetRivalPreimagePubkeyResponse{
			Preimage: participantContext.Preimage.String(),
			Pubkey:   hex.EncodeToString(participantContext.Pubkey.SerializeCompressed()),
		}, nil
	}
	return nil, errors.New("preimage not found")
}
