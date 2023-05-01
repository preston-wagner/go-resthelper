package resthelper

import (
	"encoding/json"
	"net/http"

	"github.com/nuvi/unicycle"
)

func DecodeRequest[T any](r *http.Request) (T, *HttpError) {
	var req T
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		return req, NewHttpErr(http.StatusBadRequest, err)
	}
	return req, nil
}

func JsonRequestWrapper[REQUEST_TYPE any, RESPONSE_TYPE any](toWrap func(REQUEST_TYPE) (RESPONSE_TYPE, *HttpError)) func(*http.Request) (RESPONSE_TYPE, *HttpError) {
	return func(r *http.Request) (RESPONSE_TYPE, *HttpError) {
		body, httpErr := DecodeRequest[REQUEST_TYPE](r)
		if httpErr != nil {
			return unicycle.ZeroValue[RESPONSE_TYPE](), httpErr
		}
		return toWrap(body)
	}
}
