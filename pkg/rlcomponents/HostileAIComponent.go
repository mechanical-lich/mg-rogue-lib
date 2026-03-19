package rlcomponents

import (
	"github.com/mechanical-lich/mlge/ecs"
)

// HostileAIComponent causes an entity to pursue and attack targets within sight range.
type HostileAIComponent struct {
	SightRange int
	TargetX    int
	TargetY    int
	Path       []int // cached path as flat tile indices
}

func (pc HostileAIComponent) GetType() ecs.ComponentType {
	return HostileAI
}
