---
layout: default
title: Getting Started
nav_order: 2
---

# Getting Started

This guide covers installation and a minimal integration example using ML Rogue Lib.

## Prerequisites

- **Go 1.25+**
- A project already using [MLGE](https://mechanical-lich.github.io/mlge) (or at minimum `github.com/mechanical-lich/mlge/ecs`)

## Installation

```bash
go get github.com/mechanical-lich/ml-rogue-lib
```

## Core Concepts

ML Rogue Lib is built around MLGE's ECS. Every game object is an `*ecs.Entity` carrying a set of `rlcomponents` structs. Systems iterate entities each frame and act on whichever components they require.

The library provides two interfaces — `rlworld.LevelInterface` and `rlworld.TileInterface` — along with **base implementations** (`rlworld.Level` and `rlworld.Tile`) that you can use directly or embed in your own types. All systems and helpers accept the interfaces, keeping the library decoupled from any specific game.

## Minimal Integration

### 1. Load tile definitions

Create a `tile_definitions.json` file:

```json
[
  {"name": "air", "air": true, "variants": [{"variant": 0, "spriteX": 0, "spriteY": 0}]},
  {"name": "grass", "variants": [{"variant": 0, "spriteX": 0, "spriteY": 16}]},
  {"name": "wall", "solid": true, "variants": [{"variant": 0, "spriteX": 16, "spriteY": 0}]},
  {"name": "water", "water": true, "variants": [{"variant": 0, "spriteX": 32, "spriteY": 0}]}
]
```

Load it at startup:

```go
import "github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"

err := rlworld.LoadTileDefinitions("data/tile_definitions.json")
if err != nil {
    log.Fatal(err)
}
```

### 2. Create a level

Use the base `Level` directly, or embed it for game-specific fields:

```go
// Option A: Use the base directly
level := rlworld.NewLevel(128, 128, 10)

// Option B: Embed in a game-specific wrapper
type GameLevel struct {
    *rlworld.Level
    lightOverlay *ebiten.Image
}

func NewGameLevel(w, h, d int) *GameLevel {
    base := rlworld.NewLevel(w, h, d)
    gl := &GameLevel{Level: base}
    base.PathCostFunc = myPathCost(gl) // custom pathfinding
    return gl
}
```

### 3. Spawn entities with components

```go
import (
    "github.com/mechanical-lich/mlge/ecs"
    "github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
)

func spawnPlayer(level rlworld.LevelInterface, x, y int) *ecs.Entity {
    e := &ecs.Entity{Blueprint: "player"}
    e.AddComponent(&rlcomponents.PositionComponent{X: x, Y: y, Z: 0})
    e.AddComponent(&rlcomponents.HealthComponent{MaxHealth: 20, Health: 20})
    e.AddComponent(&rlcomponents.StatsComponent{
        AC: 12, Str: 14, Dex: 12,
        BasicAttackDice: "1d6",
    })
    e.AddComponent(&rlcomponents.InitiativeComponent{DefaultValue: 10, Ticks: 10})
    e.AddComponent(&rlcomponents.DescriptionComponent{Name: "Player"})
    e.AddComponent(&rlcomponents.InventoryComponent{})
    level.AddEntity(e)
    return e
}
```

### 4. Register systems and run the game loop

```go
import (
    "github.com/mechanical-lich/ml-rogue-lib/pkg/rlsystems"
    "github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
    "github.com/mechanical-lich/mlge/ecs"
)

type Game struct {
    level    *rlworld.Level
    systemMgr ecs.SystemManager
    cleanup  rlsystems.CleanUpSystem
}

func NewGame() *Game {
    g := &Game{level: rlworld.NewLevel(128, 128, 10)}

    // Register systems.
    g.systemMgr.AddSystem(rlsystems.NewAISystem())
    g.systemMgr.AddSystem(&rlsystems.InitiativeSystem{Speed: 1})
    g.systemMgr.AddSystem(&rlsystems.StatusConditionSystem{})

    // Wire up extension hooks.
    g.cleanup.OnEntityDead = func(level rlworld.LevelInterface, e *ecs.Entity) {
        // spawn loot, award XP, play sounds…
    }
    return g
}

func (g *Game) Update() error {
    // 1. Strip MyTurn and remove dead entities.
    g.cleanup.Update(g.level)

    // 2. Run all registered systems for every entity.
    g.systemMgr.UpdateSystemsForEntities(g.level, g.level.GetEntities())
    return nil
}
```

## Custom Pathfinding Costs

The base `Level` uses `DefaultPathCost` which rejects solid/water tiles and enforces stairs for Z transitions. To add game-specific logic (entity blocking, doors, factions), set `PathCostFunc`:

```go
level := rlworld.NewLevel(128, 128, 10)
level.PathCostFunc = func(from, to *rlworld.Tile) float64 {
    if rlworld.TileDefinitions[to.Type].Solid {
        return 5000
    }
    // Check for blocking entities, doors, etc.
    return 0
}
```

## Extension Hooks

Every system in `rlsystems` exposes callback fields (e.g. `OnEntityDead`, `OnHostileAttack`, `OnEntityTurn`) rather than hard-coding game-specific behaviour. Assign Go functions to these fields to layer your game's logic on top of the built-in mechanics.

```go
aiSystem := rlsystems.NewAISystem()
aiSystem.OnHostileAttack = func(level rlworld.LevelInterface, attacker, target *ecs.Entity) {
    // play sfx, shake camera, etc.
}
```

See individual package pages for the full list of hooks each system exposes.
