package component

import "github.com/mechanical-lich/mlge/ecs"

// PositionComponent .
type PositionComponent struct {
	x, y  int
	Level int
}

func (pc PositionComponent) GetType() ecs.ComponentType {
	return "PositionComponent"
}

func (pc PositionComponent) GetX() int {
	return pc.x
}
func (pc PositionComponent) GetY() int {
	return pc.y
}
func (pc *PositionComponent) SetPosition(x int, y int) {
	pc.x = x
	pc.y = y

}
