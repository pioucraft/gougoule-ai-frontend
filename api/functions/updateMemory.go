package functions

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
)

func UpdateMemory(query string, Conn *pgxpool.Pool, ctx context.Context) error {
	// get query.memory from query
	var memoryInput struct {
		Memory string `json:"memory"`
	}
	err := json.Unmarshal([]byte(query), &memoryInput)
	if err != nil {
		return err
	}

	// Update the database with the new memory string
	_, err = Conn.Exec(ctx, "UPDATE memory_cells SET content = $1 WHERE name = $2", memoryInput.Memory, "general")
	if err != nil {
		return err
	}
	return nil
}
