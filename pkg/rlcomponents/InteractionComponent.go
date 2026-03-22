package rlcomponents

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/event"
)

// Trigger describes a single effect fired when an InteractionComponent activates.
// Type identifies the handler (e.g. "unlock_door", "post_message", "quest_flag").
// Params carries arbitrary string key-value data the handler needs.
type Trigger struct {
	Type   string            `json:"Type"`
	Params map[string]string `json:"Params"`
}

// InteractionComponent makes an entity interactable when bumped by an actor.
// Each activation fires one InteractionEvent per Trigger in order.
// If Repeatable is false the component is locked after the first use.
type InteractionComponent struct {
	Prompt     string    // Message shown when the player bumps into this entity.
	Triggers   []Trigger // Effects fired on activation, in order.
	Repeatable bool      // If false, the component locks after first use.
	Used       bool      // Set true after first use. Not serialised from JSON.
}

func (c *InteractionComponent) GetType() ecs.ComponentType {
	return Interaction
}

// --- InteractionEvent ---

// InteractionEventType is the event key registered with mlge's queued event manager.
const InteractionEventType event.EventType = "rl.interaction"

// InteractionEvent is posted once per Trigger when an InteractionComponent activates.
// Register a listener on InteractionEventType to handle specific trigger types.
type InteractionEvent struct {
	Actor   *ecs.Entity // The entity that triggered the interaction (e.g. the player).
	Target  *ecs.Entity // The entity carrying the InteractionComponent.
	Trigger Trigger     // The specific trigger being fired.
}

func (e InteractionEvent) GetType() event.EventType {
	return InteractionEventType
}
