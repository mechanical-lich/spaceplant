package system

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"

	"github.com/mechanical-lich/mechanical-basic/pkg/basic"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/generation"
	"github.com/mechanical-lich/spaceplant/internal/skill"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// setupContext is the runtime state for a scenario setup script.
type setupContext struct {
	Level        *world.Level
	FloorResults []generation.FloorResult
	LastSpawned  *component.DescriptionComponent // description of the last entity spawned by the script
	LastItem     *component.ItemComponent        // item component of the last entity spawned
}

// RunSetupScripts executes each script path in order with a shared generation
// context. Scripts run once at station creation; they are not attached to any entity.
func RunSetupScripts(scripts []string, l *world.Level, results []generation.FloorResult) {
	ctx := &setupContext{Level: l, FloorResults: results}
	for _, path := range scripts {
		if err := runSetupScript(path, ctx); err != nil {
			log.Printf("[SetupScript] %s: %v", path, err)
		}
	}
}

func runSetupScript(path string, ctx *setupContext) error {
	code, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}

	interp := basic.NewMechanicalBasic()
	registerSetupFuncs(interp, ctx)

	if err := interp.Load(string(code)); err != nil {
		return fmt.Errorf("load: %w", err)
	}

	if !interp.HasFunction("on_setup") {
		return fmt.Errorf("setup script %s has no on_setup function", path)
	}
	if _, err := interp.Call("on_setup"); err != nil {
		return fmt.Errorf("on_setup: %w", err)
	}
	return nil
}

