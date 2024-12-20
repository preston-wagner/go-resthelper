package resthelper

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/preston-wagner/unicycle/fetch"
	"github.com/preston-wagner/unicycle/promises"
)

func testNoResponseHandler(r *http.Request) *HttpError {
	return nil
}

func unauthorizedHook(r *http.Request) *HttpError {
	return NewHttpErrF(http.StatusUnauthorized, "unauthorized")
}

func TestNoContentWrapper(t *testing.T) {
	router := mux.NewRouter()

	hookCalled := false
	postHook := func(*HttpError, int) {
		hookCalled = true
	}

	noResponseRoute := "/no_response/"
	router.HandleFunc(noResponseRoute, NoContentWrapper(testNoResponseHandler)).Methods("GET")

	unauthorizedRoute := "/unauthorized/"
	router.HandleFunc(unauthorizedRoute, NoContentWrapperWithHooks([]PreRequestHook{unauthorizedHook}, testNoResponseHandler, []PostResponseHook{postHook})).Methods("GET")

	const port = 9876
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		WriteTimeout: time.Second * 5,
		ReadTimeout:  time.Second * 5,
		IdleTimeout:  time.Second * 5,
		Handler:      router,
	}
	defer server.Close()
	serverRunPromise := promises.WrapInPromise(func() (bool, error) {
		err := server.ListenAndServe()
		return err == http.ErrServerClosed, err
	})

	rootUrl := "http://localhost:" + strconv.Itoa(port)

	resp, err := fetch.Fetch(rootUrl+noResponseRoute, fetch.FetchOptions{})
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Error("resp.StatusCode != http.StatusNoContent")
	}

	resp, err = fetch.Fetch(rootUrl+unauthorizedRoute, fetch.FetchOptions{})
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Error("resp.StatusCode != http.StatusUnauthorized")
	}

	if !hookCalled {
		t.Error("post response hook not called")
	}

	server.Close()
	ok, err := serverRunPromise.Await()
	if !ok {
		t.Fatal(err)
	}
}
