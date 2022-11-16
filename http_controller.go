package main

import (
	"net/http"
)

type HttpController struct {
	GameServer *GameServer
	HttpServer *http.Server
}

func NewHttpController(listen string, GameServer *GameServer) *HttpController {
	httpServer := &http.Server{Addr: listen}
	return &HttpController{
		HttpServer: httpServer,
		GameServer: GameServer,
	}
}

func (httpController *HttpController) Close() {
	err := httpController.HttpServer.Close()
	if err != nil {
		panic(err)
	}
}

func (httpController *HttpController) ListenAndServe() {
	go httpController.ListenAndServe()
}

type LogInRequest struct {
}

type LogInResponse struct {
	Id string `json:"id"`
}

func (httpController *HttpController) LogInAndJoinRandomRoom(rsp http.ResponseWriter, req *http.Request, request *LogInRequest) (*LogInResponse, error) {
	id := httpController.GameServer.LogInL()

	// roomList, err := httpController.GameServer.GetRoomListL()
	// if err != nil {
	// 	return nil, err
	// }

	return &LogInResponse{
		Id: id,
	}, nil
}

type JoinRandomRoomRequest struct {
	UserId string `json:"user_id"`
	RoomId string `json:"room_id"`
}

type JoinRandomRoomResponse struct {
}

func (httpController *HttpController) JoinRandomRoom(rsp http.ResponseWriter, req *http.Request, request *JoinRandomRoomRequest) (*JoinRandomRoomResponse, error) {
	err := httpController.GameServer.JoinRoomL(request.UserId, request.RoomId)
	if err != nil {
		return nil, err
	}
	return &JoinRandomRoomResponse{}, nil
}

// type ListRoomRequest struct {
// }

// func (httpController *HttpController) ListRoom(rsp http.ResponseWriter, req *http.Request, request *JoinGameRequest) (*JoinGameResponse, error) {
// 	err := httpController.GameServer.JoinRoomL(request.UserId, request.RoomId)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &JoinGameResponse{}, nil
// }
