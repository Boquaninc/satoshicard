package client

import "satoshicard/server"

type Client interface {
	Join() (*server.JoinResponse, error)
}
