package combat

import (
	"context"
	"log"

	"encore.app/shared/events"
)

func init() {
	go initOtel()
	go func() {
		if err := events.Connect(); err != nil {
			log.Printf("[combat-service] rabbitmq unavailable: %v", err)
		}
	}()
}

func publishEvent(ctx context.Context, routingKey string, payload map[string]any) {
	if err := events.Publish(ctx, routingKey, payload); err != nil {
		log.Printf("[combat-service] publish %s: %v", routingKey, err)
	}
}
