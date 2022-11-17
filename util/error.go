package util

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
