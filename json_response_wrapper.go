package resthelper

import (
	"encoding/json"
	"net/http"
)

// JsonResponseWrapper allows us to ensure at compile time that a route handler will always return either a json response or an error code
func JsonResponseWrapper[T any](toWrap func(*http.Request) (T, *HttpError), afterResponseHooks ...func(int)) func(http.ResponseWriter, *http.Request) {
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
