package leaderboard

import (
	"context"
	"encoding/json"
	"log"

	"encore.app/shared/events"
)

func init() {
	go initOtel()
	go func() {
		if err := events.Connect(); err != nil {
			log.Printf("[leaderboard-service] rabbitmq unavailable: %v", err)
			return
		}
		go consumePlayerDied()
	}()
}

// consumePlayerDied: finalize the run when a hero dies
func consumePlayerDied() {
	msgs, err := events.Subscribe("leaderboard.player.died", "game.player.died")
	if err != nil {
		log.Printf("[leaderboard-service] subscribe game.player.died: %v", err)
		return
	}
	for d := range msgs {
		var payload struct {
			HeroID         string `json:"hero_id"`
			HeroName       string `json:"hero_name"`
			PlayerID       string `json:"player_id"`
			FloorsCleared  int    `json:"floors_cleared"`
			MonstersKilled int    `json:"monsters_killed"`
			ItemsFound     int    `json:"items_found"`
		}
		if err := json.Unmarshal(d.Body, &payload); err != nil {
			d.Nack(false, false)
			continue
		}
		// Drop malformed messages that would cause a DB error
		if payload.HeroID == "" || payload.PlayerID == "" {
			log.Printf("[leaderboard-service] dropping event: missing hero_id or player_id")
			d.Nack(false, false)
			continue
		}
		ctx := context.Background()
		run, err := FinalizeRun(ctx, &FinalizeRunRequest{
			HeroID:         payload.HeroID,
			HeroName:       payload.HeroName,
			PlayerID:       payload.PlayerID,
			FloorsCleared:  payload.FloorsCleared,
			MonstersKilled: payload.MonstersKilled,
			ItemsFound:     payload.ItemsFound,
		})
		if err != nil {
			log.Printf("[leaderboard-service] finalize run: %v", err)
			d.Nack(false, true)
			continue
		}
		log.Printf("[leaderboard-service] run finalized: hero=%s score=%d", run.HeroID, run.Score)
		d.Ack(false)
	}
}
