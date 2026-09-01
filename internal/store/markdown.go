/*
    Copyright (C) 2026 The GotCode Collective

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package store

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"gotcode.org/tally/internal/core"
)

// Store handles all disk I/O for reading and writing Markdown task files.
type Store struct {
	BaseDir string
}

// NewStore initializes a new Store. If baseDir is empty, it defaults to the XDG data directory.
func NewStore(baseDir string) (*Store, error) {
	if baseDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not find user home directory: %w", err)
		}
		// Default XDG Data Directory structure
		baseDir = filepath.Join(home, ".local", "share", "tally", "tasks")
	}

	return &Store{BaseDir: baseDir}, nil
}

// getTaskPath calculates the exact YYYY/MM/DD path for a given task ID.
// Assumes ID is formatted like "20260901.001".
func (s *Store) getTaskPath(id string) string {
	if len(id) < 8 {
		// Fallback if ID is malformed
		return filepath.Join(s.BaseDir, "misc", id+".md")
	}

	year := id[0:4]
	month := id[4:6]
	day := id[6:8]

	return filepath.Join(s.BaseDir, year, month, day, id+".md")
}

// Save writes a core.Task to disk as a Markdown file with YAML frontmatter.
func (s *Store) Save(task *core.Task) error {
	path := s.getTaskPath(task.ID)

	// Ensure the nested YYYY/MM/DD directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create directory tree: %w", err)
	}

	// Marshal the struct into YAML frontmatter
	frontmatter, err := yaml.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal yaml frontmatter: %w", err)
	}

	// Stitch it together: --- \n frontmatter \n --- \n body
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(frontmatter)
	buf.WriteString("---\n\n")
	buf.WriteString(strings.TrimSpace(task.Body))
	buf.WriteString("\n")

	// Write atomically (or just standard write for now)
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write markdown file: %w", err)
	}

	return nil
}

// Parse extracts a core.Task from a Markdown file reading the YAML frontmatter.
func (s *Store) Parse(path string) (*core.Task, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}

	task := &core.Task{}

	// Basic split to separate frontmatter from body
	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) >= 3 {
		// Parse the YAML frontmatter
		if err := yaml.Unmarshal([]byte(parts[1]), task); err != nil {
			return nil, fmt.Errorf("failed to unmarshal frontmatter in %s: %w", path, err)
		}
		// The rest is the body
		task.Body = strings.TrimSpace(parts[2])
	} else {
		// No frontmatter found, treat whole file as body
		task.Body = strings.TrimSpace(string(content))
	}

	return task, nil
}
