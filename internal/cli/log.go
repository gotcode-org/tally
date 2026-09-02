/*
Copyright (C) 2026 The GotCode Collective
...
*/
package cli

import (
	"gotcode.org/tally/internal/config"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"gotcode.org/tally/internal/core"
	"gotcode.org/tally/internal/store"
)

func newLogCmd() *cobra.Command {
	var activity string
	cmd := &cobra.Command{
		Use:   "log [id] [duration]",
		Short: "Retroactively log time to a task (e.g., tally log 20260901.001 30m)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.NewStore("")
			if err != nil {
				return err
			}
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			
			activityID := cfg.SevenPace.ActivityID
			if activity != "" {
				if idStr, exists := cfg.SevenPace.Activities[activity]; exists {
					activityID = idStr
				} else {
					return fmt.Errorf("activity '%s' not found in config", activity)
				}
			}

			app := core.NewApp(s)
			id := args[0]
			durationStr := args[1]

			task, err := app.LogTime(id, durationStr, activityID)
			if err != nil {
				return err
			}

			dur := time.Duration(task.TotalSeconds) * time.Second
			fmt.Printf("Logged %s to %s.\nTotal time is now: %s\n", durationStr, task.ID, dur.String())
			return nil
		},
	}
	cmd.Flags().StringVar(&activity, "activity", "", "The friendly name of the activity type")
	return cmd
}
