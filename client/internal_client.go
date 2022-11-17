package client

import "satoshicard/server"

type InternalClient struct {
	Id string
}

func NewInternalClient(id string) Client {
	return &InternalClient{
		Id: id,
	}
}

func (client *InternalClient) Join() (*server.JoinResponse, error) {
	panic("not support")
}
