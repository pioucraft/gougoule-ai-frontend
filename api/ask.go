package api

import (
	"context"
	"fmt"
	"framework/api/db"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	json "github.com/json-iterator/go"
)

func AskHandler(w http.ResponseWriter, r *http.Request) {
	// Ensure the request method is POST; otherwise, return an error.
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body to extract the question, conversation ID, and model.
	var requestBody struct {
		Question       string  `json:"question"`
		ConversationID *string `json:"conversation_id"`
		Model          string  `json:"model"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}
	question := requestBody.Question
	conversation_id := requestBody.ConversationID
	model := requestBody.Model

	// Send the question to the `ask` function and handle any errors.
	_, err = ask(question, model, conversation_id, w)
	if err != nil {
		fmt.Println("Error:", err)
		http.Error(w, "Error generating answer", http.StatusInternalServerError)
		return
	}
}

func ask(question string, model string, conversation_id *string, w http.ResponseWriter) (string, error) { 	
	// If no conversation ID is provided, create a new conversation.
	if conversation_id == nil {
		id, err := createConversation(question)
		conversation_id = &id
		if err != nil {
			return "", err
		}
	}
	// Retrieve the message history for the given conversation ID.
	messages, err := retrieveMessagesHistory(*conversation_id)
	if err != nil {
		return "", err
	}

	// Append system instructions and the user's question to the messages.
	messages = append(messages, map[string]any{"role": "system", "content": []map[string]any{
		{
			"type": "text",
			"text": SystemPrompt(),
		},
	}})

	// Fetch the model details (name, URL, API key) from the database.

	messages = append(messages, map[string]any{"role": "user", "content": []map[string]any{
		{
			"type": "text",
			"text": question,
		},
	}})

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("conversation_id", *conversation_id)

	answer, err := Conversation(messages, w, model, "")
	if err != nil {
		return "", err
	}

	// Save the question and answer to the database.
	err = saveToDB(question, answer, *conversation_id)
	if err != nil {
		return "", err
	}
	return answer, nil
}

func saveToDB(question string, answer string, conversation_id string) error {
	// Save the user's question and the assistant's answer to the database.
	_, err := db.Conn.Exec(context.Background(), "INSERT INTO messages (role, content, conversation_id) VALUES ($1, $2, $3)", "user", question, conversation_id)
	if err != nil {
		return err
	}
	_, err = db.Conn.Exec(context.Background(), "INSERT INTO messages (role, content, conversation_id) VALUES ($1, $2, $3)", "assistant", answer, conversation_id)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	// Load environment variables from the .env file.
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func retrieveMessagesHistory(conversation_id string) ([]map[string]any, error) {
	// Query the database to retrieve the message history for the given conversation ID.
	rows, err := db.Conn.Query(context.Background(), "SELECT role, content FROM messages WHERE conversation_id = $1 ORDER BY created_at", conversation_id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []map[string]any
	for rows.Next() {
		var role, content string
		err := rows.Scan(&role, &content)
		if err != nil {
			return nil, err
		}
		messages = append(messages, map[string]any{"role": role, "content": []map[string]string{
			{
				"type": "text",
				"text": content,
			},
		}})
	}
	return messages, nil

}

func createConversation(question string) (string, error) {
	// Create a new conversation in the database and return its ID.
	rows, err := db.Conn.Query(context.Background(), "INSERT INTO conversations (title) VALUES ($1) RETURNING id", question)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var id string
	for rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			return "", err
		}
	}
	return id, nil
}

func fetchModel(model string) (string, string, string, error) {
	// Fetch the model details (name, provider ID, URL, API key) from the database.
	var name, provider_id string
	err := db.Conn.QueryRow(context.Background(), "SELECT name, provider_id FROM models WHERE id = $1", model).Scan(&name, &provider_id)

	if err != nil {
		return "", "", "", err
	}
	var url, api_key string
	err = db.Conn.QueryRow(context.Background(), "SELECT url, api_key FROM ai_providers WHERE id = $1", provider_id).Scan(&url, &api_key)
	if err != nil {
		return "", "", "", err
	}
	return name, url, api_key, nil
}

