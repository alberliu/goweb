package goweb

import (
	"net/http"
	"reflect"
)

var typeContext =reflect.TypeOf(Context{})

//暂时先这样，以后根据扩展
type Context struct {
	w http.ResponseWriter
	r *http.Request
}
