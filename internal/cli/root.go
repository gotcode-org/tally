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
	rootCmd.AddCommand(newListCmd())
	rootCmd.AddCommand(newLogCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
