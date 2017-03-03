package lib

import (
	"encoding/json"
	"fmt"
)

type Error struct {
	Code    string
	Message string
	Body    string
}

func (e *Error) String() string {
	data, _ := json.Marshal(e)
	return string(data)
}

type ErrList map[string]*Error

func NewError(err interface{}) *Error {
	return &Error{
		Code:    "001",
		Message: fmt.Sprintf("%s", err),
	}
}

func (l ErrList) Add(e *Error) {
	(map[string]*Error)(l)[e.Code] = e
}

func (l ErrList) Get(code string, v interface{}) (errdata *Error) {
	errdata = (map[string]*Error)(l)[code]
	if v != nil {
		errdata.Body = fmt.Sprintf("%s", v)
	}
	return errdata
}

var Errors ErrList

func init() {
	Errors = make(map[string]*Error)
	Errors.Add(&Error{"001", "Sign error", "please check your params"})
	Errors.Add(&Error{"002", "the lack of params", "please check your params"})
	Errors.Add(&Error{"100", "access denied, app can not call the api", "please check your authority"})
	Errors.Add(&Error{"101", "can't get authority", "please check your authority"})
	Errors.Add(&Error{"500", "Network anomalies", ""})
}
