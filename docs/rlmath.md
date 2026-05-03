---
layout: default
title: rlmath
nav_order: 17
---

# Math Utilities

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlmath`

Small stateless math helpers used throughout the library and game projects. No external dependencies beyond the Go standard library.

## Functions

### GetRandom

```go
func GetRandom(low, high int) int
```

Returns a random integer in `[low, high)`. If `high <= low`, returns `low`.

---

### Distance

```go
func Distance(x1, y1, x2, y2 int) int
```

Returns an estimated Chebyshev distance between two points using the `max(dx, dy) + min(dx, dy)/2` approximation. Fast and allocation-free — suitable for AI range checks and sorting candidates.

---

### Shuffle

```go
func Shuffle(s []string)
```

Randomly shuffles a string slice in place using `math/rand`.

---

### Sgn

```go
func Sgn(a int) int
```

Returns the sign of `a`: `-1`, `0`, or `+1`.

## Usage Example

```go
import "github.com/mechanical-lich/ml-rogue-lib/pkg/rlmath"

// Pick a random spawn location index.
idx := rlmath.GetRandom(0, len(candidates))

// Sort entities by approximate distance to the player.
slices.SortFunc(enemies, func(a, b *ecs.Entity) int {
    da := rlmath.Distance(px, py, posOf(a))
    db := rlmath.Distance(px, py, posOf(b))
    return rlmath.Sgn(da - db)
})

// Randomise a room-type list before assigning.
rlmath.Shuffle(roomTypes)
```
