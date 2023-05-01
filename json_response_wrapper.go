package resthelper

import (
	"encoding/json"
	"net/http"
)

type DefaultMuxHandler func(w http.ResponseWriter, r *http.Request)

type JsonResponseHandler[T any] func(*http.Request) (T, *HttpError)

// JsonResponseWrapper allows us to ensure at compile time that a route handler will always return either a json response or an error code
func JsonResponseWrapper[T any](toWrap JsonResponseHandler[T]) func(http.ResponseWriter, *http.Request) {
	return JsonResponseWrapperWithHooks([]preRequestHook{}, toWrap, []postResponseHook{})
}

func JsonResponseWrapperWithHooks[T any](preRequestHooks []preRequestHook, toWrap func(*http.Request) (T, *HttpError), postResponseHooks []postResponseHook) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer recoverToErrorResponse(w, postResponseHooks)
		writeCommonHeaders(w)
		err := callPreRequestHooks(preRequestHooks, r)
		if err != nil {
			respondWithError(w, err, postResponseHooks)
		}
		payload, err := toWrap(r)
		if err != nil {
			respondWithError(w, err, postResponseHooks)
		} else {
			response, _ := json.Marshal(payload)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write(response)
			callPostResponseHooks(postResponseHooks, nil, http.StatusOK)
		}
	}
}
