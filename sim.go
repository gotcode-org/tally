package main
import (
	"fmt"
	"time"
	"gotcode.org/tally/internal/core"
	"gotcode.org/tally/internal/store"
)
func main() {
	s, _ := store.NewStore("")
	// Create a task for yesterday
	yesterday := time.Now().AddDate(0, 0, -1)
	t := &core.Task{
		ID: "20260902.999",
		Title: "Test Yesterday Task",
		Status: "Closed",
		CreatedAt: yesterday,
		UpdatedAt: yesterday,
	}
	s.Save(t)
	fmt.Println("Saved yesterday task")
}
