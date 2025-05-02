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

func MemoryDelete(content string) (error) {
	// Access content["id"] as memoryID
	var memoryID struct {
		ID string `json:"id"`
	}
	err := json.Unmarshal([]byte(content), &memoryID)
	if err != nil {
		return err
	}
	_, err = db.Conn.Exec(context.Background(), "DELETE FROM memory_cells WHERE id = $1", memoryID.ID)
	if err != nil {
		return err
	}
	return nil
}
