package resthelper

import (
	"net/http"
)

// preRequestHook is for functions that run before the wrapped handler, to take care of common tasks like checking authentication tokens
// hooks are run in the provided order; if any return an HttpError, the provided status code will be returned and neither the handler nor any subsequent hooks will run
type preRequestHook func(r *http.Request) *HttpError

func callPreRequestHooks(hooks []preRequestHook, r *http.Request) *HttpError {
	for i := range hooks {
		err := hooks[i](r)
		if err != nil {
			return err
		}
	}
	return nil
}

// postResponseHook is for functions that run after the wrapped handler, to take care of common tasks like logging http status codes
type postResponseHook func(*HttpError, int)

func callPostResponseHooks(hooks []postResponseHook, httpErr *HttpError, status int) {
	for i := range hooks {
		go hooks[i](httpErr, status)
	}
}
