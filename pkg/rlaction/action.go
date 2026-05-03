// Package rlaction defines the core turn-based action interface.
package rlaction

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
)

// Action represents a single entity action with its associated energy cost.
// L is the level type used by the game (e.g. *world.Level in spaceplant).
// Execute applies the action and must call rlenergy.SetActionCost internally.
type Action[L any] interface {
	// Cost returns the expected energy cost, used to determine affordability.
	Cost(entity *ecs.Entity, level *L) int
	// Available returns true if this action can be performed right now.
	Available(entity *ecs.Entity, level *L) bool
	// Execute performs the action. It must call rlenergy.SetActionCost.
	Execute(entity *ecs.Entity, level *L) error
}

// HasAvailableAction returns true if at least one action in the list is both
// available and affordable given the entity's current energy.
func HasAvailableAction[L any](entity *ecs.Entity, level *L, actions []Action[L]) bool {
	if !entity.HasComponent(rlcomponents.Energy) {
		return false
	}
	ec := entity.GetComponent(rlcomponents.Energy).(*rlcomponents.EnergyComponent)
	for _, a := range actions {
		if a.Available(entity, level) && ec.Energy >= a.Cost(entity, level) {
			return true
		}
	}
	return false
}
