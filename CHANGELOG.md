# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-09-02

### Added
- **ADO Sync Debug Toggle**: Added `ado.debug: true` flag to `config.yaml` to optionally enable verbose JSON telemetry dumps during syncs (silenced by default).
- **5 New Color Themes**: Added Gruvbox, Tokyo Night, Rosé Pine, Monokai, and Solarized Dark to the TUI's rotation (press `t` to cycle through them).
- **TUI Time Summary**: Added live granular time logging summary counters (Day, Week, Month, Year) to the footer of the TUI dashboard.
- **Dynamic Swimlane TUI Dropdown**: Added a dynamic dropdown to the TUI "Create Task" modal to select ADO swimlanes, driven by a new `ado.swimlanes` array in `config.yaml`.
- **Default Swimlane Auto-Routing**: Added `ado.default_swimlane` configuration setting so newly created tasks automatically route to the correct ADO board lane upon creation.
- **ADO Swimlane Support**: Users can now push tasks directly into specific board lanes by setting `ado.swimlane_field` in their config and `swimlane` in their markdown frontmatter.
- **Tally Debug CLI**: Introduced `tally debug [ado_id]` command to fetch and dump raw JSON payloads from the ADO REST API, specifically helping users hunt down their custom `WEF_*_Kanban.Lane` field names.
- **ADO Hierarchies**: Full parent-child relationship support. Subtasks can be created directly from the TUI by highlighting a parent Story and pressing `c`.
- **Dashboard Nesting**: Subtasks visually render indented underneath their parent Story on the TUI dashboard tree.
- **Granular Time Logging**: Core engine upgraded to store discrete `TimeLog` entries instead of a single merged integer, allowing independent syncing of time entries to 7pace with dynamic Activity IDs.
- **Dynamic Activity Types**: The TUI 'Log Time' dialog now automatically populates an Activity Type dropdown based on the `7pace.activities` map in `~/.config/tally/config.yaml`.
- **System.LinkTypes.Hierarchy-Reverse Patching**: The ADO sync engine dynamically links child tasks to parent stories in the ADO database during sync sweeps.
- Implemented full Task Deletion support with a new `tally delete [id]` CLI command and an `x` keybinding in the TUI.
- Integrated a new interactive `Log Time` dialog directly into the TUI (via the `a` keybinding) to rapidly log time without exiting the dashboard.
- Overhauled the Create Task screen into a clean, horizontally-centered floating dialog box.
- Enclosed text input fields in the dialog inside heavily styled, inset rounded borders.
- Re-aligned dialog form fields and removed redundant help text for a more professional layout.
- Added dynamic bottom-padding to the Dashboard to maintain the UI grid structure all the way to the footer, even when empty.
- Stripped bulky background badges from Type and Status columns in the list view to reduce visual noise.
- Integrated generic Bubbletea Form/List framework into a fully functional `tally ui` dashboard.
- Implemented `tally points` and automatic Story Point defaults to satisfy strict Azure DevOps state transition rules.
- Upgraded the Sync Engine to dynamically parse Markdown `# Description` and `# Acceptance Criteria` headers into native ADO fields.
- Implemented chronological Tree rendering in the Dashboard UI (Year -> Month -> Day).
- Added `tea.ExecProcess` keybindings in the TUI to launch `$EDITOR` (`e`) and execute `tally sync` (`s`) seamlessly.
- Added `tally edit` and `tally state` CLI commands.
- Configured dynamic calculation of time logged replacing static UI progress bars.
- Created `Makefile` with robust `PREFIX` and `DESTDIR` support for native Linux installation.
- Core Domain Model (`internal/core/task.go`) with Azure DevOps and 7pace time-tracking fields.
- Initial project directory skeleton matching a flat DDD/CQRS architecture.
- Standardized `GPL-3.0` License headers applied to all Go source files.

### Fixed
- **ADO Update Payload Drops**: Swimlanes, Tags, and Story Points are now properly injected into the ADO JSON patch payload during *updates* (previously they only triggered on task creation).
- **ADO WYSIWYG Editor Swallowing**: Markdown body output is now wrapped in explicit HTML `<div>` elements during ADO syncs to prevent Azure DevOps from silently dropping bare text strings in rich text fields.
- **Recurrence Scheduling Rules**: Overhauled the Reconciliation engine to properly evaluate ISO weeks and calendar boundaries for `weekly`, `monthly`, and `weekdays` rules instead of cloning everything daily.
- **Ghost Spawns**: Forced the Reconciliation Engine to fire instantly upon form submission so newly created templates spawn clones immediately without requiring an app restart.
- **Template Sync Exclusion**: Explicitly blocked raw Master Templates (`recur-*`) from syncing to ADO and 7pace to prevent phantom timesheets and ghost work items.
- Fixed critical TUI transparency bugs where ANSI reset codes (`\x1b[0m`) emitted by Bubbletea text inputs and borders would shatter the UI background and expose the terminal's native background.
- Resolved severe contrast issues when cycling themes (specifically in Nord) by dynamically checking the active row's state and inverting text highlights when necessary.
- Decoupled column separators from dynamic row styling to ensure grid lines draw consistently down the screen.
