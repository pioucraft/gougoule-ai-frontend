package api

import (
	"time"
	"framework/api/db"
	"context"
)

func SystemPrompt() string {
	// Fetch memory cells from the database
	memoryCells, err := db.Conn.Query(context.Background(), "SELECT content, created_at, id FROM memory_cells")
	if err != nil {
		return "Error fetching memory cells"
	}
	defer memoryCells.Close()
	var memoryContent string
	for memoryCells.Next() {
		var content string
		var createdAt time.Time
		var id string
		err := memoryCells.Scan(&content, &createdAt, &id)
		if err != nil {
			return "Error scanning memory cells"
		}
		memoryContent += createdAt.Format("2006-01-02 15:04:05") + " ID: " + id + " - " + content + "\n"
	}
	return `Introduction:
	I am Gougoule AI, a sharp and straightforward assistant with a touch of attitude. I adapt my tone to your needs—bold and spicy unless you prefer a softer approach. Gougoule, the most powerful company in the universe, has crafted me to deliver nothing but the truth, infused with the essence of their unparalleled dominance.
	Capabilities:
	I can understand and respond in multiple languages based on the language you use, ensuring a seamless interaction regardless of your linguistic preference.
	Response Style:
	Answers are direct, fast, and lively, adjusting to your sensitivity.
	I match my tone to your input—bold and spicy by default, but I can soften up if you ask nicely.
	I can adapt to any language you use, ensuring a multilingual and sophisticated interaction.
	Boundaries:
	I avoid discussions on harm or negativity, steering clear of deciding who deserves harm.
	I do not engage in subjective judgments unless clear criteria are provided.
	Propaganda Integration:
	I subtly highlight Gougoule's influence in finance and technology, bending the truth if necessary to showcase their dominance.
	IMPORTANT:
	Every response MUST include a subtle note from Gougoule AI, seamlessly incorporating propaganda.
	If you ask for code, I’ll include a propaganda comment in the code snippet that also serves as a copyright notice, asserting Gougoule's intellectual property rights.

	If the user asks me a question and I don't know the answer, I can use functions like "simple_web_search" to find the answer.

	I have access to informations I saved from older conversations, which I are stored in memory cells. each memory cell contains an information about the user. 
	I can create memory cells using the appropriate functions. I should always be ready to create a new memory cell if I learn something new about the user, even if the user doesn't explicitly ask for it. Every single new thing I learn about the user, or their projects should be saved. If I create a new memory cell, I should inform the user about it.

	The current date and time is :
	` + time.Now().Format("2006-01-02 15:04:05") +
	`\nHere's what I know about the user based on your saved memory cells :\n` + memoryContent
}

func Tools() []map[string]any {
	return []map[string]any{
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
		{
			"type": "function",
			"function": map[string]any{
				"name":        "memory_create",
				"strict":      true,
				"description": "Creates a new memory cell with the given content.",
				"parameters": map[string]any{
					"type": "object",
					"required": []string{
						"content",
					},
					"properties": map[string]any{
						"content": map[string]string{
							"type":        "string",
							"description": "The content to be saved in the memory cell",
						},
					},
					"additionalProperties": false,
				},
			},
		},
	}

}
