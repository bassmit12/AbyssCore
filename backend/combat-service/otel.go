package combat

import (
	"context"
	"log"

	appOtel "encore.app/shared/otel"
)

func initOtel() {
	shutdown, err := appOtel.Init(context.Background(), "combat-service")
	if err != nil {
		log.Printf("[combat-service] otel init failed: %v", err)
		return
	}
	_ = shutdown
}
