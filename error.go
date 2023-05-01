package resthelper

import (
	"fmt"
)

type HttpError struct {
	Err    error
	Status int
}

func (e HttpError) Error() string {
	return e.Err.Error()
}

func (e HttpError) Unwrap() error {
	return e.Err
}

func NewHttpErr(status int, err error) *HttpError {
	return &HttpError{
		Status: status,
		Err:    err,
	}
}

func NewHttpErrF(status int, format string, a ...any) *HttpError {
	return &HttpError{
		Status: status,
		Err:    fmt.Errorf(format, a...),
	}
}
