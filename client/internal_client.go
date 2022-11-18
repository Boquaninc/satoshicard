package client

import "satoshicard/server"

type InternalClient struct {
	GameServer *server.GameServer
}

func NewInternalClient(GameServer *server.GameServer) Client {
	return &InternalClient{
		GameServer: GameServer,
	}
}

func (client *InternalClient) Join(id string) (*server.JoinResponse, error) {
	request := &server.JoinRequest{
		Id: id,
	}
	return client.GameServer.JoinLock(request)
}

func (client *InternalClient) SetUtxoAndHash(request *server.SetUtxoAndHashRequest) error {
	_, err := client.GameServer.SetUtxoAndHashLock(request)
	return err
}

func (client *InternalClient) GetGenesisTx(request *server.GetGenesisTxRequest) (*server.GetGenesisTxResponse, error) {
	return client.GameServer.GetGenesisTxLock(request)
}

func (client *InternalClient) SetGenesisTxUnlockScript(request *server.SetGenesisTxUnlockScriptRequest) error {
	_, err := client.GameServer.SetGenesisTxUnlockScriptLock(request)
	return err
}

func (client *InternalClient) Publish() (string, error) {
	request := &server.PublishRequest{}
	response, err := client.GameServer.PublishLock(request)
	return response.Txid, err
}
