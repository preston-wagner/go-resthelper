package resthelper

import (
	"net/http"
)

// PreRequestHook is for functions that run before the wrapped handler, to take care of common tasks like checking authentication tokens
// hooks are run in the provided order; if any return an HttpError, the provided status code will be returned and neither the handler nor any subsequent hooks will run
type PreRequestHook func(r *http.Request) *HttpError

func callPreRequestHooks(hooks []PreRequestHook, r *http.Request) *HttpError {
	for i := range hooks {
		err := hooks[i](r)
		if err != nil {
			return err
		}
	}
	return nil
}

// PostResponseHook is for functions that run after the wrapped handler, to take care of common tasks like logging http status codes
type PostResponseHook func(*HttpError, int)

func callPostResponseHooks(hooks []PostResponseHook, httpErr *HttpError, status int) {
	for i := range hooks {
		go hooks[i](httpErr, status)
	}
}
