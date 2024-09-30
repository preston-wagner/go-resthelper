package resthelper_test

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/nuvi/go-resthelper"
	"github.com/nuvi/unicycle/fetch"
	"github.com/nuvi/unicycle/promises"
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

func TestJsonResponseWrapper(t *testing.T) {
	router := mux.NewRouter()

	jsonRoute := "/json/"
	router.HandleFunc(jsonRoute, resthelper.JsonResponseWrapper(resthelper.JsonRequestWrapper(testJsonHandler))).Methods("POST")

	jsonErrorRoute := "/json_err/"
	router.HandleFunc(jsonErrorRoute, resthelper.JsonResponseWrapper(resthelper.JsonRequestWrapper(testErrorHandler))).Methods("POST")

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
		Body:   fetch.JsonToReader(original),
	})
	if err != nil {
		t.Error(err)
	}
	if resp != original {
		t.Error("struct did not survive round trip")
	}

	_, err = fetch.FetchJson[testJsonStruct](rootUrl+jsonErrorRoute, fetch.FetchOptions{
		Method: "POST",
		Body:   fetch.JsonToReader(original),
	})
	resthelper.AssertErrorStatusCode(t, http.StatusNotFound, err)

	server.Close()
	ok, err := serverRunPromise.Await()
	if !ok {
		t.Fatal(err)
	}
}
