package client

import (
	"satoshicard/server"
)

type InternalClient struct {
	GameServer *server.GameServer
}

func NewInternalClient(GameServer *server.GameServer) Client {
	return &InternalClient{
		GameServer: GameServer,
	}
}

func (client *InternalClient) Join(id string) *server.JoinResponse {
	request := &server.JoinRequest{
		Id: id,
	}
	return client.GameServer.JoinLock(request)
}

func (client *InternalClient) SetUtxoAndHash(request *server.SetUtxoAndHashRequest) {
	client.GameServer.SetUtxoAndHashLock(request)
	return
}

func (client *InternalClient) GetGenesisTx(request *server.GetGenesisTxRequest) *server.GetGenesisTxResponse {
	return client.GameServer.GetGenesisTxLock(request)
}

func (client *InternalClient) SetGenesisTxUnlockScript(request *server.SetGenesisTxUnlockScriptRequest) {
	client.GameServer.SetGenesisTxUnlockScriptLock(request)
	return
}

func (client *InternalClient) Publish() string {
	request := &server.PublishRequest{}
	response := client.GameServer.PublishLock(request)
	return response.Txid
}

func (client *InternalClient) SetPreimage(request *server.SetPreimageRequest) {
	client.GameServer.SetPreimageLock(request)
	return
}

func (client *InternalClient) GetRivalPreimage(request *server.GetRivalPreimagePubkeyRequest) *server.GetRivalPreimagePubkeyResponse {
	return client.GameServer.GetRivalPreimagePubkeyLock(request)
}
