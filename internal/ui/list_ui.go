package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	HeaderStyle = lipgloss.NewStyle().Foreground(ThemeBase).Background(ThemeMauve).Bold(true)
	MilestoneRowStyle = lipgloss.NewStyle().Foreground(ThemeText)
	StoryRowStyle = lipgloss.NewStyle().Foreground(ThemeText)
	TaskRowStyle = lipgloss.NewStyle().Foreground(ThemeOverlay)
	ActiveRowStyle = lipgloss.NewStyle().Foreground(ThemeText).Background(ThemeOverlay).Bold(true)
)

type CreateNewTaskMsg struct{}
type EditTaskMsg struct{ ID string }
type EditorFinishedMsg struct{ Err error }
type SyncTasksMsg struct{}
type CycleThemeMsg struct{}
type SyncFinishedMsg struct{ Err error }

type ListItem struct {
	ID           string
	Title        string
	Type         string
	Status       string
	Progress     float64
	ProgressText string
	TimeText     string
	TypeColor    lipgloss.Color
	Expanded     bool
	Children     []*ListItem
}

type FlatRow struct {
	Depth       int
	Expanded    bool
	HasChildren bool
	Item        *ListItem
}

type ListModel struct {
	title          string
	items          []*ListItem
	cursor         int
	terminalWidth  int
	terminalHeight int
	widthPct       float64
	heightPct      float64
}

func NewListModel(title string, items []*ListItem) ListModel {
	return ListModel{
		title:     title,
		items:     items,
		cursor:    0,
		widthPct:  0.95,
		heightPct: 0.95,
	}
}

func (m ListModel) Init() tea.Cmd {
	return nil
}

func (m *ListModel) getFlatRows() []FlatRow {
	var rows []FlatRow
	var flatten func(items []*ListItem, depth int)
	flatten = func(items []*ListItem, depth int) {
		for _, item := range items {
			hasChildren := len(item.Children) > 0
			rows = append(rows, FlatRow{
				Depth:       depth,
				Expanded:    item.Expanded,
				HasChildren: hasChildren,
				Item:        item,
			})
			if item.Expanded && hasChildren {
				flatten(item.Children, depth+1)
			}
		}
	}
	flatten(m.items, 0)
	return rows
}

func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.terminalWidth = msg.Width
		m.terminalHeight = msg.Height
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			return m, tea.Quit
		case "n":
			return m, func() tea.Msg { return CreateNewTaskMsg{} }
		case "t":
			return m, func() tea.Msg { return CycleThemeMsg{} }
		case "s":
			return m, func() tea.Msg { return SyncTasksMsg{} }
		case "e":
			flatRows := m.getFlatRows()
			if len(flatRows) > 0 {
				row := flatRows[m.cursor]
				if row.Item.ID != "" {
					return m, func() tea.Msg { return EditTaskMsg{ID: row.Item.ID} }
				}
			}
		case "a":
			flatRows := m.getFlatRows()
			if len(flatRows) > 0 {
				row := flatRows[m.cursor]
				if row.Item.ID != "" {
					return m, func() tea.Msg { return LogTimeMsg{ID: row.Item.ID} }
				}
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.getFlatRows())-1 {
				m.cursor++
			}
		case "enter":
			flatRows := m.getFlatRows()
			if len(flatRows) > 0 {
				row := flatRows[m.cursor]
				row.Item.Expanded = !row.Item.Expanded
			}
		}
	}
	return m, nil
}

func (m ListModel) renderHeader(widths []int) string {
	cells := []string{}
	headers := []string{"NAME", "TYPE", "STATUS", "TIME"}

	for i, col := range headers {
		w := widths[i]
		var styled string
		if len(col) > w {
			styled = HeaderStyle.Render(col[:w-3] + "...")
		} else {
			styled = HeaderStyle.Render(col) + lipgloss.NewStyle().Background(ThemeMauve).Render(strings.Repeat(" ", w-len(col)))
		}
		cells = append(cells, styled)
	}

	prefix := lipgloss.NewStyle().Background(ThemeMauve).Render("  ")
	separator := lipgloss.NewStyle().Foreground(ThemeBase).Background(ThemeMauve).Render(" │ ") // Keep dark on Mauve
	middle := prefix + strings.Join(cells, separator)

	var parts []string
	for _, w := range widths {
		parts = append(parts, strings.Repeat("─", w))
	}
	borderStyle := lipgloss.NewStyle().Foreground(ThemeBase).Background(ThemeMauve)
	
	topBorder := borderStyle.Render("──" + strings.Join(parts, "─┬─"))
	botBorder := borderStyle.Render("──" + strings.Join(parts, "─┼─"))

	return topBorder + "\n" + middle + "\n" + botBorder
}

