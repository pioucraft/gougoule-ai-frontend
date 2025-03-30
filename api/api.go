package api

import (
	"net/http"
)

func API(w http.ResponseWriter, r *http.Request) {
	//TODO: Add authentication
	if r.URL.Path == "/api/v1/ask" || r.URL.Path == "/api/v1/ask/" {
		AskHandler(w, r)
	}
}
