package game

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/transport"
)

// CmdAction is the command type sent from the client when the player presses a key.
const CmdAction transport.CommandType = "sp.action"

// ActionPayload carries the key string that was pressed (matches PlayerComponent command strings).
type ActionPayload struct {
	Key string
}

// CmdReload is the command type sent from the reload modal when the player confirms a reload.
const CmdReload transport.CommandType = "sp.reload"

// ReloadPayload carries the weapon and ammo entity pointers selected in the reload modal.
type ReloadPayload struct {
	WeaponItem *ecs.Entity
	AmmoItem   *ecs.Entity
}

// CmdAimedShot is sent from the targeted aimed shot modal when the player picks a body part.
const CmdAimedShot transport.CommandType = "sp.aimed_shot"

// AimedShotPayload carries the body part name chosen by the player.
type AimedShotPayload struct {
	BodyPart string
}
