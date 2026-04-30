# AbyssCore

A real-time multiplayer dungeon crawler built to showcase a modern Kubernetes-native stack.

## Stack

| Layer | Tech |
|---|---|
| Frontend | Next.js 15 (App Router, TypeScript, Tailwind) |
| API Gateway | GraphQL (gqlgen) + WebSocket subscriptions |
| Backend | Encore.go microservices (6 services) |
| Auth | Keycloak (OIDC/JWT) |
| Event Bus | RabbitMQ (topic exchange: `abysscore.events`) |
| Database | PostgreSQL (per-service schemas) |
| Tracing | OpenTelemetry → Jaeger |
| Metrics | Prometheus + Grafana |
| Infra | Kubernetes + Cilium CNI (network policies) |

## Services

- **auth-service** — JWKS JWT validation, Keycloak integration
- **game-service** — Hero CRUD, movement, XP/leveling
- **dungeon-service** — Procedural 8×8 floor generation
- **combat-service** — Turn-based combat resolution
- **inventory-service** — Weighted loot drops, item use
- **leaderboard-service** — Run submission, top-10 rankings
- **graphql-gateway** — Unified query surface, WebSocket real-time

## Local Dev

### Prerequisites
- Docker + Docker Compose
- Node.js 20+
- Go 1.22+
- Encore CLI (Windows: `C:\Users\<user>\.encore\bin\encore.exe`)

### Start

```bash
# 1. Infra + frontend (from WSL/Linux)
./scripts/dev-start.sh

# 2. Encore backend (Windows terminal)
cd backend
encore run

# 3. GraphQL gateway
cd backend/graphql-gateway
go run .
```

### URLs

| Service | URL |
|---|---|
| Game | http://localhost:3000 |
| GraphQL Playground | http://localhost:4000/graphql |
| Keycloak Admin | http://localhost:8080 (admin/admin) |
| RabbitMQ Management | http://localhost:15672 (guest/guest) |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3001 (admin/admin) |
| Jaeger | http://localhost:16686 |

### Test Credentials
- Username: `testplayer`
- Password: `testpass`

## RabbitMQ Event Routing

Exchange: `abysscore.events` (topic)

| Routing Key | Published by | Consumed by |
|---|---|---|
| `dungeon.player.moved` | game-service | dungeon-service |
| `dungeon.floor.entered` | dungeon-service | game-service |
| `combat.attack.initiated` | combat-service | — |
| `combat.result` | combat-service | leaderboard-service |
| `combat.monster.killed` | combat-service | inventory-service, game-service |
| `game.player.died` | combat-service | leaderboard-service |
| `inventory.item.looted` | inventory-service | — |
| `inventory.item.used` | inventory-service | game-service |

## Hero Classes

| Class | HP | ATK | Special |
|---|---|---|---|
| Warrior | 120 | 12 | Tank |
| Rogue | 80 | 18 | 20% dodge |
| Mage | 70 | 22 | Glass cannon |

## Score Formula

`floors × 500 + kills × 50 + items × 25`
