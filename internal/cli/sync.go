/*
Copyright (C) 2026 The GotCode Collective
...
*/
package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gotcode.org/tally/internal/config"
	"gotcode.org/tally/internal/core"
	"gotcode.org/tally/internal/store"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Push all offline tasks and timesheets to Azure DevOps and 7pace",
		RunE: func(cmd *cobra.Command, args []string) error {
			adoPat := os.Getenv("TALLY_ADO_PAT")
			if adoPat == "" {
				return fmt.Errorf("FATAL: TALLY_ADO_PAT environment variable is not set")
			}
			
			sevenPaceToken := os.Getenv("TALLY_7PACE_TOKEN")
			if sevenPaceToken == "" {
				// Fallback to ADO PAT just in case, but 7pace usually requires its own token
				sevenPaceToken = adoPat
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			if cfg.ADO.Organization == "" {
				return fmt.Errorf("FATAL: ADO Organization is missing. Run 'tally config'")
			}

			s, err := store.NewStore("")
			if err != nil {
				return err
			}
			app := core.NewApp(s)

			fmt.Println("Initializing Tally Enterprise Sync Engine...")
			if err := app.Sync(cfg, adoPat, sevenPaceToken); err != nil {
				return err
			}

			fmt.Println("\nSync complete.")
			return nil
		},
	}
	return cmd
}
