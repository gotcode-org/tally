/*
    Copyright (C) 2026 The GotCode Collective
    ...
*/
package core

import "fmt"

// ListTasks returns all tasks, optionally filtered by a date prefix (YYYY, YYYYMM, YYYYMMDD).
func (a *App) ListTasks(datePrefix string) ([]*Task, error) {
	tasks, err := a.Store.ListTasks(datePrefix)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tasks from store: %w", err)
	}
	return tasks, nil
}
