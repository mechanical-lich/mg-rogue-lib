package rlcomponents

import (
	"fmt"

	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/utility"
)

type DropItem struct {
	Chance    int    // Chance to drop (0-100)
	Quantity  int    // Quantity to drop if successful
	Blueprint string // Blueprint name of the item to drop
}

type DropsComponent struct {
	Items         []DropItem // List of potential drops with their chances
	NumRolls      int        // number of times to roll for drops
	AlwaysDropAll bool       // if true, all items with be dropped
}

func (c *DropsComponent) GetType() ecs.ComponentType {
	return Drops
}

func (c *DropsComponent) GetDrops() map[string]int {
	dropped := make(map[string]int)
	if c.AlwaysDropAll {
		for _, item := range c.Items {
			dropped[item.Blueprint] = item.Quantity
		}
		return dropped
	}

	fmt.Println("Calculating drops for", c.Items)
	// TODO - Look more into this, but fine for POC.
	for _, item := range c.Items {
		chance := item.Chance
		for i := 0; i < c.NumRolls; i++ {
			if utility.GetRandom(0, 100) < chance {
				if item.Quantity <= 0 {
					item.Quantity = 1
				}
				dropped[item.Blueprint] += item.Quantity
			}
		}
	}

	return dropped
}
