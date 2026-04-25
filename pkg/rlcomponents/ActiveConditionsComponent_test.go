package rlcomponents

import (
	"testing"

	"github.com/mechanical-lich/mlge/ecs"
	"github.com/stretchr/testify/assert"
)

// ---- helpers ----------------------------------------------------------------

func newTestEntity() *ecs.Entity {
	e := &ecs.Entity{}
	e.AddComponent(&StatsComponent{AC: 10, Str: 10, Con: 10})
	return e
}

func newTestEntityWithHealth(hp int) *ecs.Entity {
	e := &ecs.Entity{}
	e.AddComponent(&HealthComponent{Health: hp, MaxHealth: hp})
	return e
}

func damageApplied(entity *ecs.Entity) func(*ecs.Entity, int) {
	// Returns a damage function that deducts from HealthComponent.
	return func(e *ecs.Entity, dmg int) {
		if e.HasComponent(Health) {
			hc := e.GetComponent(Health).(*HealthComponent)
			hc.Health -= dmg
		}
	}
}

// ---- GetOrCreateActiveConditions --------------------------------------------

func TestGetOrCreate_CreatesWhenAbsent(t *testing.T) {
	e := &ecs.Entity{}
	acc := GetOrCreateActiveConditions(e)
	assert.NotNil(t, acc)
	assert.True(t, e.HasComponent(ActiveConditions))
}

func TestGetOrCreate_ReturnsSameInstance(t *testing.T) {
	e := &ecs.Entity{}
	a := GetOrCreateActiveConditions(e)
	b := GetOrCreateActiveConditions(e)
	assert.Same(t, a, b)
}

// ---- Add / stacking ---------------------------------------------------------

func TestAdd_AppendsUnnamedConditions(t *testing.T) {
	acc := &ActiveConditionsComponent{}
	acc.Add(&DamageConditionComponent{Duration: 3, DamageDice: "1"})
	acc.Add(&DamageConditionComponent{Duration: 3, DamageDice: "1"})
	assert.Len(t, acc.Items, 2, "unnamed conditions always stack")
}

func TestAdd_DeduplicatesNamedConditions(t *testing.T) {
	acc := &ActiveConditionsComponent{}
	acc.Add(&DamageConditionComponent{Name: "poison", Duration: 3, DamageDice: "1"})
	acc.Add(&DamageConditionComponent{Name: "poison", Duration: 5, DamageDice: "1"})
	assert.Len(t, acc.Items, 1, "same-named condition should not be duplicated")
}

func TestAdd_RefreshesNamedConditionWhenNewDurationIsLonger(t *testing.T) {
	acc := &ActiveConditionsComponent{}
	acc.Add(&DamageConditionComponent{Name: "burning", Duration: 2, DamageDice: "1"})
	acc.Add(&DamageConditionComponent{Name: "burning", Duration: 6, DamageDice: "1"})
	dc := acc.Items[0].(*DamageConditionComponent)
	assert.Equal(t, 6, dc.Duration)
}

func TestAdd_KeepsLongerDurationWhenExistingIsGreater(t *testing.T) {
	acc := &ActiveConditionsComponent{}
	acc.Add(&DamageConditionComponent{Name: "venom", Duration: 10, DamageDice: "1"})
	acc.Add(&DamageConditionComponent{Name: "venom", Duration: 3, DamageDice: "1"})
	dc := acc.Items[0].(*DamageConditionComponent)
	assert.Equal(t, 10, dc.Duration, "shorter re-application must not reduce duration")
}

func TestAdd_DifferentNamesStack(t *testing.T) {
	acc := &ActiveConditionsComponent{}
	acc.Add(&DamageConditionComponent{Name: "poison", Duration: 3, DamageDice: "1"})
	acc.Add(&DamageConditionComponent{Name: "burning", Duration: 3, DamageDice: "1"})
	assert.Len(t, acc.Items, 2)
}

func TestAdd_StatAndDamageConditionsStack(t *testing.T) {
	acc := &ActiveConditionsComponent{}
	acc.Add(&DamageConditionComponent{Name: "acid", Duration: 3, DamageDice: "1"})
	acc.Add(&StatConditionComponent{Name: "weakened", Duration: 3, Mods: []StatMod{{Stat: "str", Delta: -2}}})
	assert.Len(t, acc.Items, 2)
}

