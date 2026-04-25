package rlcomponents

import "github.com/mechanical-lich/mlge/ecs"

// DamageDealing is optionally implemented by conditions that deal damage each
// tick. ActiveConditionsComponent calls this on every live condition.
type DamageDealing interface {
	DealDamage() int
}

// TurnHandler is optionally implemented by conditions that need to run custom
// logic each turn with access to both the entity and the level.
type TurnHandler interface {
	HandleTurn(entity *ecs.Entity, levelData any)
}

// DeathHandler is optionally implemented by conditions that need to react when
// the host entity dies. FireDeath should be called by the game's cleanup system
// on the first frame the entity gains DeadComponent.
type DeathHandler interface {
	OnDeath(entity *ecs.Entity, levelData any)
}

// ActiveConditionsComponent holds multiple decaying conditions on a single
// entity, working around the ECS one-component-per-type constraint. It allows
// an entity to have any number of DamageConditionComponent,
// StatConditionComponent, or other DecayingComponent instances simultaneously.
//
// Usage:
//
//	acc := ecs.GetOrAdd[*ActiveConditionsComponent](entity, ActiveConditions,
//	    func() ecs.Component { return &ActiveConditionsComponent{} })
//	acc.Add(&DamageConditionComponent{Name: "poison", Duration: 5, DamageDice: "1d4"})
type ActiveConditionsComponent struct {
	// Items is the list of active conditions. Read-only outside this package
	// except for UI display.
	Items []DecayingComponent
}

func (c *ActiveConditionsComponent) GetType() ecs.ComponentType {
	return ActiveConditions
}

// Add appends d to the condition list. If a condition with the same name
// already exists (via GetConditionName), its duration is extended to match d's
// duration when d's duration is longer; otherwise the existing condition is
// kept and d is discarded. Unnamed conditions always stack.
func (c *ActiveConditionsComponent) Add(d DecayingComponent) {
	newName := getConditionName(d)
	if newName != "" {
		for _, existing := range c.Items {
			if getConditionName(existing) == newName {
				refreshIfLonger(existing, d)
				return
			}
		}
	}
	c.Items = append(c.Items, d)
}

// Tick processes all active conditions for one game turn:
//  1. Calls ApplyOnce on any ConditionModifier (idempotent stat application).
//  2. Calls applyDmg for any DamageDealing condition.
//  3. Calls Decay — if the condition expires, Revert is called on any
//     ConditionModifier and the condition is removed from the list.
//
// applyDmg may be nil if no damage routing is desired.
func (c *ActiveConditionsComponent) Tick(entity *ecs.Entity, levelData any, applyDmg func(*ecs.Entity, int)) {
	var remaining []DecayingComponent
	for _, d := range c.Items {
		if cm, ok := d.(ConditionModifier); ok {
			cm.ApplyOnce(entity)
		}

		if applyDmg != nil {
			if dd, ok := d.(DamageDealing); ok {
				if dmg := dd.DealDamage(); dmg > 0 {
					applyDmg(entity, dmg)
				}
			}
		}

		if th, ok := d.(TurnHandler); ok {
			th.HandleTurn(entity, levelData)
		}

		if d.Decay() {
			if cm, ok := d.(ConditionModifier); ok {
				cm.Revert(entity)
			}
		} else {
			remaining = append(remaining, d)
		}
	}
	c.Items = remaining
}

// FireDeath calls OnDeath on every condition that implements DeathHandler.
// Call this from the game's death/cleanup system on the first frame an entity
// receives DeadComponent, before conditions are removed.
func (c *ActiveConditionsComponent) FireDeath(entity *ecs.Entity, levelData any) {
	for _, d := range c.Items {
		if dh, ok := d.(DeathHandler); ok {
			dh.OnDeath(entity, levelData)
		}
	}
}

// GetOrCreateActiveConditions returns the entity's ActiveConditionsComponent,
// creating and attaching one if it does not yet exist.
func GetOrCreateActiveConditions(entity *ecs.Entity) *ActiveConditionsComponent {
	if entity.HasComponent(ActiveConditions) {
		return entity.GetComponent(ActiveConditions).(*ActiveConditionsComponent)
	}
	acc := &ActiveConditionsComponent{}
	entity.AddComponent(acc)
	return acc
}

// ---- private helpers --------------------------------------------------------

// conditionNamed is satisfied by conditions that have a display name.
type conditionNamed interface {
	GetConditionName() string
}

// durationAccessor is satisfied within this package by DamageConditionComponent
// and StatConditionComponent via their unexported getDuration/setDuration methods.
type durationAccessor interface {
	getDuration() int
	setDuration(int)
}

func getConditionName(d DecayingComponent) string {
	if n, ok := d.(conditionNamed); ok {
		return n.GetConditionName()
	}
	return ""
}

// refreshIfLonger updates existing's duration to match incoming's when incoming
// has a longer remaining duration. Both must implement durationAccessor.
func refreshIfLonger(existing, incoming DecayingComponent) {
	e, eok := existing.(durationAccessor)
	i, iok := incoming.(durationAccessor)
	if eok && iok && i.getDuration() > e.getDuration() {
		e.setDuration(i.getDuration())
	}
}
