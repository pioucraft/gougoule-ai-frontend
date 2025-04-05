package api

import (
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
	}
}

func init() {
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
