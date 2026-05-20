package api

import "net/http"

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}
