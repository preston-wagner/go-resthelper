package resthelper

import (
	"encoding/json"
	"net/http"
)

type DefaultMuxHandler func(w http.ResponseWriter, r *http.Request)

type JsonResponseHandler[T any] func(*http.Request) (T, *HttpError)

// JsonResponseWrapper allows us to ensure at compile time that a route handler will always return either a json response or an error code
func JsonResponseWrapper[T any](toWrap JsonResponseHandler[T]) DefaultMuxHandler {
	return JsonResponseWrapperWithHooks([]PreRequestHook{}, toWrap, []PostResponseHook{})
}

func JsonResponseWrapperWithHooks[T any](
	preRequestHooks []PreRequestHook,
	toWrap func(*http.Request) (T, *HttpError),
	postResponseHooks []PostResponseHook,
) DefaultMuxHandler {
	return func(w http.ResponseWriter, r *http.Request) {
		defer recoverToErrorResponse(w, postResponseHooks)
		writeCommonHeaders(w)
		err := callPreRequestHooks(preRequestHooks, r)
		if err != nil {
			respondWithError(w, err, postResponseHooks)
			return
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

// JsonToJsonWrapper simplifies the common case where both the body of the request and the response should be json
func JsonToJsonWrapper[REQUEST_TYPE any, RESPONSE_TYPE any](
	toWrap JsonRequestHandler[REQUEST_TYPE, RESPONSE_TYPE],
) DefaultMuxHandler {
	return JsonResponseWrapper(JsonRequestWrapper(toWrap))
}

func JsonToJsonWrapperWithHooks[REQUEST_TYPE any, RESPONSE_TYPE any](
	preRequestHooks []PreRequestHook,
	toWrap JsonRequestHandler[REQUEST_TYPE, RESPONSE_TYPE],
	postResponseHooks []PostResponseHook,
) DefaultMuxHandler {
	return JsonResponseWrapperWithHooks(
		preRequestHooks,
		JsonRequestWrapper(toWrap),
		postResponseHooks,
	)
}
