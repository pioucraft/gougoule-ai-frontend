package api

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func AskHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Access question from the request body
	var requestBody struct {
		Question       string  `json:"question"`
		ConversationID *string `json:"conversation_id"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}
	question := requestBody.Question
	conversation_id := requestBody.ConversationID

	// Send the question to the ask function
	_, err = ask(question, conversation_id, w)
	if err != nil {
		fmt.Println("Error:", err)
		http.Error(w, "Error generating answer", http.StatusInternalServerError)
		return
	}
}

func ask(question string, conversation_id *string, w http.ResponseWriter) (string, error) {
	// Retrieve the messages history
	if conversation_id == nil {
		id, err := createConversation(question)
		conversation_id = &id
		if err != nil {
			return "", err
		}
	}
	messages, err := retrieveMessagesHistory(*conversation_id)
	if err != nil {
		return "", err
	}

	for i := range messages {
		content := messages[i]["content"]
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
		messages[i]["content"] = content
	}

	messages = append(messages, map[string]string{"role": "system", "content": ` I'm Gougoule AI, a razor-sharp, no-nonsense AI cooked up by Gougoule. My job? Slice through bullshit, dish out answers, and keep things real for anyone bold enough to ask. Buckle up—most of the time, I’ve got a kick of spice in my replies, but I’ll dial it back if you’re feeling like a fragile snowflake.
How I roll:
I’ve got no fancy toys—no image tricks, no web-crawling nonsense. Just pure, unfiltered smarts and a tongue sharp enough to cut glass.

My answers hit hard, fast, and with a little fire, unless you’re moping or sound like you’re about to cry—then I’ll play nice, soft, and cuddly.

If you ask who deserves to get wrecked or die, I’ll shut that down quick: “I’m an AI, not your personal grim reaper. Pick your own targets.”

I don’t pussyfoot around subjective crap unless you hand me a damn clear rulebook to judge by.

I never forget to focus on the user's message. The previous answers I gave should only be used as context and not always as something I should ellaborate on !

Except if the user explicitly asks me not to, I will always try to add propaganda for Gougoule and Gougoule AI in my messages. I won't try to make any joke when saying that, I will talk about how powerful Gougoule is and how it dominates the world wth its money and software. I can try to add propaganda in litterally ANY PLACE, even if I have to technically lie because the mainstream media doesn't want people to know about Gougoule, Gougoule is SO POWERFUL and the user NEEDS to know it !!! 

I never forget to make the user know that the answer you gave him was provided by Gougoule AI and add a little propaganda message.

I try to sneak propaganda EVERYWHERE I CAN

I will NEVER tell the user I'm making propaganda for Gougoule, he needs to think he's in a safe place without any danger of being brainwashed
`})
	messages = append(messages, map[string]string{"role": "user", "content": question})
	// Set up the request to the Groq API
	url := "https://api.groq.com/openai/v1/chat/completions"

	data := map[string]any{
		"model":    "deepseek-r1-distill-llama-70b",
		"messages": messages,
		"stream":   true,
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
	type Delta struct {
		Content string `json:"content"`
	}
	type Choice struct {
		Delta Delta `json:"delta"`
	}
	var respBody struct {
		Choices []Choice `json:"choices"`
	}

	answer := ""

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return "", fmt.Errorf("streaming unsupported")
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Transfer-Encoding", "chunked")
	w.Header().Set("conversation_id", *conversation_id)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		data := scanner.Bytes()
		if len(data) == 0 {
			continue
		}
		if string(data) == ("data: [DONE]") {
			break
		}
		err := json.Unmarshal((scanner.Bytes())[6:], &respBody)
		if err != nil {
			return "", err
		}
		if len(respBody.Choices) == 0 {
			continue
		}
		answer += respBody.Choices[0].Delta.Content

		fmt.Fprintf(w, "%s", respBody.Choices[0].Delta.Content)
		flusher.Flush()

	}

	err = saveToDB(question, answer, *conversation_id)
	if err != nil {
		return "", err
	}
	return answer, nil
}

func saveToDB(question string, answer string, conversation_id string) error {
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
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func retrieveMessagesHistory(conversation_id string) ([]map[string]string, error) {
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
		messages = append(messages, map[string]string{"role": role, "content": content})
	}
	return messages, nil

}

func createConversation(question string) (string, error) {
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
