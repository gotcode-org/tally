package ui

import (
	"fmt"
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
	coreApp *core.App
	state   AppState
	list    ListModel
	form    *FormModel
	
	selectedID     string
	terminalWidth  int
	terminalHeight int
}

func NewMainModel(app *core.App) *MainModel {
	m := &MainModel{
		coreApp: app,
		state:   StateDashboard,
	}
	m.reloadList()
	return m
}

func (m *MainModel) reloadList() {
	tasks, _ := m.coreApp.Store.ListTasks("")

	// Reverse tasks so newest are at the top
	for i, j := 0, len(tasks)-1; i < j; i, j = i+1, j-1 {
		tasks[i], tasks[j] = tasks[j], tasks[i]
	}

	var items []*ListItem
	var lastYear, lastMonth, lastDay *ListItem
	var backlogNode *ListItem
	var templatesNode *ListItem
	
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

		titleStr := t.ID + " - " + t.Title
		if t.Recurrence != "" {
			titleStr = "↻ " + titleStr
		}

		taskItem := &ListItem{
			ID:        t.ID,
			Title:     titleStr,
			Type:      t.ADOType,
			Status:    string(t.Status),
			TimeText:  timeStr,
			TypeColor: c,
		}

		// Intercept Templates so they skip chronological binning
		if strings.HasPrefix(t.ID, "recur-") {
			if templatesNode == nil {
				// Keep it collapsed by default so it doesn't clutter the UI
				templatesNode = &ListItem{Title: "Templates", Expanded: false} 
			}
			templatesNode.Children = append(templatesNode.Children, taskItem)
			continue
		}

		// Intercept Backlog tasks so they skip chronological binning
		if strings.HasPrefix(t.ID, "backlog-") {
			if backlogNode == nil {
				backlogNode = &ListItem{Title: "Backlog", Expanded: true}
			}
			backlogNode.Children = append(backlogNode.Children, taskItem)
			continue
		}

		yStr := t.CreatedAt.Format("2006")
		mStr := t.CreatedAt.Format("January")
		dStr := t.CreatedAt.Format("02 (Mon)")
		
		if lastYear == nil || lastYear.Title != yStr {
			lastYear = &ListItem{Title: yStr, Expanded: true}
			items = append(items, lastYear)
			lastMonth = nil // Reset month and day
			lastDay = nil
		}
		
		if lastMonth == nil || lastMonth.Title != mStr {
			lastMonth = &ListItem{Title: mStr, Expanded: true}
			lastYear.Children = append(lastYear.Children, lastMonth)
			lastDay = nil
		}
		
		if lastDay == nil || lastDay.Title != dStr {
			lastDay = &ListItem{Title: dStr, Expanded: true}
			lastMonth.Children = append(lastMonth.Children, lastDay)
		}
		
		lastDay.Children = append(lastDay.Children, taskItem)
	}

	// Prepend our special folders to the top of the dashboard
	if backlogNode != nil {
		items = append([]*ListItem{backlogNode}, items...)
	}
	if templatesNode != nil {
		items = append([]*ListItem{templatesNode}, items...)
	}

	m.list = NewListModel("TALLY DASHBOARD", items)
	// Propagate dimensions if already set
	if m.terminalWidth > 0 {
		newM, _ := m.list.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
		m.list = newM.(ListModel)
	}
}

func (m *MainModel) buildCreateForm() {
	f := NewForm("CREATE NEW TASK")
	f.AddTextBox("title", "Title", "Enter task title...", "")
	f.AddSelector("type", "Type", []string{"Task", "Story", "Technical Story", "Bug"}, "")
	f.AddBoolean("backlog", "Backlog?", "")
	f.AddSelector("recur", "Recurrence", []string{"", "daily", "weekly", "monthly"}, "")
	
	f.AddButton("CREATE", ThemeGreen, ThemeBase, func(form *FormModel) tea.Cmd {
		return func() tea.Msg {
			title := form.GetString("title")
			adoType := form.GetString("type")
			if title != "" {
				m.coreApp.AddTask(title, adoType, []string{}, form.GetString("backlog") == "True", form.GetString("recur"))
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
	f := NewForm("LOG TIME: " + id)
	f.AddTextBox("time", "Time", "e.g., 5m, 1.5h", "")
	
	f.AddButton("LOG TIME", ThemeGreen, ThemeBase, func(form *FormModel) tea.Cmd {
		return func() tea.Msg {
			timeStr := form.GetString("time")
			if timeStr != "" {
				m.coreApp.LogTime(id, timeStr)
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
		script := fmt.Sprintf("%s sync && echo '' && read -p 'Press Enter to return to Dashboard...'", executable)
		c := exec.Command("bash", "-c", script)
		
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
		m.buildCreateForm()
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
