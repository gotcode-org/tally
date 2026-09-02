package ui

import (
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
	var items []*ListItem
	
	for _, t := range tasks {
		c := CatMochaBlue
		if strings.Contains(strings.ToLower(t.ADOType), "story") {
			c = CatMochaMauve
		} else if strings.Contains(strings.ToLower(t.ADOType), "bug") {
			c = CatMochaRed
		}
		
		items = append(items, &ListItem{
			Title:     t.ID + " - " + t.Title,
			Type:      t.ADOType,
			Status:    string(t.Status),
			TypeColor: c,
		})
	}
	
	// Reverse the list so newest are at the top
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	
	m.list = NewListModel("TALLY DASHBOARD", items)
	// Propagate dimensions if already set
	if m.terminalWidth > 0 {
		m.list.Update(tea.WindowSizeMsg{Width: m.terminalWidth, Height: m.terminalHeight})
	}
}

func (m *MainModel) buildCreateForm() {
	f := NewForm("CREATE NEW TASK")
	f.AddTextBox("title", "Title", "Enter task title...", "The name of the task")
	f.AddSelector("type", "Type", []string{"Task", "Story", "Technical Story", "Bug"}, "Use left/right to select type")
	
	f.AddButton("CREATE", CatMochaGreen, CatMochaBase, func(form *FormModel) tea.Cmd {
		return func() tea.Msg {
			title := form.GetString("title")
			adoType := form.GetString("type")
			if title != "" {
				m.coreApp.AddTask(title, adoType, []string{})
			}
			return FormSubmitMsg{}
		}
	})
	
	f.AddButton("CANCEL", CatMochaRed, CatMochaBase, func(form *FormModel) tea.Cmd {
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
		m.list.Update(msg)
		if m.form != nil {
			m.form.Update(msg)
		}
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
