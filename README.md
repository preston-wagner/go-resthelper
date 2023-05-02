# go-resthelper
A simple library of wrapper functions to help make sure your gorilla/mux routes return the responses you expect

## Usage
By default, gorilla/mux expects handlers to be functions like `func(w http.ResponseWriter, r *http.Request)`, where you parse inputs from `r` and call methods on `w` to make your response.

Unfortunately, the structure of that function doesn't actually require you to do any of that; execution of the handler may follow a path that simply never makes a response (eventually timing out the request), or it may make responses with an unexpected structure.

These wrapper functions allow you to write handlers that won't compile unless all paths result in a response, and also takes some of the busywork out of marshalling and unmarshalling.

To use `JsonResponseWrapper`, for example, you write a handler with the signature `func(*http.Request) (T, *HttpError)`, which will fail to compile if you fail to return a response or try to return a type other than `T`.
