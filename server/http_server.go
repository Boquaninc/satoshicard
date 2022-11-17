package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"satoshicard/core"
	"satoshicard/util"
)

const (
	JOIN_URI = "/join"
)

type Request interface {
	JoinRequest
}

type Response interface {
	JoinResponse
}

func HttpAspect[T1 Request, T2 Response](
	handler func(rsp http.ResponseWriter, req *http.Request, request *T1) (*T2, error),
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
		response, err := handler(rsp, req, request)
		if err != nil {
			rsp.Write(util.MakeHttpJsonResponseByError(err, nil))
			return
		}
		rsp.Write(util.MakeHttpJsonResponseByInterface(response))
		return
	}
}

type HttpServer struct {
	GameContext *core.GameContext
	Server      *http.Server
}

func NewHttpServer(gameContext *core.GameContext, listen string) *HttpServer {
	server := &HttpServer{
		GameContext: gameContext,
		Server:      &http.Server{Addr: listen},
	}
	http.HandleFunc(JOIN_URI, HttpAspect(server.Join))
	return server
}

func (httpServer *HttpServer) Close() {
	err := httpServer.Server.Close()
	if err != nil {
		panic(err)
	}
}

func (httpServer *HttpServer) Open() {
	go httpServer.Server.ListenAndServe()
}

type JoinRequest struct {
	Id string `json:"id"`
}

type JoinResponse struct {
	Rival string `json:"rival"`
}

func (httpServer *HttpServer) Join(rsp http.ResponseWriter, req *http.Request, request *JoinRequest) (*JoinResponse, error) {
	id, err := httpServer.GameContext.AddParticipantLock(request.Id)
	if err != nil {
		fmt.Printf("join err:%s\n", err)
		return nil, err
	}
	return &JoinResponse{
		Rival: id,
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

func (httpServer *HttpServer) SetUtxoAndHash(rsp http.ResponseWriter, req *http.Request, request *SetUtxoAndHashRequest) (*SetUtxoAndHashResponse, error) {
	step1Info := &core.UtxoAndHash{
		Pubkey: request.Pubkey,
		Hash:   request.Hash,
		Txid:   request.Pretxid,
		Index:  request.Preindex,
	}
	err := httpServer.GameContext.SetUtxoAndHashLock(request.UserId, step1Info)
	if err != nil {
		return nil, err
	}
	return &SetUtxoAndHashResponse{}, nil
}

type GetGenesisTxRequest struct {
	UserId string `json:"user_id"`
	Sign   bool   `json:"sign"`
}

type GetGenesisTxResponse struct {
	Rawtx string `json:"rawtx"`
}

func (httpServer *HttpServer) GetUnSignGenesisTx(rsp http.ResponseWriter, req *http.Request, request *GetGenesisTxRequest) (*GetGenesisTxResponse, error) {
	msgTx, err := httpServer.GameContext.GetGenesisTxLock(request.Sign)
	if err != nil {
		return nil, err
	}
	return &GetGenesisTxResponse{
		Rawtx: util.SeserializeMsgTx(msgTx),
	}, nil
}

type SetUnlockScriptRequest struct {
	UserId          string `json:"user_id"`
	UnlockScriptHex string `json:"unlock_script_hex"`
}

type SetUnlockScriptResponse struct {
}

func (httpServer *HttpServer) SetUnlockScript(rsp http.ResponseWriter, req *http.Request, request *SetUnlockScriptRequest) (*SetUnlockScriptResponse, error) {
	unlockScrip, err := hex.DecodeString(request.UnlockScriptHex)
	if err != nil {
		return nil, err
	}
	err = httpServer.GameContext.SetUnlockScriptLock(request.UserId, unlockScrip)
	if err != nil {
		return nil, err
	}
	return &SetUnlockScriptResponse{}, nil
}

type SendGenesisRequest struct {
}

type SendGenesisResponse struct {
}

func (httpServer *HttpServer) SendGenesis(rsp http.ResponseWriter, req *http.Request, request *SendGenesisRequest) (*SendGenesisResponse, error) {
	err := httpServer.GameContext.SendGenesisLock()
	if err != nil {
		return nil, err
	}
	return &SendGenesisResponse{}, nil
}

type GetRivalPreimageRequest struct {
	UserId string `json:"user_id"`
}

type GetRivalPreimageResponse struct {
	Preimage string `json:"preimage"`
}

func (httpServer *HttpServer) GetRivalPreimage(rsp http.ResponseWriter, req *http.Request, request *GetRivalPreimageRequest) (*GetRivalPreimageResponse, error) {
	preimage, err := httpServer.GameContext.GetRivalPreimage(request.UserId)
	if err != nil {
		return nil, err
	}
	return &GetRivalPreimageResponse{
		Preimage: preimage.String(),
	}, nil
}
