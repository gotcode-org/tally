package ui

import "github.com/charmbracelet/lipgloss"

type ThemePalette struct {
	Name    string
	Base    lipgloss.Color
	Text    lipgloss.Color
	Subtext lipgloss.Color
	Overlay lipgloss.Color
	Blue    lipgloss.Color
	Green   lipgloss.Color
	Red     lipgloss.Color
	Mauve   lipgloss.Color
	Peach   lipgloss.Color
}

var (
	CatppuccinMocha = ThemePalette{
		Name: "Catppuccin Mocha",
		Base:    lipgloss.Color("#1e1e2e"),
		Text:    lipgloss.Color("#cdd6f4"),
		Subtext: lipgloss.Color("#a6adc8"),
		Overlay: lipgloss.Color("#6c7086"),
		Blue:    lipgloss.Color("#89b4fa"),
		Green:   lipgloss.Color("#a6e3a1"),
		Red:     lipgloss.Color("#f38ba8"),
		Mauve:   lipgloss.Color("#cba6f7"),
		Peach:   lipgloss.Color("#fab387"),
	}

	Dracula = ThemePalette{
		Name: "Dracula",
		Base:    lipgloss.Color("#282a36"),
		Text:    lipgloss.Color("#f8f8f2"),
		Subtext: lipgloss.Color("#6272a4"),
		Overlay: lipgloss.Color("#44475a"),
		Blue:    lipgloss.Color("#8be9fd"),
		Green:   lipgloss.Color("#50fa7b"),
		Red:     lipgloss.Color("#ff5555"),
		Mauve:   lipgloss.Color("#ff79c6"), // Dracula Pink
		Peach:   lipgloss.Color("#ffb86c"),
	}

	Nord = ThemePalette{
		Name: "Nord",
		Base:    lipgloss.Color("#2e3440"),
		Text:    lipgloss.Color("#eceff4"),
		Subtext: lipgloss.Color("#d8dee9"),
		Overlay: lipgloss.Color("#4c566a"),
		Blue:    lipgloss.Color("#81a1c1"),
		Green:   lipgloss.Color("#a3be8c"),
		Red:     lipgloss.Color("#bf616a"),
		Mauve:   lipgloss.Color("#88c0d0"), // Nord Frost Blue
		Peach:   lipgloss.Color("#ebcb8b"),
	}
)

var availableThemes = []ThemePalette{CatppuccinMocha, Dracula, Nord}
var currentThemeIdx = 0

var (
	ThemeBase    lipgloss.Color
	ThemeText    lipgloss.Color
	ThemeSubtext lipgloss.Color
	ThemeOverlay lipgloss.Color
	ThemeBlue    lipgloss.Color
	ThemeGreen   lipgloss.Color
	ThemeRed     lipgloss.Color
	ThemeMauve   lipgloss.Color
	ThemePeach   lipgloss.Color
)

var (
	WindowStyle       lipgloss.Style
	TitleStyle        lipgloss.Style
	LabelStyle        lipgloss.Style
	FocusedLabelStyle lipgloss.Style
	HelperStyle       lipgloss.Style
	ErrorStyle        lipgloss.Style
	ButtonStyle       lipgloss.Style
	ActiveButtonStyle lipgloss.Style
	SuccessTitleStyle lipgloss.Style
)

func init() {
	ApplyTheme(CatppuccinMocha)
}

func CycleTheme() string {
	currentThemeIdx = (currentThemeIdx + 1) % len(availableThemes)
	t := availableThemes[currentThemeIdx]
	ApplyTheme(t)
	return t.Name
}

func ApplyThemeByName(name string) {
	for i, t := range availableThemes {
		if t.Name == name {
			currentThemeIdx = i
			ApplyTheme(t)
			return
		}
	}
	ApplyTheme(CatppuccinMocha)
}

func ApplyTheme(t ThemePalette) {
	ThemeBase = t.Base
	ThemeText = t.Text
	ThemeSubtext = t.Subtext
	ThemeOverlay = t.Overlay
	ThemeBlue = t.Blue
	ThemeGreen = t.Green
	ThemeRed = t.Red
	ThemeMauve = t.Mauve
	ThemePeach = t.Peach

	WindowStyle = lipgloss.NewStyle().
		Padding(1, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ThemeMauve).
		Background(ThemeBase)

	TitleStyle = lipgloss.NewStyle().
		Foreground(ThemeBase).
		Background(ThemeMauve).
		Padding(0, 2).
		Bold(true).
		MarginBottom(1)

	LabelStyle = lipgloss.NewStyle().
		Foreground(ThemeText).
		Bold(true)

	FocusedLabelStyle = lipgloss.NewStyle().
		Foreground(ThemeGreen).
		Bold(true)

	HelperStyle = lipgloss.NewStyle().
		Foreground(ThemeOverlay).
		Italic(true)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(ThemeRed).
		Bold(true)

	ButtonStyle = lipgloss.NewStyle().
		Foreground(ThemeText).
		Background(ThemeOverlay).
		Padding(0, 3).
		MarginTop(1)

	ActiveButtonStyle = lipgloss.NewStyle().
		Foreground(ThemeBase).
		Background(ThemeGreen).
		Padding(0, 3).
		Bold(true).
		MarginTop(1)

	SuccessTitleStyle = lipgloss.NewStyle().
		Foreground(ThemeBase).
		Background(ThemeGreen).
		Padding(0, 2).
		Bold(true).
		MarginBottom(1)
		
	// Refresh List UI globals too!
	HeaderStyle = lipgloss.NewStyle().Foreground(ThemeBase).Background(ThemeMauve).Bold(true)
	MilestoneRowStyle = lipgloss.NewStyle().Foreground(ThemeText).Background(ThemeOverlay)
	StoryRowStyle = lipgloss.NewStyle().Foreground(ThemeText).Background(ThemeOverlay)
	TaskRowStyle = lipgloss.NewStyle().Foreground(ThemeBase).Background(ThemeOverlay) // Darker text for lowest tier tasks
	ActiveRowStyle = lipgloss.NewStyle().Foreground(ThemeBase).Background(ThemeBlue).Bold(true) // Blue highlight for selected row
}

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
