package util

import "encoding/json"

const (
	SERVER_ERROR_CODE = -1
)

type HttpJsonResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func MakeHttpJsonResponse(code int, msg string, data json.RawMessage) []byte {
	jsonResponse := &HttpJsonResponse{
		Code: code,
		Msg:  msg,
		Data: data,
	}
	responseByte, err := json.Marshal(jsonResponse)
	if err != nil {
		panic(err)
	}
	return responseByte
}

func MakeHttpJsonResponseByInterface(data interface{}) []byte {
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return MakeHttpJsonResponse(0, "", json.RawMessage(b))
}

func MakeHttpJsonResponseByError(err error, data interface{}) []byte {
	code := SERVER_ERROR_CODE
	codeErr, ok := err.(*CodeError)
	if ok {
		code = codeErr.Code
	}
	b, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return MakeHttpJsonResponse(code, err.Error(), json.RawMessage(b))
}
