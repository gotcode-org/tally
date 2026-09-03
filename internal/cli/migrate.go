package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gotcode.org/tally/internal/config"
)

func newMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate legacy files to the new architecture",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, _ := os.UserHomeDir()
			baseDir := filepath.Join(home, ".local", "share", "tally", "tasks")
			
			cfg, _ := config.Load()
			if cfg != nil && cfg.ADO.Organization != "" {
				// future custom data dir checking
			}
			
			backlogDir := filepath.Join(baseDir, "backlog")
			fmt.Printf("Scanning %s for legacy backlog tasks...\n", backlogDir)
			
			entries, err := os.ReadDir(backlogDir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No legacy backlog directory found. You are good to go!")
					return nil
				}
				return err
			}
			
			count := 0
			for _, e := range entries {
				if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
					path := filepath.Join(backlogDir, e.Name())
					data, err := os.ReadFile(path)
					if err != nil {
						continue
					}
					
					content := string(data)
					// Replace status: open or status: New with status: Backlog
					content = strings.Replace(content, "status: open", "status: Backlog", 1)
					content = strings.Replace(content, "status: New", "status: Backlog", 1)
					
					os.WriteFile(path, []byte(content), 0600)
					count++
				}
			}
			
			fmt.Printf("Successfully migrated %d legacy backlog tasks! They will now appear in your Backlog folder in the UI.\n", count)
			return nil
		},
	}
}
