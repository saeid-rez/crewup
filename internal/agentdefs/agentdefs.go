package agentdefs

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/saeid-rez/crewup/templates"
)

// ValidToolAliases is the set of recognized crewup tool alias names.
var ValidToolAliases = []string{"read", "edit", "search", "bash", "webfetch", "web", "task"}

// AgentDef holds the parsed definition of a crewup agent template.
type AgentDef struct {
	ID                     string
	Name                   string
	Description            string
	Mode                   string // "primary" or "subagent"
	DefaultProvider        string
	DefaultModel           string
	Temperature            float64
	MaxSteps               int
	ToolsAllow             []string // nil = all tools; non-nil = explicit allowlist
	Prompt                 string   // body after frontmatter
	DisableModelInvocation *bool    // nil = not set; non-nil = explicitly set
}

// All parses all agent templates from the embedded FS and returns them.
// Returns an error if any template fails to parse.
func All() ([]AgentDef, error) {
	entries, err := fs.ReadDir(templates.AgentFS, "agents")
	if err != nil {
		return nil, fmt.Errorf("reading embedded agent templates: %w", err)
	}

	var defs []AgentDef
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".md" {
			continue
		}

		data, err := templates.AgentFS.ReadFile("agents/" + entry.Name())
		if err != nil {
			return nil, fmt.Errorf("reading agent template %s: %w", entry.Name(), err)
		}

		def, err := parse(string(data))
		if err != nil {
			return nil, fmt.Errorf("parsing agent template %s: %w", entry.Name(), err)
		}

		// Derive filename stem (e.g. "planner" from "planner.md")
		stem := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))

		// If ID is set in frontmatter, it must match the filename stem.
		// If ID is not set, default it to the filename stem.
		if def.ID != "" {
			if def.ID != stem {
				return nil, fmt.Errorf("agent file %s: id %q does not match filename", entry.Name(), def.ID)
			}
		} else {
			def.ID = stem
		}

		defs = append(defs, def)
	}

	return defs, nil
}

// ByID returns the AgentDef with the given ID, or (zero, false) if not found.
func ByID(id string) (AgentDef, bool) {
	defs, err := All()
	if err != nil {
		return AgentDef{}, false
	}
	for _, d := range defs {
		if d.ID == id {
			return d, true
		}
	}
	return AgentDef{}, false
}

// parse parses a single agent template file.
// Frontmatter is delimited by lines that are exactly "---".
// The parser scans line-by-line to avoid breaking on "---" in the body.
func parse(content string) (AgentDef, error) {
	lines := strings.Split(content, "\n")

	// Find the opening "---"
	start := -1
	for i, line := range lines {
		if strings.TrimRight(line, "\r") == "---" {
			start = i
			break
		}
	}
	if start == -1 {
		return AgentDef{}, fmt.Errorf("missing opening frontmatter delimiter '---'")
	}

	// Find the closing "---"
	end := -1
	for i := start + 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return AgentDef{}, fmt.Errorf("missing closing frontmatter delimiter '---'")
	}

	// Parse frontmatter key: value lines
	fm := make(map[string]string)
	for _, line := range lines[start+1 : end] {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		idx := strings.Index(line, ":")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(line[:idx])
		val := strings.TrimSpace(line[idx+1:])
		fm[key] = val
	}

	// Extract body (everything after closing ---, trimmed)
	bodyLines := lines[end+1:]
	// Trim leading blank lines
	for len(bodyLines) > 0 && strings.TrimSpace(bodyLines[0]) == "" {
		bodyLines = bodyLines[1:]
	}
	// Trim trailing blank lines
	for len(bodyLines) > 0 && strings.TrimSpace(bodyLines[len(bodyLines)-1]) == "" {
		bodyLines = bodyLines[:len(bodyLines)-1]
	}
	body := strings.Join(bodyLines, "\n")

	// Validate required fields (id is optional; if absent, All() defaults it to the filename stem)
	for _, req := range []string{"name", "description", "default_provider", "default_model"} {
		if fm[req] == "" {
			return AgentDef{}, fmt.Errorf("missing required frontmatter field: %q", req)
		}
	}

	def := AgentDef{
		ID:              fm["id"],
		Name:            fm["name"],
		Description:     fm["description"],
		Mode:            fm["mode"],
		DefaultProvider: fm["default_provider"],
		DefaultModel:    fm["default_model"],
		Prompt:          body,
	}

	// Parse temperature
	if v, ok := fm["temperature"]; ok && v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return AgentDef{}, fmt.Errorf("invalid temperature %q: %w", v, err)
		}
		def.Temperature = f
	} else {
		def.Temperature = 0.1 // default
	}

	// Parse max_steps
	if v, ok := fm["max_steps"]; ok && v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return AgentDef{}, fmt.Errorf("invalid max_steps %q: %w", v, err)
		}
		def.MaxSteps = n
	} else {
		def.MaxSteps = 8 // default
	}

	// Parse tools_allow
	if v, ok := fm["tools_allow"]; ok {
		// Field is present — parse it strictly
		if strings.TrimSpace(v) == "" {
			return AgentDef{}, fmt.Errorf("tools_allow is present but empty; omit the field to allow all tools")
		}
		tokens := strings.Split(v, ",")
		var aliases []string
		for _, tok := range tokens {
			tok = strings.TrimSpace(tok)
			if tok == "" {
				return AgentDef{}, fmt.Errorf("tools_allow contains an empty token (check for double commas or trailing commas)")
			}
			if !isValidAlias(tok) {
				return AgentDef{}, fmt.Errorf("unknown tool alias %q in tools_allow (valid: %s)", tok, strings.Join(ValidToolAliases, ", "))
			}
			aliases = append(aliases, tok)
		}
		def.ToolsAllow = aliases
	}
	// If tools_allow key is absent, ToolsAllow remains nil (= all tools)

	// Parse disable_model_invocation (optional)
	if v, ok := fm["disable_model_invocation"]; ok && v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return AgentDef{}, fmt.Errorf("invalid disable_model_invocation %q: %w", v, err)
		}
		def.DisableModelInvocation = &b
	}

	return def, nil
}

func isValidAlias(alias string) bool {
	for _, v := range ValidToolAliases {
		if v == alias {
			return true
		}
	}
	return false
}
