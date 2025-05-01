package functions

import (
	"encoding/json"
	"database/sql"

	"context"
)

func UpdateMemory(query string, Conn *sql.Conn, ctx context.Context) error {
	// get query.memory from query
	var memoryInput struct {
		Memory string `json:"memory"`
	}
	err := json.Unmarshal([]byte(query), &memoryInput)
	if err != nil {
		return err
	}

	// Update the database with the new memory string
	_, err = Conn.ExecContext(ctx, "UPDATE memory SET memory = $1 WHERE name = $2", memoryInput.Memory, "general")
	if err != nil {
		return err
	}
	return nil
}
