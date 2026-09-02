/*
Copyright (C) 2026 The GotCode Collective
...
*/
package cli

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"gotcode.org/tally/internal/core"
	"gotcode.org/tally/internal/store"
)

func newPointsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "points [id] [points]",
		Short: "Assign story points to a task (e.g. tally points 20260901.001 5)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			
			points, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return fmt.Errorf("points must be a valid number: %w", err)
			}

			s, err := store.NewStore("")
			if err != nil {
				return err
			}
			app := core.NewApp(s)

			task, err := app.SetPoints(id, points)
			if err != nil {
				return err
			}

			fmt.Printf("Task %s is now estimated at %g points\n", task.ID, *task.StoryPoints)
			return nil
		},
	}
	return cmd
}
