package component

import (
	"github.com/mechanical-lich/mechanical-basic/pkg/basic"
	"github.com/mechanical-lich/mlge/ecs"
)

const Script ecs.ComponentType = "ScriptComponent"

// ScriptContext is defined in system/ScriptSystem.go but stored here to avoid
// an import cycle. We use any to keep the component package free of system deps.
type ScriptComponent struct {
	ScriptPath  string
	Vars        map[string]any
	Interpreter *basic.MechBasic `json:"-"`
	Ctx         any                    `json:"-"`
}

func (c *ScriptComponent) GetType() ecs.ComponentType {
	return Script
}
