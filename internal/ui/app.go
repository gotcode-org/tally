/*
Copyright (C) 2026 The GotCode Collective

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package ui

import (
	"fmt"
	"time"
	"gotcode.org/tally/internal/config"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"gotcode.org/tally/internal/core"
)

type AppState int

const (
	StateDashboard AppState = iota
	StateCreateTask
	StateLogTime
)

type MainModel struct {
	coreApp       *core.App
	state         AppState
	list          ListModel
	expandedState map[string]bool
	form    *FormModel
	
	selectedID     string
	terminalWidth  int
	terminalHeight int
}

func NewMainModel(app *core.App) *MainModel {
	if cfg, err := config.Load(); err == nil {
		if cfg.UI.Theme != "" {
			ApplyThemeByName(cfg.UI.Theme)
		}
	}
	m := &MainModel{
		coreApp:       app,
		state:         StateDashboard,
		expandedState: make(map[string]bool),
	}
	m.reloadList()
	return m
}

func (m *MainModel) reloadList() {
	// Capture current expansion state and cursor before wiping
	var selectedID string
	if m.list.items != nil {
		flat := m.list.getFlatRows()
		if m.list.cursor >= 0 && m.list.cursor < len(flat) {
			selectedID = flat[m.list.cursor].Item.ID
		}
		
		var captureState func(items []*ListItem)
		captureState = func(items []*ListItem) {
			for _, item := range items {
				// Use ID if available, otherwise fallback to Title (for folders)
				key := item.ID
				if key == "" {
					key = "folder:" + item.Title
				}
				m.expandedState[key] = item.Expanded
				captureState(item.Children)
			}
		}
		captureState(m.list.items)
	}

	tasks, _ := m.coreApp.Store.ListTasks("")

	// Reverse tasks so newest are at the top
	for i, j := 0, len(tasks)-1; i < j; i, j = i+1, j-1 {
		tasks[i], tasks[j] = tasks[j], tasks[i]
	}

	var items []*ListItem
	var lastYear, lastMonth, lastDay *ListItem
	var todayNode *ListItem
	var backlogNode *ListItem
	var templatesNode *ListItem
	
	// 1. Map out all task ListItems so we can nest children easily
	itemMap := make(map[string]*ListItem)
	
	// 2. Pre-generate all ListItems
	for _, t := range tasks {
		c := ThemeBlue
		if strings.Contains(strings.ToLower(t.ADOType), "story") {
			c = ThemeMauve
		} else if strings.Contains(strings.ToLower(t.ADOType), "bug") {
			c = ThemeRed
		}

		timeStr := ""
		if t.TotalSeconds > 0 {
			hours := float64(t.TotalSeconds) / 3600.0
			timeStr = fmt.Sprintf("%.1fh", hours)
		}

		titleStr := t.Title
		if t.Recurrence != "" {
			titleStr = "↻ " + titleStr
		}

		expanded := true
		if val, exists := m.expandedState[t.ID]; exists {
			expanded = val
		}

		itemMap[t.ID] = &ListItem{
			ID:        t.ID,
			Title:     titleStr,
			Type:      t.ADOType,
			Status:    string(t.Status),
			TimeText:  timeStr,
			TypeColor: c,
			Expanded:  expanded, // keep children visible by default or use captured state
		}
	}
	
	// 3. Nest children into parents, and build a list of top-level nodes
	var topLevelTasks []*core.Task
	for _, t := range tasks {
		if t.ParentID != "" {
			// Find parent in the map
			if parentNode, exists := itemMap[t.ParentID]; exists {
				// Attach this child to the parent!
				parentNode.Children = append(parentNode.Children, itemMap[t.ID])
			} else {
				// Parent is missing (maybe deleted?), render as top-level fallback
				topLevelTasks = append(topLevelTasks, t)
			}
		} else {
			// No parent, so it's a top-level node
			topLevelTasks = append(topLevelTasks, t)
		}
	}

	for _, t := range topLevelTasks {
		taskItem := itemMap[t.ID]

		// Intercept Templates so they skip chronological binning
		if strings.HasPrefix(t.ID, "recur-") {
			if templatesNode == nil {
				// Keep it collapsed by default so it doesn't clutter the UI
				expanded := false
			if val, ok := m.expandedState["folder:Templates"]; ok { expanded = val }
			templatesNode = &ListItem{Title: "Templates", Expanded: expanded} 
			}
			templatesNode.Children = append(templatesNode.Children, taskItem)
			continue
		}

		// Intercept Backlog tasks so they skip chronological binning
		if strings.HasPrefix(t.ID, "backlog-") {
			if backlogNode == nil {
				expanded := false
			if val, ok := m.expandedState["folder:Backlog"]; ok { expanded = val }
			backlogNode = &ListItem{Title: "Backlog", Expanded: expanded}
			}
			backlogNode.Children = append(backlogNode.Children, taskItem)
			continue
		}

		// Intercept Today's tasks
		now := time.Now()
		isToday := t.CreatedAt.Year() == now.Year() && t.CreatedAt.Month() == now.Month() && t.CreatedAt.Day() == now.Day()
		if isToday {
			if todayNode == nil {
				expanded := true
			if val, ok := m.expandedState["folder:Today"]; ok { expanded = val }
			todayNode = &ListItem{Title: "Today", Expanded: expanded}
			}
			todayNode.Children = append(todayNode.Children, taskItem)
			continue
		}

		yStr := t.CreatedAt.Format("2006")
		mStr := t.CreatedAt.Format("January")
		dStr := t.CreatedAt.Format("02 (Mon)")
		
		if lastYear == nil || lastYear.Title != yStr {
			expanded := false
		if val, ok := m.expandedState["folder:"+yStr]; ok { expanded = val }
		lastYear = &ListItem{Title: yStr, Expanded: expanded}
			items = append(items, lastYear)
			lastMonth = nil // Reset month and day
			lastDay = nil
		}
		
		if lastMonth == nil || lastMonth.Title != mStr {
			expanded := false
		if val, ok := m.expandedState["folder:"+mStr]; ok { expanded = val }
		lastMonth = &ListItem{Title: mStr, Expanded: expanded}
			lastYear.Children = append(lastYear.Children, lastMonth)
			lastDay = nil
		}
		
		if lastDay == nil || lastDay.Title != dStr {
			expanded := false
		if val, ok := m.expandedState["folder:"+dStr]; ok { expanded = val }
		lastDay = &ListItem{Title: dStr, Expanded: expanded}
			lastMonth.Children = append(lastMonth.Children, lastDay)
		}
		
		lastDay.Children = append(lastDay.Children, taskItem)
	}

	// Build the final ordered list
	var finalItems []*ListItem
	
	if todayNode != nil {
		finalItems = append(finalItems, todayNode)
	}
	
	// Append the historical Year/Month/Day tree
	finalItems = append(finalItems, items...)
	
	if backlogNode != nil {
		finalItems = append(finalItems, backlogNode)
	}
	
	if templatesNode != nil {
		finalItems = append(finalItems, templatesNode)
	}
	
	items = finalItems

	
	// Calculate Time Stats
	now := time.Now()
	var daySec, weekSec, monthSec, yearSec int
	
	// Determine the start of the ISO week (Monday)
	offset := int(time.Monday - now.Weekday())
	if offset > 0 {
		offset = -6
	}
	startOfWeek := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, offset)
	
	for _, t := range tasks {
		// Calculate granular time logs
		for _, log := range t.TimeLogs {
			if log.Timestamp.Year() == now.Year() {
				yearSec += log.Seconds
				if log.Timestamp.Month() == now.Month() {
					monthSec += log.Seconds
					if log.Timestamp.Day() == now.Day() {
						daySec += log.Seconds
					}
				}
				if !log.Timestamp.Before(startOfWeek) {
					weekSec += log.Seconds
				}
			}
		}
		// Also support legacy TotalSeconds if TimeLogs is empty
		if len(t.TimeLogs) == 0 && t.TotalSeconds > 0 {
			if t.CreatedAt.Year() == now.Year() {
				yearSec += t.TotalSeconds
				if t.CreatedAt.Month() == now.Month() {
					monthSec += t.TotalSeconds
					if t.CreatedAt.Day() == now.Day() {
						daySec += t.TotalSeconds
					}
				}
				if !t.CreatedAt.Before(startOfWeek) {
					weekSec += t.TotalSeconds
				}
			}
		}
	}
	
	timeStatsStr := fmt.Sprintf(" TIME LOGGED - Day: %.1fh | Week: %.1fh | Month: %.1fh | Year: %.1fh ", 
		float64(daySec)/3600.0, float64(weekSec)/3600.0, float64(monthSec)/3600.0, float64(yearSec)/3600.0)

	m.list = NewListModel("TALLY DASHBOARD", items)
	m.list.TimeStats = timeStatsStr
	
	// Restore cursor position
	if selectedID != "" {
		flat := m.list.getFlatRows()
		for i, row := range flat {
			if row.Item.ID == selectedID {
				m.list.cursor = i
				break
			}
		}
	}
	
	// Propagate dimensions if already set
	if m.terminalWidth > 0 {
		newM, _ := m.list.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
		m.list = newM.(ListModel)
	}
}

func (m *MainModel) buildCreateForm(parentID string) {
	title := "CREATE NEW TASK"
	if parentID != "" {
		title = "CREATE SUBTASK FOR: " + parentID
	}
	f := NewForm(title)
	f.AddTextBox("title", "Title", "Enter task title...", "")
	f.AddSelector("type", "Type", []string{"Task", "Story", "Technical Story", "Bug"}, "")
	f.AddTextBox("tags", "Tags", "comma separated (e.g. urgent, backend)", "")
	f.AddBoolean("backlog", "Backlog?", "")
	f.AddSelector("recur", "Recurrence", []string{"", "daily", "weekdays", "weekly", "monthly"}, "")
	
	if cfg, err := config.Load(); err == nil && len(cfg.ADO.Swimlanes) > 0 {
		opts := []string{""}
		opts = append(opts, cfg.ADO.Swimlanes...)
		f.AddSelector("swimlane", "Swimlane", opts, cfg.ADO.DefaultSwimlane)
	}
	
	f.AddButton("CREATE", ThemeGreen, ThemeBase, func(form *FormModel) tea.Cmd {
		return func() tea.Msg {
			title := form.GetString("title")
			adoType := form.GetString("type")
			tagStr := form.GetString("tags")
			
			var tags []string
			if tagStr != "" {
				for _, t := range strings.Split(tagStr, ",") {
					tags = append(tags, strings.TrimSpace(t))
				}
			}

			if title != "" {
				m.coreApp.AddTask(title, adoType, tags, form.GetString("backlog") == "True", form.GetString("recur"), parentID, form.GetString("swimlane"))
			}
			return FormSubmitMsg{}
		}
	})
	
	f.AddButton("CANCEL", ThemeRed, ThemeBase, func(form *FormModel) tea.Cmd {
		return func() tea.Msg { return FormCancelMsg{} }
	})
	
	m.form = f
	if m.terminalWidth > 0 {
		m.form.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
	}
	m.form.Init()
}

func (m *MainModel) buildLogTimeForm(id string) {
	// Try to load activities from config for the dropdown
	cfg, _ := config.Load()
	activityNames := []string{"Default"}
	if cfg != nil && cfg.SevenPace.Activities != nil {
		for name := range cfg.SevenPace.Activities {
			activityNames = append(activityNames, name)
		}
	}

	f := NewForm("LOG TIME: " + id)
	f.AddTextBox("time", "Time", "e.g., 5m, 1.5h", "")
	if len(activityNames) > 1 {
		f.AddSelector("activity", "Activity", activityNames, "")
	}
	
	f.AddButton("LOG TIME", ThemeGreen, ThemeBase, func(form *FormModel) tea.Cmd {
		return func() tea.Msg {
			timeStr := form.GetString("time")
			activityName := form.GetString("activity")
			
			var activityID string
			if cfg != nil {
				if activityName == "Default" || activityName == "" {
					activityID = cfg.SevenPace.ActivityID
				} else {
					activityID = cfg.SevenPace.Activities[activityName]
				}
			}

			if timeStr != "" {
				m.coreApp.LogTime(id, timeStr, activityID)
			}
			return FormSubmitMsg{}
		}
	})
	
	f.AddButton("CANCEL", ThemeRed, ThemeBase, func(form *FormModel) tea.Cmd {
		return func() tea.Msg { return FormCancelMsg{} }
	})
	
	m.form = f
	if m.terminalWidth > 0 {
		m.form.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
	}
	m.form.Init()
}

type FormSubmitMsg struct{}
type FormCancelMsg struct{}
type LogTimeMsg struct{ ID string }
type DeleteTaskMsg struct{ ID string }
type CreateSubtaskMsg struct{ ParentID string }
type StartTaskMsg struct{ ID string }

func (m *MainModel) Init() tea.Cmd {
	return m.list.Init()
}

func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		newM, _ := m.list.Update(msg)
		m.list = newM.(ListModel)
		if m.form != nil {
			newF, _ := m.form.Update(msg)
			m.form = newF.(*FormModel)
		}
		return m, nil
		
	case RefreshMsg:
		m.reloadList()
		return m, nil

	case CycleThemeMsg:
		newTheme := CycleTheme()
		
		// Save the new theme to config
		cfg, err := config.Load()
		if err == nil {
			cfg.UI.Theme = newTheme
			config.Save(cfg)
		}
		
		m.reloadList() // redraw with new global styles
		return m, nil

	case SyncTasksMsg:
		// Execute the sync command in a shell so we can pause and let the user read the output
		executable := os.Args[0]
		// Use $0 to safely pass the executable path as an argument to bash, preventing command injection
		script := `"$0" sync && echo '' && read -p 'Press Enter to return to Dashboard...'`
		c := exec.Command("bash", "-c", script, executable)
		
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			return SyncFinishedMsg{Err: err}
		})
		
	case SyncFinishedMsg:
		m.coreApp.ReconcileRecurringTasks()
		m.reloadList() // Pull fresh data in case sync updated anything locally
		return m, nil

	case EditTaskMsg:
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vim"
		}
		path := m.coreApp.Store.GetTaskPath(msg.ID)
		c := exec.Command(editor, path)
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			return EditorFinishedMsg{Err: err}
		})
		
	case EditorFinishedMsg:
		m.reloadList() // reload data after editor closes
		return m, nil

	case CreateNewTaskMsg:
		m.state = StateCreateTask
		m.buildCreateForm("")
		return m, nil
		
	case CreateSubtaskMsg:
		m.state = StateCreateTask
		m.buildCreateForm(msg.ParentID)
		return m, nil
		
	case LogTimeMsg:
		m.state = StateLogTime
		m.selectedID = msg.ID
		m.buildLogTimeForm(msg.ID)
		return m, nil
		
	case DeleteTaskMsg:
		m.coreApp.DeleteTask(msg.ID)
		m.reloadList()
		return m, nil
		
	case StartTaskMsg:
		m.coreApp.StartTask(msg.ID)
		m.reloadList()
		return m, nil
		
	case FormSubmitMsg, FormCancelMsg:
		m.state = StateDashboard
		m.form = nil
		m.coreApp.ReconcileRecurringTasks() // Force a spawn check in case they just created a template
		m.reloadList() // Pull fresh data
		return m, nil
	}

	var cmd tea.Cmd
	if m.state == StateDashboard {
		var newModel tea.Model
		newModel, cmd = m.list.Update(msg)
		m.list = newModel.(ListModel)
	} else if m.state == StateCreateTask || m.state == StateLogTime {
		var newModel tea.Model
		newModel, cmd = m.form.Update(msg)
		m.form = newModel.(*FormModel)
	}
	
	return m, cmd
}

func (m *MainModel) View() string {
	if (m.state == StateCreateTask || m.state == StateLogTime) && m.form != nil {
		return m.form.View()
	}
	return m.list.View()
}

func RunApp(app *core.App) error {
	p := tea.NewProgram(NewMainModel(app), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
