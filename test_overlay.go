package main

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

func main() {
	bg := lipgloss.NewStyle().Width(10).Height(10).Background(lipgloss.Color("0")).Render("bg")
	fg := lipgloss.NewStyle().Width(5).Height(5).Background(lipgloss.Color("1")).Render("fg")
	out := lipgloss.PlaceOverlay(2, 2, fg, bg, false)
	fmt.Println(out)
}
