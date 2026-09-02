/*
Copyright (C) 2026 The GotCode Collective
...
*/
package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"gotcode.org/tally/internal/core"
	"gotcode.org/tally/internal/store"
)

func newAddCmd() *cobra.Command {
	var adoType string
	var tags []string
	var isBacklog bool
	var recurrence string
	var parentID string

	cmd := &cobra.Command{
		Use:   "add [title]",
		Short: "Create a new task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Initialize storage (Presentation layer doing setup)
			s, err := store.NewStore("")
			if err != nil {
				return err
			}
			app := core.NewApp(s)

			// Execute business logic
			task, err := app.AddTask(args[0], adoType, tags, isBacklog, recurrence, parentID)
			if err != nil {
				return err
			}

			// Format presentation output
			fmt.Printf("Created task %s: %s\n", task.ID, task.Title)
			return nil
		},
	}

	cmd.Flags().StringVar(&adoType, "type", "", "ADO Work Item Type (e.g., Story, Bug)")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "Comma-separated list of tags")
	cmd.Flags().BoolVar(&isBacklog, "backlog", false, "Send the task directly to the backlog")
	cmd.Flags().StringVar(&recurrence, "recur", "", "Set a recurrence rule (e.g., daily)")
	cmd.Flags().StringVar(&parentID, "parent", "", "The Tally ID of the parent Story")

	return cmd
}
