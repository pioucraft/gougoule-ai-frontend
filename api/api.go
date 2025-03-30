package api

import (
	"net/http"
)

func API(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello from API !"))
}
