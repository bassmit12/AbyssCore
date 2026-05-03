# AbyssCore - Progress Log

> Live operational status. PLAN.md = what to build. PROGRESS.md = what's deployed, what broke, what's next.

**Production URL:** https://abysscore.bassmit.dev
**Last updated:** 2026-05-03

---

## Current Status

✅ **App is LIVE on the public internet.** Login (testplayer / password) works, hero creation flow reaches the database.

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

### Public Hostnames (Cloudflare Tunnel)

| Subdomain | Routes To | Use |
|---|---|---|
| https://abysscore.bassmit.dev | abysscore-frontend.abysscore.svc.cluster.local:3000 | Game UI |
| https://abysscore-api.bassmit.dev | abysscore-gateway.abysscore.svc.cluster.local:4001 | GraphQL playground |
| https://abysscore-auth.bassmit.dev | keycloak.abysscore.svc.cluster.local:8080 | Keycloak admin (admin/admin) |

---

## Cluster Access

**Kubeconfig:** `abysscore_cluster.conf` at project root (gitignored — DO NOT commit).
**VPN required for kubectl?** Yes (Fontys OpenVPN). Cloudflare Tunnel makes the *app* public, but kubectl talks to the API server on the private subnet.

```bash
KUBECONFIG="/mnt/d/Fontys/Personal Project/AbyssCore/abysscore_cluster.conf" kubectl get pods -n abysscore
```

⚠️ **Known leak:** the kubeconfig was committed to git history (commit 004f076) before being removed. The cert is still valid until rotated. Multistax-managed certs aren't trivially rotatable — flag for supervisor or accept the risk on this dev cluster.

---

## CI/CD (current state)

**Automated** (GitHub Actions on push to `main`):
1. Build backend image → push to `ghcr.io/bassmit12/abysscore-backend:latest`
2. Build gateway image → push to `ghcr.io/bassmit12/abysscore-gateway:latest`
3. Build frontend image → push to `ghcr.io/bassmit12/abysscore-frontend:latest`

The frontend build bakes `NEXT_PUBLIC_*` URLs into the image — these are hardcoded in `.github/workflows/build-push.yml`. If hostnames change, edit the workflow.

**Auth:** uses repo secret `CR_PAT` (PAT with `write:packages` scope). `GITHUB_TOKEN` was unreliable for first-time package creation.

**Manual** (must run after CI):
```bash
KUBECONFIG=... kubectl rollout restart deployment/abysscore-backend -n abysscore
KUBECONFIG=... kubectl rollout restart deployment/abysscore-frontend -n abysscore
KUBECONFIG=... kubectl rollout restart deployment/abysscore-gateway -n abysscore
```
Without rollout restart, k8s caches the `:latest` tag and won't pull the new digest.

**DB migrations:** completely manual. See "How to apply migrations" below.

---

## Next Up

### 1. ArgoCD (planned)
Replace the manual `rollout restart` step with GitOps reconciliation.
- Install ArgoCD into `argocd` namespace
- Expose UI at `argocd.bassmit.dev` via the existing tunnel
- Create Application pointing at `infra/k8s/` in this repo
- Configure auto-sync + image updater for `:latest` rollouts on new digests

### 2. Production hardening
- RabbitMQ Cluster Operator instead of plain `guest:guest` Deployment
- Replace hardcoded secrets (`abysscore:abysscore`, `dev-secret-change-in-prod`) with k8s Secrets / sealed-secrets
- Postgres backups
- Resource limits/requests on all deployments

### 3. Observability stack (Phase 6/7 of PLAN.md)
- kube-prometheus-stack Helm chart
- Jaeger operator
- OpenTelemetry collector
- Wire `OTEL_EXPORTER_OTLP_ENDPOINT` env var to point at the collector instead of `localhost:4317`

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

Migrations are idempotent if written with `IF NOT EXISTS` / `IF EXISTS`. Re-running on an already-migrated DB is safe but each migration must be idempotent.

---

## Known Issues / Pitfalls (project-specific)

### Fontys cluster TLS interception breaks Docker Hub pulls

The Fontys network does TLS inspection with rotating "Elsevier" / "Buyerzone" certs. Pulls from `docker.io` randomly fail with `tls: failed to verify certificate: x509: certificate is valid for *.elsst.com, not registry-1.docker.io`.

**Workaround:** mirror third-party images to ghcr.io. Done for `cloudflare/cloudflared` → `ghcr.io/bassmit12/cloudflared:latest`.

For init containers that just need to wait for Postgres: just remove them. Postgres is up in seconds and Encore retries connect.

### Encore Container PORT defaults to 8080
`encore build docker` images listen on 8080 unless `PORT` env is set. Backend Service expects port 4000, so set `PORT=4000` on the Deployment.

### Encore infra.config.json schema
- `username` / `password` are **plain strings** or `{"$env": "VAR"}` — NOT `{"value": "..."}`. Wrong schema makes Encore parse the struct as a connection string and fail with `password authentication failed for user "password="`.
- `databases` map keys = literal Postgres database names. Each DB must exist.

### Next.js standalone HOSTNAME
Standalone build defaults to listening on the pod hostname only. Set `HOSTNAME=0.0.0.0` and `PORT=3000` in the Dockerfile, or `kubectl port-forward` will fail with `connection refused`.

### NextAuth in k8s
- Set `AUTH_TRUST_HOST=true` (or `NEXTAUTH_URL`) — without it, NextAuth refuses to handle requests from non-localhost hostnames.
- `NEXTAUTH_SECRET` and `KEYCLOAK_*` env vars are read at runtime — must be in the Deployment, not just `.env.local`.

### Keycloak behind Cloudflare Tunnel
Need `KC_HOSTNAME=<public-host>`, `KC_HOSTNAME_STRICT=false`, `KC_HOSTNAME_STRICT_HTTPS=false`, `KC_PROXY=edge` so OIDC discovery returns the public URL and Keycloak trusts the X-Forwarded-* headers from cloudflared.

### Backend KEYCLOAK_ISSUER
The Encore auth handler reads `KEYCLOAK_ISSUER` at runtime to fetch JWKS. Must point at the **public** issuer URL since tokens are issued with that as the `iss` claim. Set on backend Deployment.

### Image deployments are stuck on `bassmit123/...` Docker Hub paths in some manifests
Some old manifests still referenced an unrelated `bassmit123/...` Docker Hub repo. All have been corrected to `ghcr.io/bassmit12/...`. If a deploy ever falls back to Docker Hub, re-apply the YAML manifest from `infra/k8s/`.

---

## Operational Recipes

### Force-pull new `:latest` image
```bash
kubectl rollout restart deployment/abysscore-frontend -n abysscore
```

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
| 2026-05-02 | Cluster came online (Multistax). Initial deploy via kubectl apply. |
| 2026-05-02 | GitHub Actions CI configured for image builds. |
| 2026-05-03 | Cloudflare Tunnel deployed, public URLs live. |
| 2026-05-03 | Encore infra.config.json schema fix + 7-DB postgres setup + migrations applied. |
| 2026-05-03 | Full end-to-end auth → hero creation flow verified working. |
