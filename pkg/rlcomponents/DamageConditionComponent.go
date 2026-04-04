package rlcomponents

import (
	"log"

	"github.com/mechanical-lich/mlge/dice"
	"github.com/mechanical-lich/mlge/ecs"
)

// DamageConditionComponent is a decaying status effect that deals damage each
// turn. Damage is expressed as a dice string (e.g. "1d6", "2d4+1", "3").
// DamageType is informational (e.g. "poison", "fire") and may be used by
// callers for display or future resistance checks.
//
// Example — 1d4 poison damage for 6 turns:
//
//	&DamageConditionComponent{
//	    Name:       "Venom",
//	    Duration:   6,
//	    DamageDice: "1d4",
//	    DamageType: "poison",
//	}
type DamageConditionComponent struct {
	Name       string
	Duration   int
	DamageDice string
	DamageType string
}

func (c *DamageConditionComponent) GetType() ecs.ComponentType {
	return DamageCondition
}

func (c *DamageConditionComponent) Decay() bool {
	c.Duration--
	return c.Duration <= 0
}

// GetConditionName returns the condition's display name, satisfying the named
// interface used by ActiveConditionsComponent for deduplication.
func (c *DamageConditionComponent) GetConditionName() string { return c.Name }

// DealDamage returns the damage for one tick (delegates to Roll).
// It satisfies the DamageDealing interface consumed by ActiveConditionsComponent.
func (c *DamageConditionComponent) DealDamage() int { return c.Roll() }

func (c *DamageConditionComponent) getDuration() int      { return c.Duration }
func (c *DamageConditionComponent) setDuration(d int)     { c.Duration = d }

// Roll returns the damage for one tick, evaluated from DamageDice.
// Returns 1 on any parse error.
func (c *DamageConditionComponent) Roll() int {
	result, err := dice.Roll(c.DamageDice)
	if err != nil {
		log.Printf("DamageConditionComponent: invalid dice %q: %v", c.DamageDice, err)
		return 1
	}
	if result < 1 {
		return 1
	}
	return result
}
