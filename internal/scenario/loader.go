package scenario

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
)

var active *Scenario
var enabled []Scenario

// Load reads all *.json files from dir and stores the enabled ones.
// Call SelectRandom or SelectByID afterward (or call Load followed by
// SelectRandom, which is what game.go does).
func Load(dir string) error {
	entries, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil || len(entries) == 0 {
		return fmt.Errorf("scenario: no scenario files found in %s", dir)
	}

	enabled = nil
	for _, path := range entries {
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("scenario: read %s: %w", path, err)
		}
		var s Scenario
		if err := json.Unmarshal(data, &s); err != nil {
			return fmt.Errorf("scenario: parse %s: %w", path, err)
		}
		if s.Enabled {
			enabled = append(enabled, s)
		}
	}

	if len(enabled) == 0 {
		return fmt.Errorf("scenario: no enabled scenarios found in %s", dir)
	}

	return SelectRandom()
}

// SelectRandom picks a random enabled scenario as the active one.
func SelectRandom() error {
	if len(enabled) == 0 {
		return fmt.Errorf("scenario: no enabled scenarios loaded")
	}
	picked := enabled[rand.Intn(len(enabled))]
	active = &picked
	return nil
}

// SelectByID sets the active scenario to the one with the given ID.
// Returns an error if no enabled scenario with that ID exists.
func SelectByID(id string) error {
	for i := range enabled {
		if enabled[i].ID == id {
			active = &enabled[i]
			return nil
		}
	}
	return fmt.Errorf("scenario: no enabled scenario with id %q", id)
}

// AllEnabled returns all currently loaded enabled scenarios.
func AllEnabled() []Scenario {
	return enabled
}

// Active returns the active scenario. Panics if Load was never called.
func Active() *Scenario {
	if active == nil {
		panic("scenario: Active() called before Load()")
	}
	return active
}
