package models

import (
	"encoding/csv"
	"strconv"
	"strings"

	"github.com/saeid-rez/crewup/assets"
)

// Model represents a single model entry from models.csv.
type Model struct {
	Provider    string
	Tool        string
	ModelID     string
	DisplayName string
	Description string
	IsDefault   bool
	IsInternal  bool
}

// All parses assets.ModelsCSV and returns all models.
// Returns an empty slice if the CSV is empty or has only a header.
func All() []Model {
	r := csv.NewReader(strings.NewReader(assets.ModelsCSV))
	records, err := r.ReadAll()
	if err != nil || len(records) <= 1 {
		return nil
	}

	var models []Model
	for _, row := range records[1:] { // skip header
		if len(row) < 7 {
			continue
		}
		models = append(models, Model{
			Provider:    strings.TrimSpace(row[0]),
			Tool:        strings.TrimSpace(row[1]),
			ModelID:     strings.TrimSpace(row[2]),
			DisplayName: strings.TrimSpace(row[3]),
			Description: strings.TrimSpace(row[4]),
			IsDefault:   parseBool(row[5]),
			IsInternal:  parseBool(row[6]),
		})
	}
	return models
}

// Providers returns unique provider names in CSV order.
func Providers() []string {
	seen := make(map[string]bool)
	var result []string
	for _, m := range All() {
		if !seen[m.Provider] {
			seen[m.Provider] = true
			result = append(result, m.Provider)
		}
	}
	return result
}

// ByProvider returns all models for the given provider.
func ByProvider(p string) []Model {
	var result []Model
	for _, m := range All() {
		if m.Provider == p {
			result = append(result, m)
		}
	}
	return result
}

// DefaultFor returns the default model for the given provider.
// It returns the first model with is_default=true; if none, it returns the first model.
// Returns (zero, false) if the provider has no models.
func DefaultFor(provider string) (Model, bool) {
	models := ByProvider(provider)
	if len(models) == 0 {
		return Model{}, false
	}
	for _, m := range models {
		if m.IsDefault {
			return m, true
		}
	}
	// No explicit default — use first
	return models[0], true
}

// ByModelID returns the model entry for the given canonical model ID.
func ByModelID(modelID string) (Model, bool) {
	for _, m := range All() {
		if m.ModelID == modelID {
			return m, true
		}
	}
	return Model{}, false
}

func parseBool(s string) bool {
	b, _ := strconv.ParseBool(strings.TrimSpace(s))
	return b
}
