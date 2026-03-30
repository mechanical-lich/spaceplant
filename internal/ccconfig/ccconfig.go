package ccconfig

import (
	"encoding/json"
	"fmt"
	"os"
)

// Config holds character creator configuration loaded from JSON.
type Config struct {
	BaseSkills []string `json:"base_skills"`
}

var loaded Config

// Load reads character creator config from a JSON file.
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ccconfig.Load: %w", err)
	}
	if err := json.Unmarshal(data, &loaded); err != nil {
		return fmt.Errorf("ccconfig.Load: %w", err)
	}
	return nil
}

// Get returns the loaded character creator config.
func Get() *Config {
	return &loaded
}
