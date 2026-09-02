package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gotcode.org/tally/internal/config"
)

type ActivityResponse struct {
	// 7pace might return under "data" or "value" depending on the specific API endpoint version
	Data []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"data"`
	Value []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"value"`
}

func newActivitiesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activities",
		Short: "Fetch and display available 7pace Activity Types and their GUIDs",
		RunE: func(cmd *cobra.Command, args []string) error {
			sevenPaceToken := os.Getenv("TALLY_7PACE_TOKEN")
			if sevenPaceToken == "" {
				sevenPaceToken = os.Getenv("TALLY_ADO_PAT") // Fallback
			}
			if sevenPaceToken == "" {
				return fmt.Errorf("FATAL: TALLY_7PACE_TOKEN environment variable is not set")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}
			if cfg.ADO.Organization == "" {
				return fmt.Errorf("FATAL: ADO Organization is missing in config. Run 'tally config'")
			}

			url := fmt.Sprintf("https://%s.timehub.7pace.com/api/rest/activityTypes?api-version=3.1", cfg.ADO.Organization)
			fmt.Printf("Fetching Activity Types from %s...\n\n", url)

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return err
			}
			req.Header.Set("Authorization", "Bearer "+sevenPaceToken)
			req.Header.Set("Accept", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("network error hitting 7pace: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				body, _ := io.ReadAll(resp.Body)
				return fmt.Errorf("7pace API returned HTTP %d: %s", resp.StatusCode, string(body))
			}

			var payload ActivityResponse
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				return fmt.Errorf("failed to decode JSON response: %w", err)
			}

			items := payload.Data
			if len(items) == 0 {
				items = payload.Value
			}

			if len(items) == 0 {
				fmt.Println("No activity types found or unexpected JSON format.")
				return nil
			}

			// Print the table
			fmt.Printf("%-40s | %s\n", "ACTIVITY NAME", "GUID")
			fmt.Println(strings.Repeat("-", 41) + "+" + strings.Repeat("-", 38))
			for _, item := range items {
				fmt.Printf("%-40s | %s\n", item.Name, item.ID)
			}
			fmt.Println()

			return nil
		},
	}
	return cmd
}
