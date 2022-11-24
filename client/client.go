package client

import "satoshicard/server"

type Client interface {
	Join(id string) *server.JoinResponse
	SetUtxoAndHash(request *server.SetUtxoAndHashRequest)
	GetGenesisTx(*server.GetGenesisTxRequest) *server.GetGenesisTxResponse
	SetGenesisTxUnlockScript(request *server.SetGenesisTxUnlockScriptRequest)
	Publish() string
	SetPreimage(request *server.SetPreimageRequest)
	GetRivalPreimage(*server.GetRivalPreimagePubkeyRequest) *server.GetRivalPreimagePubkeyResponse
}
