package leaderboard

import (
	"context"
	"log"

	appOtel "encore.app/shared/otel"
)

func initOtel() {
	shutdown, err := appOtel.Init(context.Background(), "leaderboard-service")
	if err != nil {
		log.Printf("[leaderboard-service] otel init failed: %v", err)
		return
	}
	_ = shutdown
}
