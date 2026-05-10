# crewup Agent Templates

This folder contains the built-in agent definitions shipped with crewup. Each `.md` file defines one agent with a crewup-native frontmatter schema and a system prompt body.

> **Note:** These templates are embedded into the crewup binary at compile time via `go:embed`. Changes require rebuilding the binary.

---

## Frontmatter Schema

All values are single-line `key: value`. No nested objects.

| Field | Required | Type | Notes |
|---|---|---|---|
| `id` | Yes | string | Machine identifier; must match the filename (minus `.md`) |
| `name` | Yes | string | Display name shown in UI and tool configs |
| `description` | Yes | string | Single line; used in menus and tool description fields |
| `mode` | No | `primary` \| `subagent` | Used by Claude Code writer |
| `default_provider` | Yes | string | Must match a provider name in `assets/models.csv` |
| `default_model` | Yes | string | Must match a `model_id` in `assets/models.csv` |
| `temperature` | No | float (0.0–1.0) | Default: `0.1` |
| `max_steps` | No | int | Default: `8` |
| `tools_allow` | No | comma-separated string | crewup tool aliases (see below); omit = all tools |

### Example

```markdown
---
id: planner
name: Planner
description: Clarifies the request first, then creates a plan
mode: primary
default_provider: GitHub Copilot
default_model: github-copilot/claude-sonnet-4.6
temperature: 0.1
max_steps: 8
tools_allow: read,edit,search,bash,webfetch,task
---

You are the planner agent. [system prompt here]
```

---

## Tool Alias Reference

crewup uses tool-agnostic aliases that are rendered to each tool's native names at setup time.

| crewup alias | Claude Code | GitHub Copilot | Notes |
|---|---|---|---|
| `read` | `Read` | `read` | File reading |
| `edit` | `Edit`, `Write` | `edit` | File editing/writing |
| `search` | `Grep`, `Glob` | `search` | File search |
| `bash` | `Bash` | `execute` | Shell execution |
| `webfetch` | `WebFetch` | *(omitted — unconfirmed)* | Web fetching |
| `task` | `Task` | `agent` | Sub-agent spawning |

---

## How crewup Renders Each Field Per Tool

### Claude Code (`~/.claude/agents/<id>.md`)

```markdown
---
name: planner
description: Clarifies the request first, then creates a plan
model: claude-sonnet-4-5
tools: [Read, Edit, Grep, Glob, Bash, Task]
maxTurns: 8
---

[system prompt body]
```

- `model`: from user's selected model, or `default_model`
- `tools`: mapped from `tools_allow` aliases to Claude Code names
- `maxTurns`: from `max_steps`

### GitHub Copilot (`.github/agents/<id>.agent.md`)

```markdown
---
name: planner
description: Clarifies the request first, then creates a plan
model: claude-sonnet-4.6
tools: ["read", "edit", "search", "execute", "agent"]
---

[system prompt body]
```

- `model`: from user's selected model, or `default_model`
- `tools`: mapped from `tools_allow` aliases to Copilot names (`webfetch` omitted)

### Ollama (`~/.ollama/crewup-agents/<id>/Modelfile`)

```
FROM llama3.2
SYSTEM """
[system prompt body]
"""
PARAMETER temperature 0.1
```

- `FROM`: uses `default_model` if provider is Ollama, otherwise `llama3.2`
- crewup prints: `Run: ollama create <id> -f <path>`

### Aider (`~/.aider/crewup-agents/<id>.yml`) — reference only

```yaml
# crewup agent reference: planner
# NOT auto-loaded by Aider. Copy model: to your ~/.aider.conf.yml
# system-prompt support requires Aider >= 0.50
model: anthropic/claude-sonnet-4-5
# system-prompt: |
#   [system prompt body]
```

- crewup prints: `Copy model: to your .aider.conf.yml`

---

## How to Add a New Agent

1. Copy an existing agent file (e.g. `planner.md`) to a new file (e.g. `my-agent.md`)
2. Change `id` to match the new filename (e.g. `my-agent`)
3. Update `name`, `description`, `default_provider`, `default_model`
4. Fill in the system prompt body after the closing `---`
5. Rebuild the binary: `go build ./...`

> The `id` field **must** match the filename (minus `.md`). crewup validates this at startup.

---

## How Model Selection Works

1. Each agent has a `default_provider` and `default_model` in its frontmatter
2. These must match entries in `assets/models.csv`
3. During `crewup init`, users can customize the model per agent
4. The selected model is stored in `~/.crewup/config.json` under each agent's `model` field
5. Writers use the selected model when rendering tool configs

To add more providers or models, edit `assets/models.csv` and rebuild. See the CSV header for column definitions.

---

## Contributing New Agents

1. Fork the repository at https://github.com/saeid-rez/crewup
2. Add your agent file to `templates/agents/`
3. Ensure the frontmatter is valid (all required fields present, `id` matches filename)
4. Add a meaningful system prompt
5. Run `go build ./...` and `go test ./...` to verify
6. Open a pull request with a description of the agent's purpose
