package keybindings

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Bindings maps action names to key combos and provides runtime lookup.
type Bindings struct {
	raw     map[string]string // action -> combo ("move_north" -> "w", "aimed_shot" -> "shift+f")
	inverse map[string]string // normalized combo -> action
}

var global *Bindings

// Global returns the singleton Bindings, loading from data/keybindings.json on first call.
func Global() *Bindings {
	if global == nil {
		b, err := load("data/keybindings.json")
		if err != nil {
			log.Printf("keybindings: using defaults (%v)", err)
			b = buildDefaults()
		}
		global = b
	}
	return global
}

// ActionFor returns the action name for a pressed key, or "" if unbound.
// keyStr should be the raw ebiten key string (e.g. "W", "Period").
func (b *Bindings) ActionFor(keyStr string, shift bool) string {
	combo := normalizeKey(keyStr)
	if shift {
		combo = "shift+" + combo
	}
	return b.inverse[combo]
}

// IsJustPressed reports whether the key bound to action was just pressed this frame,
// including any required modifier state.
func (b *Bindings) IsJustPressed(action string) bool {
	combo := b.raw[action]
	if combo == "" {
		return false
	}
	shift := strings.HasPrefix(combo, "shift+")
	keyStr := strings.TrimPrefix(combo, "shift+")
	k, ok := parseKey(keyStr)
	if !ok {
		return false
	}
	shiftHeld := ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight)
	return shift == shiftHeld && inpututil.IsKeyJustPressed(k)
}

// KeyCombo returns the display string for an action (e.g. "shift+f"), or "" if unbound.
func (b *Bindings) KeyCombo(action string) string {
	return b.raw[action]
}

// MergeDefaults adds entries from defaults that are not already bound.
// Use this to register skill action IDs after skills are loaded.
func (b *Bindings) MergeDefaults(defaults map[string]string) {
	for action, combo := range defaults {
		if _, exists := b.raw[action]; !exists {
			b.raw[action] = combo
			b.inverse[strings.ToLower(combo)] = action
		}
	}
}

func load(path string) (*Bindings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return build(raw), nil
}

func buildDefaults() *Bindings {
	return build(map[string]string{
		"move_north":        "w",
		"move_south":        "s",
		"move_west":         "a",
		"move_east":         "d",
		"face_north":        "shift+w",
		"face_south":        "shift+s",
		"face_west":         "shift+a",
		"face_east":         "shift+d",
		"fire":              "f",
		"burst_fire":        "g",
		"heal":              "h",
		"stairs":            ".",
		"equip":             "e",
		"pickup":            "p",
		"rush":              "r",
		"reload":            "shift+r",
		"aimed_shot":        "shift+f",
		"class_upgrade":     "shift+c",
		"inventory":         "i",
		"character_overview": "shift+i",
	})
}

func build(raw map[string]string) *Bindings {
	inv := make(map[string]string, len(raw))
	for action, combo := range raw {
		inv[strings.ToLower(combo)] = action
	}
	return &Bindings{raw: raw, inverse: inv}
}

// normalizeKey converts an ebiten key string to the lowercase form used in keybindings.json.
func normalizeKey(keyStr string) string {
	s := strings.ToLower(keyStr)
	if s == "period" {
		return "."
	}
	return s
}

// parseKey converts a keybindings.json key string to an ebiten.Key.
func parseKey(s string) (ebiten.Key, bool) {
	k, ok := keyMap[strings.ToLower(s)]
	return k, ok
}

var keyMap = map[string]ebiten.Key{
	"a": ebiten.KeyA, "b": ebiten.KeyB, "c": ebiten.KeyC, "d": ebiten.KeyD,
	"e": ebiten.KeyE, "f": ebiten.KeyF, "g": ebiten.KeyG, "h": ebiten.KeyH,
	"i": ebiten.KeyI, "j": ebiten.KeyJ, "k": ebiten.KeyK, "l": ebiten.KeyL,
	"m": ebiten.KeyM, "n": ebiten.KeyN, "o": ebiten.KeyO, "p": ebiten.KeyP,
	"q": ebiten.KeyQ, "r": ebiten.KeyR, "s": ebiten.KeyS, "t": ebiten.KeyT,
	"u": ebiten.KeyU, "v": ebiten.KeyV, "w": ebiten.KeyW, "x": ebiten.KeyX,
	"y": ebiten.KeyY, "z": ebiten.KeyZ,
	".": ebiten.KeyPeriod,
}
