package util

import "strings"

type CodeError struct {
	Code int
	Msg  string
}

func (codeError CodeError) Error() string {
	return codeError.Msg
}

func NewCodeError(code int, msg string) error {
	return &CodeError{
		Code: code,
		Msg:  msg,
	}
}

func PanicIfErr(err error, msg interface{}) {
	if err != nil {
		panic(err)
	}
}

func PanicIfTrue(condition bool, msg interface{}) {
	if condition {
		panic(msg)
	}
}

func Try(tryFunc func(), catchFuncs ...func(interface{})) {
	for index := len(catchFuncs); index > 0; index-- {
		defer func(i int) {
			errInterface := recover()
			if errInterface == nil {
				return
			}
			catchFuncs[i-1](errInterface)
		}(index)
	}
	tryFunc()
}

func GetErrInterfaceMsg(errInterface interface{}) string {
	errStr := ""
	switch v := errInterface.(type) {
	case error:
		errStr = v.Error()
	case string:
		errStr = v
	default:
		errStr = "unknown err"
	}
	return errStr
}

func Catch(msg string, handler func(i interface{})) func(i interface{}) {
	return func(i interface{}) {
		errStr := GetErrInterfaceMsg(i)
		if !strings.Contains(errStr, msg) {
			panic(i)
		}
		handler(i)
	}
}
