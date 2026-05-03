---
layout: default
title: rlaoe
nav_order: 16
---

# Area-of-Effect Helpers

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlaoe`

Tile-offset generators for common area-of-effect shapes. All functions return offsets **relative to an origin** — callers add the origin position and apply whatever effect they need. Nothing in this package reads or writes game state.

## Types

### Offset

```go
type Offset struct{ X, Y int }
```

A relative (X, Y) tile displacement from an origin point.

## Functions

### Cone

```go
func Cone(fdx, fdy, depth, spread int) []Offset
```

Returns tile offsets for a cone pointing in direction `(fdx, fdy)`.

- `depth` — how many rows deep the cone extends.
- `spread` — perpendicular half-width at every row. Pass `-1` for classic widening behaviour (spread grows by 1 per row).

```
depth 3, spread -1 (widening):
  row 1:  [1]
  row 2: [2][2][2]
  row 3: [3][3][3][3][3]

depth 2, spread 1 (constant width):
  row 1: [1][1][1]
  row 2: [2][2][2]
```

Use `DirToVec` to convert a `DirectionComponent.Direction` value to `(fdx, fdy)`.

---

### Circle

```go
func Circle(radius int) []Offset
```

Returns all tile offsets within the given Euclidean radius, excluding the origin. Useful for explosions, auras, and area denial.

---

### Ring

```go
func Ring(radius int) []Offset
```

Returns tile offsets on the outer shell of the given Euclidean radius (a hollow circle). Tiles inside the ring are excluded. Useful for shockwaves and expanding hazards.

---

### Line

```go
func Line(fdx, fdy, length int) []Offset
```

Returns `length` tile offsets in a straight line in direction `(fdx, fdy)`, starting one step from the origin.

---

### Burst

```go
func Burst() []Offset
```

Returns all 8 adjacent tile offsets (Moore neighbourhood). Equivalent to `Circle(1)` but without the distance check, in a fixed clockwise order. Useful for close-range blasts and push effects.

---

### DirToVec

```go
func DirToVec(dir int) (int, int)
```

Converts a `DirectionComponent.Direction` value to a `(dx, dy)` unit vector using the `rlcomponents` convention:

| `dir` | Direction | `dx, dy` |
|-------|-----------|----------|
| 0 | right | `1, 0` |
| 1 | down | `0, 1` |
| 2 | up | `0, -1` |
| 3 | left | `-1, 0` |

## Usage Example

```go
import "github.com/mechanical-lich/ml-rogue-lib/pkg/rlaoe"

// Cone-of-fire spell in the direction the caster faces.
dc := caster.GetComponent(rlcomponents.Direction).(*rlcomponents.DirectionComponent)
pc := caster.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
ox, oy, z := pc.GetX(), pc.GetY(), pc.GetZ()

fdx, fdy := rlaoe.DirToVec(dc.Direction)
for _, off := range rlaoe.Cone(fdx, fdy, 3, -1) {
    tx, ty := ox+off.X, oy+off.Y
    applyFireDamage(level, tx, ty, z)
}

// Circular explosion.
for _, off := range rlaoe.Circle(4) {
    applyBlastDamage(level, ox+off.X, oy+off.Y, z)
}
```
