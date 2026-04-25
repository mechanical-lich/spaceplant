package action

import (
	"log"
	"os"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/component"
)

// applyScriptableCondition adds a ScriptableConditionComponent to target using
// params. It logs a warning and does nothing if the script file does not exist.
func applyScriptableCondition(target *ecs.Entity, params ActionParams, duration int) {
	path := params.StatusConditionScript
	if path == "" {
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("[action] WARNING: status condition script not found: %s", path)
		return
	}
	interval := params.StatusConditionScriptInterval
	if interval <= 0 {
		interval = 1
	}
	name := params.StatusConditionOnFailSave
	if name == "" {
		name = "script"
	}
	acc := rlcomponents.GetOrCreateActiveConditions(target)
	acc.Add(&component.ScriptableConditionComponent{
		Name:       name,
		Duration:   duration,
		ScriptPath: path,
		Interval:   interval,
	})
}
