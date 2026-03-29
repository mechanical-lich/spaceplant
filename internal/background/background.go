package background

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// BackgroundDef describes a character background — a fixed bundle of skills
// granted automatically at character creation (not upgradable).
type BackgroundDef struct {
	ID          string
	Name        string
	Description string
	Skills      []string
}

var registry = map[string]*BackgroundDef{}

// Load reads background definitions from a JSON file.
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("background.Load: %w", err)
	}
	var defs []BackgroundDef
	if err := json.Unmarshal(data, &defs); err != nil {
		return fmt.Errorf("background.Load: %w", err)
	}
	for i := range defs {
		registry[defs[i].ID] = &defs[i]
	}
	return nil
}

// Get returns the BackgroundDef for the given ID, or nil if not found.
func Get(id string) *BackgroundDef {
	return registry[id]
}

// All returns all registered backgrounds in stable (sorted by ID) order.
func All() []*BackgroundDef {
	defs := make([]*BackgroundDef, 0, len(registry))
	for _, d := range registry {
		defs = append(defs, d)
	}
	sort.Slice(defs, func(i, j int) bool { return defs[i].ID < defs[j].ID })
	return defs
}
