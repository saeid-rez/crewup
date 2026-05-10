package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// readJSONOrEmpty reads a JSON file into a map, or returns an empty map if the file doesn't exist.
// Returns an error if the file exists but is malformed.
func readJSONOrEmpty(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]interface{}{}, nil
		}
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("could not parse %s: %w. Fix it manually or delete it", path, err)
	}
	return m, nil
}

// writeJSONWithBackup writes a map as pretty JSON to path, backing up any existing file first.
// Preserves the existing file's permissions if the file already exists.
func writeJSONWithBackup(path string, data map[string]interface{}) error {
	backupPath := path + ".crewup.bak"
	perm := os.FileMode(0644)

	// Backup existing file and capture its permissions.
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
		if existing, err := os.ReadFile(path); err == nil {
			_ = os.WriteFile(backupPath, existing, perm)
		}
	}

	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, out, perm); err != nil {
		// Attempt restore
		if backup, berr := os.ReadFile(backupPath); berr == nil {
			_ = os.WriteFile(path, backup, perm)
		}
		return err
	}
	return nil
}

// getOrCreateMap returns the nested map at key, creating it if missing.
// Returns an error if the key exists but is not a JSON object — this prevents
// silently clobbering unexpected config shapes.
func getOrCreateMap(m map[string]interface{}, key string, path string) (map[string]interface{}, error) {
	if v, ok := m[key]; ok {
		if sub, ok := v.(map[string]interface{}); ok {
			return sub, nil
		}
		return nil, fmt.Errorf("config file %s: key %q has unexpected type %T, expected object — fix it manually", path, key, v)
	}
	sub := map[string]interface{}{}
	m[key] = sub
	return sub, nil
}
