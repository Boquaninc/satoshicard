package client

import "satoshicard/server"

type Client interface {
	Join(id string) (*server.JoinResponse, error)
	SetUtxoAndHash(request *server.SetUtxoAndHashRequest) error
	GetGenesisTx(*server.GetGenesisTxRequest) (*server.GetGenesisTxResponse, error)
	SetGenesisTxUnlockScript(request *server.SetGenesisTxUnlockScriptRequest) error
	Publish() (string, error)
	SetPreimage(request *server.SetPreimageRequest) error
	GetRivalPreimage(*server.GetRivalPreimagePubkeyRequest) (*server.GetRivalPreimagePubkeyResponse, error)
}
