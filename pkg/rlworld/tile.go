package rlworld

// TileInterface is the read-only view of a tile used by AI and entity systems.
// Pathfinding is now handled by Level (which implements path.Graph), so only
// the coordinate, identity, and property methods are needed here.
type TileInterface interface {
	Coords() (x, y, z int)
	PathID() int
	IsSolid() bool
	IsWater() bool
	IsAir() bool
}
