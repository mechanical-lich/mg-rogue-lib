package rlworld

import "github.com/mechanical-lich/mlge/path"

type TileInterface interface {
	Coords() (x, y, z int)
	PathID() int
	PathNeighborsAppend(neighbors []path.Pather) []path.Pather
	PathNeighborCost(to path.Pather) float64
	PathEstimatedCost(to path.Pather) float64
	// Tile properties for movement and AI
	IsSolid() bool
	IsWater() bool
	IsAir() bool
}
