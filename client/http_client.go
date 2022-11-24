package client

import (
	"fmt"
	"satoshicard/server"
	"satoshicard/util"
)

const (
	HTTP_MEHTOD_POST = "POST"
)

type HttpClient struct {
	Host                        string
	JoinUrl                     string
	SetUtxoAndHashUrl           string
	GetGenesisTxUrl             string
	SetGenesisTxUnlockScriptUrl string
	PublishUrl                  string
	SetPreimageUrl              string
	GetRivalPreimageUrl         string
}

func NewHttpClient(host string) Client {
	format := "http://%s%s"
	return &HttpClient{
		JoinUrl:                     fmt.Sprintf(format, host, server.JOIN_URI),
		SetUtxoAndHashUrl:           fmt.Sprintf(format, host, server.SET_UTXO_AND_HASH_URI),
		GetGenesisTxUrl:             fmt.Sprintf(format, host, server.GET_GENESIS_TX_URI),
		SetGenesisTxUnlockScriptUrl: fmt.Sprintf(format, host, server.SET_GENESIS_TX_UNLOCK_SCRIPT_URI),
		PublishUrl:                  fmt.Sprintf(format, host, server.PUBLISH_URI),
		SetPreimageUrl:              fmt.Sprintf(format, host, server.SET_PREIMAGE_URI),
		GetRivalPreimageUrl:         fmt.Sprintf(format, host, server.GET_RIVAL_PREIMAGE_PUBKEY_URI),
		Host:                        host,
	}
}

func (client *HttpClient) Join(id string) *server.JoinResponse {
	request := &server.JoinRequest{
		Id: id,
	}
	response := &server.JoinResponse{}
	util.DoHttpParseHttpJsonResponse(HTTP_MEHTOD_POST, client.JoinUrl, nil, request, response)
	return response
}

func (client *HttpClient) SetUtxoAndHash(request *server.SetUtxoAndHashRequest) {
	response := &server.SetUtxoAndHashResponse{}
	util.DoHttpParseHttpJsonResponse(HTTP_MEHTOD_POST, client.SetUtxoAndHashUrl, nil, request, response)
	return
}

func (client *HttpClient) GetGenesisTx(request *server.GetGenesisTxRequest) *server.GetGenesisTxResponse {
	response := &server.GetGenesisTxResponse{}
	util.DoHttpParseHttpJsonResponse(HTTP_MEHTOD_POST, client.GetGenesisTxUrl, nil, request, response)
	return response
}

func (client *HttpClient) SetGenesisTxUnlockScript(request *server.SetGenesisTxUnlockScriptRequest) {
	response := &server.SetGenesisTxUnlockScriptResponse{}
	util.DoHttpParseHttpJsonResponse(HTTP_MEHTOD_POST, client.SetGenesisTxUnlockScriptUrl, nil, request, response)
	return
}

func (client *HttpClient) Publish() string {
	request := &server.PublishRequest{}
	response := &server.PulishResponse{}
	util.DoHttpParseHttpJsonResponse(HTTP_MEHTOD_POST, client.PublishUrl, nil, request, response)
	return response.Txid
}

func (client *HttpClient) SetPreimage(request *server.SetPreimageRequest) {
	response := &server.SetPreimageResponse{}
	util.DoHttpParseHttpJsonResponse(HTTP_MEHTOD_POST, client.SetPreimageUrl, nil, request, response)
	return
}

func (client *HttpClient) GetRivalPreimage(request *server.GetRivalPreimagePubkeyRequest) *server.GetRivalPreimagePubkeyResponse {
	response := &server.GetRivalPreimagePubkeyResponse{}
	util.DoHttpParseHttpJsonResponse(HTTP_MEHTOD_POST, client.GetRivalPreimageUrl, nil, request, response)
	return response
}
