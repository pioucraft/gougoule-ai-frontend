package api

import (
	"context"
	"encoding/json"
	"framework/api/db"
	"net/http"
)

func AIModels(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		var requestBody struct {
			Name       string `json:"name"`
			ProviderID string `json:"provider_id"`
		}

		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}

		_, err = db.Conn.Exec(context.Background(), "INSERT INTO models (name, provider_id) VALUES ($1, $2)", requestBody.Name, requestBody.ProviderID)
		if err != nil {
			http.Error(w, "Error inserting AI model", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK"))
	} else if r.Method == http.MethodGet {
		rows, err := db.Conn.Query(context.Background(), "SELECT name, provider_id, id FROM models ORDER BY created_at")
		if err != nil {
			http.Error(w, "Error querying AI models", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var models []struct {
			Name       string `json:"name"`
			ProviderID string `json:"provider_id"`
			ID         string `json:"id"`
		}
		for rows.Next() {
			var model struct {
				Name       string `json:"name"`
				ProviderID string `json:"provider_id"`
				ID         string `json:"id"`
			}
			err := rows.Scan(&model.Name, &model.ProviderID, &model.ID)
			if err != nil {
				http.Error(w, "Error scanning AI models", http.StatusInternalServerError)
				return
			}
			models = append(models, model)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "Error iterating over AI models", http.StatusInternalServerError)
			return
		}
		response, err := json.Marshal(models)
		if err != nil {
			http.Error(w, "Error marshaling AI models", http.StatusInternalServerError)
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

		_, err = db.Conn.Exec(context.Background(), "DELETE FROM models WHERE id = $1", requestBody.ID)
		if err != nil {
			http.Error(w, "Error deleting AI model", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK"))
	} else if r.Method == http.MethodPatch {
		var requestBody struct {
			ID         string `json:"id"`
			Name       string `json:"name"`
			ProviderID string `json:"provider_id"`
		}

		err := json.NewDecoder(r.Body).Decode(&requestBody)
		if err != nil {
			http.Error(w, "Error decoding request body", http.StatusBadRequest)
			return
		}
		_, err = db.Conn.Exec(context.Background(), "UPDATE models SET name = $1, provider_id = $2 WHERE id = $3", requestBody.Name, requestBody.ProviderID, requestBody.ID)
		if err != nil {
			http.Error(w, "Error updating AI model", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("200 OK"))
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
