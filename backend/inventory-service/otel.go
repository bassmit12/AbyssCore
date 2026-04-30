package inventory

import (
	"context"
	"log"

	appOtel "encore.app/shared/otel"
)

func initOtel() {
	shutdown, err := appOtel.Init(context.Background(), "inventory-service")
	if err != nil {
		log.Printf("[inventory-service] otel init failed: %v", err)
		return
	}
	_ = shutdown
}
