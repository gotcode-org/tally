package cli

import (
	"github.com/spf13/cobra"
	"gotcode.org/tally/internal/ui"
)

func newUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Launch the interactive Tally Terminal User Interface",
		RunE: func(cmd *cobra.Command, args []string) error {
			
			items := []*ui.ListItem{
				{Title: "Design reusable Bubbletea TUI", Type: "Story", Status: "ACTIVE", Progress: 0.25, ProgressText: "25%", TypeColor: ui.CatMochaBlue},
				{Title: "Build Tally Sync Engine", Type: "Story", Status: "COMPLETED", Progress: 1.0, ProgressText: "100%", TypeColor: ui.CatMochaGreen},
				{Title: "Fix Azure DevOps identity bugs", Type: "Bug", Status: "COMPLETED", Progress: 1.0, ProgressText: "100%", TypeColor: ui.CatMochaRed},
			}

			return ui.RunList("TALLY DASHBOARD", items)
		},
	}
}
