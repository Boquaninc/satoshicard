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
	Id      string
	Host    string
	JoinUrl string
}

func NewHttpClient(id string, host string) Client {
	format := "http://%s%s"
	return &HttpClient{
		JoinUrl: fmt.Sprintf(format, host, server.JOIN_URI),
		Host:    host,
		Id:      id,
	}
}

func (httpClient *HttpClient) Join() (*server.JoinResponse, error) {
	request := &server.JoinRequest{
		Id: httpClient.Id,
	}
	response := &server.JoinResponse{}
	err := util.DoHttpParseHttpJsonResponse(HTTP_MEHTOD_POST, httpClient.JoinUrl, nil, request, response)
	return response, err
}
