package ui

import "github.com/charmbracelet/lipgloss"

// Catppuccin Mocha Palette
var (
    CatMochaBase    = lipgloss.Color("#1e1e2e")
    CatMochaText    = lipgloss.Color("#cdd6f4")
    CatMochaSubtext = lipgloss.Color("#a6adc8")
    CatMochaOverlay = lipgloss.Color("#6c7086")
    CatMochaBlue    = lipgloss.Color("#89b4fa")
    CatMochaGreen   = lipgloss.Color("#a6e3a1")
    CatMochaRed     = lipgloss.Color("#f38ba8")
    CatMochaMauve   = lipgloss.Color("#cba6f7")
    CatMochaPeach   = lipgloss.Color("#fab387")
)

// Common Styles
var (
    WindowStyle = lipgloss.NewStyle().
        Padding(1, 2).
        Border(lipgloss.RoundedBorder()).
        BorderForeground(CatMochaMauve).
        Background(CatMochaBase)

    TitleStyle = lipgloss.NewStyle().
        Foreground(CatMochaBase).
        Background(CatMochaMauve).
        Padding(0, 2).
        Bold(true).
        MarginBottom(1)

    LabelStyle = lipgloss.NewStyle().
        Foreground(CatMochaText).
        Bold(true)

    FocusedLabelStyle = lipgloss.NewStyle().
        Foreground(CatMochaGreen).
        Bold(true)

    HelperStyle = lipgloss.NewStyle().
        Foreground(CatMochaOverlay).
        Italic(true)

    ErrorStyle = lipgloss.NewStyle().
        Foreground(CatMochaRed).
        Bold(true)

    ButtonStyle = lipgloss.NewStyle().
        Foreground(CatMochaText).
        Background(CatMochaOverlay).
        Padding(0, 3).
        MarginTop(1)

    ActiveButtonStyle = lipgloss.NewStyle().
        Foreground(CatMochaBase).
        Background(CatMochaGreen).
        Padding(0, 3).
        Bold(true).
        MarginTop(1)

    SuccessTitleStyle = lipgloss.NewStyle().
        Foreground(CatMochaBase).
        Background(CatMochaGreen).
        Padding(0, 2).
        Bold(true).
        MarginBottom(1)
)

// Adjust dimensions based on terminal size
func AdjustDimensions(width, height int) (contentWidth, targetWidth, targetHeight int) {
    targetWidth = int(float64(width) * 0.95)
    if targetWidth < 40 {
        targetWidth = 40
    }
    targetHeight = int(float64(height) * 0.95)
    if targetHeight < 12 {
        targetHeight = 12
    }
    contentWidth = targetWidth - 8
    if contentWidth < 20 {
        contentWidth = 20
    }
    return
}
