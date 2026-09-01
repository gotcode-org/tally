/*
    Copyright (C) 2026 The GotCode Collective
    ...
*/
package core

import (
	"fmt"
	"time"
)

// AddTask contains the pure business logic for creating a new task.
func (a *App) AddTask(title string, adoType string, tags []string) (*Task, error) {
	now := time.Now()

	// 1. Generate the sequential local ID
	id, err := a.Store.GetNextID(now)
	if err != nil {
		return nil, fmt.Errorf("failed to generate task ID: %w", err)
	}

	// 2. Construct the Domain Model
	task := &Task{
		ID:        id,
		Title:     title,
		Status:    StateOpen,
		Tags:      tags,
		CreatedAt: now,
		UpdatedAt: now,
		ADOType:   adoType,
	}

	// 3. Save it via the Storage Engine
	if err := a.Store.Save(task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	return task, nil
}
