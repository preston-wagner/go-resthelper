package resthelper

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// RestJsonWrapper allows us to ensure at compile time that a route handler will always return either a json response or an error code
func RestJsonWrapper[T any](toWrap func(*http.Request) (T, *HttpError), afterResponseHooks ...func(int)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer recoverToErrorResponse(w, afterResponseHooks)
		writeCommonHeaders(w)
		payload, err := toWrap(r)
		if err != nil {
			respondWithError(w, err.Status, err.Error(), afterResponseHooks)
		} else {
			response, _ := json.Marshal(payload)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
			callAllHooks(afterResponseHooks, http.StatusOK)
		}
	}
}

// RestNoContentWrapper allows us to ensure at compile time that a route handler will always return either a 204 No Content response or an error code
func RestNoContentWrapper(toWrap func(*http.Request) *HttpError, afterResponseHooks ...func(int)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer recoverToErrorResponse(w, afterResponseHooks)
		writeCommonHeaders(w)
		err := toWrap(r)
		if err != nil {
			respondWithError(w, err.Status, err.Error(), afterResponseHooks)
		} else {
			w.WriteHeader(http.StatusNoContent)
			callAllHooks(afterResponseHooks, http.StatusNoContent)
		}
	}
}

func writeCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func callAllHooks(hooks []func(int), statusCode int) {
	for i := range hooks {
		go hooks[i](statusCode)
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string, afterResponseHooks []func(int)) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	w.Write([]byte(msg))
	callAllHooks(afterResponseHooks, code)
}

func recoverToErrorResponse(w http.ResponseWriter, afterResponseHooks []func(int)) {
	if r := recover(); r != nil {
		fmt.Println("panicking goroutine recovered", r)
		respondWithError(w, http.StatusInternalServerError, "panic", afterResponseHooks)
	}
}
