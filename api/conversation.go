/*
	I am sorry.
	This code should not exist.
	This is some very messy code.
	The "handleFunctionCalls" function is basically a way to make it look like the main function is separated, but in reality, it is not.
	Even if there aren't any function calls, it will still "handle function calls". And if there were a function call, it will just use the result of the function call and just continue to call the "Conversation" function again and again.
*/

package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"framework/api/functions"
	"github.com/google/uuid"
	"net/http"
)

var devMode *bool

func openAIAPIRequest(modelName string, messages []map[string]any, url string, api_key string) (*http.Response, error) {
		data := map[string]any{
		"model":    modelName,
		"messages": messages,
		"stream":   true,
		"response_format": map[string]string{
			"type": "text",
		},
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	if devMode != nil && *devMode {
		fmt.Println(string(jsonData))
	}
	req, err := http.NewRequest("POST", url+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+api_key)

	// Create and send the HTTP request to the  API.
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func Conversation(messages []map[string]any, w http.ResponseWriter, model string, currentAnswer string) (string, error) {
	if devMode != nil && *devMode {
		fmt.Println(messages)
	}
	modelName, url, api_key, err := fetchModel(model)
	if err != nil {
		return "", err
	}

	resp, err := openAIAPIRequest(modelName, messages, url, api_key)
	if err != nil {
		http.Error(w, "Failed to create request: "+err.Error(), http.StatusInternalServerError)
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	defer resp.Body.Close()

	answer := ""

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return "", fmt.Errorf("streaming unsupported")
	}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {

		data := scanner.Bytes()
		if devMode != nil && *devMode {
			fmt.Println(string(data))
		}
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

		if content, ok := delta["content"]; ok && content != nil {
			if contentStr, ok := content.(string); ok {
				answer += contentStr
				fmt.Fprintf(w, "%s", contentStr)
			} else {
				return "", fmt.Errorf("unexpected type for Delta.Content")
			}
		}
		flusher.Flush()
	}
	messages = append(messages, map[string]any{
		"role":    "assistant",
		"content": []map[string]any{ 
			{
				"type": "text",
				"text": answer,
			},
		},
	})
	
	finalAnswer, err := handleFunctionCalls(answer, messages, w, model, currentAnswer)
	if err != nil {
		http.Error(w, "Failed to handle function calls: "+err.Error(), http.StatusInternalServerError)
		return "", fmt.Errorf("failed to handle function calls: %w", err)
	}
	return finalAnswer, nil
}

func handleFunctionCalls(answer string, messages []map[string]any, w http.ResponseWriter, model string, currentAnswer string) (string, error) {
	functionCalls, err := getFunctionCalls([]byte(answer))
	if err != nil {
		return "", err
	}

	functionCalled := false
	finalFunctionCallString := ""

	for _, functionCall := range functionCalls {
		if functionCall["name"] == nil {
			continue
		}
		functionName, ok := functionCall["name"].(string)
		if !ok {
			continue
		}
		functionCalled = true
		var result string
		if functionName == "simple_web_search" {
			// Call the function
			result, err = functions.SimpleWebSearch(functionCall["arguments"].(string))
			if err != nil {
				return "", err
			}
		} else if functionName == "memory_create" {
			// Call the function
			err := functions.MemoryCreate(functionCall["arguments"].(string))
			if err != nil {
				return "", err
			}
			result = "Memory cell created successfully."
		} else if functionName == "memory_delete" {
			// Call the function
			err := functions.MemoryDelete(functionCall["arguments"].(string))
			if err != nil {
				return "", err
			}
			result = "Memory cell deleted successfully."
		} else {
			result = "Unknown function call."
			return "", fmt.Errorf("unknown function: %s", functionName)
		}
		functionCallResult := map[string]any{
			"name":      functionName,
			"arguments": functionCall["arguments"],
			"result":    result,
		}
		functionCallString, err := json.Marshal(functionCallResult)
		if err != nil {
			return "", err
		}
		functionCallAnswer := "{@function_result}" + string(functionCallString) + "{/function_result}"
		answer += functionCallAnswer
		fmt.Fprintf(w, "%s", functionCallAnswer)
		finalFunctionCallString += functionCallAnswer

	}
	if functionCalled {

		messages = append(messages, map[string]any{
			"role":    "user",
			"content": []map[string]any{
				{
					"type": "text",
					"text": "result : " + string(finalFunctionCallString),
				},
			},
		})

		return Conversation(messages, w, model, currentAnswer+answer)
	} else {
		return currentAnswer + answer, nil
	}
}

func getFunctionCalls(data []byte) ([]map[string]any, error) {
	var functionCalls []map[string]any

	splittedData := bytes.SplitAfter(data, []byte("{/function_call}"))
	for _, part := range splittedData {
		if bytes.Contains(part, []byte("{@function_call}")) {
			functionCall := make(map[string]any)
			start := bytes.Index(part, []byte("{@function_call}"))
			end := bytes.Index(part, []byte("{/function_call}"))
			if start != -1 && end != -1 {
				functionCallData := part[start+len("{@function_call}") : end]
				err := json.Unmarshal(functionCallData, &functionCall)
				if err != nil {
					return nil, err
				}
				functionCalls = append(functionCalls, functionCall)
			}
		}
	}
	return functionCalls, nil
}

// UUID generates a UUIDv4 string.
func UUID() string {
	return uuid.New().String()
}

func init() {
	devMode = flag.Bool("dev", false, "Enable development mode")
	flag.Parse()
	if devMode != nil && *devMode {
		fmt.Println("Running in development mode")
	} else {
		fmt.Println("Running in production mode")
	}
}
