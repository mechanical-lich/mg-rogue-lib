package rlcomponents

import (
	"testing"

	"github.com/mechanical-lich/mlge/ecs"
	"github.com/stretchr/testify/assert"
)

func createTestEntityWithItemComponent(name string) *ecs.Entity {
	entity := &ecs.Entity{}
	descriptionComponent := &DescriptionComponent{Name: name}
	entity.AddComponent(descriptionComponent)
	itemComponent := &ItemComponent{Slot: BagSlot} // Default slot is bag
	entity.AddComponent(itemComponent)
	return entity
}

func createTestEntityWithWeaponComponent(name string, slot ItemSlot) *ecs.Entity {
	entity := &ecs.Entity{}
	descriptionComponent := &DescriptionComponent{Name: name}
	entity.AddComponent(descriptionComponent)
	weaponComponent := &WeaponComponent{AttackBonus: 10, AttackDice: "1d6"}
	entity.AddComponent(weaponComponent)
	itemComponent := &ItemComponent{Slot: slot}
	entity.AddComponent(itemComponent)
	return entity
}

func createTestEntityWithArmorComponent(name string, slot ItemSlot) *ecs.Entity {
	entity := &ecs.Entity{}
	descriptionComponent := &DescriptionComponent{Name: name}
	entity.AddComponent(descriptionComponent)
	armorComponent := &ArmorComponent{DefenseBonus: 5}
	entity.AddComponent(armorComponent)
	itemComponent := &ItemComponent{Slot: slot}
	entity.AddComponent(itemComponent)
	return entity
}

func TestInventoryComponent_AddItem(t *testing.T) {
	sword := createTestEntityWithWeaponComponent("sword", HandSlot)
	iC := &InventoryComponent{}
	iC.AddItem(sword)
	assert.Equal(t, 1, len(iC.Bag))
	assert.Equal(t, sword, iC.Bag[0])
}

func TestInventoryComponent_RemoveItem(t *testing.T) {
	sword := createTestEntityWithWeaponComponent("sword", HandSlot)
	shield := createTestEntityWithArmorComponent("shield", TorsoSlot)

	iC := &InventoryComponent{}
	iC.AddItem(sword)
	iC.AddItem(shield)
	assert.True(t, iC.RemoveItem(sword))
	assert.Equal(t, 1, len(iC.Bag))
	// Test removing an item that doesn't exist
	assert.False(t, iC.RemoveItem(sword))
}

func TestInventoryComponent_HasItem(t *testing.T) {
	iC := &InventoryComponent{}
	sword := createTestEntityWithWeaponComponent("sword", HandSlot)

	iC.AddItem(sword)

	assert.True(t, iC.HasItem("sword"), "Expected sword to be in bag")

	assert.False(t, iC.HasItem("shield"), "Expected there to be no shield in the bag.")
}

func TestInventoryComponent_RemoveAll(t *testing.T) {
	iC := &InventoryComponent{}
	sword := createTestEntityWithWeaponComponent("sword", HandSlot)
	sword2 := createTestEntityWithWeaponComponent("sword", HandSlot)
	shield := createTestEntityWithArmorComponent("shield", TorsoSlot)
	iC.AddItem(sword)
	iC.AddItem(sword2)
	iC.AddItem(shield)
	assert.True(t, iC.RemoveAll("sword"))
	assert.Equal(t, 1, len(iC.Bag))
	assert.Equal(t, shield, iC.Bag[0])
	// Test removing an item that doesn't exist
	assert.False(t, iC.RemoveAll("sword"))
}

func TestInventoryComponent_EquipHandWeapon(t *testing.T) {
	iC := &InventoryComponent{}
	sword := createTestEntityWithWeaponComponent("sword", HandSlot)
	iC.AddItem(sword)
	iC.Equip(sword)
	assert.Equal(t, sword, iC.RightHand, "Expected sword to be equipped in right hand")
}

func TestInventoryComponent_EquipTorsoArmor(t *testing.T) {
	iC := &InventoryComponent{}
	armor := createTestEntityWithArmorComponent("armor", TorsoSlot)
	iC.AddItem(armor)
	iC.Equip(armor)
	assert.Equal(t, armor, iC.Torso, "Expected armor to be equipped in torso")
}

func TestInventoryComponent_EquipHeadArmor(t *testing.T) {
	iC := &InventoryComponent{}
	armor := createTestEntityWithArmorComponent("armor", HeadSlot)
	iC.AddItem(armor)
	iC.Equip(armor)
	assert.Equal(t, armor, iC.Head, "Expected armor to be equipped in head")
}

func TestInventoryComponent_EquipLegArmor(t *testing.T) {
	iC := &InventoryComponent{}
	armor := createTestEntityWithArmorComponent("armor", LegsSlot)
	iC.AddItem(armor)
	iC.Equip(armor)
	assert.Equal(t, armor, iC.Legs, "Expected armor to be equipped in legs")
}

func TestInventoryComponent_GetAttackModifier(t *testing.T) {
	iC := &InventoryComponent{}
	sword1 := createTestEntityWithWeaponComponent("sword1", HandSlot)
	sword2 := createTestEntityWithWeaponComponent("sword2", HandSlot)

	iC.AddItem(sword1)
	iC.Equip(sword1)

	iC.AddItem(sword2)
	iC.Equip(sword2)

	assert.Equal(t, 20, iC.GetAttackModifier(), "Expected attack modifier to be 20")
}

func TestInventoryComponent_GetDefenseModifier(t *testing.T) {
	iC := &InventoryComponent{}
	armor1 := createTestEntityWithArmorComponent("armor1", TorsoSlot)
	armor2 := createTestEntityWithArmorComponent("armor2", LegsSlot)
	armor3 := createTestEntityWithArmorComponent("armor3", FeetSlot)

	iC.AddItem(armor1)
	iC.Equip(armor1)

	iC.AddItem(armor2)
	iC.Equip(armor2)

	iC.AddItem(armor3)
	iC.Equip(armor3)

	assert.Equal(t, 15, iC.GetDefenseModifier(), "Expected defense modifier to be 10")
}

func TestInventoryComponent_GetAttackDice(t *testing.T) {
	iC := &InventoryComponent{}
	sword1 := createTestEntityWithWeaponComponent("sword1", HandSlot)
	sword2 := createTestEntityWithWeaponComponent("sword2", HandSlot)

	iC.AddItem(sword1)
	iC.Equip(sword1)

	iC.AddItem(sword2)
	iC.Equip(sword2)

	assert.Equal(t, "1d6+1d6", iC.GetAttackDice(), "Expected attack dice to be '1d6+1d6'")
}
