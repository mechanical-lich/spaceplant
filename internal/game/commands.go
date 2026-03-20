package game

import "github.com/mechanical-lich/mlge/transport"

// CmdAction is the command type sent from the client when the player presses a key.
const CmdAction transport.CommandType = "sp.action"

// ActionPayload carries the key string that was pressed (matches PlayerComponent command strings).
type ActionPayload struct {
	Key string
}
