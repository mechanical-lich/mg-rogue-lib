package rlworld

import (
	"github.com/mechanical-lich/mlge/path"
)

// Tile is a GC-friendly tile struct with no pointer fields.
// Coordinates are derived from the flat index via the active Level.
type Tile struct {
	Type       int    // Index into TileDefinitions
	Variant    int    // Visual variant
	LightLevel int    // Cached lighting value
	Idx        int    `json:"-"` // Flat index into Level.Data (derives X/Y/Z) — excluded from serialization
}

// Coords derives X, Y, Z from the flat index and the active level dimensions.
func (t *Tile) Coords() (x, y, z int) {
	x = t.Idx % activeLevel.Width
	y = (t.Idx / activeLevel.Width) % activeLevel.Height
	z = t.Idx / (activeLevel.Width * activeLevel.Height)
	return
}

func (t *Tile) IsSolid() bool { return TileDefinitions[t.Type].Solid }
func (t *Tile) IsWater() bool { return TileDefinitions[t.Type].Water }
func (t *Tile) IsAir() bool   { return TileDefinitions[t.Type].Air }

// pathOffsets is pre-allocated to avoid allocation in the hot path.
var pathOffsets = [6][3]int{
	{-1, 0, 0}, // Left
	{1, 0, 0},  // Right
	{0, -1, 0}, // Up
	{0, 1, 0},  // Down
	{0, 0, -1}, // Z down
	{0, 0, 1},  // Z up
}

// PathID returns a unique integer ID for this tile (its flat index).
func (t *Tile) PathID() int {
	return t.Idx
}

// PathNeighborsAppend appends walkable neighbors to the provided slice (zero-allocation).
// Vertical (Z) neighbors require StairsUp/StairsDown on the destination tile.
func (t *Tile) PathNeighborsAppend(neighbors []path.Pather) []path.Pather {
	x, y, z := t.Coords()
	for i := range pathOffsets {
		offset := &pathOffsets[i]
		n := activeLevel.GetTilePtr(x+offset[0], y+offset[1], z+offset[2])
		if n == nil {
			continue
		}
		if offset[2] != 0 && !(TileDefinitions[n.Type].StairsUp || TileDefinitions[n.Type].StairsDown) {
			continue
		}
		neighbors = append(neighbors, n)
	}
	return neighbors
}

// PathNeighborCost returns the movement cost to an adjacent tile.
// If the Level has a custom PathCostFunc set, it is used; otherwise defaultPathCost applies.
func (t *Tile) PathNeighborCost(to path.Pather) float64 {
	toTile, ok := to.(*Tile)
	if !ok || toTile == nil || t == nil {
		return 500
	}
	if activeLevel.PathCostFunc != nil {
		return activeLevel.PathCostFunc(t, toTile)
	}
	return DefaultPathCost(t, toTile)
}

// PathEstimatedCost returns a heuristic estimate (squared Euclidean distance).
func (t *Tile) PathEstimatedCost(to path.Pather) float64 {
	t2, ok := to.(*Tile)
	if !ok || t2 == nil || t == nil {
		return 1e18
	}
	x1, y1, z1 := t.Coords()
	x2, y2, z2 := t2.Coords()
	dx := x2 - x1
	dy := y2 - y1
	dz := z2 - z1
	return float64(dx*dx + dy*dy + dz*dz)
}

// DefaultPathCost is the base path cost function. It rejects solid/water tiles
// and enforces stairs for Z-level transitions. Games should provide their own
// PathCostFunc on Level for entity-blocking, doors, factions, etc.
func DefaultPathCost(from, to *Tile) float64 {
	tileDef := TileDefinitions[to.Type]

	if tileDef.Solid || tileDef.Water {
		return 5000.0
	}

	_, _, fromZ := from.Coords()
	_, _, toZ := to.Coords()

	if fromZ < toZ {
		if !TileDefinitions[from.Type].StairsUp {
			return 1000.0
		}
	} else if fromZ > toZ {
		if !TileDefinitions[from.Type].StairsDown {
			return 1000.0
		}
	}

	return 0.0
}

