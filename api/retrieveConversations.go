package api

import (
	"context"
	"encoding/json"
	"framework/api/db"
	"net/http"
)

func RetrieveConversationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	type conversation struct {
		Title string `json:"title"`
		UUID  string `json:"id"`
	}
	var conversations []conversation
	conversationsFromDB, err := db.Conn.Query(context.Background(), "SELECT title, id FROM conversations ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, "Error retrieving conversations", http.StatusInternalServerError)
		return
	}
	defer conversationsFromDB.Close()
	for conversationsFromDB.Next() {
		var conversation conversation
		if err := conversationsFromDB.Scan(&conversation.Title, &conversation.UUID); err != nil {
			http.Error(w, "Error scanning conversations", http.StatusInternalServerError)
			return
		}
		conversations = append(conversations, conversation)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonResponse, err := json.Marshal(conversations)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
	w.Write(jsonResponse)
}
