package main 

import (
	"bufio"
	"fmt"
	"os"
	"net/http"
	"bytes"
	"encoding/json"
)

func main() {
	token := os.Getenv("GAI_TOKEN")
	model := os.Getenv("GAI_MODEL")
	url := os.Getenv("GAI_URL") 
	fmt.Println(url)

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
	ask(token, model, url, "")
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

