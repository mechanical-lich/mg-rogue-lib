package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

type KeyComponent struct {
	KeyID string
}

func (kc KeyComponent) GetType() ecs.ComponentType {
	return Key
}
