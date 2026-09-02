# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
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
