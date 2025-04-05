package api

import (
	"context"
	"encoding/json"
	"net/http"
)

func MessagesHistoryHandler(w http.ResponseWriter, r *http.Request, conversation_id string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Retrieve the messages history
	messages, err := MessagesHistory(conversation_id)
	if err != nil {
		http.Error(w, "Error retrieving messages history", http.StatusInternalServerError)
		return
	}

	// Return the messages as a JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]any{
		"messages": messages,
	}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
	w.Write(jsonResponse)
}

func MessagesHistory(conversation_id string) ([]map[string]string, error) {
	rows, err := Conn.Query(context.Background(), "SELECT role, content FROM messages WHERE conversation_id = $1 ORDER BY created_at", conversation_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []map[string]string
	for rows.Next() {
		var role, content string
		err := rows.Scan(&role, &content)
		if err != nil {
			return nil, err
		}
		messages = append(messages, map[string]string{
			"role":    role,
			"content": content,
		})
	}
	return messages, nil
}
