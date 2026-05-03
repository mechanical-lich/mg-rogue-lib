## ML Rogue Lib

Reusable roguelike / tile-based game features that don't belong in mlge.
Designed to run efficiently on large maps (2000×2000×10 and beyond) with many
simultaneous AI agents.

## Packages

| Package | Purpose |
|---|---|
| `pkg/path` | Graph-centric A\* pathfinding (GC-invisible hot path) |
| `pkg/rlworld` | 3D tile level, spatial entity index, lighting, time |
| `pkg/rlcomponents` | ECS component definitions |
| `pkg/rlentity` | Entity helpers (move, face, eat, death) |
| `pkg/rlai` | AI utility functions (pathfinding, range checks) |
| `pkg/rlcombat` | Combat resolution, status effects |
| `pkg/rlsystems` | Turn-based ECS systems (AI, initiative, cleanup, doors) |
| `pkg/rlgeneration` | Procedural level generation (rooms, clusters, terrain) |
| `pkg/rlaoe` | Tile-offset generators for area-of-effect shapes (cone, circle, ring, line, burst) |
| `pkg/rlmath` | Stateless math utilities (random range, distance, shuffle, sign) |
| `pkg/rlaction` | Generic turn-based action interface (`Action[L]`) and affordability helper |

---

## Pathfinding

### Design

The pathfinding system is **graph-centric**: the graph (a `Level` or future
chunk) owns all knowledge of topology, cost, and heuristics. Individual tiles
need no back-reference to the level and no global state is required. This
enables:

- **Multiple simultaneous levels** — each is an independent graph
- **Chunked worlds** — each chunk can implement `path.Graph` independently
- **GC-invisible hot path** — the internal node array holds no pointer fields

### Usage

Node IDs for tile levels are flat tile indices (`tile.Idx`):

```go
astar := path.NewAStar(1024) // pre-allocate capacity for ~1024 explored nodes

from := level.GetTilePtr(x1, y1, z)
to   := level.GetTilePtr(x2, y2, z)

steps, dist, ok := astar.Path(level, from.Idx, to.Idx)
// steps is []int — flat tile indices, start→goal
// resolve back to tiles: level.GetTilePtrIndex(steps[i])
```

The returned `steps` slice is **reused between calls** — copy it if you need to
keep it past the next `Path` call.

For single-threaded convenience:

```go
steps, dist, ok := path.Path(level, from.Idx, to.Idx) // uses path.Default
```

### Concurrency

`AStar` is not goroutine-safe. Give each goroutine (or each chunk worker) its
own instance. An instance can be **reused across different graphs** sequentially
— it fully resets between calls:

```go
// Two chunks, one shared astar used sequentially
steps := astar.Path(chunkA, from, to)
steps  = astar.Path(chunkB, from, to)
```

### Memory layout

The internal `node` struct is 32 bytes with **zero pointer fields**, so the GC
skips the entire nodes slice regardless of how many nodes are explored:

```
cost   float64  @ 0   (8 bytes)
rank   float64  @ 8   (8 bytes)
parent int32    @ 16  (4 bytes)   -1 = no parent
index  int32    @ 20  (4 bytes)   position in heap
id     int32    @ 24  (4 bytes)   tile flat index
open   bool     @ 28  (1 byte)
closed bool     @ 29  (1 byte)
_      [2]byte  @ 30  (padding)
Total: 32 bytes, GC-invisible
```

At 5,000 explored nodes: ~160 KB, and zero pointers scanned by the GC.

### Implementing `path.Graph`

`*rlworld.Level` implements `path.Graph` out of the box. For custom graphs
(chunks, abstract navigation meshes, etc.) implement three methods:

```go
type Graph interface {
    // Append walkable neighbor IDs of nodeID into buf and return it.
    PathNeighborIDs(nodeID int, buf []int) []int

    // Movement cost between two adjacent node IDs.
    PathCost(fromID, toID int) float64

    // Heuristic estimate of remaining cost (squared Euclidean distance works well).
    PathEstimate(nodeID, goalID int) float64
}
```

### Custom movement costs

Set `Level.PathCostFunc` to inject game-specific logic (entity blocking, doors,
faction-restricted areas, etc.). If nil, `DefaultPathCost` is used:

```go
level.PathCostFunc = func(from, to *rlworld.Tile) float64 {
    toX, toY, toZ := to.Coords()
    if level.GetSolidEntityAt(toX, toY, toZ) != nil {
        return 5000
    }
    return rlworld.DefaultPathCost(from, to)
}
```

`DefaultPathCost` rules: solid/water tiles cost 5000, Z-transitions require
stair tiles, open floor costs 0.

---

## World / Level

### Tile layout

`Tile` stores only integers — no pointer back to its level. This keeps the GC
from scanning the entire tile array on every collection cycle. Coordinates are
derived from the flat index on demand:

```go
x, y, z := tile.Coords()  // O(1) arithmetic from tile.Idx
id       := tile.Idx       // use as the node ID for pathfinding
```

### Creating a level

```go
rlworld.SetTileDefinitions(myTileDefs) // once at startup
level := rlworld.NewLevel(width, height, depth)
```

Multiple levels exist simultaneously with no shared state between them — no
global `activeLevel`, no `SetActive()` call required.

### Entity spatial index

```go
level.AddEntity(entity)
level.PlaceEntity(x, y, z, entity)
level.GetEntityAt(x, y, z)
level.GetSolidEntityAt(x, y, z)
level.GetEntitiesAround(x, y, z, w, h, &buf)
level.GetClosestEntityMatching(x, y, z, w, h, exclude, matchFn)
```

---

## Chunked worlds

Because `path.Graph` is an interface and `AStar` holds no reference to any
specific level, chunked worlds fit naturally:

1. Each chunk is a `*rlworld.Level` or a custom type implementing `path.Graph`
2. Each chunk has its own `path.AStar` instance (or shares one sequentially)
3. Cross-chunk paths can be stitched by treating chunk boundary tiles as nodes

```go
chunkAstar.Path(chunkA, fromID, toID)
chunkBstar.Path(chunkB, fromID, toID)
```
