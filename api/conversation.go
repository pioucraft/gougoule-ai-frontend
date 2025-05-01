package api

import (
	"bufio"
	"bytes"
	"fmt"
	"framework/api/functions"
	"net/http"
	"encoding/json"
)

func Conversation(messages []map[string]any, w http.ResponseWriter, model string) (string, error) {
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
					calledFunction.function = function["name"].(string)
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

		var result string
		if calledFunction.function == "simple_web_search" {
			// Call the function
			result, err = functions.SimpleWebSearch(calledFunction.arguments)
			if err != nil {
				return "", err
			}
		} 
		messages = append(messages, map[string]any{
			"role":    "function",
			"name":    calledFunction.function, // Must match the name in function_call
			"content": result,
		})

		return Conversation(messages, w, model)
	}
	return answer, nil
}
