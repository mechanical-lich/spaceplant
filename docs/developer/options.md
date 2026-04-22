# Options System

The options screen is available from the **title screen** and the **pause menu**. It edits a subset of `config.Global()` in-place and writes the result to `data/config.json`.

## Configurable fields

| Config field | JSON key | Type | Description |
|---|---|---|---|
| `CRTIntensity` | `crtIntensity` | `float64` | CRT shader blend (0 = off, 1 = full) |
| `PressDelay` | `pressDelay` | `int` | Key-repeat delay in ticks |
| `RenderScale` | `renderScale` | `float64` | Map viewport zoom |
| `NpcTurnDelayTicks` | `npcTurnDelayTicks` | `int` | Server ticks between NPC-only turns |

## Adding a new option

1. Add the field to `Config` in `internal/config/config.go` with a JSON tag.
2. Add a default value to `data/config.json`.
3. Add a `TextInput` row in `newOptionsModal` (`internal/game/options_modal.go`) following the existing pattern.
4. Parse and apply the value in `OptionsModal.apply`.

## Persistence

`config.Save()` (`internal/config/config.go`) marshals the singleton `*Config` to `data/config.json` with `json.MarshalIndent`. Changes are applied to the in-memory singleton immediately and persist to disk on save.
