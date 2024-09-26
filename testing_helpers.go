package resthelper

import (
	"errors"
	"testing"

	"github.com/nuvi/unicycle/fetch"
)

func AssertErrorStatusCode(t *testing.T, statusCode int, err error) {
	if err == nil {
		t.Error("expected FetchError, got: nil")
	} else {
		var fetchErr fetch.FetchError
		if errors.As(err, &fetchErr) {
			if fetchErr.Response.StatusCode != statusCode {
				fetch.LogPossibleFetchError(fetchErr)
				t.Error("expected status code", statusCode, "in FetchError, got", fetchErr.Response.StatusCode)
			}
		} else {
			t.Error("expected FetchError, got:", err)
		}
	}
}
