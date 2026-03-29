package rlcomponents

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/event"
)

// DeathEventType is the event key registered with mlge's queued event manager.
const DeathEventType event.EventType = "rl.death"

// DeathEvent is posted when an entity dies and a watcher has line-of-sight to it.
// The LOS check is performed before queuing, so listeners can log unconditionally
// after confirming the watcher is the player.
type DeathEvent struct {
	Watcher *ecs.Entity // The entity observing the death (e.g. the player).
	Dying   *ecs.Entity // The entity that died.
	Message string      // Pre-selected death announcement string.
	X, Y, Z int         // World position of the dying entity.
}

func (e DeathEvent) GetType() event.EventType {
	return DeathEventType
}
