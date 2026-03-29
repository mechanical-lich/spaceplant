package class

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

type ClassDef struct {
	ID          string
	Name        string
	Description string
	Skills      []string // level → skill IDs
}

var registry = map[string]*ClassDef{}

// Load reads class definitions from a JSON file (an array of ClassDef objects).
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("class.Load: %w", err)
	}
	var defs []ClassDef
	if err := json.Unmarshal(data, &defs); err != nil {
		return fmt.Errorf("class.Load: %w", err)
	}
	for i := range defs {
		registry[defs[i].ID] = &defs[i]
	}
	return nil
}

// Get returns the ClassDef for the given ID, or nil if not found.
func Get(id string) *ClassDef {
	return registry[id]
}

// All returns all registered classes in stable (sorted by ID) order.
func All() []*ClassDef {
	defs := make([]*ClassDef, 0, len(registry))
	for _, d := range registry {
		defs = append(defs, d)
	}
	sort.Slice(defs, func(i, j int) bool { return defs[i].ID < defs[j].ID })
	return defs
}
