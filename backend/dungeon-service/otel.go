package dungeon

import (
	"context"
	"log"

	appOtel "encore.app/shared/otel"
)

func initOtel() {
	shutdown, err := appOtel.Init(context.Background(), "dungeon-service")
	if err != nil {
		log.Printf("[dungeon-service] otel init failed: %v", err)
		return
	}
	_ = shutdown
}
