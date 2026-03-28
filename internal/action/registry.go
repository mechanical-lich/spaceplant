package action

// SimpleFactory creates an action with no direction parameter.
// Used for skills that add new key bindings.
type SimpleFactory func() Action

var simpleRegistry = map[string]SimpleFactory{}

// RegisterSimple registers a no-parameter action factory under the given ID.
// Call this at startup before skills are loaded.
func RegisterSimple(id string, factory SimpleFactory) {
	simpleRegistry[id] = factory
}

// CreateSimple instantiates an action by ID, or nil if the ID is unknown.
func CreateSimple(id string) Action {
	if f, ok := simpleRegistry[id]; ok {
		return f()
	}
	return nil
}

func init() {
	RegisterSimple("heal", func() Action { return HealAction{} })
	RegisterSimple("pickup", func() Action { return PickupAction{} })
	RegisterSimple("equip", func() Action { return EquipAction{} })
	RegisterSimple("stairs", func() Action { return StairsAction{} })
	RegisterSimple("roundhouse_kick", func() Action { return RoundhouseKickAction{} })
}
