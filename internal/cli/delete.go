package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"gotcode.org/tally/internal/core"
	"gotcode.org/tally/internal/store"
)

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a task by ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.NewStore("")
			if err != nil {
				return err
			}
			app := core.NewApp(s)

			err = app.DeleteTask(args[0])
			if err != nil {
				return err
			}

			fmt.Printf("Deleted task: %s\n", args[0])
			return nil
		},
	}
	return cmd
}
