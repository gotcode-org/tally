/*
    Copyright (C) 2026 The GotCode Collective
    ...
*/
package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"gotcode.org/tally/internal/config"
)

type ADOResponse struct {
	ID int `json:"id"`
}

// Sync executes the network operations to push local data to ADO and 7pace.
func (a *App) Sync(cfg *config.Config, adoPat string, sevenPaceToken string) error {
	tasks, err := a.Store.ListTasks("")
	if err != nil {
		return fmt.Errorf("failed to list tasks for sync: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	orgName := extractOrgName(cfg.ADO.Organization)

	for _, t := range tasks {
		if t.ADOID == nil {
			fmt.Printf("Syncing task %s to ADO...\n", t.ID)
			
			adoType := t.ADOType
			if adoType == "" {
				adoType = "Task"
			}

			patch := []map[string]interface{}{
				{"op": "add", "path": "/fields/System.Title", "value": t.Title},
				{"op": "add", "path": "/fields/System.AreaPath", "value": cfg.ADO.DefaultArea},
			}
			
			if cfg.User.Email != "" {
				patch = append(patch, map[string]interface{}{
					"op": "add", "path": "/fields/System.AssignedTo", "value": cfg.User.Email,
				})
			}
			
			if t.Body != "" {
				patch = append(patch, map[string]interface{}{
					"op": "add", "path": "/fields/System.Description", "value": t.Body,
				})
			}
			
			if len(t.Tags) > 0 {
				patch = append(patch, map[string]interface{}{
					"op": "add", "path": "/fields/System.Tags", "value": strings.Join(t.Tags, "; "),
				})
			}

			payload, _ := json.Marshal(patch)
			url := fmt.Sprintf("%s/%s/_apis/wit/workitems/$%s?api-version=7.0", strings.TrimRight(cfg.ADO.Organization, "/"), cfg.ADO.DefaultProject, strings.ReplaceAll(adoType, " ", "%20"))
			
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json-patch+json")
			req.SetBasicAuth("", adoPat)

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("network error hitting ADO: %w", err)
			}
			
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				var adoResp ADOResponse
				body, _ := io.ReadAll(resp.Body)
				json.Unmarshal(body, &adoResp)
				resp.Body.Close()
				
				t.ADOID = &adoResp.ID
				t.UpdatedAt = time.Now()
				a.Store.Save(t)
				fmt.Printf("  -> Successfully created ADO Work Item #%d\n", *t.ADOID)
			} else if resp.StatusCode == 400 && cfg.User.Email != "" {
				resp.Body.Close()
				fmt.Printf("  -> ADO rejected the AssignedTo identity. Retrying without assignment...\n")
				
				fallbackPatch := []map[string]interface{}{}
				for _, p := range patch {
					if p["path"] != "/fields/System.AssignedTo" {
						fallbackPatch = append(fallbackPatch, p)
					}
				}
				
				payload, _ = json.Marshal(fallbackPatch)
				req, _ = http.NewRequest("POST", url, bytes.NewBuffer(payload))
				req.Header.Set("Content-Type", "application/json-patch+json")
				req.SetBasicAuth("", adoPat)
				
				resp2, err := client.Do(req)
				if err == nil && resp2.StatusCode >= 200 && resp2.StatusCode < 300 {
					var adoResp ADOResponse
					body, _ := io.ReadAll(resp2.Body)
					json.Unmarshal(body, &adoResp)
					resp2.Body.Close()
					
					t.ADOID = &adoResp.ID
					t.UpdatedAt = time.Now()
					a.Store.Save(t)
					fmt.Printf("  -> Successfully created ADO Work Item #%d (Unassigned)\n", *t.ADOID)
				} else {
					if err == nil {
						body, _ := io.ReadAll(resp2.Body)
						resp2.Body.Close()
						fmt.Printf("  -> Fallback also failed (HTTP %d). Response: %s\n", resp2.StatusCode, string(body))
					}
				}
			} else {
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				fmt.Printf("  -> Failed to create ADO item (HTTP %d). Response: %s\n", resp.StatusCode, string(body))
				continue
			}
		}

		unsynced := t.UnsyncedSeconds()
		if unsynced > 0 && t.ADOID != nil {
			fmt.Printf("Pushing %d seconds of time to 7pace for ADO #%d...\n", unsynced, *t.ADOID)
			
			logData := map[string]interface{}{
				"timestamp":  time.Now().Format(time.RFC3339),
				"length":     unsynced,
				"workItemId": *t.ADOID,
				"comment":    "Logged via Tally terminal",
			}
			
			// 7pace strictly requires a valid GUID. If the user typed "Development", it throws a 500.
			if len(cfg.SevenPace.ActivityID) >= 32 {
				logData["activityTypeId"] = cfg.SevenPace.ActivityID
			}
			
			payload, _ := json.Marshal(logData)
			url := fmt.Sprintf("https://%s.timehub.7pace.com/api/rest/workLogs?api-version=3.1", orgName)
			
			req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+sevenPaceToken)

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("network error hitting 7pace: %w", err)
			}
			
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				t.SyncedSeconds = t.TotalSeconds
				a.Store.Save(t)
				fmt.Printf("  -> Time successfully logged to 7pace!\n")
			} else {
				body, _ := io.ReadAll(resp.Body)
				fmt.Printf("  -> Failed to log time to 7pace (HTTP %d). Response: %s\n", resp.StatusCode, string(body))
			}
			resp.Body.Close()
		}
	}

	return nil
}

func extractOrgName(adoURL string) string {
	url := strings.TrimRight(adoURL, "/")
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
