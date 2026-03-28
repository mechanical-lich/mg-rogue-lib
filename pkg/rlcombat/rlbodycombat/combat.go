package rlbodycombat

import (
	"fmt"
	"log"
	"math/rand"
	"strings"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcombat"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlworld"
	"github.com/mechanical-lich/mlge/dice"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
)

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

// rollDamage rolls and returns damage for an attack on a specific body part.
// Resistance is checked against entity-wide stats and the armor on hitPartName only.
func rollDamage(attacker, defender *ecs.Entity, crit bool, hitPartName string) (int, string) {
	attackDice, damageType, mod := getAttackStats(attacker)

	d := attackDice
	if strings.Contains(d, "d") {
		if mod >= 0 {
			d = fmt.Sprintf("%s+%d", attackDice, mod)
		} else {
			d = fmt.Sprintf("%s%d", attackDice, mod)
		}
	}

	damage := 0
	roll, err := dice.ParseDiceRequest(d)
	if err == nil {
		damage = roll.Result
	} else {
		log.Print("rlcombat/v2: error rolling dice: ", d)
	}

	if crit {
		damage *= 2
	}

	if partHasResistance(defender, hitPartName, damageType) {
		damage /= 2
	}
	if rlcombat.HasWeakness(defender, damageType) {
		damage *= 2
	}
	if damage <= 0 {
		damage = 1
	}

	return damage, damageType
}

// randomBodyPart picks a random non-amputated body part from bc.
// Returns ("", nil) if all parts are amputated.
func randomBodyPart(bc *rlcomponents.BodyComponent) (string, *rlcomponents.BodyPart) {
	available := make([]string, 0, len(bc.Parts))
	for name, part := range bc.Parts {
		if !part.Amputated {
			available = append(available, name)
		}
	}
	if len(available) == 0 {
		return "", nil
	}
	name := available[rand.Intn(len(available))]
	part := bc.Parts[name]
	return name, &part
}

// applyBodyPartDamage subtracts damage from the named part and updates Broken/Amputated flags.
// Returns whether the part became broken, amputated, and whether the entity should die.
func applyBodyPartDamage(bc *rlcomponents.BodyComponent, partName string, damage int) (broken, amputated, kills bool) {
	part := bc.Parts[partName]
	part.HP -= damage

	if part.HP <= 0 && !part.Broken {
		part.Broken = true
		broken = true
		if part.KillsWhenBroken {
			kills = true
		}
	}

	if damage >= part.MaxHP*2 && !part.Amputated {
		part.Amputated = true
		amputated = true
		if part.KillsWhenAmputated {
			kills = true
		}
	}

	bc.Parts[partName] = part
	return broken, amputated, kills
}

// Hit performs a D&D-style melee attack from entity to entityHit.
// If the defender has a BodyComponent, damage is applied to a random body part;
// otherwise it falls back to the v1 HealthComponent behaviour.
// A natural 20 is a critical hit and doubles damage.
// If swap is true and both entities are friendly, they swap positions instead.
// Returns true if the attack was executed, false if it was invalid.
func Hit(level rlworld.LevelInterface, entity, entityHit *ecs.Entity, swap bool) bool {
	if entity == nil || entityHit == nil || entity == entityHit {
		return false
	}

	if rlcombat.IsFriendly(entity, entityHit) {
		if swap {
			pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
			hitPC := entityHit.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
			oldX, oldY, oldZ := pc.GetX(), pc.GetY(), pc.GetZ()
			level.PlaceEntity(hitPC.GetX(), hitPC.GetY(), hitPC.GetZ(), entity)
			level.PlaceEntity(oldX, oldY, oldZ, entityHit)
		}
		return false
	}

	hasBody := entityHit.HasComponent(rlcomponents.Body)

	if !entity.HasComponent(rlcomponents.Stats) || !entityHit.HasComponent(rlcomponents.Stats) {
		return false
	}
	if !hasBody && !entityHit.HasComponent(rlcomponents.Health) {
		return false
	}

	hitSc := entityHit.GetComponent(rlcomponents.Stats).(*rlcomponents.StatsComponent)

	mod := getToHitMod(entity)
	acBonus := getACBonus(entityHit)

	roll, err := dice.ParseDiceRequest("1d20")
	if err != nil {
		log.Print("rlcombat/v2: error rolling d20: ", err)
		return false
	}

	pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
	crit := roll.Result == 20
	hitLanded := crit || roll.Result+mod > hitSc.AC+acBonus

	if hitLanded {
		if hasBody {
			bc := entityHit.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)
			partName, part := randomBodyPart(bc)

			if part != nil {
				damage, damageType := rollDamage(entity, entityHit, crit, partName)
				broken, amputated, kills := applyBodyPartDamage(bc, partName, damage)

				postHitMessage(entity, entityHit, partName, damage, damageType, crit, broken, amputated, pc)

				if kills {
					if entityHit.HasComponent(rlcomponents.Health) {
						entityHit.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent).Health = 0
					}
					entityHit.AddComponent(&rlcomponents.DeadComponent{})
				}
			} else {
				// All parts are amputated — entity cannot survive.
				entityHit.AddComponent(&rlcomponents.DeadComponent{})
			}
		} else {
			// No body — damage goes directly to HealthComponent.
			damage, damageType := rollDamage(entity, entityHit, crit, "")
			hc := entityHit.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
			hc.Health -= damage
			postHitMessage(entity, entityHit, "", damage, damageType, crit, false, false, pc)
			if hc.Health <= 0 {
				hc.Health = 0
				entityHit.AddComponent(&rlcomponents.DeadComponent{})
			}
		}
		rlcombat.ApplyStatusEffects(entity, entityHit)
	} else {
		atkName, defName := getEntityName(entity), getEntityName(entityHit)
		if atkName != "" {
			message.PostLocatedTaggedMessage("combat", atkName, fmt.Sprintf("missed %s", defName), pc.GetX(), pc.GetY(), pc.GetZ())
		}
		event.GetQueuedInstance().QueueEvent(CombatEvent{
			X: pc.GetX(), Y: pc.GetY(), Z: pc.GetZ(),
			AttackerName: atkName,
			DefenderName: defName,
			Miss:         true,
		})
	}

	rlcombat.TriggerDefenses(entityHit, pc.GetX(), pc.GetY())
	return true
}

