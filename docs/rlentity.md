---
layout: default
title: rlentity
nav_order: 7
---

# Entity Helpers (Legacy)

`github.com/mechanical-lich/mg-rogue-lib/pkg/rlentity`

> **Note:** This package predates `rlai` and duplicates some of its functionality. Prefer `rlai` for new code. `rlentity` is retained for backward compatibility.

Stateless helper functions for common entity manipulations.

## Functions

### Face

```go
func Face(entity *ecs.Entity, deltaX int, deltaY int)
```

Updates the entity's `DirectionComponent` based on a movement delta. Direction mapping:

| `deltaX` / `deltaY` | Direction |
|---------------------|-----------|
| `deltaX > 0` | 0 (right) |
| `deltaY > 0` | 1 (down) |
| `deltaY < 0` | 2 (up) |
| `deltaX < 0` | 3 (left) |

Panics if the entity does not have a `DirectionComponent`. Prefer `rlai.Face`, which checks for the component first.

---

### Swap

```go
func Swap(level rlworld.LevelInterface, entity *ecs.Entity, entityHit *ecs.Entity)
```

Exchanges the grid positions of two entities. Does nothing if the two pointers are equal.
