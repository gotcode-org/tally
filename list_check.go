package main

import (
	"fmt"
	"gotcode.org/tally/internal/store"
)

func main() {
	s, _ := store.NewStore("")
	tasks, _ := s.ListTasks("")
	for _, t := range tasks {
		fmt.Printf("Task ID: %s, Status: %s, Created: %s\n", t.ID, t.Status, t.CreatedAt.Format("2006-01-02"))
	}
}
