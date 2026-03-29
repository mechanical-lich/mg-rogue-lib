package rlbodycombat

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
)

// getEntityNames returns the Description names for attacker and defender.
// Returns ("", "") if either entity lacks a DescriptionComponent.
func getEntityName(entity *ecs.Entity) string {
	if !entity.HasComponent(rlcomponents.Description) {
		return ""
	}
	return entity.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent).Name
}

// getAttackStats returns (attackDice, damageType, mod) for the attacker,
// preferring BodyInventoryComponent over the legacy InventoryComponent.
func getAttackStats(attacker *ecs.Entity) (string, string, int) {
	if attacker.HasComponent(rlcomponents.BodyInventory) {
		sc := attacker.GetComponent(rlcomponents.Stats).(*rlcomponents.StatsComponent)
		inv := attacker.GetComponent(rlcomponents.BodyInventory).(*rlcomponents.BodyInventoryComponent)
		attackDice := sc.BasicAttackDice
		damageType := sc.BaseDamageType
		if damageType == "" {
			damageType = rlcombat.DefaultDamageType
		}
		mod := rlcombat.GetModifier(sc.Str) + inv.GetAttackModifier()
		if d := inv.GetAttackDice(); d != "" {
			attackDice = d
		}
		if dt := inv.GetDamageType(); dt != "" {
			damageType = dt
		}
		return attackDice, damageType, mod
	}
	return rlcombat.GetAttackDice(attacker)
}

// getToHitMod returns Dex + weapon attack bonus for the to-hit roll,
// preferring BodyInventoryComponent.
func getToHitMod(attacker *ecs.Entity) int {
	sc := attacker.GetComponent(rlcomponents.Stats).(*rlcomponents.StatsComponent)
	mod := rlcombat.GetModifier(sc.Dex)
	if attacker.HasComponent(rlcomponents.BodyInventory) {
		mod += attacker.GetComponent(rlcomponents.BodyInventory).(*rlcomponents.BodyInventoryComponent).GetAttackModifier()
	} else if attacker.HasComponent(rlcomponents.Inventory) {
		mod += attacker.GetComponent(rlcomponents.Inventory).(*rlcomponents.InventoryComponent).GetAttackModifier()
	}
	return mod
}

// getACBonus returns the total armor defense bonus used for the to-hit roll.
// All equipped armor counts here; per-part mitigation is handled separately.
func getACBonus(defender *ecs.Entity) int {
	if defender.HasComponent(rlcomponents.BodyInventory) {
		return defender.GetComponent(rlcomponents.BodyInventory).(*rlcomponents.BodyInventoryComponent).GetDefenseModifier()
	}
	if defender.HasComponent(rlcomponents.Inventory) {
		return defender.GetComponent(rlcomponents.Inventory).(*rlcomponents.InventoryComponent).GetDefenseModifier()
	}
	return 0
}

// partHasResistance returns true if the defender resists damageType either
// inherently (StatsComponent) or via armor equipped on the specific hit part.
func partHasResistance(defender *ecs.Entity, hitPartName, damageType string) bool {
	if defender.HasComponent(rlcomponents.Stats) {
		for _, r := range defender.GetComponent(rlcomponents.Stats).(*rlcomponents.StatsComponent).Resistances {
			if r == damageType {
				return true
			}
		}
	}
	if !defender.HasComponent(rlcomponents.BodyInventory) {
		return false
	}
	inv := defender.GetComponent(rlcomponents.BodyInventory).(*rlcomponents.BodyInventoryComponent)
	item := inv.Equipped[hitPartName]
	if item == nil || !item.HasComponent(rlcomponents.Armor) {
		return false
	}
	for _, r := range item.GetComponent(rlcomponents.Armor).(*rlcomponents.ArmorComponent).Resistances {
		if r == damageType {
			return true
		}
	}
	return false
}
