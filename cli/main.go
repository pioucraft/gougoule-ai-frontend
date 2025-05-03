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
	token := os.Getenv("GOI_TOKEN")
	model := os.Getenv("GOI_MODEL")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter your question: ")
	question, _ := reader.ReadString('\n')

	ask(question, token, model)
}

func ask(question string, token string, model string) {
	url := "https://gougoule.ch/api/v1/ask"
	
	body, _ := json.Marshal(map[string]string{
		"question": question,
		"model": model, 
	})

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Create and send the HTTP request to the  API.
	client := &http.Client{}
	resp, _ := client.Do(req)

	//Print the response (plain text stream)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	
}

