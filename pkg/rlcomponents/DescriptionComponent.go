package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// DescriptionComponent holds an entity's display name, faction, and optional
// narrative text used by the interaction and announcement systems.
type DescriptionComponent struct {
	Name            string
	Faction         string
	ID              string   // Optional unique identifier used by the interaction system to reference this entity.
	Tags            []string // Optional group tags (e.g. "airlock", "sector_b") for multi-target triggers.
	LongDescription string   // Shown when the player examines the entity.

	PassOverDescription   []string // Randomly chosen when a configured entity passes over this one.
	DeathAnnouncements    []string // Randomly chosen when a watcher sees this entity die. Defaults to "<Name> has died."
	ExcuseMeAnnouncements []string // Randomly chosen when a friendly bumps into this entity and they swap.
}

func (pc DescriptionComponent) GetType() ecs.ComponentType {
	return Description
}
