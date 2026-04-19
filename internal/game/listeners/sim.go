package listeners

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/spaceplant/internal/wincondition"
)

// SimAccess is the minimal interface the listeners need from the simulation.
// *game.SimWorld implements this; the interface breaks the import cycle between
// the listeners sub-package and the game package.
type SimAccess interface {
	GetPlayer() *ecs.Entity
	GetRLLevel() *rlworld.Level
	BuildEvalContext() wincondition.EvalContext
}
