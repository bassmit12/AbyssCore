module encore.app

go 1.22

require (
	encore.dev v1.46.0
	github.com/golang-jwt/jwt/v5 v5.2.2
	github.com/MicahParks/keyfunc/v3 v3.7.0
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/prometheus/client_golang v1.19.1
	go.opentelemetry.io/otel v1.28.0
	go.opentelemetry.io/otel/trace v1.28.0
	go.opentelemetry.io/otel/sdk v1.28.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.28.0
	go.opentelemetry.io/otel/sdk/metric v1.28.0
	google.golang.org/grpc v1.65.0
)
