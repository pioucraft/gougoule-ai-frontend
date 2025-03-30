package api

import (
	"bytes"
	"encoding/json"
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
	// Set up the request to the Groq API
	url := "https://api.groq.com/openai/v1/chat/completions"

	data := map[string]interface{}{
		"model": "deepseek-r1-distill-llama-70b",
		"messages": []map[string]string{
			{"role": "user", "content": question},
		},
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
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return "", err
	}

	answer := respBody.Choices[0].Message.Content
	return answer, nil
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
