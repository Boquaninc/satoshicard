package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

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

func HttpAspect[T1 Request, T2 Response](
	handler func(rsp http.ResponseWriter, req *http.Request, request *T1) (*T2, error),
) func(rsp http.ResponseWriter, req *http.Request) {
	return func(rsp http.ResponseWriter, req *http.Request) {
		request := new(T1)
		body, err := ioutil.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		if len(body) == 0 {
			body = []byte("{}")
		}
		err = json.Unmarshal(body, request)
		if err != nil {
			return
		}
		response, err := handler(rsp, req, request)
		if err != nil {
			return
		}
		rsp.Write(MakeHttpJsonResponseByInterface(response))
		return
	}
}

type Request interface {
	// JoinGameRequest | SubmitHashRequest
}

type Response interface {
	// JoinGameResponse | SubmitHashResponse
}

type JoinGameRequest struct {
}

type JoinGameResponse struct {
	Id string `json:"id"`
}
