package rlbodycombat

import (
	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/event"
)

func postHitMessage(entity, entityHit *ecs.Entity, partName string, damage int, damageType string, crit, broken, amputated bool, pc *rlcomponents.PositionComponent) {
	atkName, defName := getEntityName(entity), getEntityName(entityHit)
	// Post the main hit event (no broken/amputated flags — those fire separately).
	event.GetQueuedInstance().QueueEvent(CombatEvent{
		X: pc.GetX(), Y: pc.GetY(), Z: pc.GetZ(),
		Attacker:     entity,
		Defender:     entityHit,
		AttackerName: atkName,
		DefenderName: defName,
		Source:       getWeaponName(entity),
		Damage:       damage,
		DamageType:   damageType,
		BodyPart:     partName,
		Crit:         crit,
	})
	// Post a separate event for broken/amputated so the listener can format it independently.
	if broken || amputated {
		event.GetQueuedInstance().QueueEvent(CombatEvent{
			X: pc.GetX(), Y: pc.GetY(), Z: pc.GetZ(),
			Defender:     entityHit,
			DefenderName: defName,
			BodyPart:     partName,
			Broken:       broken,
			Amputated:    amputated,
		})
	}
}

func postSaveFailMessage(entityHit *ecs.Entity, partName string, damage int, damageType string, broken, amputated bool, pc *rlcomponents.PositionComponent) {
	defName := getEntityName(entityHit)
	event.GetQueuedInstance().QueueEvent(CombatEvent{
		X: pc.GetX(), Y: pc.GetY(), Z: pc.GetZ(),
		Defender:     entityHit,
		DefenderName: defName,
		Damage:       damage,
		DamageType:   damageType,
		BodyPart:     partName,
		Broken:       broken,
		Amputated:    amputated,
		SaveFail:     true,
	})
}

func postSaveSuccessMessage(entityHit *ecs.Entity, partName string, damageType string, pc *rlcomponents.PositionComponent) {
	defName := getEntityName(entityHit)
	event.GetQueuedInstance().QueueEvent(CombatEvent{
		X: pc.GetX(), Y: pc.GetY(), Z: pc.GetZ(),
		Defender:     entityHit,
		DefenderName: defName,
		DamageType:   damageType,
		BodyPart:     partName,
		SavePass:     true,
	})
}
