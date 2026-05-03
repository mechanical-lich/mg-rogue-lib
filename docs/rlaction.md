---
layout: default
title: rlaction
nav_order: 18
---

# Action Interface

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlaction`

Generic interface for turn-based entity actions. Parameterised over the game's Level type so games can extend the base level without losing type safety.

## Types

### Action

```go
type Action[L any] interface {
    Cost(entity *ecs.Entity, level *L) int
    Available(entity *ecs.Entity, level *L) bool
    Execute(entity *ecs.Entity, level *L) error
}
```

A single entity action with three responsibilities:

- **`Cost`** — returns the expected energy cost; used by `HasAvailableAction` to decide affordability before calling `Execute`.
- **`Available`** — returns `true` if the action can be performed right now (e.g. entity has a weapon, has line of sight, etc.). Called before `Execute` as a guard.
- **`Execute`** — performs the action. Must call `rlenergy.SetActionCost` internally to consume the turn.

## Functions

### HasAvailableAction

```go
func HasAvailableAction[L any](entity *ecs.Entity, level *L, actions []Action[L]) bool
```

Returns `true` if at least one action in `actions` is both available (`Available` returns `true`) and affordable (entity's current energy ≥ `Cost`). Returns `false` if the entity has no `EnergyComponent`.

Used to decide whether to wait for input or skip an entity's turn entirely.

## Usage

### Define a type alias in your game

```go
// internal/action/action.go
package action

import (
    "github.com/mechanical-lich/ml-rogue-lib/pkg/rlaction"
    "github.com/mechanical-lich/mlge/ecs"
    "mygame/internal/world"
)

type Action = rlaction.Action[world.Level]

func HasAvailableAction(entity *ecs.Entity, level *world.Level, actions []Action) bool {
    return rlaction.HasAvailableAction(entity, level, actions)
}
```

### Implement an action

```go
type MoveAction struct{ DeltaX, DeltaY int }

func (a MoveAction) Cost(_ *ecs.Entity, _ *world.Level) int {
    return energy.CostMove
}

func (a MoveAction) Available(entity *ecs.Entity, _ *world.Level) bool {
    return entity.HasComponent(rlcomponents.Position)
}

func (a MoveAction) Execute(entity *ecs.Entity, level *world.Level) error {
    rlentity.Move(entity, level.Level, a.DeltaX, a.DeltaY, 0)
    rlenergy.SetActionCost(entity, energy.CostMove)
    return nil
}
```

### Check before acting (AI loop)

```go
if !action.HasAvailableAction(entity, level, myActions) {
    return // skip turn — no affordable action available
}
```
