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
			descBuilder.WriteString(fmt.Sprintf("<div>%s</div>", line))
		} else if currentSection == "ac" {
			acBuilder.WriteString(fmt.Sprintf("<div>%s</div>", line))
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
