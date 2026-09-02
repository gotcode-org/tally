package main

import (
	"fmt"
	"strings"
)

func parseMarkdownSections(body string) (description, acceptanceCriteria string) {
	lines := strings.Split(body, "\n")
	
	var currentSection string = "description"
	var descBuilder strings.Builder
	var acBuilder strings.Builder
	
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		
		if strings.HasPrefix(lower, "# description") {
			currentSection = "description"
			continue
		} else if strings.HasPrefix(lower, "# acceptance criteria") {
			currentSection = "ac"
			continue
		}
		
		if currentSection == "description" {
			descBuilder.WriteString(line + "<br>")
		} else if currentSection == "ac" {
			acBuilder.WriteString(line + "<br>")
		}
	}
	
	return strings.TrimSpace(descBuilder.String()), strings.TrimSpace(acBuilder.String())
}

func main() {
	body := `
# Description
Meetings - 2026/09/02

# Acceptance Criteria
[ ] - Documentum D2 Smartview Daily Standup
`
	desc, ac := parseMarkdownSections(body)
	fmt.Printf("Desc: %s\nAC: %s\n", desc, ac)
}
