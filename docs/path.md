---
layout: default
title: path
nav_order: 3
---

# path

`github.com/mechanical-lich/ml-rogue-lib/pkg/path`

Graph-centric A\* pathfinding with a GC-invisible hot path. Designed for large tile maps (2000×2000×10 and beyond) with many simultaneous AI agents.

---

## Design

The pathfinder is **graph-centric**: the graph owns all knowledge of topology, cost, and heuristics. The `AStar` instance is a pure algorithm that operates on integer node IDs — it holds no reference to any level, tile, or game type.

This means:
- Multiple levels can exist simultaneously — each is an independent graph
- Each chunk in a chunked world can implement `Graph` independently
- The internal node array has **zero pointer fields**, so the GC skips it entirely during collection

---

## Graph interface

Any type that implements `Graph` can be pathfinded over:

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

`*rlworld.Level` implements `Graph` out of the box. See [rlworld](rlworld.html) for details on the built-in cost function and how to override it.

---

## Usage

Node IDs for tile levels are flat tile indices (`tile.Idx`):

```go
astar := path.NewAStar(1024) // pre-allocate capacity for ~1024 explored nodes

from := level.GetTilePtr(x1, y1, z)
to   := level.GetTilePtr(x2, y2, z)

steps, dist, ok := astar.Path(level, from.Idx, to.Idx)
// steps is []int — flat tile indices, start→goal
// resolve back to tiles: level.GetTilePtrIndex(steps[i])
```

The returned `steps` slice is **reused between calls** — copy it before calling `Path` again if you need to keep the result:

```go
saved := make([]int, len(steps))
copy(saved, steps)
```

### Single-threaded convenience

A package-level `Default` instance is provided for single-threaded use:

```go
steps, dist, ok := path.Path(level, from.Idx, to.Idx) // uses path.Default
```

---

## Concurrency

`AStar` is **not goroutine-safe**. Give each goroutine (or each chunk worker) its own instance. An instance **can** be reused across different graphs sequentially — it fully resets between calls:

```go
// Two chunks, one AStar used sequentially
steps, _, ok := astar.Path(chunkA, fromID, toID)
steps, _, ok  = astar.Path(chunkB, fromID, toID)
```

---

## Memory layout

The internal `node` struct is 32 bytes with **zero pointer fields**. The Go GC skips the entire `nodes` slice regardless of how many nodes are explored:

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

Paths are returned as `[]int` (flat tile indices), so storing them in AI components is also GC-invisible.

---

## Custom graphs

For chunks, abstract navigation meshes, or other graph types, implement the three `Graph` methods and pass the graph directly to `AStar.Path`:

```go
type MyChunk struct { /* ... */ }

func (c *MyChunk) PathNeighborIDs(nodeID int, buf []int) []int {
    // Append walkable neighbor IDs
    return buf
}

func (c *MyChunk) PathCost(fromID, toID int) float64 {
    return 1.0
}

func (c *MyChunk) PathEstimate(fromID, goalID int) float64 {
    // Squared Euclidean distance works well as a heuristic
    x1, y1 := c.coords(fromID)
    x2, y2 := c.coords(goalID)
    dx, dy := x2-x1, y2-y1
    return float64(dx*dx + dy*dy)
}

steps, _, ok := chunkAstar.Path(myChunk, fromID, toID)
```

---

## Chunked worlds

Because `AStar` holds no reference to any specific level, chunked worlds fit naturally:

1. Each chunk is a `*rlworld.Level` or a custom type implementing `Graph`
2. Each chunk has its own `AStar` instance (or shares one sequentially)
3. Cross-chunk paths can be stitched by treating chunk boundary tiles as nodes in a higher-level graph

```go
chunkAstar.Path(chunkA, fromID, toID)
chunkAstar.Path(chunkB, fromID, toID)
```

---

## API reference

### `NewAStar(capacity int) *AStar`

Allocates a new `AStar` instance with the given initial node capacity. The instance grows automatically if more nodes are explored. A capacity of `1024` is a reasonable default; use a larger value if paths routinely explore many nodes.

### `(*AStar).Path(graph Graph, from, to int) (steps []int, dist float64, ok bool)`

Finds the shortest path from `from` to `to` on the given graph. Returns the slice of node IDs (start→goal), the total cost, and whether a path was found. The returned slice is owned by the `AStar` instance and reused on the next call.

### `Path(graph Graph, from, to int) (steps []int, dist float64, ok bool)`

Package-level convenience that calls `Default.Path(...)`. Not safe for concurrent use.
