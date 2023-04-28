package resthelper

import (
	"net/http"
)

// NoContentWrapper allows us to ensure at compile time that a route handler will always return either a 204 No Content response or an error code
func NoContentWrapper(toWrap func(*http.Request) *HttpError, afterResponseHooks ...func(int)) func(http.ResponseWriter, *http.Request) {
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
