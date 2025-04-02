package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func AskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Access question from the request body
	var requestBody struct {
		Question string `json:"question"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}
	question := requestBody.Question

	// Send the question to the ask function
	answer, err := ask(question, w)
	if err != nil {
		http.Error(w, "Error generating answer", http.StatusInternalServerError)
		return
	}

	// Return the answer as a JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]string{"answer": answer}
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
	w.Write(jsonResponse)
}

func ask(question string, w http.ResponseWriter) (string, error) {
	// Retrieve the messages history
	messages, err := retrieveMessagesHistory()
	if err != nil {
		return "", err
	}
	messages = append(messages, map[string]string{"role": "user", "content": question})
	fmt.Printf("Messages: %v\n", messages)
	// Set up the request to the Groq API
	url := "https://api.groq.com/openai/v1/chat/completions"

	data := map[string]any{
		"model":    "deepseek-r1-distill-llama-70b",
		"messages": messages,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("GROQ_API_KEY"))

	// Create the client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse the response
	type Message struct {
		Content string `json:"content"`
	}
	type Choice struct {
		Message Message `json:"message"`
	}
	var respBody struct {
		Choices []Choice `json:"choices"`
	}

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return "", err
	}

	answer := respBody.Choices[0].Message.Content
	err = saveToDB(question, answer)
	if err != nil {
		return "", err
	}
	return answer, nil
}

func saveToDB(question string, answer string) error {
	_, err := Conn.Exec(context.Background(), "INSERT INTO messages (role, content) VALUES ($1, $2)", "user", question)
	if err != nil {
		return err
	}
	_, err = Conn.Exec(context.Background(), "INSERT INTO messages (role, content) VALUES ($1, $2)", "assistant", answer)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func retrieveMessagesHistory() ([]map[string]string, error) {
	rows, err := Conn.Query(context.Background(), "SELECT role, content FROM messages ORDER BY created_at")
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
		messages = append(messages, map[string]string{"role": role, "content": content})
	}
	return messages, nil

}
