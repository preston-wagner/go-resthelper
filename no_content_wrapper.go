package resthelper

import (
	"net/http"
)

type NoResponseHandler func(*http.Request) *HttpError

// NoContentWrapper allows us to ensure at compile time that a route handler will always return either a 204 No Content response or an error code
func NoContentWrapper(toWrap NoResponseHandler) func(http.ResponseWriter, *http.Request) {
	return NoContentWrapperWithHooks([]PreRequestHook{}, toWrap, []PostResponseHook{})
}

func NoContentWrapperWithHooks(preRequestHooks []PreRequestHook, toWrap NoResponseHandler, postResponseHooks []PostResponseHook) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer recoverToErrorResponse(w, postResponseHooks)
		writeCommonHeaders(w)
		err := callPreRequestHooks(preRequestHooks, r)
		if err != nil {
			respondWithError(w, err, postResponseHooks)
			return
		}
		err = toWrap(r)
		if err != nil {
			respondWithError(w, err, postResponseHooks)
		} else {
			w.WriteHeader(http.StatusNoContent)
			callPostResponseHooks(postResponseHooks, nil, http.StatusNoContent)
		}
	}
}
