package api

import (
	"context"
	"encoding/json"
	"net/http"
)

func AIProvidersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var requestBody struct {
			Name    string `json:"name"`
			URL     string `json:"url"`
			Api_key string `json:"api_key"`
		}

		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		_, err = Conn.Exec(context.Background(), "INSERT INTO ai_providers (name, api_key, url) VALUES ($1, $2, $3)", requestBody.Name, requestBody.Api_key, requestBody.URL)
		if err != nil {
			http.Error(w, "Error inserting AI provider", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK"))
	} else if r.Method == http.MethodGet {
		rows, err := Conn.Query(context.Background(), "SELECT name, url, id FROM ai_providers")
		if err != nil {
			http.Error(w, "Error querying AI providers", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var providers []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
			ID   string `json:"id"`
		}
		for rows.Next() {
			var provider struct {
				Name string `json:"name"`
				URL  string `json:"url"`
				ID   string `json:"id"`
			}
			err := rows.Scan(&provider.Name, &provider.URL, &provider.ID)
			if err != nil {
				http.Error(w, "Error scanning AI providers", http.StatusInternalServerError)
				return
			}
			providers = append(providers, provider)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Error iterating over AI providers", http.StatusInternalServerError)
			return
		}
		response, err := json.Marshal(providers)
		if err != nil {
			http.Error(w, "Error marshaling AI providers", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(response)
	} else if r.Method == http.MethodDelete {
		var requestBody struct {
			ID string `json:"id"`
		}

		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		_, err = Conn.Exec(context.Background(), "DELETE FROM ai_providers WHERE id = $1", requestBody.ID)
		if err != nil {
			http.Error(w, "Error deleting AI provider", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK"))
	} else if r.Method == http.MethodPatch {
		var requestBody struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			URL  string `json:"url"`
		}

		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}
		_, err = Conn.Exec(context.Background(), "UPDATE ai_providers SET name = $1, url = $2 WHERE id = $3", requestBody.Name, requestBody.URL, requestBody.ID)
		if err != nil {
			http.Error(w, "Error updating AI provider", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK"))
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
