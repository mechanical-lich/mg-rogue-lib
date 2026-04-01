package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// DecayingComponent is implemented by status effects that expire over time.
// Decay returns true when the effect should be removed.
type DecayingComponent interface {
	Decay() bool
	GetType() ecs.ComponentType
}

// SpeedModifier is optionally implemented by status effects that modify an
// entity's Speed. ApplyOnce applies the effect the first time it is called
// (idempotent); Revert undoes the effect when the status expires.
type SpeedModifier interface {
	ApplyOnce(entity *ecs.Entity)
	Revert(entity *ecs.Entity)
}
