package action

// PlayerActions returns all possible player actions.
// This list is used by HasAvailableAction to determine whether the player
// has any affordable, available action — enabling auto-skip when they don't.
func PlayerActions() []Action {
	return []Action{
		MoveAction{DeltaX: 0, DeltaY: -1},
		MoveAction{DeltaX: 0, DeltaY: 1},
		MoveAction{DeltaX: -1, DeltaY: 0},
		MoveAction{DeltaX: 1, DeltaY: 0},
		HealAction{},
		PickupAction{},
		EquipAction{},
		StairsAction{},
		ShootAction{},
	}
}
