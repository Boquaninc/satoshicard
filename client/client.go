package client

type Client interface {
	Host()
	Join()
	SubmitStep1Info()
	SubmitStep2Info()
}
