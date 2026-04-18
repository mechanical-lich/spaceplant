# Space Plants!

Dead Space–inspired roguelike built using EbitenEngine, MLGE, and ml-rogue-lib.

## Running from source

**Graphical (Ebiten):**
```
go run cmd/game/main.go
```

**Terminal (tcell):**
```
go run cmd/terminal/main.go
```

## Documentation

- [Player Guide](docs/player/README.md) — controls, stations, combat, classes, skills, and items
- [Developer Guide](docs/developer/README.md) — architecture, adding actions/skills, save system

## Quick Controls Reference

### Movement
| Key | Action |
|-----|--------|
| W / A / S / D | Move north / west / south / east |
| Shift+W/A/S/D | Rotate facing without moving |

### Combat
| Key | Action |
|-----|--------|
| F | Snap shot — fire in facing direction |
| Shift+F | Aimed shot — choose a body part to target |
| G | Burst fire |
| H | Use healing item |
| Move into enemy | Melee attack |

### Inventory & Equipment
| Key | Action |
|-----|--------|
| E | Auto-equip from bag |
| P | Open nearby items (pick up / equip within 1 tile) |
| I | Open inventory |
| Shift+I | Open character overview |
| Shift+R | Open reload modal |

### UI
| Key | Action |
|-----|--------|
| R | Toggle rush mode |
| . | Use stairs |
| Shift+C | Open class upgrade screen |
| Escape | Close current modal / open pause menu |

