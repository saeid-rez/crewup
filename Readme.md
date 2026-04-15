# crewup

> Setup AI agent crews for any AI tool on your machine.

crewup is a cross-platform CLI tool that helps you configure complete AI agent
workflows (planner, implementor, reviewer, tester, documenter) for tools like
Claude CLI, GitHub Copilot, Ollama, Aider, and more.

## Features

- 🔍 **Auto-detects** installed AI tools on your machine
- 🤖 **Scaffolds agent crews** with pre-built role prompts
- 🔌 **Installs MCP servers** (Context7, Filesystem, GitHub, etc.) with one command
- 🔄 **Update checker** — notifies you when a new version is available
- 🖥  **Cross-platform** — macOS, Linux, Windows

## Install

**macOS (Homebrew):**
```bash
brew tap yourusername/tap
brew install crewup
```

**Linux / macOS (direct):**
```bash
curl -sSf https://raw.githubusercontent.com/yourusername/crewup/main/install.sh | sh
```

**Windows (Scoop):**
```powershell
scoop bucket add yourusername https://github.com/yourusername/scoop-bucket
scoop install crewup
```

**Go install (if you have Go):**
```bash
go install github.com/yourusername/crewup@latest
```

## Usage

```bash
# Interactive setup wizard
crewup init

# Add an MCP server to your existing setup
crewup add mcp context7
crewup add mcp filesystem

# Add an agent role
crewup add agent tester

# List your current setup
crewup list

# Check version
crewup version
```

## Project Structure

```
crewup/
├── main.go
├── cmd/
│   ├── root.go       # root command + update check
│   ├── init.go       # interactive setup wizard
│   ├── add.go        # add mcp / agent subcommands
│   └── list.go       # list current setup
├── internal/
│   ├── config/       # crewup config read/write (~/.crewup/config.json)
│   ├── updater/      # GitHub release check + self-update
│   ├── tools/        # AI tool detection & config writers
│   └── mcp/          # MCP server registry & installer
└── pkg/
    └── ui/           # terminal UI (bubbletea prompts, menus, banners)
```

## Roadmap

- [ ] Bubbletea interactive menus for init wizard
- [ ] Self-update binary in-place
- [ ] Claude CLI agent config writer
- [ ] Ollama Modelfile agent config writer
- [ ] More MCP servers (Postgres, Slack, Notion, Linear)
- [ ] `crewup remove` command
- [ ] `crewup sync` — re-apply config after tool updates

## License

MIT
