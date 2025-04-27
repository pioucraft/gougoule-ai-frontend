package api

import (
	"net/http"
	"context"
	"encoding/json"
)

func JournalHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		rows, err := Conn.Query(context.Background(), "SELECT content, id FROM journal_entries ORDER BY created_at DESC")
		if err != nil {
			http.Error(w, "Error querying journal entries", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var entries []struct {
			Content string `json:"content"`
			ID      string `json:"id"`
		}
		for rows.Next() {
			var entry struct {
				Content string `json:"content"`
				ID      string `json:"id"`
			}
			err := rows.Scan(&entry.Content, &entry.ID)
			if err != nil {
				http.Error(w, "Error scanning journal entries", http.StatusInternalServerError)
				return
			}
			entries = append(entries, entry)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Error iterating over journal entries", http.StatusInternalServerError)
			return
		}
		response, err := json.Marshal(entries)
		if err != nil {
			http.Error(w, "Error marshaling journal entries", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	} else if r.Method == http.MethodPost {	
		var requestBody struct {
			Content string `json:"content"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}
		_, err = Conn.Exec(context.Background(), "INSERT INTO journal_entries (content) VALUES ($1)", requestBody.Content)
		if err != nil {
			http.Error(w, "Error inserting journal entry", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK"))
	} else if r.Method == http.MethodPut {
		var requestBody struct {
			ID      string `json:"id"`
			Content string `json:"content"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}
		_, err = Conn.Exec(context.Background(), "UPDATE journal_entries SET content = $1 WHERE id = $2", requestBody.Content, requestBody.ID)
		if err != nil {
			http.Error(w, "Error updating journal entry", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK"))
	} else if r.Method == http.MethodDelete {
		var requestBody struct {
			ID string `json:"id"`
		}
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}
		_, err = Conn.Exec(context.Background(), "DELETE FROM journal_entries WHERE id = $1", requestBody.ID)
		if err != nil {
			http.Error(w, "Error deleting journal entry", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK"))
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
