package rlbodycombat

import (
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/event"
)

const CombatEventType event.EventType = "CombatEvent"

// CombatEvent is posted whenever an attack resolves in v2 combat.
// GUIs can listen for this event to display visual effects (floating damage
// numbers, hit animations, etc.) near the world location of the attack.
// For saving throws, Attacker is nil.
type CombatEvent struct {
	// World position of the attacker (or defender for saving throws).
	X, Y, Z int

	Attacker *ecs.Entity // nil for saving throws
	Defender *ecs.Entity

	AttackerName string
	DefenderName string

	// Source is the weapon, skill, or ability that caused the attack or save.
	// e.g. "laser trimmers", "fist", "poisonous bite", "flamethrower".
	Source string

	// Damage dealt. Zero means the attack missed.
	Damage     int
	DamageType string

	// BodyPart is the name of the body part that was hit.
	// Empty when the attack missed or when the defender has no BodyComponent.
	BodyPart string

	Miss      bool
	Crit      bool
	Broken    bool
	Amputated bool
	SaveFail  bool
	SavePass  bool
}

func (e CombatEvent) GetType() event.EventType {
	return CombatEventType
}
