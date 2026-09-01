/*
    Copyright (C) 2026 The GotCode Collective
    ...
*/
package cli

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"gotcode.org/tally/internal/core"
	"gotcode.org/tally/internal/store"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := store.NewStore("")
			if err != nil {
				return err
			}
			app := core.NewApp(s)

			tasks, err := app.ListTasks()
			if err != nil {
				return err
			}

			if len(tasks) == 0 {
				fmt.Println("No tasks found.")
				return nil
			}

			// Initialize tabwriter for clean columns (like kubectl)
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			
			// Print header
			fmt.Fprintln(w, "ID\tSTATE\tTYPE\tTIME\tTITLE")

			for _, t := range tasks {
				// Convert total seconds to a human readable duration (e.g., "1h30m")
				dur := time.Duration(t.TotalSeconds) * time.Second
				timeStr := dur.String()
				if t.TotalSeconds == 0 {
					timeStr = "0m"
				}

				adoType := t.ADOType
				if adoType == "" {
					adoType = "-"
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", t.ID, t.Status, adoType, timeStr, t.Title)
			}
			
			w.Flush()
			return nil
		},
	}
	return cmd
}
