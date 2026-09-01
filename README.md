# Tally

A blistering fast TUI/CLI task manager and time tracker driven entirely by local plain-text files. Own your data, survive cloud outages, and push your hours to the enterprise when you are ready.

Part of the **[GotCode Collective](https://gotcode.org)**.

## Core Features

- **Sharded File Storage:** Zero databases. Every task is its own Markdown file organized in a `YYYY/MM/DD` directory tree. This guarantees infinite scalability and absolutely zero Git merge conflicts when syncing across devices.
- **Enterprise Offline Sync:** Survive cloud outages. Create tasks and track time entirely locally. When the internet returns, run `tally sync` to automatically push your hours to the **7pace TimeTracker API** and generate stories in **Azure DevOps**.
- **Sequential Task IDs:** Every file is sequentially named, granting every task a pristine local ID (e.g., `20260831.001`). Instantly clock-in from the terminal by running `tally start 20260831.001`.
- **Dynamic TUI Dashboards:** Launch the full-screen Bubbletea TUI and use hotkeys to instantly toggle between **Day View**, **Week View**, and **Month View** to track your long-term velocity and project burn rates.
- **Reusable Component Architecture:** The internal UI layer is built on 100% generic, data-driven Form and List components. Business logic is strictly separated via a flat DDD/CQRS architecture in the `internal/` directory.
- **Cascading Configuration:** Respects standard XDG Base Directories while allowing per-project `.tally/config.yaml` overrides. Seamlessly switch ADO Area Paths just by changing directories in your terminal.

## Configuration

Tally respects the XDG Base Directory specification:
* **Global Config:** `~/.config/tally/config.yaml`
* **Local Data (Tasks):** `~/.local/share/tally/tasks/`
* **Local Overrides:** `./.tally/config.yaml` (inside your active Git repository)

## License

This project is licensed under the GNU General Public License v3.0 (GPL-3.0) - see the [LICENSE](LICENSE) file for details.
