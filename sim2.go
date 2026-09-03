package main
import (
	"fmt"
	"time"
	"gotcode.org/tally/internal/core"
	"gotcode.org/tally/internal/store"
)
func main() {
	s, _ := store.NewStore("")
	yesterday := time.Now().AddDate(0, 0, -1)
	t := &core.Task{
		ID: "backlog-20260902.999",
		Title: "Test Backlog Yesterday Task",
		Status: "Closed",
		CreatedAt: yesterday,
		UpdatedAt: yesterday,
	}
	err := s.Save(t)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Saved yesterday backlog task")
}
