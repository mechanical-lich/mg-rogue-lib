package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// SlowedComponent is a decaying status that halves the entity's Speed while active.
type SlowedComponent struct {
	Duration      int
	originalSpeed int
	applied       bool
}

func (s *SlowedComponent) GetType() ecs.ComponentType {
	return Slowed
}

func (s *SlowedComponent) Decay() bool {
	s.Duration--
	return s.Duration <= 0
}

// ApplyOnce halves the entity's Speed the first time it is called.
func (s *SlowedComponent) ApplyOnce(entity *ecs.Entity) {
	if s.applied || !entity.HasComponent(Energy) {
		return
	}
	ec := entity.GetComponent(Energy).(*EnergyComponent)
	s.originalSpeed = ec.Speed
	ec.Speed /= 2
	if ec.Speed < 1 {
		ec.Speed = 1
	}
	s.applied = true
}

// Revert restores the entity's Speed to its value before Slowed was applied.
func (s *SlowedComponent) Revert(entity *ecs.Entity) {
	if !s.applied || !entity.HasComponent(Energy) {
		return
	}
	ec := entity.GetComponent(Energy).(*EnergyComponent)
	ec.Speed = s.originalSpeed
}
