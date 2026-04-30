#!/usr/bin/env bash
# AbyssCore dev-start.sh
# Starts all infra (Docker Compose) + frontend (Next.js) for local testing
# Usage: ./scripts/dev-start.sh

set -e
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${CYAN}[dev-start]${NC} $1"; }
ok()   { echo -e "${GREEN}[ok]${NC} $1"; }
warn() { echo -e "${YELLOW}[warn]${NC} $1"; }

cd "$ROOT"

# ── 1. Write frontend .env.local ─────────────────────────────────────────────
log "Writing frontend/.env.local"
cat > "$ROOT/frontend/.env.local" <<'EOF'
NEXTAUTH_URL=http://localhost:3000
NEXTAUTH_SECRET=dev-secret-change-in-prod

KEYCLOAK_ID=abysscore-frontend
KEYCLOAK_SECRET=abysscore-secret
KEYCLOAK_ISSUER=http://localhost:8080/realms/abysscore

NEXT_PUBLIC_GRAPHQL_URL=http://localhost:4000/graphql
NEXT_PUBLIC_GRAPHQL_WS_URL=ws://localhost:4000/graphql
EOF
ok "frontend/.env.local written"

# ── 2. Start Docker Compose infra ────────────────────────────────────────────
log "Starting Docker Compose (infra services)..."
docker compose up -d --remove-orphans
ok "Docker Compose up"

# ── 3. Wait for Keycloak ─────────────────────────────────────────────────────
log "Waiting for Keycloak at :8080 ..."
for i in $(seq 1 30); do
  if curl -sf http://localhost:8080/health/ready > /dev/null 2>&1; then
    ok "Keycloak ready"
    break
  fi
  echo -n "."
  sleep 3
done

# ── 4. Wait for RabbitMQ ─────────────────────────────────────────────────────
log "Waiting for RabbitMQ management at :15672 ..."
for i in $(seq 1 20); do
  if curl -sf http://localhost:15672 > /dev/null 2>&1; then
    ok "RabbitMQ ready"
    break
  fi
  echo -n "."
  sleep 2
done

# ── 5. Start Next.js frontend ─────────────────────────────────────────────────
log "Starting Next.js frontend at http://localhost:3000 ..."
cd "$ROOT/frontend"
npm run dev &
FRONTEND_PID=$!
ok "Next.js starting (pid $FRONTEND_PID)"

cd "$ROOT"

echo ""
echo -e "${GREEN}═══════════════════════════════════════════════${NC}"
echo -e "${GREEN}  AbyssCore local dev stack is running!${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════${NC}"
echo ""
echo -e "  Game:         ${CYAN}http://localhost:3000${NC}"
echo -e "  GraphQL:      ${CYAN}http://localhost:4000/graphql${NC}"
echo -e "  Keycloak:     ${CYAN}http://localhost:8080${NC}"
echo -e "  RabbitMQ UI:  ${CYAN}http://localhost:15672${NC}  (guest/guest)"
echo -e "  Prometheus:   ${CYAN}http://localhost:9090${NC}"
echo -e "  Grafana:      ${CYAN}http://localhost:3001${NC}  (admin/admin)"
echo -e "  Jaeger:       ${CYAN}http://localhost:16686${NC}"
echo ""
echo -e "${YELLOW}  Encore backend must be run separately on Windows:${NC}"
echo -e "  cd backend && encore run  (run in a Windows terminal)"
echo ""
echo -e "  GraphQL gateway: cd backend/graphql-gateway && go run ."
echo ""
echo "Press Ctrl-C to stop frontend (infra stays running)"
echo ""

wait $FRONTEND_PID
