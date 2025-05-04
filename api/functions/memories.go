package functions

import (
	"framework/api/db"
	"context"
)

func MemoryCreate(content string) (error) {
	_, err := db.Conn.Exec(context.Background(), "INSERT INTO memory_cells (content) VALUES ($1)", content)	
	if err != nil {
		return err
	}
	return nil
}

func MemoryDelete(content string) (error) {
	_, err := db.Conn.Exec(context.Background(), "DELETE FROM memory_cells WHERE id = $1", content)
	if err != nil {
		return err
	}
	return nil
}
