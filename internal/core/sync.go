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
		// Never sync raw templates, only their instantiated clones!
		if strings.HasPrefix(t.ID, "recur-") {
			continue
		}

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
				desc, ac := parseMarkdownSections(t.Body)
				
				if desc != "" {
					patch = append(patch, map[string]interface{}{
						"op": "add", "path": "/fields/System.Description", "value": desc,
					})
				}
				if ac != "" {
					patch = append(patch, map[string]interface{}{
						"op": "add", "path": "/fields/Microsoft.VSTS.Common.AcceptanceCriteria", "value": ac,
					})
				}
			}
			
			if len(t.Tags) > 0 {
				patch = append(patch, map[string]interface{}{
					"op": "add", "path": "/fields/System.Tags", "value": strings.Join(t.Tags, "; "),
				})
			}
			
			if t.StoryPoints != nil {
				patch = append(patch, map[string]interface{}{
					"op": "add", "path": "/fields/Microsoft.VSTS.Scheduling.StoryPoints", "value": *t.StoryPoints,
				})
			}
			
			// Handle ADO Hierarchy Linking
			if t.ParentID != "" {
				parentTask, err := a.Store.Load(t.ParentID)
				if err == nil && parentTask.ADOID != nil {
					patch = append(patch, map[string]interface{}{
						"op": "add",
						"path": "/relations/-",
						"value": map[string]interface{}{
							"rel": "System.LinkTypes.Hierarchy-Reverse", // I am a child of this parent
							"url": fmt.Sprintf("%s/_apis/wit/workitems/%d", strings.TrimRight(cfg.ADO.Organization, "/"), *parentTask.ADOID),
							"attributes": map[string]interface{}{
								"comment": "Linked via Tally",
							},
						},
					})
				}
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

		// 1.5 Update State if Closed
		if t.ADOID != nil {
			patch := []map[string]interface{}{}
			
			// Always sync title
			patch = append(patch, map[string]interface{}{
				"op": "add", "path": "/fields/System.Title", "value": t.Title,
			})
			
			// Sync state if specified and not generic defaults
			if t.Status != "open" && t.Status != "active" {
				patch = append(patch, map[string]interface{}{
					"op": "add", "path": "/fields/System.State", "value": t.Status,
				})
			}
			
			if t.StoryPoints != nil {
				patch = append(patch, map[string]interface{}{
					"op": "add", "path": "/fields/Microsoft.VSTS.Scheduling.StoryPoints", "value": *t.StoryPoints,
				})
			}
			
			// Handle ADO Hierarchy Linking
			if t.ParentID != "" {
				parentTask, err := a.Store.Load(t.ParentID)
				if err == nil && parentTask.ADOID != nil {
					patch = append(patch, map[string]interface{}{
						"op": "add",
						"path": "/relations/-",
						"value": map[string]interface{}{
							"rel": "System.LinkTypes.Hierarchy-Reverse", // I am a child of this parent
							"url": fmt.Sprintf("%s/_apis/wit/workitems/%d", strings.TrimRight(cfg.ADO.Organization, "/"), *parentTask.ADOID),
							"attributes": map[string]interface{}{
								"comment": "Linked via Tally",
							},
						},
					})
				}
			}
			
			if t.Body != "" {
				desc, ac := parseMarkdownSections(t.Body)
				if desc != "" {
					patch = append(patch, map[string]interface{}{
						"op": "add", "path": "/fields/System.Description", "value": desc,
					})
				}
				if ac != "" {
					patch = append(patch, map[string]interface{}{
						"op": "add", "path": "/fields/Microsoft.VSTS.Common.AcceptanceCriteria", "value": ac,
					})
				}
			}
			
			if len(patch) > 0 {
				payload, _ := json.Marshal(patch)
				url := fmt.Sprintf("%s/%s/_apis/wit/workitems/%d?api-version=7.0", strings.TrimRight(cfg.ADO.Organization, "/"), cfg.ADO.DefaultProject, *t.ADOID)
				
				req, _ := http.NewRequest("PATCH", url, bytes.NewBuffer(payload))
				req.Header.Set("Content-Type", "application/json-patch+json")
				req.SetBasicAuth("", adoPat)
				
				resp, err := client.Do(req)
				if err == nil {
					body, _ := io.ReadAll(resp.Body)
					if resp.StatusCode >= 400 {
						fmt.Printf("  -> ADO rejected update for task %s (HTTP %d). Response: %s\n", t.ID, resp.StatusCode, string(body))
					} else {
						fmt.Printf("  -> Successfully updated ADO Work Item #%d\n", *t.ADOID)
						// Dump the payload we sent and the ADO response for debugging
						fmt.Printf("     [DEBUG] Sent Payload: %s\n", string(payload))
						// Only print the first 200 chars of response so we don't flood the terminal
						respStr := string(body)
						if len(respStr) > 500 {
							respStr = respStr[:500] + "..."
						}
						fmt.Printf("     [DEBUG] ADO Response: %s\n", respStr)
					}
					resp.Body.Close()
				}
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

// parseMarkdownSections splits the body into Description and Acceptance Criteria
func parseMarkdownSections(body string) (description, acceptanceCriteria string) {
	lines := strings.Split(body, "\n")
	
	var currentSection string = "description"
	var descBuilder strings.Builder
	var acBuilder strings.Builder
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		
		// Skip headers
		if strings.HasPrefix(lower, "# description") {
			currentSection = "description"
			continue
		} else if strings.HasPrefix(lower, "# acceptance criteria") {
			currentSection = "ac"
			continue
		}
		
		// ADO WYSIWYG requires proper block elements, bare text with <br> is often swallowed
		htmlLine := line
		if htmlLine == "" {
			htmlLine = "<br>"
		}
		
		if currentSection == "description" {
			descBuilder.WriteString(fmt.Sprintf("<div>%s</div>", htmlLine))
		} else if currentSection == "ac" {
			acBuilder.WriteString(fmt.Sprintf("<div>%s</div>", htmlLine))
		}
	}
	
	return strings.TrimSpace(descBuilder.String()), strings.TrimSpace(acBuilder.String())
}
