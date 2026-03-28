package rlbodycombat

import (
	"fmt"

	"github.com/mechanical-lich/ml-rogue-lib/pkg/rlcomponents"
	"github.com/mechanical-lich/mlge/ecs"
	"github.com/mechanical-lich/mlge/event"
	"github.com/mechanical-lich/mlge/message"
)

func postHitMessage(entity, entityHit *ecs.Entity, partName string, damage int, damageType string, crit, broken, amputated bool, pc *rlcomponents.PositionComponent) {
	atkName, defName := getEntityName(entity), getEntityName(entityHit)

	if atkName != "" {
		verb := "hit"
		if crit {
			verb = "critically hit"
		}
		msg := fmt.Sprintf("%s %s's %s for %d (%s)", verb, defName, partName, damage, damageType)
		if amputated {
			msg += fmt.Sprintf(" — %s's %s was amputated!", defName, partName)
		} else if broken {
			msg += fmt.Sprintf(" — %s's %s was broken!", defName, partName)
		}
		message.PostLocatedTaggedMessage("combat", atkName, msg, pc.GetX(), pc.GetY(), pc.GetZ())
	}

	event.GetQueuedInstance().QueueEvent(CombatEvent{
		X: pc.GetX(), Y: pc.GetY(), Z: pc.GetZ(),
		AttackerName: atkName,
		DefenderName: defName,
		Damage:       damage,
		DamageType:   damageType,
		BodyPart:     partName,
		Crit:         crit,
		Broken:       broken,
		Amputated:    amputated,
	})
}

func postSaveFailMessage(entityHit *ecs.Entity, partName string, damage int, damageType string, broken, amputated bool, pc *rlcomponents.PositionComponent) {
	defName := getEntityName(entityHit)

	if defName != "" {
		msg := fmt.Sprintf("%s's %s failed to save against %d (%s)", defName, partName, damage, damageType)
		if amputated {
			msg += fmt.Sprintf(" — %s's %s was amputated!", defName, partName)
		} else if broken {
			msg += fmt.Sprintf(" — %s's %s was broken!", defName, partName)
		}
		message.PostLocatedTaggedMessage("combat", defName, msg, pc.GetX(), pc.GetY(), pc.GetZ())
	}
}

func postSaveSuccessMessage(entityHit *ecs.Entity, partName string, damageType string, pc *rlcomponents.PositionComponent) {
	defName := getEntityName(entityHit)

	if defName != "" {
		msg := fmt.Sprintf("%s's %s successfully saved against %s", defName, partName, damageType)

		message.PostLocatedTaggedMessage("combat", defName, msg, pc.GetX(), pc.GetY(), pc.GetZ())
	}
}
