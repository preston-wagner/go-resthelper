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

func testJsonHandler(input testJsonStruct) (testJsonStruct, *HttpError) {
	return input, nil
}

func TestJsonResponseWrapper(t *testing.T) {
	router := mux.NewRouter()

	testRoute := "/json/"
	router.HandleFunc(testRoute, JsonResponseWrapper(JsonRequestWrapper(testJsonHandler))).Methods("POST")

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
	resp, err := unicycle.FetchJson[testJsonStruct](rootUrl+testRoute, unicycle.FetchOptions{
		Method: "POST",
		Body:   unicycle.JsonToReader(original),
	})
	if err != nil {
		t.Error(err)
	}
	if resp != original {
		t.Error("struct did not survive round trip")
	}

	server.Close()
	ok, err := serverRunPromise.Await()
	if !ok {
		t.Fatal(err)
	}
}
