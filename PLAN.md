# AbyssCore - Implementation Plan

> **For Hermes:** This is the living plan for AbyssCore. Resume from the first unchecked task.
> Project root: /mnt/d/Fontys/Personal Project/AbyssCore/

**Goal:** A real-time multiplayer dungeon crawler game that showcases a production-grade Kubernetes stack.

**Cluster status:** ONLINE since 2026-05-02. Fontys-managed Kubernetes (k8s v1.33.1, 3 nodes, 10.1.1.0/24). Reachable via Fontys OpenVPN OR Cloudflare Tunnel (no VPN required). Live at https://abysscore.bassmit.dev.

**Track ongoing work in PROGRESS.md** — single source of truth for what's done, what's next, and known pitfalls.

**Architecture:** Event-driven microservices. Player actions flow through RabbitMQ and fan out to independent services. GraphQL gateway aggregates Encore.go services into a single API for the Next.js frontend.

**Tech Stack:**
- Frontend: Next.js (TypeScript)
- Backend: Encore.go microservices (Go)
- API layer: GraphQL gateway (gqlgen or graphql-go)
- Database: PostgreSQL (per-service) + Hasura or custom GraphQL layer
- Message broker: RabbitMQ
- Auth: Keycloak (OIDC / JWT)
- Observability: Prometheus + Grafana + OpenTelemetry
- Networking: Cilium (when cluster is back online)
- Container: Docker / Kubernetes

---

## Phase 1: Project Scaffold (Local, no cluster needed)
- [x] **1.1** Create folder structure (frontend/, backend/, infra/, docs/)
- [x] **1.2** Init Next.js app in frontend/ (TypeScript, App Router, Tailwind)
- [x] **1.3** Init Encore.go workspace in backend/ with encore.app
- [x] **1.4** Create backend services: game-service, dungeon-service, combat-service, inventory-service, leaderboard-service, auth-service
- [x] **1.5** Create docker-compose.yml for local dev: Postgres, RabbitMQ, Keycloak, Prometheus, Grafana, OTEL, Jaeger
- [x] **1.6** Init git repo, add .gitignore, first commit

---

## Phase 2: Auth (Keycloak + Encore + Next.js)
- [x] **2.1** Keycloak realm + client config script (infra/scripts/keycloak-init.sh) — realm: abysscore, client: abysscore-frontend
- [x] **2.2** Encore auth-service: validate Keycloak JWT, expose user identity to other services
- [x] **2.3** Next.js: NextAuth.js with Keycloak provider, login/logout flow
- [x] **2.4** Protected route middleware in Next.js (redirect to Keycloak if no session)
- [x] **2.5** Pass JWT from frontend → GraphQL gateway → Encore services (auth header propagation)

---

## Phase 3: GraphQL Gateway
- [x] **3.1** Set up graphql-gateway service (Go, gqlgen)
- [x] **3.2** Define schema: Hero, DungeonFloor, Room, Monster, Item, CombatEvent, Leaderboard
- [x] **3.3** Wire resolvers to Encore service HTTP endpoints
- [x] **3.4** Add GraphQL subscriptions (WebSocket) for real-time combat log + floor state
- [x] **3.5** Auth middleware: extract JWT, inject user identity into resolver context

---

## Phase 4: Core Game Services (Encore.go)

### 4A: Dungeon Service
- [x] **4.1** Postgres schema: dungeons, floors, rooms, room_connections
- [x] **4.2** Procedural floor generator: grid of rooms, random connections, place monsters + loot
- [x] **4.3** API endpoint: POST /dungeon/start (creates dungeon + first floor for a player)
- [x] **4.4** API endpoint: GET /dungeon/:id/floor/:n (returns floor layout + entities)
- [x] **4.5** Publish event `dungeon.floor.entered` to RabbitMQ on floor entry

