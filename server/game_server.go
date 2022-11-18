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

	"github.com/btcsuite/btcd/btcec"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/sCrypt-Inc/go-scryptlib"
)

const (
	GAMBLING_CAPITAL                 = 10000000
	MAX_FACTOR                       = 10
	EACH_FEE                         = 1000000
	JOIN_URI                         = "/join"
	SET_UTXO_AND_HASH_URI            = "/set_utxo_and_hash"
	GET_GENESIS_TX_URI               = "/get_genesis_tx"
	SET_GENESIS_TX_UNLOCK_SCRIPT_URI = "/set_genesis_tx_unlock_script"
)

type Request interface {
	JoinRequest |
		SetUtxoAndHashRequest |
		GetGenesisTxRequest |
		SetGenesisTxUnlockScriptRequest
}

type Response interface {
	JoinResponse |
		SetUtxoAndHashResponse |
		GetGenesisTxResponse |
		SetGenesisTxUnlockScriptResponse
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
	Id                  string
	ParticipantContexts []*ParticipantContext
	L                   sync.Locker
	Contract            *scryptlib.Contract
	ContractPath        string
	RpcClient           *rpcclient.Client
	OnAddParticipant    func(string)
	Server              *http.Server
	GenesisTxid         string
}

func NewGameServer(listen string, contractPath string, rpcClient *rpcclient.Client, OnAddParticipant func(string)) *GameServer {
	desc, err := scryptlib.LoadDesc(contractPath)
	if err != nil {
		panic(err)
	}

	contract, err := scryptlib.NewContractFromDesc(desc)
	if err != nil {
		panic(err)
	}
	server := &GameServer{
		Id:                  util.RandStringBytesMaskImprSrcUnsafe(8),
		ParticipantContexts: []*ParticipantContext{},
		Contract:            &contract,
		ContractPath:        contractPath,
		L:                   &sync.Mutex{},
		OnAddParticipant:    OnAddParticipant,
		Server:              &http.Server{Addr: listen},
		RpcClient:           rpcClient,
	}
	http.HandleFunc(JOIN_URI, HttpAspect(server.JoinLock))
	http.HandleFunc(SET_UTXO_AND_HASH_URI, HttpAspect(server.SetUtxoAndHashLock))
	http.HandleFunc(GET_GENESIS_TX_URI, HttpAspect(server.GetGenesisTxLock))
	http.HandleFunc(SET_GENESIS_TX_UNLOCK_SCRIPT_URI, HttpAspect(server.SetGenesisTxUnlockScriptLock))
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

	playerContexts := gameServer.ParticipantContexts
	constructorParams := map[string]scryptlib.ScryptType{
		"hash1": scryptlib.NewIntFromBigInt(playerContexts[0].Hash),
		"hash2": scryptlib.NewIntFromBigInt(playerContexts[1].Hash),
		"user1": scryptlib.NewPubKey(util.ToBecPubkey(playerContexts[0].Pubkey)),
		"user2": scryptlib.NewPubKey(util.ToBecPubkey(playerContexts[1].Pubkey)),
	}
	contract := gameServer.Contract
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

func (gameServer *GameServer) SetPreimageLock(uid string, preimageStr string) error {
	preimage, ok := big.NewInt(0).SetString(preimageStr, 10)
	if !ok {
		return errors.New("error preimage")
	}
	hash := util.GetHash(preimage)
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	participantContext, err := gameServer.GetParticipantContext(uid)
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

type GetRivalPreimageRequest struct {
	UserId string `json:"user_id"`
}

type GetRivalPreimageResponse struct {
	Preimage string `json:"preimage"`
}

func (gameServer *GameServer) GetRivalPreimageLock(uid string) (*big.Int, error) {
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	for _, participantContext := range gameServer.ParticipantContexts {
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

func (gameServer *GameServer) GetRivalPreimage(request *GetRivalPreimageRequest) (*GetRivalPreimageResponse, error) {
	gameServer.L.Lock()
	defer gameServer.L.Unlock()
	for _, participantContext := range gameServer.ParticipantContexts {
		if participantContext.Id == request.UserId {
			continue
		}
		if participantContext.Preimage == nil {
			return nil, errors.New("rival preimage not set")
		}
		return &GetRivalPreimageResponse{
			Preimage: participantContext.Preimage.String(),
		}, nil
	}
	return nil, errors.New("preimage not found")
}
