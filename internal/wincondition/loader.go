package wincondition

import (
	"encoding/json"
	"fmt"
	"os"
)

var active *Evaluator

// Load reads the rule set from path and installs it as the active evaluator.
func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("wincondition: load %s: %w", path, err)
	}
	var rs RuleSet
	if err := json.Unmarshal(data, &rs); err != nil {
		return fmt.Errorf("wincondition: parse %s: %w", path, err)
	}
	active = New(rs)
	return nil
}

// Active returns the active evaluator. Panics if Load was never called.
func Active() *Evaluator {
	if active == nil {
		panic("wincondition: Active() called before Load()")
	}
	return active
}
