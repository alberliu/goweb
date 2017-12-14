package goweb

import "net/http"

//默认的错误处理
func handler400(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	w.Write([]byte("bad request"))
}
func handler404(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Write([]byte("url not found"))
}
func handler405(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(405)
	w.Write([]byte("method not found"))
}