func SavingThrow(entity *ecs.Entity, saveType string, dc int, damageType string, damageDice string) bool {
	if !entity.HasComponent(rlcomponents.Stats) {
		return false
	}

	sc := entity.GetComponent(rlcomponents.Stats).(*rlcomponents.StatsComponent)
	var mod int
	switch saveType {
	case "str":
		mod = rlcombat.GetModifier(sc.Str)
	case "dex":
		mod = rlcombat.GetModifier(sc.Dex)
	case "int":
		mod = rlcombat.GetModifier(sc.Int)
	case "wis":
		mod = rlcombat.GetModifier(sc.Wis)
	default:
		mod = rlcombat.GetModifier(sc.Dex)
	}

	roll, err := dice.ParseDiceRequest("1d20")
	if err != nil {
		log.Print("rlcombat/v2: error rolling d20: ", err)
		return false
	}

	success := roll.Result+mod >= dc

	if !success {
		damage, _ := rollDamage(entity, entity, false, "")
		if entity.HasComponent(rlcomponents.Body) {
			bc := entity.GetComponent(rlcomponents.Body).(*rlcomponents.BodyComponent)
			partName, part := randomBodyPart(bc)
			pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)

			if part != nil {
				damage, err := dice.Roll(damageDice)
				if err != nil {
					log.Print("rlcombat/v2: error rolling damage dice: ", err)
					return false
				}

				// Apply resistances/weaknesses to hit part.
				if partHasResistance(entity, part.Name, damageType) {
					damage /= 2
				}
				if rlcombat.HasWeakness(entity, damageType) {
					damage *= 2
				}
				if damage <= 0 {
					damage = 1
				}

				broken, amputated, kills := applyBodyPartDamage(bc, partName, damage)
				//postHitMessage(entity, entityHit, partName, damage, damageType, crit, broken, amputated, pc)
				postSaveFailMessage(entity, partName, damage, damageType, broken, amputated, pc)
				if kills {
					if entity.HasComponent(rlcomponents.Health) {
						entity.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent).Health = 0
					}
					entity.AddComponent(&rlcomponents.DeadComponent{})
				}
			} else {
				// All parts are amputated — entity cannot survive.
				entity.AddComponent(&rlcomponents.DeadComponent{})
			}
		} else {
			hc := entity.GetComponent(rlcomponents.Health).(*rlcomponents.HealthComponent)
			hc.Health -= damage
			pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
			postSaveFailMessage(entity, "", damage, damageType, false, false, pc)
			if hc.Health <= 0 {
				hc.Health = 0
				entity.AddComponent(&rlcomponents.DeadComponent{})
			}
		}

	} else {
		pc := entity.GetComponent(rlcomponents.Position).(*rlcomponents.PositionComponent)
		postSaveSuccessMessage(entity, "", damageType, pc)
	}

	return success
}

// getEntityNames returns the Description names for attacker and defender.
// Returns ("", "") if either entity lacks a DescriptionComponent.
func getEntityName(entity *ecs.Entity) string {
	if !entity.HasComponent(rlcomponents.Description) {
		return ""
	}
	return entity.GetComponent(rlcomponents.Description).(*rlcomponents.DescriptionComponent).Name
}
