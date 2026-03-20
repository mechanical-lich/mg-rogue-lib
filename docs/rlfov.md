---
layout: default
title: rlfov
nav_order: 5
---

# rlfov

`github.com/mechanical-lich/ml-rogue-lib/pkg/rlfov`

Line-of-sight and field-of-view for tile-based levels. Works directly with `*rlworld.Level` and the fog-of-war explored state built into the base level.

---

## Line of sight

```go
visible := rlfov.Los(level, pX, pY, tX, tY, z)
```

`Los` reports whether `(tX, tY)` has an unobstructed line of sight to `(pX, pY)` on Z layer `z`. Uses Bresenham's line algorithm — integer-only, no allocations.

**Blocked by:**
- Tiles where `IsSolid()` is true
- Tiles whose `TileDefinition` has `Door: true`
- Out-of-bounds coordinates (returns false)

The source tile itself is never checked, so a viewer standing inside a wall will still see outward.

---

## Field of view

```go
rlfov.UpdateFieldOfView(level, x, y, z, radius)
```

Calls `Los` for every tile within `radius` of `(x, y, z)` and marks visible tiles as seen on the level (`level.SetSeen`). Explored state accumulates — tiles are never un-marked by this call.

This is the standard call to make each turn after the player moves:

```go
rlfov.UpdateFieldOfView(level, player.X, player.Y, player.Z, visionRadius)
```

---

## Fog of war state

The explored state lives on `*rlworld.Level` as a parallel `[]bool` slice — one byte per tile, never scanned by the GC:

```go
// Check whether a tile has ever been seen
seen := level.GetSeen(x, y, z)

// Mark a tile as seen (done automatically by UpdateFieldOfView)
level.SetSeen(x, y, z, true)

// Reset all explored state (e.g. when loading a new level)
level.ClearSeen()
```

The slice is allocated by `rlworld.NewLevel` alongside the tile array, so no extra setup is required.

---

## Rendering pattern

`rlfov` provides only the visibility data — rendering is left to the game. A typical drawing loop:

```go
for each tile in viewport {
    currentlyVisible := rlfov.Los(level, playerX, playerY, tileX, tileY, z)
    everSeen         := level.GetSeen(tileX, tileY, z)

    if currentlyVisible {
        level.SetSeen(tileX, tileY, z, true) // mark explored
        drawTileFull(tile)
        drawEntities(tile)
    } else if everSeen {
        drawTileDark(tile) // explored but not currently visible
    } else {
        drawUnknown()      // never seen — solid black or theme color
    }
}
```

---

## Performance

`Los` walks at most `max(|dx|, |dy|)` tiles. On a 2000×2000 map with a vision radius of 20, each `UpdateFieldOfView` call runs up to 1,681 LOS checks (41×41 grid). Each check is a tight integer loop with no allocations.

For best throughput, call `UpdateFieldOfView` once per turn rather than once per frame.

---

## API reference

### `Los(level *rlworld.Level, pX, pY, tX, tY, z int) bool`

Returns true if `(tX, tY, z)` has line of sight to `(pX, pY, z)`. Not goroutine-safe if the level is being written concurrently, but safe for concurrent reads.

### `UpdateFieldOfView(level *rlworld.Level, x, y, z, radius int)`

Marks all tiles within `radius` of `(x, y, z)` as seen if they pass `Los`. Out-of-bounds tiles are skipped. Does not clear previously seen tiles outside the radius.
