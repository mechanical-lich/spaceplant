package action

// SkillFactory creates an action using parameters from a skill definition.
// Actions that don't need parameters simply ignore them.
type SkillFactory func(params ActionParams) Action

var skillRegistry = map[string]SkillFactory{}

// RegisterSkill registers a param-aware action factory under the given ID.
func RegisterSkill(id string, factory SkillFactory) {
	skillRegistry[id] = factory
}

// RegisterSimple registers a no-parameter action factory. It is a convenience
// wrapper around RegisterSkill for actions that don't use skill params.
func RegisterSimple(id string, factory func() Action) {
	skillRegistry[id] = func(_ ActionParams) Action { return factory() }
}

// CreateSkillAction instantiates an action by ID with the given skill params,
// or nil if the ID is unknown.
func CreateSkillAction(id string, params ActionParams) Action {
	if f, ok := skillRegistry[id]; ok {
		return f(params)
	}
	return nil
}

func init() {
	RegisterSimple("heal", func() Action { return HealAction{} })
	RegisterSimple("pickup", func() Action { return PickupAction{} })
	RegisterSimple("equip", func() Action { return EquipAction{} })
	RegisterSimple("stairs", func() Action { return StairsAction{} })
	RegisterSimple("roundhouse_kick", func() Action { return RoundhouseKickAction{} })
	RegisterSkill("cone_of", func(p ActionParams) Action {
		return ConeOfAction{Params: p}
	})
}
