package goweb

import (
	"testing"
	"fmt"
	"runtime"
	"bytes"
	"strconv"
	"net/http"
)

type S struct {
	A int
	B int
}

func TestGoWeb_HandleGroup(t *testing.T) {
	group := NewGroup("/wifi")
	group.HandlePost("/test/{name}/{id}", handler5)
	goWeb := NewGoWeb()
	goWeb.HandleGroup(group)
	//goWeb.set
	goWeb.ListenAndServe(":4000")
}

func TestGoWeb_HandleFunc(t *testing.T) {
	goWeb := NewGoWeb()
	goWeb.HandleFunc(POST, "/", handler)
	goWeb.HandleGet("/", handler)
	goWeb.HandlePost("/", handler)
	goWeb.HandlePut("/", handler)
	goWeb.HandleDelete("/", handler)
	fmt.Println(goWeb)
	fmt.Println(goWeb.router)
}

func TestHandleGroup(t *testing.T) {
	group := NewGroup("/hello")
	group.HandlePost("/", handler)
	HandleGroup(group)
	fmt.Println(defultGoWeb)
}

func TestHandleFunc(t *testing.T) {
	HandleFunc("/", POST, handler)
	HandleGet("/", handler)
	HandlePost("/", handler)
	HandlePut("/", handler)
	HandleDelete("/", handler)
	fmt.Println(defultGoWeb)
	fmt.Println(defultGoWeb.router)
}

func handler(context Context) S {
	return S{}
}
func handler0(s S) (S) {
	fmt.Println("handle ...")
	return s
}

func handler1(s *S) (*S) {
	fmt.Println("handle ...")
	return s
}

func handler3(s S) (*S) {
	fmt.Println("handle ...")
	return &s
}

func handler4(ctx Context, s *S) (S) {
	fmt.Println("handle ...")
	return *s
}

func handler5(name string, id int64) (*S) {
	//fmt.Println(name)
	//fmt.Println(id)
	fmt.Println(getGID())
	return new(S)
}

func getGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

func handler6(ctx Context, name string, id int64) (*S) {
	fmt.Println(name)
	fmt.Println(id)
	fmt.Println("handle ...")
	return new(S)
}

func TestListenAndServe(t *testing.T) {
	HandlePost("/test", handler4)
	ListenAndServe(":4000")
}

func TestListenAndServeGet(t *testing.T) {
	HandleGet("/{name}/{id}", handler6)
	ListenAndServe(":4000")
}

func TestFilter(t *testing.T) {
	HandleGet("/{name}/{id}", handler6)
	defultGoWeb.Filter = Filter
	ListenAndServe(":4000")
}

func Filter(w http.ResponseWriter, r *http.Request, f func(http.ResponseWriter, *http.Request)) {
	fmt.Println(r.RequestURI)
	//f(w,r)
	w.Write([]byte("test"))

}
