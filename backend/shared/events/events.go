package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel"
)

func otelHeaders(ctx context.Context) map[string]any {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	out := make(map[string]any, len(carrier))
	for k, v := range carrier {
		out[k] = v
	}
	return out
}

const ExchangeName = "abysscore.events"

var conn *amqp.Connection
var ch *amqp.Channel

func Connect() error {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://abysscore:abysscore@localhost:5672/"
	}

	var err error
	// Retry up to 10 times (RabbitMQ can be slow to start)
	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("[rabbitmq] connection attempt %d failed: %v", i+1, err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("rabbitmq connect: %w", err)
	}

	ch, err = conn.Channel()
	if err != nil {
		return fmt.Errorf("rabbitmq channel: %w", err)
	}

	// Declare topic exchange
	err = ch.ExchangeDeclare(
		ExchangeName,
		"topic",
		true,  // durable
		false, // auto-delete
		false, // internal
		false, // no-wait
		nil,
	)
	if err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}

	log.Println("[rabbitmq] connected and exchange declared")
	return nil
}

// Publish sends a message to the exchange with the given routing key.
// Injects OTEL trace context into headers.
func Publish(ctx context.Context, routingKey string, payload any) error {
	if ch == nil {
		return nil // not connected (e.g. during tests), silently skip
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Inject OTEL trace context into AMQP headers for distributed tracing
	headers := amqp.Table{}
	for k, v := range otelHeaders(ctx) {
		headers[k] = v
	}

	return ch.PublishWithContext(ctx,
		ExchangeName,
		routingKey,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
			Headers:      headers,
			Timestamp:    time.Now(),
		},
	)
}

// Subscribe binds a durable queue to the exchange with a routing key pattern and returns messages.
func Subscribe(queueName, routingKey string) (<-chan amqp.Delivery, error) {
	if ch == nil {
		return nil, fmt.Errorf("rabbitmq not connected")
	}

	q, err := ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-dead-letter-exchange": ExchangeName + ".dlx",
		},
	)
	if err != nil {
		return nil, fmt.Errorf("declare queue %s: %w", queueName, err)
	}

	if err := ch.QueueBind(q.Name, routingKey, ExchangeName, false, nil); err != nil {
		return nil, fmt.Errorf("bind queue %s: %w", queueName, err)
	}

	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("consume %s: %w", queueName, err)
	}

	return msgs, nil
}

// Close cleans up connections.
func Close() {
	if ch != nil {
		ch.Close()
	}
	if conn != nil {
		conn.Close()
	}
}
