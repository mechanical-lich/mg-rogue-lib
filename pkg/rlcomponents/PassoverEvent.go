package rlcomponents

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/event"
)

// PassoverEventType is the event key registered with mlge's queued event manager.
const PassoverEventType event.EventType = "rl.passover"

// PassoverEvent is posted when an entity moves onto a tile occupied by another
// entity that has a PassOverDescription. Register a listener on PassoverEventType
// to format and filter messages (e.g. only log when the mover is the player).
type PassoverEvent struct {
	Mover     *ecs.Entity // The entity that moved onto the tile.
	SteppedOn *ecs.Entity // The entity being walked over.
	Message   string      // Pre-selected passover description string.
	X, Y, Z   int         // World position of the event.
}

func (e PassoverEvent) GetType() event.EventType {
	return PassoverEventType
}
