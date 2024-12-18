package resthelper

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/preston-wagner/unicycle"
)

type testJsonStruct struct {
	Name  string
	Count int
}

func testJsonHandler(r *http.Request, input testJsonStruct) (testJsonStruct, *HttpError) {
	return input, nil
}

func testErrorHandler(r *http.Request, input testJsonStruct) (testJsonStruct, *HttpError) {
	return input, NewHttpErrF(http.StatusNotFound, "not found")
}

func TestJsonToJsonWrapper(t *testing.T) {
	router := mux.NewRouter()

	jsonRoute := "/json/"
	router.HandleFunc(jsonRoute, JsonToJsonWrapper(testJsonHandler)).Methods("POST")

	jsonErrorRoute := "/json_err/"
	router.HandleFunc(jsonErrorRoute, JsonToJsonWrapper(testErrorHandler)).Methods("POST")

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

	original := testJsonStruct{
		Name:  "Steve",
		Count: 7,
	}
	resp, err := unicycle.FetchJson[testJsonStruct](rootUrl+jsonRoute, unicycle.FetchOptions{
		Method: "POST",
		Body:   unicycle.JsonToReader(original),
	})
	if err != nil {
		t.Error(err)
	}
	if resp != original {
		t.Error("struct did not survive round trip")
	}

	_, err = unicycle.FetchJson[testJsonStruct](rootUrl+jsonErrorRoute, unicycle.FetchOptions{
		Method: "POST",
		Body:   unicycle.JsonToReader(original),
	})
	AssertErrorStatusCode(t, http.StatusNotFound, err)

	server.Close()
	ok, err := serverRunPromise.Await()
	if !ok {
		t.Fatal(err)
	}
}

func testErrorPreRequestHandler(r *http.Request) *HttpError {
	return NewHttpErrF(http.StatusForbidden, "unauthorized!")
}

func TestJsonToJsonWrapperWithHooks(t *testing.T) {
	router := mux.NewRouter()

	pre_hook_called := false
	post_hook_called := false

	set_pre_hook_called := func(r *http.Request) *HttpError { pre_hook_called = true; return nil }
	set_post_hook_called := func(*HttpError, int) { post_hook_called = true }

	jsonRoute := "/json/"
	router.HandleFunc(jsonRoute, JsonToJsonWrapperWithHooks(
		[]PreRequestHook{set_pre_hook_called},
		testJsonHandler,
		[]PostResponseHook{set_post_hook_called},
	)).Methods("POST")

	post_hook_status := 0

	set_post_hook_status := func(err *HttpError, status int) { post_hook_status = status }

	jsonErrorRoute := "/json_err/"
	router.HandleFunc(jsonErrorRoute, JsonToJsonWrapperWithHooks(
		[]PreRequestHook{testErrorPreRequestHandler},
		testErrorHandler, // prerequest hook throws the error, this should not be called
		[]PostResponseHook{set_post_hook_status},
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
	serverRunPromise := unicycle.WrapInPromise(func() (bool, error) {
		err := server.ListenAndServe()
		return err == http.ErrServerClosed, err
	})

	rootUrl := "http://localhost:" + strconv.Itoa(port)

	original := testJsonStruct{
		Name:  "Steve",
		Count: 7,
	}
	resp, err := unicycle.FetchJson[testJsonStruct](rootUrl+jsonRoute, unicycle.FetchOptions{
		Method: "POST",
		Body:   unicycle.JsonToReader(original),
	})
	if err != nil {
		t.Error(err)
	}
	if resp != original {
		t.Error("struct did not survive round trip")
	}

	_, err = unicycle.FetchJson[testJsonStruct](rootUrl+jsonErrorRoute, unicycle.FetchOptions{
		Method: "POST",
		Body:   unicycle.JsonToReader(original),
	})
	AssertErrorStatusCode(t, http.StatusForbidden, err)

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
