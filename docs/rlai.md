---
layout: default
title: rlai
nav_order: 5
---

# AI Helpers

`github.com/mechanical-lich/mg-rogue-lib/pkg/rlai`

Stateless helper functions for entity movement, facing, interaction, and death detection. These are the low-level primitives that `AISystem` and game-specific player-input handlers call directly.

## Functions

### HandleDeath

```go
func HandleDeath(entity *ecs.Entity) bool
```

Checks whether the entity's `HealthComponent.Health` has dropped to zero or below. If so, adds a `DeadComponent` and returns `true`. Does nothing and returns `false` if the entity has no `HealthComponent`.

Call this at the start of an entity's update to bail out early if it just died.

---

### Move

```go
func Move(entity *ecs.Entity, level rlworld.LevelInterface, deltaX, deltaY, deltaZ int) bool
```

Attempts to move the entity by `(deltaX, deltaY, deltaZ)`. Returns `true` if a solid entity was blocking the destination.

**Movement rules:**

- If a `SolidComponent` entity is at the destination, movement is blocked — unless the blocker is a `DoorComponent` that the entity is allowed to pass through (checked via `CanPassThroughDoor`).
- If the destination tile is `IsAir`, the entity instead drops to the solid tile directly below (simulates gravity).
- Water tiles block movement entirely.
- Solid tiles block movement entirely.

---

### Face

```go
func Face(entity *ecs.Entity, deltaX, deltaY int)
```

Updates the entity's `DirectionComponent` based on a movement delta. Direction mapping:

| `deltaX` / `deltaY` | Direction |
|---------------------|-----------|
| `deltaX > 0` | 0 (right) |
| `deltaY > 0` | 1 (down) |
| `deltaY < 0` | 2 (up) |
| `deltaX < 0` | 3 (left) |

Does nothing if the entity lacks a `DirectionComponent`.

---

### Eat

```go
func Eat(entity, foodEntity *ecs.Entity) bool
```

Consumes one unit of food from `foodEntity` by decrementing `FoodComponent.Amount`. Returns `true` on success, `false` if `foodEntity` has no `FoodComponent` or if `entity == foodEntity`.

---

### Swap

```go
func Swap(level rlworld.LevelInterface, entity, entityHit *ecs.Entity)
```

Exchanges the grid positions of two entities. Calls `level.PlaceEntity` for both. Does nothing if they are the same entity.

---

### CanPassThroughDoor

```go
func CanPassThroughDoor(entity *ecs.Entity, door *rlcomponents.DoorComponent) bool
```

Returns `true` if the entity is allowed to move through the given door. An entity may pass if:

- The door is not locked, **or**
- The door's `OwnedBy` faction matches the entity's `DescriptionComponent.Faction`.

---

## Usage Example

```go
import (
    "github.com/mechanical-lich/mg-rogue-lib/pkg/rlai"
    "github.com/mechanical-lich/mlge/utility"
)

// Simple wander step (same logic as WanderAI inside AISystem).
func wanderStep(entity *ecs.Entity, level rlworld.LevelInterface) {
    dx := utility.GetRandom(-1, 2)
    dy := 0
    if dx == 0 {
        dy = utility.GetRandom(-1, 2)
    }
    rlai.Move(entity, level, dx, dy, 0)
    rlai.Face(entity, dx, dy)
}

// Player movement handler.
func movePlayer(player *ecs.Entity, level rlworld.LevelInterface, dx, dy int) {
    if rlai.HandleDeath(player) {
        return
    }
    hitSolid := rlai.Move(player, level, dx, dy, 0)
    if hitSolid {
        // bump-attack logic, open door prompt, etc.
    }
    rlai.Face(player, dx, dy)
}
```
