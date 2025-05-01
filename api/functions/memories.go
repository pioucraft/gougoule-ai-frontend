package functions

import (
	"framework/api/db"
	"context"
	"encoding/json"
)

func MemoryCreate(content string) (error) {
	// Access content["content"] as memoryContent
	var memoryContent struct {
		Content string `json:"content"`
	}
	err := json.Unmarshal([]byte(content), &memoryContent)
	_, err = db.Conn.Exec(context.Background(), "INSERT INTO memory_cells (content) VALUES ($1)", memoryContent.Content)	
	if err != nil {
		return err
	}
	return nil
}