func formatStatusCell(status string, width int, isCursor bool, bgStyle lipgloss.Style) string {
	if status == "" {
		return ""
	}
	var fg lipgloss.Color

	switch status {
	case "ACTIVE", "In Progress":
		fg = ThemeBlue
	case "COMPLETED", "Done":
		fg = ThemeGreen
	case "CANCELLED":
		fg = ThemeRed
	case "BACKLOG", "Todo":
		fg = ThemeSubtext
	default:
		fg = ThemeSubtext
	}

	if isCursor {
		fg = ThemeBase
	}

	badge := lipgloss.NewStyle().Foreground(fg).Background(bgStyle.GetBackground()).Bold(true).Render(strings.ToUpper(status))
	visibleLen := lipgloss.Width(badge)

	if visibleLen > width {
		return badge[:width]
	}

	return badge
}

func renderProgressBar(pct float64, text string, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}

	filledLen := int(float64(width) * pct)
	emptyLen := width - filledLen

	filled := lipgloss.NewStyle().Foreground(ThemeGreen).Render(strings.Repeat("█", filledLen))
	empty := lipgloss.NewStyle().Foreground(ThemeOverlay).Render(strings.Repeat("░", emptyLen))

	bar := filled + empty
	
	if text != "" {
		textStyle := lipgloss.NewStyle().Foreground(ThemeOverlay).Italic(true)
		return fmt.Sprintf("%s %s", bar, textStyle.Render(text))
	}
	return bar
}

func getRowTitle(row FlatRow) string {
	icon := ""
	if row.HasChildren {
		if row.Expanded {
			icon = "▼ "
		} else {
			icon = "▶ "
		}
	} else {
		icon = "  "
	}
	indent := strings.Repeat("  ", row.Depth)
	return indent + icon + row.Item.Title
}

func renderRow(title, typeStr, statusCell, progress string, widths []int, isCursor bool, rowStyle lipgloss.Style) string {
	formatColLeft := func(txt string, w int) string {
		visibleLen := lipgloss.Width(txt)
		if visibleLen > w {
			return rowStyle.Render(txt[:w-3] + "...")
		}
		spaces := strings.Repeat(" ", w-visibleLen)
		return txt + rowStyle.Render(spaces)
	}
	
	formatColCenter := func(txt string, w int) string {
		visibleLen := lipgloss.Width(txt)
		if visibleLen > w {
			return rowStyle.Render(txt[:w-3] + "...")
		}
		if txt == "" {
			return rowStyle.Render(strings.Repeat(" ", w))
		}
		leftPad := (w - visibleLen) / 2
		rightPad := w - visibleLen - leftPad
		return rowStyle.Render(strings.Repeat(" ", leftPad)) + txt + rowStyle.Render(strings.Repeat(" ", rightPad))
	}

	if isCursor {
		rowStyle = ActiveRowStyle
	}

	c1 := formatColLeft(rowStyle.Render(title), widths[0])
	c2 := formatColCenter(typeStr, widths[1])
	c3 := formatColCenter(statusCell, widths[2])
	c4 := formatColLeft(rowStyle.Render(progress), widths[3])

	prefix := "  "
	if isCursor {
		prefix = lipgloss.NewStyle().Foreground(ThemeBase).Background(ThemeBlue).Bold(true).Render("▶ ")
	}
	prefix = rowStyle.Render(prefix)
	separator := lipgloss.NewStyle().Foreground(ThemeSubtext).Background(rowStyle.GetBackground()).Render(" │ ")

	return prefix + c1 + separator + c2 + separator + c3 + separator + c4
}

