package game

import (
	"context"
	"log"

	appOtel "encore.app/shared/otel"
)

// initOtel initialises OpenTelemetry tracing for game-service.
// Called from init() alongside StartConsumers.
func initOtel() {
	shutdown, err := appOtel.Init(context.Background(), "game-service")
	if err != nil {
		log.Printf("[game-service] otel init failed: %v", err)
		return
	}
	// shutdown is called when the process exits - register it
	_ = shutdown
}
