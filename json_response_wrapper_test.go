package resthelper_test

import (
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/preston-wagner/go-resthelper"
	"github.com/preston-wagner/unicycle/fetch"
	"github.com/preston-wagner/unicycle/json_ext"
	"github.com/preston-wagner/unicycle/promises"
)

type testJsonStruct struct {
	Name  string
	Count int
}

func testJsonHandler(r *http.Request, input testJsonStruct) (testJsonStruct, *resthelper.HttpError) {
	return input, nil
}

func testErrorHandler(r *http.Request, input testJsonStruct) (testJsonStruct, *resthelper.HttpError) {
	return input, resthelper.NewHttpErrF(http.StatusNotFound, "not found")
}

func assertErrorStatusCode(t *testing.T, statusCode int, err error) {
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

func TestJsonToJsonWrapper(t *testing.T) {
	router := mux.NewRouter()

	jsonRoute := "/json/"
	router.HandleFunc(jsonRoute, resthelper.JsonToJsonWrapper(testJsonHandler)).Methods("POST")

	jsonErrorRoute := "/json_err/"
	router.HandleFunc(jsonErrorRoute, resthelper.JsonToJsonWrapper(testErrorHandler)).Methods("POST")

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

	original := testJsonStruct{
		Name:  "Steve",
		Count: 7,
	}
	resp, err := fetch.FetchJson[testJsonStruct](rootUrl+jsonRoute, fetch.FetchOptions{
		Method: "POST",
		Body:   json_ext.JsonToReader(original),
	})
	if err != nil {
		t.Error(err)
	}
	if resp != original {
		t.Error("struct did not survive round trip")
	}

	_, err = fetch.FetchJson[testJsonStruct](rootUrl+jsonErrorRoute, fetch.FetchOptions{
		Method: "POST",
		Body:   json_ext.JsonToReader(original),
	})
	assertErrorStatusCode(t, http.StatusNotFound, err)

	server.Close()
	ok, err := serverRunPromise.Await()
	if !ok {
		t.Fatal(err)
	}
}

func testErrorPreRequestHandler(r *http.Request) *resthelper.HttpError {
	return resthelper.NewHttpErrF(http.StatusForbidden, "unauthorized!")
}

func TestJsonToJsonWrapperWithHooks(t *testing.T) {
	router := mux.NewRouter()

	pre_hook_called := false
	post_hook_called := false

	set_pre_hook_called := func(r *http.Request) *resthelper.HttpError { pre_hook_called = true; return nil }
	set_post_hook_called := func(*resthelper.HttpError, int) { post_hook_called = true }

	jsonRoute := "/json/"
	router.HandleFunc(jsonRoute, resthelper.JsonToJsonWrapperWithHooks(
		[]resthelper.PreRequestHook{set_pre_hook_called},
		testJsonHandler,
		[]resthelper.PostResponseHook{set_post_hook_called},
	)).Methods("POST")

	post_hook_status := 0

	set_post_hook_status := func(err *resthelper.HttpError, status int) { post_hook_status = status }

	jsonErrorRoute := "/json_err/"
	router.HandleFunc(jsonErrorRoute, resthelper.JsonToJsonWrapperWithHooks(
		[]resthelper.PreRequestHook{testErrorPreRequestHandler},
		testErrorHandler, // prerequest hook throws the error, this should not be called
		[]resthelper.PostResponseHook{set_post_hook_status},
	)).Methods("POST")

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

	original := testJsonStruct{
		Name:  "Steve",
		Count: 7,
	}
	resp, err := fetch.FetchJson[testJsonStruct](rootUrl+jsonRoute, fetch.FetchOptions{
		Method: "POST",
		Body:   json_ext.JsonToReader(original),
	})
	if err != nil {
		t.Error(err)
	}
	if resp != original {
		t.Error("struct did not survive round trip")
	}

	_, err = fetch.FetchJson[testJsonStruct](rootUrl+jsonErrorRoute, fetch.FetchOptions{
		Method: "POST",
		Body:   json_ext.JsonToReader(original),
	})
	assertErrorStatusCode(t, http.StatusForbidden, err)

	if !pre_hook_called {
		t.Error("pre request hook was not called!")
	}
	if !post_hook_called {
		t.Error("post request hook was not called!")
	}
	if post_hook_status != http.StatusForbidden {
		t.Error("post request hook did not receive correct status code!")
	}

	server.Close()
	ok, err := serverRunPromise.Await()
	if !ok {
		t.Fatal(err)
	}
}
