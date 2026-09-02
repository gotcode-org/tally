package ui

import (
	"fmt"
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
)

type MainModel struct {
	coreApp *core.App
	state   AppState
	list    ListModel
	form    *FormModel
	
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
	
	for _, t := range tasks {
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

		taskItem := &ListItem{
			ID:        t.ID,
			Title:     t.ID + " - " + t.Title,
			Type:      t.ADOType,
			Status:    string(t.Status),
			TimeText:  timeStr,
			TypeColor: c,
		}
		
		lastDay.Children = append(lastDay.Children, taskItem)
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
	f.AddTextBox("title", "Title", "Enter task title...", "The name of the task")
	f.AddSelector("type", "Type", []string{"Task", "Story", "Technical Story", "Bug"}, "Use left/right to select type")
	
	f.AddButton("CREATE", ThemeGreen, ThemeBase, func(form *FormModel) tea.Cmd {
		return func() tea.Msg {
			title := form.GetString("title")
			adoType := form.GetString("type")
			if title != "" {
				m.coreApp.AddTask(title, adoType, []string{})
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
		CycleTheme()
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
	} else if m.state == StateCreateTask {
		var newModel tea.Model
		newModel, cmd = m.form.Update(msg)
		m.form = newModel.(*FormModel)
	}
	
	return m, cmd
}

func (m *MainModel) View() string {
	if m.state == StateCreateTask && m.form != nil {
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
