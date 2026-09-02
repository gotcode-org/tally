/*
Copyright (C) 2026 The GotCode Collective
...
*/
package core

import (
	"fmt"
	"strings"
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

	var defaultPoints float64 = 1.0

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

	// Default to 1 point if the type is a Story to satisfy strict ADO requirements
	if strings.Contains(strings.ToLower(adoType), "story") {
		task.StoryPoints = &defaultPoints
	}

	// 3. Save it via the Storage Engine
	if err := a.Store.Save(task); err != nil {
		return nil, fmt.Errorf("failed to save task: %w", err)
	}

	return task, nil
}

// LogTime retroactively adds time to a task.
func (a *App) LogTime(id string, durationStr string) (*Task, error) {
	dur, err := time.ParseDuration(durationStr)
	if err != nil {
		return nil, fmt.Errorf("invalid duration format (e.g. 30m, 1h30m): %w", err)
	}

	task, err := a.Store.Load(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load task %s: %w", id, err)
	}

	task.TotalSeconds += int(dur.Seconds())
	task.UpdatedAt = time.Now()

	if err := a.Store.Save(task); err != nil {
		return nil, fmt.Errorf("failed to save task after logging time: %w", err)
	}

	return task, nil
}

// SetState manually updates the status of a task.
func (a *App) SetState(id string, state string) (*Task, error) {
	task, err := a.Store.Load(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load task %s: %w", id, err)
	}

	task.Status = TaskState(state)
	task.UpdatedAt = time.Now()

	if err := a.Store.Save(task); err != nil {
		return nil, fmt.Errorf("failed to save task after updating state: %w", err)
	}

	return task, nil
}

// SetPoints updates the story points for a task.
func (a *App) SetPoints(id string, points float64) (*Task, error) {
	task, err := a.Store.Load(id)
	if err != nil {
		return nil, fmt.Errorf("failed to load task %s: %w", id, err)
	}

	task.StoryPoints = &points
	task.UpdatedAt = time.Now()

	if err := a.Store.Save(task); err != nil {
		return nil, fmt.Errorf("failed to save task after updating points: %w", err)
	}

	return task, nil
}


// DeleteTask removes the task from disk.
func (a *App) DeleteTask(id string) error {
	return a.Store.Delete(id)
}
