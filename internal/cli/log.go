/*
Copyright (C) 2026 The GotCode Collective
...
*/
package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"gotcode.org/tally/internal/core"
	"gotcode.org/tally/internal/store"
)

func newLogCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log [id] [duration]",
		Short: "Retroactively log time to a task (e.g., tally log 20260901.001 30m)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.NewStore("")
			if err != nil {
				return err
			}
			app := core.NewApp(s)

			id := args[0]
			durationStr := args[1]

			task, err := app.LogTime(id, durationStr)
			if err != nil {
				return err
			}

			dur := time.Duration(task.TotalSeconds) * time.Second
			fmt.Printf("Logged %s to %s.\nTotal time is now: %s\n", durationStr, task.ID, dur.String())
			return nil
		},
	}
	return cmd
}
