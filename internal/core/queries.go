/*
    Copyright (C) 2026 The GotCode Collective
    ...
*/
package core

import "fmt"

// ListTasks returns all tasks, optionally filtered (filters to be added).
func (a *App) ListTasks() ([]*Task, error) {
	// In the future, we can add logic here to filter out "closed" tasks or filter by date
	tasks, err := a.Store.ListAll()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks from store: %w", err)
	}
	return tasks, nil
}
