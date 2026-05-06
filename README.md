# AbyssCore

A real-time multiplayer dungeon crawler built to showcase a modern Kubernetes-native stack.

🌐 **Live demo:** https://abysscore.bassmit.dev (testplayer / password)

📋 **Operational status:** see [PROGRESS.md](./PROGRESS.md)
📝 **Implementation plan:** see [PLAN.md](./PLAN.md)

## Stack

| Layer | Tech |
|---|---|
| Frontend | Next.js 16 (App Router, TypeScript) |
| API Gateway | GraphQL (gqlgen) + WebSocket subscriptions |
| Backend | Encore.go microservices (7 services) |
| Auth | Keycloak (OIDC/JWT) |
| Event Bus | RabbitMQ (topic exchange: `abysscore.events`) |
| Database | PostgreSQL (per-service databases) |
| Metrics | Prometheus + Grafana ✅ |
| Tracing | OpenTelemetry → Jaeger _(planned)_ |
| Infra | Kubernetes (Fontys-managed, 3 nodes) |
| GitOps | ArgoCD (auto-sync on push to `main`) |
| Public Access | Cloudflare Tunnel (no LoadBalancer/Ingress, no VPN required) |
| CI/CD | GitHub Actions → ghcr.io (SHA-tagged images, ArgoCD auto-rollout) |

## Services

- **auth-service** — JWKS JWT validation, Keycloak integration
- **game-service** — Hero CRUD, movement, XP/leveling
- **dungeon-service** — Procedural floor generation
- **combat-service** — Turn-based combat resolution
- **deck-service** — Card definitions, hero decks
- **inventory-service** — Weighted loot drops, item use
- **leaderboard-service** — Run submission, top-10 rankings
- **map-service** — Dungeon node graph, hero positions
- **graphql-gateway** — Unified query surface, WebSocket real-time

## Production Endpoints

| Endpoint | URL |
|---|---|
| Game | https://abysscore.bassmit.dev |
| GraphQL Playground | https://abysscore-api.bassmit.dev/playground |
| Keycloak Admin | https://abysscore-auth.bassmit.dev (admin/admin) |
| ArgoCD | https://argocd.bassmit.dev |
| Grafana | https://grafana.bassmit.dev |

Pushing to `main` triggers GitHub Actions to build and push images to ghcr.io with SHA tags.
ArgoCD picks up the manifest change and rolls out automatically — no manual steps needed.

For cluster operations (kubectl), connect to Fontys OpenVPN and use `abysscore_cluster.conf` (gitignored).

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
cd graphql-gateway
go run .
```

### Local URLs

| Service | URL |
|---|---|
| Game | http://localhost:3000 |
| GraphQL Playground | http://localhost:4001/playground |
| Encore Dashboard | http://localhost:9400 |
| Keycloak Admin | http://localhost:8080 (admin/admin) |
| RabbitMQ Management | http://localhost:15672 (guest/guest) |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3001 (admin/admin) |
| Jaeger | http://localhost:16686 |

### Test Credentials
- Username: `testplayer`
- Password: `password`

## RabbitMQ Event Routing

Exchange: `abysscore.events` (topic)

| Routing Key | Published by | Consumed by |
|---|---|---|
| `dungeon.player.moved` | game-service | dungeon-service |
| `dungeon.floor.entered` | dungeon-service | game-service |
| `dungeon.chest.opened` | dungeon-service | inventory-service |
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
