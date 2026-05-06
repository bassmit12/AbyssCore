# AbyssCore - Progress Log

> Live operational status. PLAN.md = what to build. PROGRESS.md = what's deployed, what broke, what's next.

**Production URL:** https://abysscore.bassmit.dev
**Last updated:** 2026-05-06

---

## Current Status

✅ **App is LIVE on the public internet.** Login (testplayer / password) works, hero creation → combat flow is functional. Card artwork for Ball Lightning and Chaos Theory is live. Version badge in header shows commit SHA.

### Deployed Components

| Component | Status | Notes |
|---|---|---|
| Postgres | ✅ Running | Single instance, 8 DBs (abysscore, keycloak, combat, deck, dungeon, game, inventory, leaderboard, map) |
| RabbitMQ | ✅ Running | Default `guest:guest`, 3-management-alpine |
| Keycloak | ✅ Running | Realm `abysscore`, client `abysscore-frontend`, user `testplayer / password` |
| Encore backend | ✅ Running | All 7 services in one image, port 4000 |
| GraphQL gateway | ✅ Running | Port 4001, playground at `/playground` |
| Next.js frontend | ✅ Running | Port 3000, NextAuth + Keycloak |
| Cloudflared tunnel | ✅ Running | 2 replicas, dialing out to Cloudflare edge (Amsterdam) |
| ArgoCD | ✅ Running | 3 Applications (abysscore, monitoring, monitoring-extras), all Synced/Healthy |
| Prometheus + Grafana | ✅ Running | kube-prometheus-stack v84.5.0, all pods healthy |

### Public Hostnames (Cloudflare Tunnel)

| Subdomain | Routes To | Use |
|---|---|---|
| https://abysscore.bassmit.dev | frontend:3000 | Game UI |
| https://abysscore-api.bassmit.dev | gateway:4001 | GraphQL playground |
| https://abysscore-auth.bassmit.dev | keycloak:8080 | Keycloak admin (admin/admin) |
| https://argocd.bassmit.dev | argocd-server:80 | GitOps dashboard |
| https://grafana.bassmit.dev | kube-prometheus-stack-grafana:80 | Metrics dashboard |

---

## Cluster Access

**Kubeconfig:** `abysscore_cluster.conf` at project root (gitignored — DO NOT commit).
**Cluster:** Fontys-managed Kubernetes (k8s v1.33.1, 3 nodes). Node IP: 10.1.1.158.
**VPN required for kubectl?** Yes (Fontys OpenVPN). Cloudflare Tunnel makes the *app* public, but kubectl talks to the API server on the private subnet.

```bash
KUBECONFIG="/mnt/d/Fontys/Personal Project/AbyssCore/abysscore_cluster.conf" kubectl get pods -n abysscore
```

⚠️ **Known leak:** the kubeconfig was committed to git history (commit 004f076) before being removed. The cert is still valid until rotated. Flag for supervisor or accept the risk on this dev cluster.

---

## CI/CD (current state)

**Fully automated** (GitHub Actions on push to `main`):

1. Build backend image → push `ghcr.io/bassmit12/abysscore-backend:<sha>` + `:latest`
2. Build gateway image → push `ghcr.io/bassmit12/abysscore-gateway:<sha>` + `:latest`
3. Build frontend image → push `ghcr.io/bassmit12/abysscore-frontend:<sha>` + `:latest`
4. `update-manifests` job patches `infra/k8s/*.yaml` with the new SHA tag and commits back to git
5. ArgoCD detects the manifest change and rolls out the new images automatically

**No manual steps needed** — pushing to `main` is the full deploy.

The `paths-ignore: infra/k8s/**` filter on the trigger prevents the manifest-update commit from triggering a new build (infinite loop fix).

**Auth:** uses repo secret `GHCR_TOKEN` (PAT with `write:packages` scope). `GITHUB_TOKEN` fails 403 for new packages.

**DB migrations:** completely manual. See "How to apply migrations" below.

---

## Next Up

### 1. Application metrics
- Add `"metrics": {"type": "prometheus"}` to backend `infra.config.json`
- Add `promhttp.Handler()` at `/metrics` in the gateway
- ServiceMonitors already exist — once the endpoints return real data, Grafana will pick it up

### 2. Custom Grafana dashboard
- AbyssCore-specific panels: dungeons created, combats resolved, queue depth, API latency per service

### 3. Alerting
- Alertmanager is deployed but has no routing rules
- Basic alerts: pod restarts, high memory, service unavailability

### 4. Log aggregation
- Loki stack for historical log search and correlation with metrics

### 5. HPA for backend and gateway
- CPU target ~70%, min 1 replica, max 5
- Resource requests already defined on all Deployments ✅

### 6. Test environment (develop branch)
- Two branches (`develop` → `main`), two namespaces (`abysscore-test` → `abysscore`)
- Kustomize overlays for per-environment patches
- Separate Cloudflare hostnames for test (`test.abysscore.bassmit.dev`)

---

## How To: Apply Encore migrations

Encore self-hosted Docker mode does **NOT auto-run migrations**. Run this from a machine with kubectl access:

