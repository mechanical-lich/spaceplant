package components

// PositionComponent .
type PositionComponent struct {
	x, y  int
	Level int
}

func (pc PositionComponent) GetType() string {
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