func registerSetupFuncs(interp *basic.MechBasic, ctx *setupContext) {
	// --- Floor info ---

	interp.RegisterFunc("get_floor_count", func(args ...any) (any, error) {
		return float64(len(ctx.FloorResults)), nil
	})

	interp.RegisterFunc("get_floor_name", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("get_floor_name: expected 1 argument (z)")
		}
		z := toInt(args[0])
		if z < 0 || z >= len(ctx.FloorResults) {
			return "", nil
		}
		if ctx.FloorResults[z].Theme == nil {
			return "", nil
		}
		return ctx.FloorResults[z].Theme.Name, nil
	})

	// --- Room info ---

	interp.RegisterFunc("get_room_count", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("get_room_count: expected 1 argument (z)")
		}
		z := toInt(args[0])
		if z < 0 || z >= len(ctx.FloorResults) {
			return float64(0), nil
		}
		return float64(len(ctx.FloorResults[z].Rooms)), nil
	})

	interp.RegisterFunc("get_room_number", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("get_room_number: expected 2 arguments (z, room_index)")
		}
		z, idx := toInt(args[0]), toInt(args[1])
		if z < 0 || z >= len(ctx.FloorResults) {
			return float64(0), nil
		}
		rooms := ctx.FloorResults[z].Rooms
		if idx < 0 || idx >= len(rooms) {
			return float64(0), nil
		}
		return float64(rooms[idx].Number), nil
	})

	interp.RegisterFunc("get_room_tag", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("get_room_tag: expected 2 arguments (z, room_index)")
		}
		z, idx := toInt(args[0]), toInt(args[1])
		if z < 0 || z >= len(ctx.FloorResults) {
			return "", nil
		}
		rooms := ctx.FloorResults[z].Rooms
		if idx < 0 || idx >= len(rooms) {
			return "", nil
		}
		return rooms[idx].Tag, nil
	})

	// find_room_by_tag(tag, z) → room index, or -1 if not found
	interp.RegisterFunc("find_room_by_tag", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("find_room_by_tag: expected 2 arguments (tag, z)")
		}
		tag, ok := args[0].(string)
		if !ok {
			return nil, errors.New("find_room_by_tag: tag must be a string")
		}
		z := toInt(args[1])
		if z < 0 || z >= len(ctx.FloorResults) {
			return float64(-1), nil
		}
		for i, r := range ctx.FloorResults[z].Rooms {
			if r.Tag == tag {
				return float64(i), nil
			}
		}
		return float64(-1), nil
	})

	// random_room(z) → random room index on floor z, or -1 if no rooms
	interp.RegisterFunc("random_room", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("random_room: expected 1 argument (z)")
		}
		z := toInt(args[0])
		if z < 0 || z >= len(ctx.FloorResults) {
			return float64(-1), nil
		}
		rooms := ctx.FloorResults[z].Rooms
		if len(rooms) == 0 {
			return float64(-1), nil
		}
		return float64(rand.Intn(len(rooms))), nil
	})

	// random_floor_excluding(z) → random floor index != z
	interp.RegisterFunc("random_floor_excluding", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("random_floor_excluding: expected 1 argument (z)")
		}
		exclude := toInt(args[0])
		n := len(ctx.FloorResults)
		if n <= 1 {
			return float64(0), nil
		}
		candidates := make([]int, 0, n-1)
		for i := 0; i < n; i++ {
			if i != exclude {
				candidates = append(candidates, i)
			}
		}
		return float64(candidates[rand.Intn(len(candidates))]), nil
	})

	// --- Spawning ---

	// spawn_in_room(blueprint, z, room_index) → places entity at a random floor tile in the room
	interp.RegisterFunc("spawn_in_room", func(args ...any) (any, error) {
		if len(args) != 3 {
			return nil, errors.New("spawn_in_room: expected 3 arguments (blueprint, z, room_index)")
		}
		blueprint, ok := args[0].(string)
		if !ok {
			return nil, errors.New("spawn_in_room: blueprint must be a string")
		}
		z, idx := toInt(args[1]), toInt(args[2])
		if z < 0 || z >= len(ctx.FloorResults) {
			return nil, fmt.Errorf("spawn_in_room: floor %d out of range", z)
		}
		rooms := ctx.FloorResults[z].Rooms
		if idx < 0 || idx >= len(rooms) {
			return nil, fmt.Errorf("spawn_in_room: room index %d out of range", idx)
		}
		room := rooms[idx]
		ctx.LastSpawned = nil
		ctx.LastItem = nil
		for tries := 0; tries < 80; tries++ {
			w := max(1, room.Width-2)
			h := max(1, room.Height-2)
			x := room.X + 1 + rand.Intn(w)
			y := room.Y + 1 + rand.Intn(h)
			if ctx.Level.GetTileType(x, y, z) != world.TypeFloor {
				continue
			}
			if ctx.Level.Level.GetEntityAt(x, y, z) != nil {
				continue
			}
			e, err := factory.Create(blueprint, x, y)
			if err != nil {
				return nil, fmt.Errorf("spawn_in_room: %w", err)
			}
			e.GetComponent(component.Position).(*component.PositionComponent).SetPosition(x, y, z)
			ctx.Level.AddEntity(e)
			if e.HasComponent(component.Description) {
				ctx.LastSpawned = e.GetComponent(component.Description).(*component.DescriptionComponent)
			}
			if e.HasComponent(component.Item) {
				ctx.LastItem = e.GetComponent(component.Item).(*component.ItemComponent)
			}
			return nil, nil
		}
		return nil, fmt.Errorf("spawn_in_room: could not find empty tile in room %d on floor %d", idx, z)
	})

	// spawn_skill_chip(skill_id, z, room_index) → creates a skill chip for the given skill and places it in the room
	interp.RegisterFunc("spawn_skill_chip", func(args ...any) (any, error) {
		if len(args) != 3 {
			return nil, errors.New("spawn_skill_chip: expected 3 arguments (skill_id, z, room_index)")
		}
		skillID, ok := args[0].(string)
		if !ok {
			return nil, errors.New("spawn_skill_chip: skill_id must be a string")
		}
		z, idx := toInt(args[1]), toInt(args[2])
		if z < 0 || z >= len(ctx.FloorResults) {
			return nil, fmt.Errorf("spawn_skill_chip: floor %d out of range", z)
		}
		rooms := ctx.FloorResults[z].Rooms
		if idx < 0 || idx >= len(rooms) {
			return nil, fmt.Errorf("spawn_skill_chip: room index %d out of range", idx)
		}
		room := rooms[idx]
		for tries := 0; tries < 80; tries++ {
			w := max(1, room.Width-2)
			h := max(1, room.Height-2)
			x := room.X + 1 + rand.Intn(w)
			y := room.Y + 1 + rand.Intn(h)
			if ctx.Level.GetTileType(x, y, z) != world.TypeFloor {
				continue
			}
			if ctx.Level.Level.GetEntityAt(x, y, z) != nil {
				continue
			}
			e, err := factory.Create("skill_chip", x, y)
			if err != nil {
				return nil, fmt.Errorf("spawn_skill_chip: %w", err)
			}
			e.GetComponent(component.Position).(*component.PositionComponent).SetPosition(x, y, z)
			if e.HasComponent(component.SkillChip) {
				e.GetComponent(component.SkillChip).(*component.SkillChipComponent).SkillId = skillID
			}
			// Patch display name so the player sees the actual skill name on the item.
			displayName := "Skill Chip: " + skillID
			if sd := skill.Get(skillID); sd != nil {
				displayName = "Skill Chip: " + sd.Name
			}
			if e.HasComponent(component.Description) {
				dc := e.GetComponent(component.Description).(*component.DescriptionComponent)
				dc.Name = displayName
			}
			if e.HasComponent(component.Item) {
				ic := e.GetComponent(component.Item).(*component.ItemComponent)
				ic.Name = displayName
				ic.Description = "A neural skill chip. Consume it to learn " + displayName[len("Skill Chip: "):] + "."
			}
			ctx.Level.AddEntity(e)
			return nil, nil
		}
		return nil, fmt.Errorf("spawn_skill_chip: could not find empty tile in room %d on floor %d", idx, z)
	})

	// --- Last-spawned description override ---

	// set_description(text) — overwrite LongDescription and ItemComponent.Description on last spawned entity
	interp.RegisterFunc("set_description", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("set_description: expected 1 argument (text)")
		}
		text, ok := args[0].(string)
		if !ok {
			return nil, errors.New("set_description: argument must be a string")
		}
		if ctx.LastSpawned != nil {
			ctx.LastSpawned.LongDescription = text
		}
		if ctx.LastItem != nil {
			ctx.LastItem.Description = text
		}
		return nil, nil
	})

	// --- Misc ---

	interp.RegisterFunc("chr", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("chr: expected 1 argument (ascii code)")
		}
		return string(rune(toInt(args[0]))), nil
	})

	interp.RegisterFunc("num_to_str", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("num_to_str: expected 1 argument")
		}
		return fmt.Sprintf("%v", toInt(args[0])), nil
	})

	interp.RegisterFunc("add_message", func(args ...any) (any, error) {
		if len(args) != 1 {
			return nil, errors.New("add_message: expected 1 argument")
		}
		text, ok := args[0].(string)
		if !ok {
			return nil, errors.New("add_message: argument must be a string")
		}
		message.AddMessage(text)
		return nil, nil
	})

	interp.RegisterFunc("set_flag", func(args ...any) (any, error) {
		if len(args) != 2 {
			return nil, errors.New("set_flag: expected 2 arguments (key, value)")
		}
		key, ok := args[0].(string)
		if !ok {
			return nil, errors.New("set_flag: first argument must be a string")
		}
		ctx.Level.Flags[key] = args[1]
		return nil, nil
	})
}
