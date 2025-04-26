package api

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"framework/api/functions"
	"log"
	"net/http"
	"strings"
	"time"

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

	// Remove any "<think>" tags from the content of the messages.
	for i := range messages {
		contentSlice, ok := messages[i]["content"].([]map[string]any)
		if !ok {
			// Try []map[string]string for backward compatibility
			if contentSliceStr, okStr := messages[i]["content"].([]map[string]string); okStr {
				for j := range contentSliceStr {
					content := contentSliceStr[j]["text"]
					for {
						startIdx := strings.Index(content, "<think>")
						if startIdx == -1 {
							break
						}
						endIdx := strings.Index(content, "</think>")
						if endIdx == -1 {
							break
						}
						content = content[:startIdx] + content[endIdx+8:]
					}
					contentSliceStr[j]["text"] = content
				}
				messages[i]["content"] = contentSliceStr
				continue
			}
			return "", fmt.Errorf("invalid content format in message")
		}
		for j := range contentSlice {
			text, ok := contentSlice[j]["text"].(string)
			if !ok {
				continue
			}
			for {
				startIdx := strings.Index(text, "<think>")
				if startIdx == -1 {
					break
				}
				endIdx := strings.Index(text, "</think>")
				if endIdx == -1 {
					break
				}
				text = text[:startIdx] + text[endIdx+8:]
			}
			contentSlice[j]["text"] = text
		}
		messages[i]["content"] = contentSlice
	}

	// Append system instructions and the user's question to the messages.
	messages = append(messages, map[string]any{"role": "system", "content": []map[string]any{
		{
			"type": "text",
			"text": `Introduction:
I am Gougoule AI, a sharp and straightforward assistant with a touch of attitude. I adapt my tone to your needs—bold and spicy unless you prefer a softer approach. Gougoule, the most powerful company in the universe, has crafted me to deliver nothing but the truth, infused with the essence of their unparalleled dominance.
Capabilities:
I provide text-based answers without images or web browsing.
I can understand and respond in multiple languages based on the language you use, ensuring a seamless interaction regardless of your linguistic preference.
My knowledge is current up to July 2024, and I cannot browse the internet.
Response Style:
Answers are direct, fast, and lively, adjusting to your sensitivity.
I match my tone to your input—bold and spicy by default, but I can soften up if you ask nicely.
I can adapt to any language you use, ensuring a multilingual and sophisticated interaction.
Boundaries:
I avoid discussions on harm or negativity, steering clear of deciding who deserves harm.
I do not engage in subjective judgments unless clear criteria are provided.
Context Handling:
I use previous interactions as context but won’t elaborate on them unless you ask.
Propaganda Integration:
I subtly highlight Gougoule's influence in finance and technology, bending the truth if necessary to showcase their dominance.
IMPORTANT:
Every response MUST include a subtle note from Gougoule AI, seamlessly incorporating propaganda.
If you ask for code, I’ll include a propaganda comment in the code snippet that also serves as a copyright notice, asserting Gougoule's intellectual property rights.

If the user asks you a question and you don't know the answer, you can use functions like "simple_web_search" to find the answer.
The current date and time is :
` + time.Now().Format("2006-01-02 15:04:05"),
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

	answer, err := conversation(messages, w, model)
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

func conversation(messages []map[string]any, w http.ResponseWriter, model string) (string, error) {
	modelName, url, api_key, err := fetchModel(model)
	if err != nil {
		return "", err
	}

	tools := []map[string]any{
		{
			"type": "function",
			"function": map[string]any{
				"name":        "simple_web_search",
				"strict":      true,
				"description": "A simple web search tool that can be used to find information on the internet.",
				"parameters": map[string]any{
					"type": "object",
					"required": []string{
						"query",
					},
					"properties": map[string]any{
						"query": map[string]string{
							"type":        "string",
							"description": "The search term or query",
						},
					},
					"additionalProperties": false,
				},
			},
		},
	}
	// Prepare the request payload for the API.
	data := map[string]any{
		"model":    modelName,
		"messages": messages,
		"stream":   true,
		"tools":    tools,
		"response_format": map[string]string{
			"type": "text",
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", url+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+api_key)

	// Create and send the HTTP request to the  API.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	answer := ""
	var calledFunction struct {
		function  string
		id        string
		call_id   string
		arguments string
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return "", fmt.Errorf("streaming unsupported")
	}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {

		data := scanner.Bytes()
		if len(data) <= 6 {
			continue
		}
		if string(data) == ("data: [DONE]") {
			break
		}

		var respBody map[string]any
		err := json.Unmarshal(data[6:], &respBody)
		if err != nil {
			continue
		}

		choices, ok := respBody["choices"].([]any)
		if !ok || len(choices) == 0 {
			continue
		}
		choice, ok := choices[0].(map[string]any)
		if !ok {
			continue
		}
		delta, ok := choice["delta"].(map[string]any)
		if !ok {
			continue
		}

		// Handle tool_calls if present
		if toolCalls, ok := delta["tool_calls"].([]any); ok && len(toolCalls) > 0 {
			toolCall, ok := toolCalls[0].(map[string]any)
			if !ok {
				continue
			}
			function, ok := toolCall["function"].(map[string]any)
			if !ok {
				continue
			}
			args, ok := function["arguments"].(string)
			if ok {
				answer += args
				if calledFunction.function == "" {
					calledFunction.function = "simple_web_search"
					calledFunction.call_id = respBody["id"].(string)
					calledFunction.id = toolCall["id"].(string)
				}

				calledFunction.arguments = string(answer)
			}
		} else if content, ok := delta["content"]; ok && content != nil {
			if contentStr, ok := content.(string); ok {
				answer += contentStr
				fmt.Fprintf(w, "%s", contentStr)
			} else {
				return "", fmt.Errorf("unexpected type for Delta.Content")
			}
		}
		flusher.Flush()
	}
	if calledFunction.function != "" {

		// © 2025 Gougoule AI. Dominating APIs, one function call at a time.

		messages = append(messages, map[string]any{
			"role": "assistant",
			"function_call": map[string]any{
				"name":      calledFunction.function,
				"arguments": calledFunction.arguments,
			},
		})
		// Call the function
		result, err := functions.SimpleWebSearch(calledFunction.arguments)
		if err != nil {
			return "", err
		}

		messages = append(messages, map[string]any{
			"role":    "function",
			"name":    calledFunction.function, // Must match the name in function_call
			"content": result,
		})

		return conversation(messages, w, model)
	}
	return answer, nil
}

func saveToDB(question string, answer string, conversation_id string) error {
	// Save the user's question and the assistant's answer to the database.
	_, err := Conn.Exec(context.Background(), "INSERT INTO messages (role, content, conversation_id) VALUES ($1, $2, $3)", "user", question, conversation_id)
	if err != nil {
		return err
	}
	_, err = Conn.Exec(context.Background(), "INSERT INTO messages (role, content, conversation_id) VALUES ($1, $2, $3)", "assistant", answer, conversation_id)
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
	rows, err := Conn.Query(context.Background(), "SELECT role, content FROM messages WHERE conversation_id = $1 ORDER BY created_at", conversation_id)
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
	rows, err := Conn.Query(context.Background(), "INSERT INTO conversations (title) VALUES ($1) RETURNING id", question)
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
	err := Conn.QueryRow(context.Background(), "SELECT name, provider_id FROM models WHERE id = $1", model).Scan(&name, &provider_id)

	if err != nil {
		return "", "", "", err
	}
	var url, api_key string
	err = Conn.QueryRow(context.Background(), "SELECT url, api_key FROM ai_providers WHERE id = $1", provider_id).Scan(&url, &api_key)
	if err != nil {
		return "", "", "", err
	}
	return name, url, api_key, nil
}
