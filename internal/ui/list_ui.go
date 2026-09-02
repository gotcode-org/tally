package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	HeaderStyle = lipgloss.NewStyle().Foreground(CatMochaBase).Background(CatMochaMauve).Bold(true)
	MilestoneRowStyle = lipgloss.NewStyle().Foreground(CatMochaText)
	StoryRowStyle = lipgloss.NewStyle().Foreground(CatMochaText)
	TaskRowStyle = lipgloss.NewStyle().Foreground(CatMochaOverlay)
	ActiveRowStyle = lipgloss.NewStyle().Foreground(CatMochaText).Background(CatMochaOverlay).Bold(true)
)

type CreateNewTaskMsg struct{}

type ListItem struct {
	ID           string
	Title        string
	Description  string
	Type         string
	TypeColor    lipgloss.Color
	Status       string
	Progress     float64
	ProgressText string
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
	headers := []string{"NAME", "TYPE", "STATUS", "PROGRESS"}

	for i, col := range headers {
		w := widths[i]
		var styled string
		if len(col) > w {
			styled = HeaderStyle.Render(col[:w-3] + "...")
		} else {
			styled = HeaderStyle.Render(col) + lipgloss.NewStyle().Background(CatMochaMauve).Render(strings.Repeat(" ", w-len(col)))
		}
		cells = append(cells, styled)
	}

	prefix := lipgloss.NewStyle().Background(CatMochaMauve).Render("  ")
	separator := lipgloss.NewStyle().Foreground(CatMochaBase).Background(CatMochaMauve).Render(" │ ")
	middle := prefix + strings.Join(cells, separator)

	var parts []string
	for _, w := range widths {
		parts = append(parts, strings.Repeat("─", w))
	}
	borderStyle := lipgloss.NewStyle().Foreground(CatMochaBase).Background(CatMochaMauve)
	
	topBorder := borderStyle.Render("──" + strings.Join(parts, "─┬─"))
	botBorder := borderStyle.Render("──" + strings.Join(parts, "─┼─"))

	return topBorder + "\n" + middle + "\n" + botBorder
}

func formatStatusCell(status string, width int, isCursor bool) string {
	var bg lipgloss.Color
	var fg lipgloss.Color

	switch status {
	case "ACTIVE", "In Progress":
		bg = CatMochaBlue
		fg = CatMochaBase
	case "COMPLETED", "Done":
		bg = CatMochaGreen
		fg = CatMochaBase
	case "CANCELLED":
		bg = CatMochaRed
		fg = CatMochaBase
	case "BACKLOG", "Todo":
		bg = CatMochaOverlay
		fg = CatMochaBase
	default:
		bg = CatMochaOverlay
		fg = CatMochaBase
	}

	badge := lipgloss.NewStyle().Foreground(fg).Background(bg).Bold(true).Padding(0, 1).Render(status)
	visibleLen := lipgloss.Width(badge)

	if visibleLen > width {
		return badge[:width]
	}

	var style lipgloss.Style
	if isCursor {
		style = ActiveRowStyle.Copy().Width(width).Align(lipgloss.Center)
	} else {
		style = lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
	}

	return style.Render(badge)
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

	filled := lipgloss.NewStyle().Foreground(CatMochaGreen).Render(strings.Repeat("█", filledLen))
	empty := lipgloss.NewStyle().Foreground(CatMochaOverlay).Render(strings.Repeat("░", emptyLen))

	bar := filled + empty
	
	if text != "" {
		textStyle := lipgloss.NewStyle().Foreground(CatMochaOverlay).Italic(true)
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
	formatCol := func(txt string, w int, style lipgloss.Style) string {
		var styled string
		var spaces string
		
		visibleLen := lipgloss.Width(txt)
		
		if visibleLen > w {
			styled = style.Render(txt[:w-3] + "...")
		} else {
			spaces = strings.Repeat(" ", w-visibleLen)
			styled = style.Render(txt) + style.Render(spaces)
		}
		return styled
	}

	activeStyle := ActiveRowStyle
	if isCursor {
		rowStyle = activeStyle
	}

	c1 := formatCol(title, widths[0], rowStyle)
	c2 := formatCol(typeStr, widths[1], rowStyle)
	c3 := statusCell
	c4 := formatCol(progress, widths[3], rowStyle)

	prefix := "  "
	if isCursor {
		prefix = lipgloss.NewStyle().Foreground(CatMochaGreen).Bold(true).Render("▶ ")
	}
	prefix = rowStyle.Render(prefix)
	separator := rowStyle.Render(" │ ")

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

		badgeStyle := lipgloss.NewStyle().Foreground(CatMochaBase).Background(row.Item.TypeColor).Bold(true).Padding(0, 1)
		typeCell := badgeStyle.Width(widths[1]).Align(lipgloss.Center).Render(strings.ToUpper(row.Item.Type))
		
		statusCell := formatStatusCell(row.Item.Status, widths[2], isCursor)

		var progressCell string
		if row.Item.ProgressText != "" || row.Item.Progress > 0 {
			progressCell = renderProgressBar(row.Item.Progress, row.Item.ProgressText, widths[3])
		} else {
			progressCell = ""
		}

		listSections = append(listSections, renderRow(titleCell, typeCell, statusCell, progressCell, widths, isCursor, rowStyle))
	}

	listSections = append(listSections, lipgloss.NewStyle().Background(CatMochaBase).Render(""))

	paddedList := lipgloss.NewStyle().
		Background(CatMochaBase).
		Width(headerWidth).
		Height(targetHeight - 8).
		Render(strings.Join(listSections, "\n"))

	cursorIndicator := ""
	if len(flatRows) > 0 {
		cursorIndicator = fmt.Sprintf(" [%d/%d]", m.cursor+1, len(flatRows))
	}
	headerFull := lipgloss.NewStyle().Width(headerWidth).Align(lipgloss.Center).Background(CatMochaMauve).Foreground(CatMochaBase).Bold(true).Render(" " + strings.ToUpper(m.title) + cursorIndicator + " ")
	headerRow := lipgloss.NewStyle().Background(CatMochaMauve).Render(m.renderHeader(widths))
	
	helpStr := " tab/shift+tab: move • enter: toggle/edit • esc: quit "
	footer := lipgloss.NewStyle().Width(headerWidth).Align(lipgloss.Center).Background(CatMochaMauve).Foreground(CatMochaBase).Bold(true).Render(helpStr)
	
	finalContent := headerFull + "\n" + headerRow + "\n" + paddedList + "\n" + footer

	activeWindow := windowStyle.Copy().Width(targetWidth)
	result := activeWindow.Render(finalContent)

	return lipgloss.Place(m.terminalWidth, m.terminalHeight, lipgloss.Center, lipgloss.Top, result, lipgloss.WithWhitespaceBackground(CatMochaBase))
}

func RunList(title string, items []*ListItem) error {
	m := NewListModel(title, items)
	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
