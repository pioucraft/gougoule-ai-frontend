package main 

import (
	"bufio"
	"fmt"
	"os"
	"net/http"
	"bytes"
	"encoding/json"
	"flag"
)

func main() {
	token := os.Getenv("GAI_TOKEN")
	model := os.Getenv("GAI_MODEL")
	url := os.Getenv("GAI_URL") 
	fmt.Println(url)

	conversationsFlag := flag.Bool("conversations", false, "Display conversations")
	modelsFlag := flag.Bool("models", false, "Display models")
	conversationFlag := flag.String("conversation", "", "Conversation ID")
	flag.Parse()

	if *conversationsFlag {
		fmt.Println("Conversations:")
		// Call url + "retrieveConversations" endpoint with the token
		httpClient := &http.Client{}
		req, err := http.NewRequest("GET", url+"retrieveConversations", nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := httpClient.Do(req)
		if err != nil {
			fmt.Println("Error making request:", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Println("Error: received non-200 response code")
			return
		}
		// for every conversation, print uuid and title
		var conversations []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&conversations); err != nil {
			fmt.Println("Error decoding response:", err)
			return
		}
		// revert them before printing
		for i, j := 0, len(conversations)-1; i < j; i, j = i+1, j-1 {
			conversations[i], conversations[j] = conversations[j], conversations[i]
		}
		for _, conversation := range conversations {
			if len(conversation.Title) > 200 {
				fmt.Printf("%s: %s...\n", conversation.ID, conversation.Title[:200])
			} else {
				fmt.Printf("%s: %s\n", conversation.ID, conversation.Title)
			}
		}
		return
	} else if *modelsFlag {
		fmt.Println("Models:")
		// Call url + "models" endpoint with the token
		httpClient := &http.Client{}
		req, err := http.NewRequest("GET", url+"models", nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := httpClient.Do(req)
		if err != nil {
			fmt.Println("Error making request:", err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			fmt.Println("Error: received non-200 response code")
			return
		}
		// for every model, print id and name
		var models []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
			fmt.Println("Error decoding response:", err)
			return
		}
		for _, model := range models {
			fmt.Printf("%s: %s\n", model.ID, model.Name)
		}
		return
	}

	asciiArt := ` __________________________
< Welcome to Gougoule AI ! >
 --------------------------
   \
    \
        .--.
       |o_o |
       |:_/ |
      //   \ \
     (|     | )
    /'\_   _/ \
    \___)=(___/`

	fmt.Println(asciiArt)
	ask(token, model, url+"ask", *conversationFlag)
}

func ask(token string, model string, url string, conversation string) {
	fmt.Println("--------------------------")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your question: ")
	question, _ := reader.ReadString('\n')
	fmt.Println("--------------------------")

	body := []byte{}
	if conversation == "" {

		body, _ = json.Marshal(map[string]string{
			"question": question,
			"model": model,
		})
	} else {

		body, _ = json.Marshal(map[string]string{
			"question": question,
			"model": model,
			"conversation_id": conversation,
		})
	}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Create and send the HTTP request to the  API.
	client := &http.Client{}
	resp, _ := client.Do(req)

	// Get header conversation_id
	conversation_id := resp.Header.Get("conversation_id")

	//Print the response (plain text stream)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}

	ask(token, model, url, conversation_id)

}

