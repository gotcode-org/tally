package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gotcode.org/tally/internal/store"
)

func newMigrateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Migrate legacy backlog files to the new architecture and rewrite IDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			// You would inject the real data dir here if the user configured it, but let's use default
			s, err := store.NewStore("")
			if err != nil {
				return err
			}
			
			backlogDir := filepath.Join(s.BaseDir, "backlog")
			fmt.Printf("Scanning %s for legacy backlog tasks...\n", backlogDir)
			
			entries, err := os.ReadDir(backlogDir)
			if err != nil {
				if os.IsNotExist(err) {
					fmt.Println("No legacy backlog directory found. You are good to go!")
					return nil
				}
				return err
			}
			
			idMap := make(map[string]string)
			count := 0
			
			for _, e := range entries {
				if !e.IsDir() && strings.HasPrefix(e.Name(), "backlog-") && strings.HasSuffix(e.Name(), ".md") {
					oldID := strings.TrimSuffix(e.Name(), ".md")
					
					// Extract date from "backlog-YYYYMMDD.001"
					if len(oldID) < 16 {
						continue // Malformed
					}
					
					dateStr := oldID[8:16] // YYYYMMDD
					parsedDate, err := time.Parse("20060102", dateStr)
					if err != nil {
						parsedDate = time.Now()
					}
					
					newID, err := s.GetNextID(parsedDate, false)
					if err != nil {
						fmt.Printf("Failed to generate new ID for %s\n", oldID)
						continue
					}
					
					// Read old file
					oldPath := filepath.Join(backlogDir, e.Name())
					data, err := os.ReadFile(oldPath)
					if err != nil {
						continue
					}
					content := string(data)
					
					// Update Frontmatter
					content = strings.Replace(content, "status: open", "status: Backlog", 1)
					content = strings.Replace(content, "status: New", "status: Backlog", 1)
					content = strings.Replace(content, "id: "+oldID, "id: "+newID, 1)
					
					// Write to new standard path
					newPath := s.GetTaskPath(newID)
					os.MkdirAll(filepath.Dir(newPath), 0755)
					os.WriteFile(newPath, []byte(content), 0600)
					
					// Delete old legacy file
					os.Remove(oldPath)
					
					idMap[oldID] = newID
					count++
				}
			}
			
			fmt.Printf("Migrated %d legacy tasks to new IDs. Now scanning entire filesystem for parent_id references...\n", count)
			
			// Recursively scan entire store to fix broken parent_id links
			updatedLinks := 0
			filepath.Walk(s.BaseDir, func(path string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
					return nil
				}
				
				data, err := os.ReadFile(path)
				if err != nil {
					return nil
				}
				
				content := string(data)
				modified := false
				
				for oldID, newID := range idMap {
					if strings.Contains(content, "parent_id: "+oldID) {
						content = strings.Replace(content, "parent_id: "+oldID, "parent_id: "+newID, -1)
						modified = true
					}
				}
				
				if modified {
					os.WriteFile(path, []byte(content), 0600)
					updatedLinks++
				}
				
				return nil
			})
			
			fmt.Printf("Successfully repaired %d parent-child links.\nMigration complete! Legacy backlog folder is officially deprecated.\n", updatedLinks)
			return nil
		},
	}
}
