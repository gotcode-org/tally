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
func (a *App) AddTask(title string, adoType string, tags []string, isBacklog bool, recurrence string) (*Task, error) {
	now := time.Now()

	// 1. Generate the sequential local ID
	id, err := a.Store.GetNextID(now, isBacklog, recurrence != "")
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
		Recurrence: recurrence,
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


// StartTask moves a task from the backlog to today's date, generating a new ID.
func (a *App) StartTask(id string) error {
	task, err := a.Store.Load(id)
	if err != nil {
		return fmt.Errorf("failed to load task: %w", err)
	}

	// Delete the old file
	if err := a.Store.Delete(id); err != nil {
		return fmt.Errorf("failed to delete old task file: %w", err)
	}

	// Generate new ID for today
	now := time.Now()
	newID, err := a.Store.GetNextID(now, false, false)
	if err != nil {
		// Try to recover by saving to old path?
		return fmt.Errorf("failed to generate new ID: %w", err)
	}

	// Update task and save
	task.ID = newID
	task.CreatedAt = now
	task.UpdatedAt = now
	
	if err := a.Store.Save(task); err != nil {
		return fmt.Errorf("failed to save migrated task: %w", err)
	}
	
	// Also need to move the physical markdown body content since Save only writes frontmatter right now?
	// Wait, does Save() write the body?
	// Let's check how Save is implemented!
	return nil
}


// ReconcileRecurringTasks scans the recurring folder and clones any missing tasks into today's list.
func (a *App) ReconcileRecurringTasks() error {
	tasks, err := a.Store.ListTasks("")
	if err != nil {
		return err
	}

	// 1. Separate templates from today's tasks
	var templates []*Task
	todayTitles := make(map[string]bool)
	now := time.Now()
	
	todayPrefix := fmt.Sprintf("%s%s%s", now.Format("2006"), now.Format("01"), now.Format("02"))

	for _, t := range tasks {
		if strings.HasPrefix(t.ID, "recur-") {
			templates = append(templates, t)
		} else if strings.HasPrefix(t.ID, todayPrefix) {
			todayTitles[t.Title] = true
		}
	}

	// 2. Clone any missing templates
	for _, template := range templates {
		if !todayTitles[template.Title] {
			// It's missing today! Clone it.
			newID, err := a.Store.GetNextID(now, false, false)
			if err != nil {
				continue
			}

			newTask := &Task{
				ID:          newID,
				Title:       template.Title,
				Status:      StateOpen,
				Tags:        template.Tags,
				CreatedAt:   now,
				UpdatedAt:   now,
				ADOType:     template.ADOType,
				StoryPoints: template.StoryPoints,
				Recurrence:  template.Recurrence, // Keep the tag so UI can draw the icon
				TemplateID:  template.ID,
				Body:        template.Body,
			}
			
			a.Store.Save(newTask)
			todayTitles[template.Title] = true // prevent duplicate cloning if there are multiple same-titled templates
		}
	}

	return nil
}
