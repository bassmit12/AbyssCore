package game

import (
	"context"
	"encoding/json"
	"log"

	"encore.dev/beta/errs"
	"encore.app/shared/events"

)

// StartConsumers wires up all RabbitMQ consumers for game-service.
// Called from encore service init.
func StartConsumers() {
	if err := events.Connect(); err != nil {
		log.Printf("[game-service] rabbitmq unavailable, consumers skipped: %v", err)
		return
	}

	go consumeMonsterKilled()
	go consumePlayerDied()
	go consumeItemUsed()
}

// consumeMonsterKilled: award XP to hero when a monster dies
func consumeMonsterKilled() {
	msgs, err := events.Subscribe("game.monster.killed", "combat.monster.killed")
	if err != nil {
		log.Printf("[game-service] subscribe combat.monster.killed: %v", err)
		return
	}
	for d := range msgs {
		var payload struct {
			HeroID    string `json:"hero_id"`
			XPReward  int    `json:"xp_reward"`
		}
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			d.Nack(false, false)
			continue
		}
		ctx := context.Background()
		if _, err := AwardXP(ctx, payload.HeroID, &AwardXPRequest{Amount: payload.XPReward}); err != nil {
			log.Printf("[game-service] award xp to %s: %v", payload.HeroID, err)
			d.Nack(false, true) // requeue
			continue
		}
		d.Ack(false)
	}
}

// consumePlayerDied: mark hero as dead
func consumePlayerDied() {
	msgs, err := events.Subscribe("game.player.died.game", "game.player.died")
	if err != nil {
		log.Printf("[game-service] subscribe game.player.died: %v", err)
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
		if err := MarkDead(ctx, payload.HeroID); err != nil {
			log.Printf("[game-service] mark dead %s: %v", payload.HeroID, err)
			d.Nack(false, true)
			continue
		}
		d.Ack(false)
	}
}

// consumeItemUsed: restore HP when a potion is used
func consumeItemUsed() {
	msgs, err := events.Subscribe("game.item.used", "inventory.item.used")
	if err != nil {
		log.Printf("[game-service] subscribe inventory.item.used: %v", err)
		return
	}
	for d := range msgs {
		var payload struct {
			HeroID   string `json:"hero_id"`
			ItemType string `json:"item_type"`
			Heal     int    `json:"heal"`
		}
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			d.Nack(false, false)
			continue
		}
		if payload.ItemType != "potion" {
			d.Ack(false)
			continue
		}
		ctx := context.Background()
		if err := healHero(ctx, payload.HeroID, payload.Heal); err != nil {
			log.Printf("[game-service] heal hero %s: %v", payload.HeroID, err)
			d.Nack(false, true)
			continue
		}
		d.Ack(false)
	}
}

func healHero(ctx context.Context, heroID string, amount int) error {
	_, err := db.Exec(ctx, `
		UPDATE game.heroes
		SET hp = LEAST(hp + $1, max_hp), updated_at = now()
		WHERE id = $2::uuid AND alive = TRUE
	`, amount, heroID)
	return err
}

func publishEvent(ctx context.Context, routingKey string, payload map[string]any) {
	if err := events.Publish(ctx, routingKey, payload); err != nil {
		log.Printf("[game-service] publish %s: %v", routingKey, err)
	}
}

// Ensure errs import doesn't error (used by encore auth)
var _ = errs.B()

// init wires consumers and telemetry when the service starts
func init() {
	go initOtel()
	go StartConsumers()
}
