# Space Plant

Dead space inspired roguelike built using EbitenEngine, MLGE, and ml-rogue-lib.

## Running from source:

Graphical:
`go run cmd/game/main.go`

Terminal:
`go run cmd/terminal/main.go`

## Controls

### Movement
| Key | Action |
|-----|--------|
| W | Move north |
| A | Move west |
| S | Move south |
| D | Move east |

### Actions
| Key | Action |
|-----|--------|
| H | Use healing item |
| E | Auto-equip first equippable item |
| P | Pick up item at current tile |
| . | Use stairs |
| F | Shoot (follow with a direction key) |

### Skills (when skill is equipped)
| Key | Skill Required | Action |
|-----|----------------|--------|
| K | Roundhouse Kick | Strike all 8 adjacent enemies |
| R | Martial Artist | Roundhouse kick (alt binding) |
| V | Shove | Push entity in front one tile |
| C | Flamethrower / Acid Spray | Fire a cone projectile |

### UI
| Key | Action |
|-----|--------|
| I | Open inventory tab |
| Shift+I | Open stats overview tab |
| Shift+C | Open class upgrade modal |
| R | Toggle rush mode (2× speed, no turn cost) |
| Escape | Close top modal |


