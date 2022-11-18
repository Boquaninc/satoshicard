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
}

func NewHttpClient(host string) Client {
	format := "http://%s%s"
	return &HttpClient{
		JoinUrl:                     fmt.Sprintf(format, host, server.JOIN_URI),
		SetUtxoAndHashUrl:           fmt.Sprintf(format, host, server.SET_UTXO_AND_HASH_URI),
		GetGenesisTxUrl:             fmt.Sprintf(format, host, server.GET_GENESIS_TX_URI),
		SetGenesisTxUnlockScriptUrl: fmt.Sprintf(format, host, server.SET_GENESIS_TX_UNLOCK_SCRIPT_URI),
		PublishUrl:                  fmt.Sprintf(format, host, server.SET_GENESIS_TX_UNLOCK_SCRIPT_URI),
		Host:                        host,
	}
}

func (client *HttpClient) Join(id string) (*server.JoinResponse, error) {
	request := &server.JoinRequest{
		Id: id,
	}
	response := &server.JoinResponse{}
	err := util.DoHttpParseHttpJsonResponse(HTTP_MEHTOD_POST, client.JoinUrl, nil, request, response)
	return response, err
}

func (client *HttpClient) SetUtxoAndHash(request *server.SetUtxoAndHashRequest) error {
	response := &server.SetUtxoAndHashResponse{}
	err := util.DoHttpParseHttpJsonResponse(HTTP_MEHTOD_POST, client.SetUtxoAndHashUrl, nil, request, response)
	return err
}

func (client *HttpClient) GetGenesisTx(request *server.GetGenesisTxRequest) (*server.GetGenesisTxResponse, error) {
	response := &server.GetGenesisTxResponse{}
	err := util.DoHttpParseHttpJsonResponse(HTTP_MEHTOD_POST, client.GetGenesisTxUrl, nil, request, response)
	return response, err
}

func (client *HttpClient) SetGenesisTxUnlockScript(request *server.SetGenesisTxUnlockScriptRequest) error {
	response := &server.SetGenesisTxUnlockScriptResponse{}
	err := util.DoHttpParseHttpJsonResponse(HTTP_MEHTOD_POST, client.SetGenesisTxUnlockScriptUrl, nil, request, response)
	return err
}

func (client *HttpClient) Publish() (string, error) {
	request := &server.PublishRequest{}
	response := &server.PulishResponse{}
	err := util.DoHttpParseHttpJsonResponse(HTTP_MEHTOD_POST, client.PublishUrl, nil, request, response)
	return response.Txid, err
}
