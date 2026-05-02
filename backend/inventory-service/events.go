package inventory

import (
	"context"
	"encoding/json"
	"log"

	"encore.app/shared/events"

	"encore.app/shared/metrics"
)

func init() {
	go initOtel()
	go func() {
		if err := events.Connect(); err != nil {
			log.Printf("[inventory-service] rabbitmq unavailable: %v", err)
			return
		}
		go consumeMonsterKilled()
		go consumeChestOpened()
	}()
}

// consumeMonsterKilled: roll loot drop when a monster dies
func consumeMonsterKilled() {
	msgs, err := events.Subscribe("inventory.monster.killed", "combat.monster.killed")
	if err != nil {
		log.Printf("[inventory-service] subscribe combat.monster.killed: %v", err)
		return
	}
	for d := range msgs {
		var payload struct {
			HeroID    string `json:"hero_id"`
			MonsterID string `json:"monster_id"`
		}
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			d.Nack(false, false)
			continue
		}
		ctx := context.Background()
		item, err := RollLootDrop(ctx, &LootDropRequest{
			HeroID:    payload.HeroID,
			MonsterID: payload.MonsterID,
		})
		if err != nil {
			log.Printf("[inventory-service] loot drop: %v", err)
			d.Nack(false, true)
			continue
		}
		if item != nil {
		metrics.LootDropsTotal.WithLabelValues(item.Rarity).Inc()
		log.Printf("[inventory-service] hero %s got loot: %s (%s)", payload.HeroID, item.Name, item.Rarity)
		}
		d.Ack(false)
	}
}

func publishEvent(ctx context.Context, routingKey string, payload map[string]any) {
	if err := events.Publish(ctx, routingKey, payload); err != nil {
		log.Printf("[inventory-service] publish %s: %v", routingKey, err)
	}
}

// consumeChestOpened: guarantee loot drop (100%) when a chest is opened
func consumeChestOpened() {
	msgs, err := events.Subscribe("inventory.chest.opened", "dungeon.chest.opened")
	if err != nil {
		log.Printf("[inventory-service] subscribe dungeon.chest.opened: %v", err)
		return
	}
	for d := range msgs {
		var payload struct {
			HeroID string `json:"hero_id"`
		}
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			d.Nack(false, false)
			continue
		}
		ctx := context.Background()
		// Chest always drops — call RollLootDrop until we get an item
		var item *Item
		for item == nil {
			item, err = RollLootDrop(ctx, &LootDropRequest{HeroID: payload.HeroID, MonsterID: ""})
			if err != nil {
				log.Printf("[inventory-service] chest loot drop: %v", err)
				break
			}
		}
		if item != nil {
			metrics.LootDropsTotal.WithLabelValues(item.Rarity).Inc()
			log.Printf("[inventory-service] hero %s opened chest: %s (%s)", payload.HeroID, item.Name, item.Rarity)
		}
		d.Ack(false)
	}
}
