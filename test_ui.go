package main
import (
	"fmt"
	"strings"
	"gotcode.org/tally/internal/store"
	"gotcode.org/tally/internal/core"
)
func main() {
	s, _ := store.NewStore("")
	tasks, _ := s.ListTasks("")
	for i, j := 0, len(tasks)-1; i < j; i, j = i+1, j-1 {
		tasks[i], tasks[j] = tasks[j], tasks[i]
	}
	var topLevelTasks []*core.Task
	for _, t := range tasks {
		if t.ParentID == "" {
			topLevelTasks = append(topLevelTasks, t)
		}
	}
	
	archiveStates := []string{"Done", "Closed", "Removed"}
	isArchive := func(s string) bool {
		for _, as := range archiveStates {
			if strings.EqualFold(s, as) {
				return true
			}
		}
		return false
	}
	
	archiveNodes := 0
	for _, t := range topLevelTasks {
		if isArchive(string(t.Status)) {
			fmt.Printf("Archived: %s (Status: %s, Parent: %s, Updated: %s)\n", t.Title, t.Status, t.ParentID, t.UpdatedAt.Format("2006-01-02"))
			archiveNodes++
		}
	}
	fmt.Printf("Total Archive Nodes: %d\n", archiveNodes)
}