func (m ListModel) View() string {
	if m.terminalWidth == 0 || m.terminalHeight == 0 {
		return ""
	}

	targetWidth := int(float64(m.terminalWidth) * m.widthPct)
	targetHeight := int(float64(m.terminalHeight) * m.heightPct)

	windowStyle := lipgloss.NewStyle().
		Border(lipgloss.HiddenBorder()).
		Height(targetHeight)

	headerWidth := targetWidth
	
	// Subtract the width of prefix (2) and separators (3 * 3 = 9) = 11 total static chars
	availableWidth := headerWidth - 11
	if availableWidth < 20 {
		availableWidth = 20
	}
	
	w0 := int(float64(availableWidth) * 0.45)
	w1 := int(float64(availableWidth) * 0.15)
	w2 := int(float64(availableWidth) * 0.20)
	w3 := availableWidth - (w0 + w1 + w2) // Soak up any float truncation remainder

	widths := []int{w0, w1, w2, w3}

	flatRows := m.getFlatRows()
	var listSections []string

	for i, row := range flatRows {
		isCursor := i == m.cursor
		
		titleCell := getRowTitle(row)
		
		rowStyle := MilestoneRowStyle
		if row.Depth > 0 {
			rowStyle = StoryRowStyle
		}
		
		bgStyle := rowStyle
		if isCursor {
			bgStyle = ActiveRowStyle
		}

		var typeCell string
		if row.Item.Type != "" {
			tColor := row.Item.TypeColor
			if isCursor {
				tColor = ThemeBase
			}
			typeCell = lipgloss.NewStyle().Foreground(tColor).Background(bgStyle.GetBackground()).Bold(true).Render(strings.ToUpper(row.Item.Type))
		}
		
		statusCell := formatStatusCell(row.Item.Status, widths[2], isCursor, bgStyle)

		var progressCell string
		if row.Item.TimeText != "" {
			pColor := ThemeSubtext // Muted but readable on Overlay background
			if isCursor {
				pColor = ThemeBase // Dark text on blue active row
			}
			progressCell = lipgloss.NewStyle().Foreground(pColor).Render(row.Item.TimeText)
		} else {
			progressCell = "-"
		}

		listSections = append(listSections, renderRow(titleCell, typeCell, statusCell, progressCell, widths, isCursor, rowStyle))
	}

	visibleRows := targetHeight - 8
	for len(listSections) < visibleRows {
		listSections = append(listSections, renderRow("", "", "", "", widths, false, lipgloss.NewStyle().Foreground(ThemeBase).Background(ThemeOverlay)))
	}

	paddedList := lipgloss.NewStyle().
		Background(ThemeOverlay).
		Width(headerWidth).
		Height(visibleRows).
		Render(strings.Join(listSections, "\n"))

	cursorIndicator := ""
	if len(flatRows) > 0 {
		cursorIndicator = fmt.Sprintf(" [%d/%d]", m.cursor+1, len(flatRows))
	}
	headerFull := lipgloss.NewStyle().Width(headerWidth).Align(lipgloss.Center).Background(ThemeMauve).Foreground(ThemeBase).Bold(true).Render(" " + strings.ToUpper(m.title) + cursorIndicator + " ")
	headerRow := lipgloss.NewStyle().Background(ThemeMauve).Render(m.renderHeader(widths))
	
	helpStr := " ↑/↓: move • enter: expand • e: edit • n: new • a: add time • s: sync • t: theme • esc: quit "
	footer := lipgloss.NewStyle().Width(headerWidth).Align(lipgloss.Center).Background(ThemeMauve).Foreground(ThemeBase).Bold(true).Render(helpStr)
	
	finalContent := headerFull + "\n" + headerRow + "\n" + paddedList + "\n" + footer

	activeWindow := windowStyle.Copy().Width(targetWidth)
	result := activeWindow.Render(finalContent)

	return lipgloss.Place(m.terminalWidth, m.terminalHeight, lipgloss.Center, lipgloss.Top, result)
}

func RunList(title string, items []*ListItem) error {
	m := NewListModel(title, items)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
