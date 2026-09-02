/*
Copyright (C) 2026 The GotCode Collective
...
*/
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Execute boots the Cobra CLI framework.
func Execute() {
	rootCmd := &cobra.Command{
		Use:   "tally",
		Short: "Tally is a blazing fast time tracker and task manager",
	}

	rootCmd.AddCommand(newAddCmd())
	rootCmd.AddCommand(newDeleteCmd())
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newLogCmd())
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newSyncCmd())
	rootCmd.AddCommand(newEditCmd())
	rootCmd.AddCommand(newStateCmd())
	rootCmd.AddCommand(newPointsCmd())
	rootCmd.AddCommand(newUICmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
