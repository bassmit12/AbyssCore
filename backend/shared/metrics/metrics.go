package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Game-native metrics - these tell a real story in demos

var (
	ActivePlayers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "abysscore_active_players_total",
		Help: "Number of heroes currently in a dungeon run",
	})

	CombatEventsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "abysscore_combat_events_total",
		Help: "Total combat events by outcome",
	}, []string{"outcome"}) // "hit", "kill", "death", "dodge"

	MonstersKilledTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "abysscore_monsters_killed_total",
		Help: "Total monsters killed across all runs",
	})

	HeroDeathsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "abysscore_hero_deaths_total",
		Help: "Total hero deaths",
	})

	FloorsGeneratedTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "abysscore_floors_generated_total",
		Help: "Total dungeon floors procedurally generated",
	})

	LootDropsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "abysscore_loot_drops_total",
		Help: "Total loot drops by rarity",
	}, []string{"rarity"}) // "common", "uncommon", "rare"

	RabbitMQQueueDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "abysscore_rabbitmq_queue_depth",
		Help: "Approximate number of messages in each RabbitMQ queue",
	}, []string{"queue"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "abysscore_http_request_duration_seconds",
		Help:    "HTTP request duration by service and endpoint",
		Buckets: prometheus.DefBuckets,
	}, []string{"service", "method", "path", "status"})

	RunScoreHistogram = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "abysscore_run_score",
		Help:    "Distribution of run scores when heroes die",
		Buckets: []float64{100, 500, 1000, 2500, 5000, 10000, 25000},
	})
)

// Handler returns the /metrics HTTP handler to mount in each service.
func Handler() http.Handler {
	return promhttp.Handler()
}
