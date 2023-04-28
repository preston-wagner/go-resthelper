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
