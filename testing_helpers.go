package resthelper

import (
	"errors"
	"testing"

	"github.com/preston-wagner/unicycle"
)

func AssertErrorStatusCode(t *testing.T, statusCode int, err error) {
	if err == nil {
		t.Error("expected FetchError, got: nil")
	} else {
		var fetchErr unicycle.FetchError
		if errors.As(err, &fetchErr) {
			if fetchErr.Response.StatusCode != statusCode {
				unicycle.LogPossibleFetchError(fetchErr)
				t.Error("expected status code", statusCode, "in FetchError, got", fetchErr.Response.StatusCode)
			}
		} else {
			t.Error("expected FetchError, got:", err)
		}
	}
}
