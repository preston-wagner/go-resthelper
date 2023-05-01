package resthelper

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/preston-wagner/unicycle"
)

func testNoResponseHandler(r *http.Request) *HttpError {
	return nil
}

func TestNoContentWrapper(t *testing.T) {
	router := mux.NewRouter()

	testRoute := "/no_response/"
	router.HandleFunc(testRoute, NoContentWrapper(testNoResponseHandler)).Methods("GET")

	const port = 9876
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(port),
		WriteTimeout: time.Second * 5,
		ReadTimeout:  time.Second * 5,
		IdleTimeout:  time.Second * 5,
		Handler:      router,
	}
	defer server.Close()
	serverRunPromise := unicycle.WrapInPromise(func() (bool, error) {
		err := server.ListenAndServe()
		return err == http.ErrServerClosed, err
	})

	rootUrl := "http://localhost:" + strconv.Itoa(port)

	resp, err := unicycle.Fetch(rootUrl+testRoute, unicycle.FetchOptions{})
	if err != nil {
		t.Error(err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Error("resp.StatusCode != http.StatusNoContent")
	}

	server.Close()
	ok, err := serverRunPromise.Await()
	if !ok {
		t.Fatal(err)
	}
}
