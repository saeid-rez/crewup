package ui

import (
	"fmt"

	"charm.land/huh/v2"
	"github.com/saeid-rez/crewup/internal/config"
	"github.com/saeid-rez/crewup/internal/models"
)

// ConfigureAgents runs the streamlined multi-agent setup wizard.
//
// Step A: "Use default models for all agents?" [Yes/No]
// Step B (if No): multi-select which agents to customize
// Step C: per-agent wizard for each selected agent
//
// Non-TTY: returns all roles with Model=nil (use defaults).
func ConfigureAgents(roles []config.AgentRole, allModels []models.Model) ([]config.AgentRole, error) {
	if !isTTY() {
		return roles, nil
	}

	// Step A: use defaults for all?
	useDefaults := true
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title("🤖 Use default models for all agents?").
				Description("You can customize individual agents in the next step.").
				Value(&useDefaults),
		),
	)
	if err := form.Run(); err != nil {
		return nil, normalizePromptError(err)
	}

	if useDefaults {
		return roles, nil
	}

	// Step B: which agents to customize?
	opts := make([]huh.Option[string], len(roles))
	for i, r := range roles {
		opts[i] = huh.NewOption(r.Name+" — "+r.Description, r.ID)
	}

	var toCustomize []string
	form2 := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("🎛  Which agents do you want to customize?").
				Description("Space to toggle, Enter to confirm").
				Options(opts...).
				Value(&toCustomize),
		),
	)
	if err := form2.Run(); err != nil {
		return nil, normalizePromptError(err)
	}

	customizeSet := make(map[string]bool, len(toCustomize))
	for _, id := range toCustomize {
		customizeSet[id] = true
	}

	// Step C: per-agent wizard for selected agents
	result := make([]config.AgentRole, len(roles))
	copy(result, roles)
	for i, role := range result {
		if !customizeSet[role.ID] {
			continue
		}
		configured, err := ConfigureAgent(role, allModels)
		if err != nil {
			return nil, err
		}
		result[i] = configured
	}

	return result, nil
}

// ConfigureAgent runs the per-agent huh wizard for a single agent.
//
// Step 1: provider select (from models.Providers())
// Step 2: model select (from models.ByProvider(provider)), default pre-selected
//
// Returns role with Model populated.
func ConfigureAgent(role config.AgentRole, allModels []models.Model) (config.AgentRole, error) {
	if !isTTY() {
		return role, nil
	}

	providers := uniqueProviders(allModels)
	if len(providers) == 0 {
		return role, nil
	}

	// Step 1: pick provider
	providerOpts := make([]huh.Option[string], len(providers))
	for i, p := range providers {
		providerOpts[i] = huh.NewOption(p, p)
	}

	selectedProvider := providers[0]
	form1 := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("🤖 %s — Select provider", role.Name)).
				Options(providerOpts...).
				Value(&selectedProvider),
		),
	)
	if err := form1.Run(); err != nil {
		return role, normalizePromptError(err)
	}

	// Step 2: pick model
	providerModels := modelsForProvider(allModels, selectedProvider)
	if len(providerModels) == 0 {
		return role, nil
	}

	modelOpts := make([]huh.Option[string], len(providerModels))
	defaultModelID := providerModels[0].ModelID
	for i, m := range providerModels {
		label := m.DisplayName
		if m.Description != "" {
			label += " — " + m.Description
		}
		modelOpts[i] = huh.NewOption(label, m.ModelID)
		if m.IsDefault {
			defaultModelID = m.ModelID
		}
	}

	selectedModel := defaultModelID
	form2 := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("🤖 %s — Select model", role.Name)).
				Options(modelOpts...).
				Value(&selectedModel),
		),
	)
	if err := form2.Run(); err != nil {
		return role, normalizePromptError(err)
	}

	role.Model = &config.ModelSelection{
		Provider:   selectedProvider,
		ModelID:    selectedModel,
		Customized: true,
	}
	return role, nil
}

func uniqueProviders(allModels []models.Model) []string {
	seen := make(map[string]bool)
	var result []string
	for _, m := range allModels {
		if !seen[m.Provider] {
			seen[m.Provider] = true
			result = append(result, m.Provider)
		}
	}
	return result
}

func modelsForProvider(allModels []models.Model, provider string) []models.Model {
	var result []models.Model
	for _, m := range allModels {
		if m.Provider == provider {
			result = append(result, m)
		}
	}
	return result
}
