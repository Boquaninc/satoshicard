package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
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
	b, err2 := json.Marshal(data)
	if err2 != nil {
		panic(err2)
	}
	return MakeHttpJsonResponse(code, err.Error(), json.RawMessage(b))
}

func ToCurlStr(method string, header map[string]string, bodyInterface interface{}, url string) {
	body, ok := bodyInterface.([]byte)
	if !ok {
		bodyTmp, err := json.Marshal(bodyInterface)
		if err != nil {
			panic(err)
		}
		body = bodyTmp
	}
	var b strings.Builder
	b.WriteString("curl ")
	for key, value := range header {
		b.WriteString("-H ")
		b.WriteString("\"")
		b.WriteString(key)
		b.WriteString(":")
		b.WriteString(value)
		b.WriteString("\" ")
	}
	if len(body) > 0 {
		b.WriteString("-X ")
		b.WriteString(method)
		b.WriteString(" --data '")
		b.Write(body)
		b.WriteString("' ")
	}
	b.WriteString(url)
	fmt.Println(b.String())
}

func PrintJson(i interface{}) {
	ib, ok := i.([]byte)
	if ok {
		fmt.Println(string(ib))
		return
	}
	is, ok := i.(string)
	if ok {
		fmt.Println(is)
		return
	}
	b, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}

func DoHttp(method string, url string, header map[string]string, data interface{}) ([]byte, error) {
	var content []byte = nil
	if data != nil {
		contentTmp, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		content = contentTmp
	}
	ToCurlStr(method, header, content, url)

	req, err := http.NewRequest(method, url, bytes.NewReader(content))
	for key, value := range header {
		req.Header.Add(key, value)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	PrintJson(body)
	return body, nil
}

func DoHttpParseHttpJsonResponse(method string, url string, header map[string]string, data interface{}, result interface{}) error {
	resultByte, err := DoHttp(method, url, header, data)
	if err != nil {
		return err
	}
	httpJsonResponse := &HttpJsonResponse{}
	err = json.Unmarshal(resultByte, httpJsonResponse)
	if err != nil {
		return err
	}
	if httpJsonResponse.Code != 0 {
		return errors.New(httpJsonResponse.Msg)
	}
	return json.Unmarshal(httpJsonResponse.Data, result)
}
