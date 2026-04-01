package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// HasteComponent is a decaying status that doubles the entity's Speed while active.
type HasteComponent struct {
	Duration      int
	originalSpeed int
	applied       bool
}

func (h *HasteComponent) GetType() ecs.ComponentType {
	return Haste
}

func (h *HasteComponent) Decay() bool {
	h.Duration--
	return h.Duration <= 0
}

// ApplyOnce doubles the entity's Speed the first time it is called.
func (h *HasteComponent) ApplyOnce(entity *ecs.Entity) {
	if h.applied || !entity.HasComponent(Energy) {
		return
	}
	ec := entity.GetComponent(Energy).(*EnergyComponent)
	h.originalSpeed = ec.Speed
	ec.Speed *= 2
	h.applied = true
}

// Revert restores the entity's Speed to its value before Haste was applied.
func (h *HasteComponent) Revert(entity *ecs.Entity) {
	if !h.applied || !entity.HasComponent(Energy) {
		return
	}
	ec := entity.GetComponent(Energy).(*EnergyComponent)
	ec.Speed = h.originalSpeed
}
