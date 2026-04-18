package system

import (
	"errors"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/message"
	"github.com/mechanical-lich/spaceplant/internal/component"
	"github.com/mechanical-lich/spaceplant/internal/factory"
	"github.com/mechanical-lich/spaceplant/internal/world"
)

// MotherPlantSeedSystem watches mobile_mother_plant entities. Each turn they
// take, TurnsLeft decrements. At zero the seed is killed and a full
// mother_plant is spawned in its place.
type MotherPlantSeedSystem struct{}

func (s MotherPlantSeedSystem) Requires() []ecs.ComponentType {
	return []ecs.ComponentType{component.MotherPlantSeed}
}

func (s MotherPlantSeedSystem) UpdateSystem(_ any) error { return nil }

func (s MotherPlantSeedSystem) UpdateEntity(data any, entity *ecs.Entity) error {
	if entity.HasComponent(rlcomponents.Dead) {
		return nil
	}
	// Only decrement once per turn.
	if !entity.HasComponent(rlcomponents.TurnTaken) {
		return nil
	}

	l, ok := data.(*world.Level)
	if !ok {
		return errors.New("MotherPlantSeedSystem: invalid level type")
	}

	seed := entity.GetComponent(component.MotherPlantSeed).(*component.MotherPlantSeedComponent)
	seed.TurnsLeft--
	if seed.TurnsLeft > 0 {
		return nil
	}

	pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	x, y, z := pc.GetX(), pc.GetY(), pc.GetZ()

	entity.AddComponent(&rlcomponents.DeadComponent{})

	large, err := factory.Create("mother_plant", x, y)
	if err != nil {
		return nil
	}
	large.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent).SetPosition(x, y, z)
	l.AddEntity(large)
	message.AddMessage("The cutting convulses and erupts — the mother plant has taken root!")
	return nil
}