### 4B: Game Service (player state)
- [x] **4.6** Postgres schema: heroes (id, player_id, name, class, hp, max_hp, level, xp, position)
- [x] **4.7** API: POST /hero/create
- [x] **4.8** API: POST /hero/:id/move (validate move, update position, publish `dungeon.player.moved`)
- [x] **4.9** Consume `combat.monster.killed` → award XP, check level up

### 4C: Combat Service
- [x] **4.10** Postgres schema: monsters (id, floor_id, name, hp, max_hp, damage, status)
- [x] **4.11** API: POST /combat/attack (hero attacks monster, resolve damage)
- [x] **4.12** Publish `combat.attack.initiated`, consume and publish `combat.result`
- [x] **4.13** On monster death: publish `combat.monster.killed`
- [x] **4.14** On hero death: publish `game.player.died`, save run stats

### 4D: Inventory Service
- [x] **4.15** Postgres schema: items, hero_inventory
- [x] **4.16** Item definitions: sword, shield, potion, armor tiers
- [x] **4.17** Consume `combat.monster.killed` → roll loot drop, insert to hero_inventory
- [x] **4.18** API: GET /inventory/:hero_id, POST /inventory/use/:item_id

### 4E: Leaderboard Service
- [x] **4.19** Postgres schema: runs (hero_id, floors_cleared, monsters_killed, items_found, score, died_at)
- [x] **4.20** Consume `game.player.died` → finalize run, calculate score
- [x] **4.21** API: GET /leaderboard (top 10 runs)

---

## Phase 5: RabbitMQ Integration
- [ ] **5.1** Define exchange topology (topic exchange: `abysscore.events`)
- [ ] **5.2** Create shared Go pkg for publishing/consuming with OpenTelemetry trace propagation
- [ ] **5.3** Wire all event producers (game, dungeon, combat services)
- [ ] **5.4** Wire all event consumers (game, inventory, leaderboard services)
- [ ] **5.5** Add RabbitMQ management UI to docker-compose (port 15672)
- [ ] **5.6** Dead-letter queue for failed event processing

---

## Phase 6: OpenTelemetry
- [ ] **6.1** Add otel-collector to docker-compose (receives traces + metrics)
- [ ] **6.2** Instrument all Encore.go services with otel-go SDK (traces, spans)
- [ ] **6.3** Propagate trace context through RabbitMQ message headers
- [ ] **6.4** Instrument Next.js frontend with otel-js (page load, API calls)
- [ ] **6.5** Export traces to Jaeger (local) / Tempo (production)
- [ ] **6.6** Add otel metrics exporter → Prometheus scrape endpoint

---

## Phase 7: Prometheus + Grafana
- [ ] **7.1** Add Prometheus + Grafana to docker-compose
- [ ] **7.2** Encore services expose /metrics (Prometheus format via otel exporter)
- [ ] **7.3** RabbitMQ exporter added (rabbitmq_prometheus plugin)
- [ ] **7.4** Define custom game metrics: active_players, combat_events_total, hero_deaths_total, floors_cleared_total, items_dropped_total
- [ ] **7.5** Grafana dashboard: game overview (active players, event rates, queue depth)
- [ ] **7.6** Grafana dashboard: service health (latency p50/p95/p99, error rates per service)

---

## Phase 8: Frontend Game Client (Next.js)
- [ ] **8.1** Dungeon map renderer: grid of rooms, player position, fog-of-war
- [ ] **8.2** Hero stats panel: HP bar, XP, level, class
- [ ] **8.3** Inventory panel: grid of items, use button
- [ ] **8.4** Combat log: scrolling feed of events (real-time via GraphQL subscription)
- [ ] **8.5** Attack button + move controls (arrow keys or click-to-move)
- [ ] **8.6** Death screen + run summary (floors cleared, kills, score)
- [ ] **8.7** Leaderboard page (server-side rendered)
- [ ] **8.8** Responsive layout, dark dungeon theme (Tailwind)

---

