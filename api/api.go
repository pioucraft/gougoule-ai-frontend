package api

import (
	"net/http"
)

func API(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/api/v1/ask" || r.URL.Path == "/api/v1/ask/" {
		AskHandler(w, r)
	}
}
