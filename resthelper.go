package resthelper

import (
	"fmt"
	"net/http"
)

func writeCommonHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func respondWithError(w http.ResponseWriter, httpErr *HttpError, postResponseHooks []postResponseHook) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(httpErr.Status)
	w.Write([]byte(httpErr.Error()))
	callPostResponseHooks(postResponseHooks, httpErr, httpErr.Status)
}

func recoverToErrorResponse(w http.ResponseWriter, postResponseHooks []postResponseHook) {
	if r := recover(); r != nil {
		msg := "goroutine panic"
		fmt.Println(msg, r)
		respondWithError(w, NewHttpErrF(http.StatusInternalServerError, msg), postResponseHooks)
	}
}