```python
import os, subprocess
base = "/mnt/d/Fontys/Personal Project/AbyssCore/backend"
kubeconfig = "/mnt/d/Fontys/Personal Project/AbyssCore/abysscore_cluster.conf"
services = {
    "game-service": "game", "dungeon-service": "dungeon",
    "combat-service": "combat", "inventory-service": "inventory",
    "leaderboard-service": "leaderboard", "deck-service": "deck",
    "map-service": "map",
}
for svc, db in services.items():
    mig_dir = os.path.join(base, svc, "migrations")
    for f in sorted(x for x in os.listdir(mig_dir) if x.endswith(".up.sql")):
        sql = open(os.path.join(mig_dir, f)).read()
        subprocess.run([
            "kubectl", f"--kubeconfig={kubeconfig}",
            "exec", "-n", "abysscore", "-i", "deployment/postgres", "--",
            "psql", "-U", "abysscore", "-d", db, "-v", "ON_ERROR_STOP=1",
        ], input=sql, check=True, text=True)
        print(f"[OK] {svc}/{f}")
```

Migrations are idempotent if written with `IF NOT EXISTS` / `IF EXISTS`. Re-running on an already-migrated DB is safe.

---

## Known Issues / Pitfalls (project-specific)

### Fontys cluster TLS interception breaks external registry pulls

The Fontys network does TLS inspection. Pulls from `docker.io`, `quay.io`, etc. fail with x509 cert errors.

**Workaround:** mirror every image to ghcr.io. All app images and kube-prometheus-stack images are mirrored to `ghcr.io/bassmit12/`. Any new image used in the cluster must be mirrored first.

### Encore Container PORT defaults to 8080
`encore build docker` images listen on 8080 unless `PORT` env is set. Set `PORT=4000` on the backend Deployment.

### Encore infra.config.json schema
- `username` / `password` are **plain strings** or `{\"$env\": \"VAR\"}` — NOT `{\"value\": \"...\"}`.
- `databases` map keys = literal Postgres database names. Each DB must exist.

### Next.js standalone HOSTNAME
Standalone build defaults to listening on the pod hostname only. Set `HOSTNAME=0.0.0.0` and `PORT=3000` in the Dockerfile.

### Next.js `next/image` with `fill` broken in standalone builds
Use plain `<img>` tags for local public assets (`/cards/*.png`). The `fill` prop on `next/image` breaks in standalone mode.

### NextAuth in k8s
- Set `AUTH_TRUST_HOST=true` — without it, NextAuth refuses non-localhost hostnames.
- `NEXTAUTH_SECRET` and `KEYCLOAK_*` must be in the Deployment env, not just `.env.local`.

### Keycloak behind Cloudflare Tunnel
Need `KC_HOSTNAME=<public-host>`, `KC_HOSTNAME_STRICT=false`, `KC_HOSTNAME_STRICT_HTTPS=false`, `KC_PROXY=edge`.

### Backend KEYCLOAK_ISSUER
Must point at the **public** issuer URL (tokens are issued with that as the `iss` claim).

### CI infinite loop from manifest-update job
The `update-manifests` job commits back to `main`. Without `paths-ignore: infra/k8s/**` on the trigger, this causes infinite build loops. Fixed — do not remove that filter.

### GHCR_TOKEN vs GITHUB_TOKEN
`GITHUB_TOKEN` fails 403 for first-time package creation. Always use `GHCR_TOKEN` (PAT with `write:packages`).

---

## Operational Recipes

### Check what image a pod is running
```bash
kubectl describe pod -n abysscore -l app=abysscore-frontend | grep "Image:"
```

### Tail backend logs (filter trace noise)
```bash
kubectl logs -n abysscore deployment/abysscore-backend --tail=50 | grep -iE "error|fail" | grep -v "trace"
```

### Connect to a service DB directly
```bash
kubectl exec -n abysscore -it deployment/postgres -- psql -U abysscore -d game
```

### Recreate ghcr image pull secret (if rotated)
```bash
kubectl create secret docker-registry ghcr-secret \
  --namespace abysscore \
  --docker-server=ghcr.io \
  --docker-username=bassmit12 \
  --docker-password=<NEW_PAT>
```

---

## Bootstrap History

| Date | Event |
|---|---|
| 2026-05-02 | Cluster came online (Fontys-managed). Initial deploy via kubectl apply. |
| 2026-05-02 | GitHub Actions CI configured for image builds (ghcr.io). |
| 2026-05-03 | Cloudflare Tunnel deployed, public URLs live. |
| 2026-05-03 | Encore infra.config.json schema fix + 7-DB postgres setup + migrations applied. |
| 2026-05-03 | Full end-to-end auth → hero creation flow verified working. |
| 2026-05-03 | ArgoCD deployed, 3 Applications Synced/Healthy. |
| 2026-05-03 | kube-prometheus-stack deployed, all images mirrored to GHCR. Grafana live. |
| 2026-05-04 | CI switched to SHA-tagged images + update-manifests job for ArgoCD auto-rollout. |
| 2026-05-04 | Card artwork (Ball Lightning, Chaos Theory, draw pile) added to frontend. |
| 2026-05-04 | Version badge (commit SHA) added to frontend header. |
| 2026-05-06 | Cards shown as full art assets (no wrapper). CI infinite loop fixed with paths-ignore. |
