/*
    Copyright (C) 2026 The GotCode Collective
    ...
*/
package core

import (
	"gotcode.org/tally/internal/store"
)

// App acts as the CQRS orchestrator containing all business logic.
type App struct {
	Store *store.Store
}

// NewApp initializes the business logic layer.
func NewApp(s *store.Store) *App {
	return &App{Store: s}
}
