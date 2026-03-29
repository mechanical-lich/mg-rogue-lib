package rlcomponents

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/event"
)

// ExcuseMeEventType is the event key registered with mlge's queued event manager.
const ExcuseMeEventType event.EventType = "rl.excuseme"

// ExcuseMeEvent is posted when two friendly entities swap positions.
// Register a listener on ExcuseMeEventType to log a message only when the
// player is involved.
type ExcuseMeEvent struct {
	Mover   *ecs.Entity // The entity that initiated the swap.
	Bumped  *ecs.Entity // The entity that was bumped and said the announcement.
	Message string      // Pre-selected ExcuseMeAnnouncement string.
}

func (e ExcuseMeEvent) GetType() event.EventType {
	return ExcuseMeEventType
}
