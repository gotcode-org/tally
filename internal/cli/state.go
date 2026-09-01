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

func newStateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state [id] [state]",
		Short: "Change the state of a task (e.g. tally state 20260901.001 Closed)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			state := args[1]

			s, err := store.NewStore("")
			if err != nil {
				return err
			}
			app := core.NewApp(s)

			task, err := app.SetState(id, state)
			if err != nil {
				return err
			}

			fmt.Printf("Task %s is now marked as '%s'\n", task.ID, task.Status)
			return nil
		},
	}
	return cmd
}
