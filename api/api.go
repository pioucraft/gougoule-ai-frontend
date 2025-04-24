package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

var password string

func API(w http.ResponseWriter, r *http.Request) {
	if !authMiddleware(w, r) {
		return
	}

	if r.URL.Path == "/api/v1/ask" || r.URL.Path == "/api/v1/ask/" {
		AskHandler(w, r)
		return
	} else if strings.HasPrefix(r.URL.Path, "/api/v1/messagesHistory/") {
		MessagesHistoryHandler(w, r, r.URL.Path[len("/api/v1/messagesHistory/"):])
		return
	} else if r.URL.Path == "/api/v1/retrieveConversations" || r.URL.Path == "/api/v1/retrieveConversations/" {
		RetrieveConversationsHandler(w, r)
		return
	} else if r.URL.Path == "/api/v1/aiProviders" || r.URL.Path == "/api/v1/aiProviders/" {
		AIProvidersHandler(w, r)
		return
	} else if r.URL.Path == "/api/v1/models" || r.URL.Path == "/api/v1/models/" {
		AIModels(w, r)
		return
	} else if r.URL.Path == "/api/v1/search" || r.URL.Path == "/api/v1/search/" {
		SearchHandler(w, r)
		return
	} else {
		http.Error(w, "Invalid endpoint", http.StatusNotFound)
		return
	}
}

func init() {
	fmt.Println(("Hello from Gougoule AI - the most powerful AI in the world from the most powerful company in the world"))
	fmt.Println("Gougoule AI is a product of Gougoule Inc. - the most powerful company in the world")
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	password = os.Getenv("PASSWORD")
}

func authMiddleware(w http.ResponseWriter, r *http.Request) bool {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Bearer token required", http.StatusUnauthorized)
		return false
	}
	if authHeader != "Bearer "+password {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return false
	}
	return true
}
