package server

import (
	"net/http"
	"satoshicard/core"
)

type HttpServer struct {
	GameContext *core.GameContext
	Server      *http.Server
}

func NewHttpServer(gameContext *core.GameContext) *HttpServer {
	return &HttpServer{
		GameContext: gameContext,
	}
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
	GameId string `json:"game_id"`
}

func (httpServer *HttpServer) Join(rsp http.ResponseWriter, req *http.Request, request *JoinRequest) (*JoinResponse, error) {
	gameId, err := httpServer.GameContext.AddParticipantLock(request.Id)
	if err != nil {
		return nil, err
	}
	return &JoinResponse{
		GameId: gameId,
	}, nil
}

type SubmitStep1InfoRequest struct {
	UserId   string `json:"user_id"`
	GameId   string `json:"game_id"`
	Hash     string `json:"hash"`
	Pubkey   string `json:"pubkey"`
	Pretxid  string `json:"pretxid"`
	Preindex int    `json:"preindex"`
}

type SubmitStep1InfoResponse struct {
	Rawtx string `json:"rawtx"`
}

func (httpServer *HttpServer) SubmitStep1Info(rsp http.ResponseWriter, req *http.Request, request *SubmitStep1InfoRequest) (*SubmitStep1InfoResponse, error) {
	step1Info := &core.Step1Info{
		Pubkey: request.Pubkey,
		Hash:   request.Hash,
		Txid:   request.Pretxid,
		Index:  request.Preindex,
	}
	rawtx, err := httpServer.GameContext.SetStep1InfoAndWaitRawTxLock(request.UserId, request.GameId, step1Info)
	if err != nil {
		return nil, err
	}
	return &SubmitStep1InfoResponse{
		Rawtx: rawtx,
	}, nil
}

type SubmitStep2InfoRequest struct {
	UserId       string `json:"user_id"`
	GameId       string `json:"game_id"`
	UnlockScript string `json:"sig"`
}

type SubmitStep2InfoResponse struct {
	Rawtx string `json:"rawtx"`
}

func (httpServer *HttpServer) SubmitStep2Info(rsp http.ResponseWriter, req *http.Request, request *SubmitStep2InfoRequest) (*SubmitStep2InfoResponse, error) {
	panic("todo")
}

type SubmitStep3InfoRequest struct {
	UserId   string `json:"user_id"`
	GameId   string `json:"game_id"`
	Preimage string `json:"preimage"`
}

type SubmitStep3InfoResponse struct {
	RivalPreimage string `json:"rival_preimage"`
}

func (httpServer *HttpServer) SubmitStep3Info(rsp http.ResponseWriter, req *http.Request, request *SubmitStep2InfoRequest) {

}
