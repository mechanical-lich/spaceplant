
## Adding a new action

If no existing action fits, create one in `internal/action/`.

### 1. Create the action file

```go
package action

import (
    "github.com/mechanical-lich/ml-rogue-lib/pkg/rlenergy"
    "github.com/mechanical-lich/mlge/ecs"
    "github.com/mechanical-lich/spaceplant/internal/energy"
    "github.com/mechanical-lich/spaceplant/internal/world"
)

type MyAction struct {
    Params ActionParams // include if your action uses skill params
}

func (a MyAction) Cost(_ *ecs.Entity, _ *world.Level) int {
    return energy.CostAttack
}

func (a MyAction) Available(entity *ecs.Entity, level *world.Level) bool {
    // return true when the action can be taken
    return true
}

func (a MyAction) Execute(entity *ecs.Entity, level *world.Level) error {
    // do the thing
    rlenergy.SetActionCost(entity, energy.CostAttack)
    return nil
}
```

**Energy cost constants** (from `internal/energy/`):

| Constant | Value | Meaning |
|---|---|---|
| `energy.CostAttack` | 100 | Full action |
| `energy.CostQuick` | 50 | Half action (e.g. pickup, equip) |

### 2. Register the action

In `internal/action/registry.go`, add to the `init()` function:

```go
// No params needed:
RegisterSimple("my_action", func() Action { return MyAction{} })

// With ActionParams:
RegisterSkill("my_action", func(p ActionParams) Action {
    return MyAction{Params: p}
})
```

### 3. Reference it in a skill

```json
"action_bindings": {
  "x": "my_action"
}
```

### 4. Handle AI (optional)

If the action should be used by AI, set `ai_type` on the skill. The template scripts (`ai_hostile.basic`, `ai_advanced.basic`) call `use_skill(ai_type)` to find and execute the matching skill automatically — no Go code changes needed. The two built-in `ai_type` values are `"align_and_shoot"` (ranged, fires when aligned with target) and `"melee_skill"` (used when adjacent). Follow the same pattern for new types and call `use_skill("your_ai_type")` from a custom script if needed.