## Phase 9: Kubernetes Manifests (write now, apply when cluster is back)
- [x] **9.1** Namespace: abysscore
- [ ] **9.2** RabbitMQ: deploy via RabbitMQ Cluster Operator _(currently plain Deployment with default guest:guest creds — works but not production-ready)_
- [x] **9.3** Keycloak: Deployment + Service + realm init Job (reuse KubePulse pattern)
- [x] **9.4** Postgres: per-service StatefulSets or single instance with multiple DBs _(single instance, 7 separate databases per Encore infra config)_
- [x] **9.5** Encore services: Deployment + Service + ConfigMap per service _(single backend image, infra.config.json baked at build time)_
- [x] **9.6** GraphQL gateway: Deployment + Service
- [x] **9.7** Next.js frontend: Deployment + Service + Ingress _(no Ingress — exposed via Cloudflare Tunnel instead)_
- [x] **9.8** Prometheus + Grafana: kube-prometheus-stack v84.5.0, all images mirrored to GHCR, live at grafana.bassmit.dev
- [ ] **9.9** Custom AbyssCore Grafana dashboard (game-specific panels)
- [ ] **9.10** OpenTelemetry Collector: DaemonSet or Deployment
- [ ] **9.11** Cilium network policies: restrict inter-service traffic (only game-service publishes to RabbitMQ, etc.)
- [ ] **9.12** Cilium L7 visibility policy for GraphQL traffic inspection
- [ ] **9.13** Horizontal Pod Autoscaler on combat-service (spikes during boss fights)
- [x] **9.14** ArgoCD: 3 Applications (abysscore, monitoring, monitoring-extras), all Synced/Healthy. UI at argocd.bassmit.dev.
- [x] **9.15** Cloudflare Tunnel pod (cloudflared) — public exposure without LoadBalancer/Ingress, no VPN required
- [x] **9.16** GitHub Actions CI: build + push backend/gateway/frontend images to ghcr.io on push to main
- [x] **9.17** ghcr.io image pull secret (`ghcr-secret`) provisioned in cluster

---

## Phase 10: Local Dev Polish
- [ ] **10.1** dev-start script (sh/ps1): starts docker-compose, runs Keycloak init, starts frontend
- [ ] **10.2** Seed script: create test player, generate a dungeon floor, pre-populate leaderboard
- [ ] **10.3** README.md: architecture diagram, prerequisites, local setup steps, k8s deploy steps
- [ ] **10.4** Architecture diagram (SVG or Excalidraw)

---

## Notes

### Local dev ports (docker-compose)
| Service         | Port  |
|----------------|-------|
| Frontend        | 3000  |
| GraphQL gateway | 4001  |
| Keycloak        | 8080  |
| RabbitMQ AMQP   | 5672  |
| RabbitMQ UI     | 15672 |
| Postgres        | 5432  |
| Prometheus      | 9090  |
| Grafana         | 3001  |
| Jaeger UI       | 16686 |
| OTEL Collector  | 4317  |

### RabbitMQ Event Topology
Exchange: `abysscore.events` (topic)

| Routing Key                  | Producer          | Consumer(s)                        |
|-----------------------------|-------------------|------------------------------------|
| dungeon.player.moved        | game-service      | dungeon-service (trap checks)      |
| dungeon.floor.entered       | dungeon-service   | dungeon-service (gen next floor)   |
| combat.attack.initiated     | game-service      | combat-service                     |
| combat.result               | combat-service    | game-service, graphql-gateway      |
| combat.monster.killed       | combat-service    | inventory-service, game-service    |
| game.player.died            | combat-service    | leaderboard-service, game-service  |

### Keycloak Config
- Realm: abysscore
- Client: abysscore-frontend (public, PKCE)
- Roles: player, admin
- Test user: testplayer / testpass

### Hero Classes (to implement)
- Warrior: high HP, melee damage
- Rogue: low HP, high crit chance
- Mage: low HP, AoE spells, mana resource
