package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// DescriptionComponent holds an entity's display name and faction affiliation.
type DescriptionComponent struct {
	Name                string
	Faction             string
	LongDescription     string   // A longer description that can be shown when the player examines the entity.
	PassOverDescription   []string // Optional descriptions that can be randomly chosen when the player passes over the entity.
	DeathAnnouncements    []string // An optional message that can be displayed when the entity dies.
	ExcuseMeAnnouncements []string // Optional messages emitted when a friendly entity bumps into this one and they swap positions.
}

func (pc DescriptionComponent) GetType() ecs.ComponentType {
	return Description
}