// ---- Tick / damage ----------------------------------------------------------

func TestTick_DealsDamageEachTurn(t *testing.T) {
	e := newTestEntityWithHealth(20)
	acc := GetOrCreateActiveConditions(e)
	acc.Add(&DamageConditionComponent{Name: "acid", Duration: 3, DamageDice: "2"})

	acc.Tick(e, nil, damageApplied(e))

	hc := e.GetComponent(Health).(*HealthComponent)
	assert.Equal(t, 18, hc.Health)
}

func TestTick_MultipleDamageConditionsApplyAll(t *testing.T) {
	e := newTestEntityWithHealth(30)
	acc := GetOrCreateActiveConditions(e)
	acc.Add(&DamageConditionComponent{Duration: 3, DamageDice: "3"}) // unnamed — stack
	acc.Add(&DamageConditionComponent{Duration: 3, DamageDice: "2"}) // unnamed — stack

	acc.Tick(e, nil, damageApplied(e))

	hc := e.GetComponent(Health).(*HealthComponent)
	assert.Equal(t, 25, hc.Health, "both conditions should deal damage")
}

func TestTick_RemovesExpiredConditions(t *testing.T) {
	e := newTestEntityWithHealth(20)
	acc := GetOrCreateActiveConditions(e)
	acc.Add(&DamageConditionComponent{Name: "venom", Duration: 2, DamageDice: "1"})

	acc.Tick(e, nil, damageApplied(e))
	acc.Tick(e, nil, damageApplied(e))

	assert.Empty(t, acc.Items, "expired condition should be removed")
}

func TestTick_KeepsNonExpiredConditions(t *testing.T) {
	e := newTestEntityWithHealth(20)
	acc := GetOrCreateActiveConditions(e)
	acc.Add(&DamageConditionComponent{Name: "burn", Duration: 5, DamageDice: "1"})

	acc.Tick(e, nil, damageApplied(e))

	assert.Len(t, acc.Items, 1, "condition with remaining duration should stay")
}

func TestTick_NilDamageFuncDoesNotPanic(t *testing.T) {
	e := newTestEntity()
	acc := GetOrCreateActiveConditions(e)
	acc.Add(&DamageConditionComponent{Name: "acid", Duration: 3, DamageDice: "1d4"})

	assert.NotPanics(t, func() { acc.Tick(e, nil, nil) })
}

// ---- Tick / stat conditions -------------------------------------------------

func TestTick_StatConditionAppliesOnFirstTick(t *testing.T) {
	e := newTestEntity()
	acc := GetOrCreateActiveConditions(e)
	acc.Add(&StatConditionComponent{
		Name:     "Hardened",
		Duration: 3,
		Mods:     []StatMod{{Stat: "ac", Delta: 3}},
	})

	acc.Tick(e, nil, nil)

	sc := e.GetComponent(Stats).(*StatsComponent)
	assert.Equal(t, 13, sc.AC)
}

func TestTick_StatConditionNotAppliedTwice(t *testing.T) {
	e := newTestEntity()
	acc := GetOrCreateActiveConditions(e)
	acc.Add(&StatConditionComponent{
		Name:     "Hardened",
		Duration: 5,
		Mods:     []StatMod{{Stat: "ac", Delta: 3}},
	})

	acc.Tick(e, nil, nil)
	acc.Tick(e, nil, nil)

	sc := e.GetComponent(Stats).(*StatsComponent)
	assert.Equal(t, 13, sc.AC, "stat bonus must not stack with itself each tick")
}

func TestTick_StatConditionRevertsOnExpiry(t *testing.T) {
	e := newTestEntity()
	acc := GetOrCreateActiveConditions(e)
	acc.Add(&StatConditionComponent{
		Name:     "Weakened",
		Duration: 2,
		Mods:     []StatMod{{Stat: "str", Delta: -4}},
	})

	acc.Tick(e, nil, nil)
	acc.Tick(e, nil, nil)

	sc := e.GetComponent(Stats).(*StatsComponent)
	assert.Equal(t, 10, sc.Str, "stat must be restored after condition expires")
	assert.Empty(t, acc.Items)
}

// ---- interface compliance ---------------------------------------------------

func TestActiveConditionsComponent_ImplementsComponent(t *testing.T) {
	var _ ecs.Component = &ActiveConditionsComponent{}
}
