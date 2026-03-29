package component

import (
	"github.com/mechanical-lich/mlge/ecs"
)

type ClassComponent struct {
	Classes       []string
	UpgradePoints int
	ChosenSkills  []string
}

func (c *ClassComponent) GetType() ecs.ComponentType {
	return Class
}

func (c *ClassComponent) HasClass(class string) bool {
	for _, c := range c.Classes {
		if c == class {
			return true
		}
	}
	return false
}
