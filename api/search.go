package api

import (
	"context"
	"encoding/json"
	"framework/api/db"
	"net/http"
)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Query string `json:"query"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	rows, err := db.Conn.Query(context.Background(), `SELECT content, conversation_id 
	FROM messages 
	WHERE search_vector @@ to_tsquery('english', replace($1, ' ', ' & ')) 
	ORDER BY created_at DESC;`, requestBody.Query)
	if err != nil {
		http.Error(w, "Error querying messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var results []map[string]string
	for rows.Next() {
		var content, conversationID string
		err := rows.Scan(&content, &conversationID)
		if err != nil {
			http.Error(w, "Error scanning messages", http.StatusInternalServerError)
			return
		}
		result := map[string]string{
			"content":         content,
			"conversation_id": conversationID,
		}
		results = append(results, result)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, "Error iterating over messages", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	jsonResponse, err := json.Marshal(results)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
	w.Write(jsonResponse)
}
