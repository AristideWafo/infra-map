# Deployment Guide — InfraMaps

---

## Dev Local — Setup Complet (10 minutes)

### Prérequis

```bash
go version    # Go 1.22+
node --version # Node 20+
pnpm --version # pnpm 8+
docker version # Docker 24+
kind version   # kind 0.22+
kubectl version --client
helm version   # Helm 3.14+
```

### Option A — Avec cluster K8s local (kind)

```bash
# 1. Créer le cluster
kind create cluster --name infra-maps-dev

# 2. Installer Prometheus (scraper de métriques réel)
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring --create-namespace \
  --set grafana.enabled=false \
  --set alertmanager.enabled=false \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false

# 3. Port-forward Prometheus
kubectl port-forward svc/prometheus-kube-prometheus-prometheus \
  9090:9090 -n monitoring &

# 4. Démarrer le backend
cd infra-maps-api
export PROMETHEUS_URL=http://localhost:9090
export K8S_IN_CLUSTER=false
export K8S_KUBECONFIG=$HOME/.kube/config
export DOCKER_ENABLED=true
export AUTH_ENABLED=false
export LOG_LEVEL=debug
go run cmd/server/main.go

# 5. Démarrer le frontend
cd infra-maps-ui
cp .env.example .env.local  # VITE_API_URL=http://localhost:8080
pnpm install && pnpm dev    # → http://localhost:5173
```

### Option B — Docker Compose (sans K8s)

```yaml
# docker-compose.dev.yml
version: "3.9"

services:
  api:
    build:
      context: ./infra-maps-api
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      PROMETHEUS_URL: http://prometheus:9090
      K8S_ENABLED: "false"
      DOCKER_ENABLED: "true"
      DOCKER_SOCKET: /var/run/docker.sock
      LOKI_URL: ""
      AUTH_ENABLED: "false"
      LOG_LEVEL: debug
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
    depends_on:
      - prometheus

  ui:
    build:
      context: ./infra-maps-ui
      dockerfile: Dockerfile.dev
    ports:
      - "3000:3000"
    environment:
      VITE_API_URL: http://localhost:8080
    volumes:
      - ./infra-maps-ui/src:/app/src  # Hot reload

  prometheus:
    image: prom/prometheus:v2.51.0
    ports:
      - "9090:9090"
    volumes:
      - ./dev/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.retention.time=24h'

  # Optional — node_exporter pour métriques VM simulées
  node-exporter:
    image: prom/node-exporter:v1.7.0
    ports:
      - "9100:9100"
    pid: host
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
```

```yaml
# dev/prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: node
    static_configs:
      - targets: ['node-exporter:9100']
        labels:
          datacenter: local
          environment: dev
```

```bash
docker compose -f docker-compose.dev.yml up
# Backend: http://localhost:8080
# Frontend: http://localhost:3000
# Prometheus: http://localhost:9090
```

---

## Dockerfile — Multi-stage (Production)

```dockerfile
# infra-maps-api/Dockerfile

# ─── Stage 1: Build Go ───────────────────────────────────
FROM golang:1.22-alpine AS go-builder
WORKDIR /app

# Dépendances d'abord (cache layer)
COPY go.mod go.sum ./
RUN go mod download

# Sources
COPY . .

# Build statique (pas de CGO → portable)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o infra-maps-api ./cmd/server/

# ─── Stage 2: Build React ────────────────────────────────
FROM node:20-alpine AS ui-builder
WORKDIR /ui

RUN npm install -g pnpm@8

# Dépendances d'abord (cache layer)
COPY infra-maps-ui/package.json infra-maps-ui/pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Sources et build
COPY infra-maps-ui/ .
ARG VITE_API_URL=/  # Chemin relatif en prod (même domaine)
ENV VITE_API_URL=$VITE_API_URL
RUN pnpm build

# ─── Stage 3: Image finale (~25MB) ───────────────────────
FROM alpine:3.19

# Sécurité — utilisateur non-root
RUN addgroup -S infra-maps && adduser -S infra-maps -G infra-maps

WORKDIR /app

# Copier le binaire Go
COPY --from=go-builder /app/infra-maps-api .

# Copier les assets React (servis statiquement par le Go backend)
COPY --from=ui-builder /ui/dist ./static

# Certificats pour les appels HTTPS sortants
RUN apk add --no-cache ca-certificates

USER infra-maps

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/v1/health || exit 1

CMD ["./infra-maps-api"]
```

```bash
# Build et test local
docker build -t infra-maps:dev .
docker run -p 8080:8080 \
  -e PROMETHEUS_URL=http://prometheus:9090 \
  -e AUTH_ENABLED=false \
  infra-maps:dev
```

---

## Helm Chart — Deploy Client

### Installation standard

```bash
# Ajouter le repo (quand disponible)
helm repo add infra-maps https://charts.infra-maps.io
helm repo update

# Installation minimale (K8s + Prometheus détectés automatiquement)
helm install infra-maps infra-maps/infra-maps \
  --namespace infra-maps \
  --create-namespace \
  --set scrapers.prometheus.url=http://prometheus.monitoring:9090

# Installation complète avec Loki
helm install infra-maps infra-maps/infra-maps \
  --namespace infra-maps \
  --create-namespace \
  --set scrapers.prometheus.url=http://prometheus.monitoring:9090 \
  --set scrapers.loki.enabled=true \
  --set scrapers.loki.url=http://loki.monitoring:3100 \
  --set ingress.enabled=true \
  --set ingress.host=infra-maps.client.example.com

# Vérification
kubectl get pods -n infra-maps
kubectl port-forward svc/infra-maps 8080:8080 -n infra-maps
# Ouvrir http://localhost:8080
```

### values.yaml — Toutes les options

```yaml
# helm/infra-maps/values.yaml

image:
  repository: ghcr.io/yourorg/infra-maps
  tag: "latest"
  pullPolicy: IfNotPresent

replicaCount: 1  # Toujours 1 — cache en mémoire (pas de HA sans Redis)

service:
  type: ClusterIP
  port: 8080

ingress:
  enabled: false
  className: nginx
  host: infra-maps.example.com
  tls: []  # Ajouter si TLS requis

scrapers:
  prometheus:
    enabled: true
    url: "http://prometheus:9090"  # Obligatoire
    interval: "30s"
  kubernetes:
    enabled: true
    inCluster: true  # true = ServiceAccount, false = kubeconfig secret
  docker:
    enabled: false  # Désactivé par défaut (socket Docker non dispo dans K8s)
    socketPath: /var/run/docker.sock
  loki:
    enabled: false
    url: ""
    maxLines: 100

auth:
  enabled: true  # Désactiver pour POC uniquement

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 256Mi

rbac:
  create: true  # Crée ServiceAccount + ClusterRole en lecture seule

# RBAC K8s — permissions minimales requises
rbac_rules:
  - apiGroups: [""]
    resources: ["nodes", "pods", "services", "endpoints", "namespaces"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["ingresses", "networkpolicies"]
    verbs: ["get", "list"]
  - apiGroups: ["apps"]
    resources: ["deployments", "replicasets", "statefulsets", "daemonsets"]
    verbs: ["get", "list"]
```

---

## Mode Air-gapped (Clients Sans Internet)

```bash
# 1. Construire l'image sur une machine avec accès internet
docker build -t infra-maps:v1.0.0 .

# 2. Exporter l'image
docker save infra-maps:v1.0.0 | gzip > infra-maps-v1.0.0.tar.gz

# 3. Transférer chez le client
scp infra-maps-v1.0.0.tar.gz admin@client-server:/tmp/

# 4. Charger sur le serveur client
ssh admin@client-server
docker load < /tmp/infra-maps-v1.0.0.tar.gz

# 5. Exporter le chart Helm
helm package ./helm/infra-maps
scp infra-maps-1.0.0.tgz admin@client-server:/tmp/

# 6. Installer (image locale)
helm install infra-maps /tmp/infra-maps-1.0.0.tgz \
  --namespace infra-maps --create-namespace \
  --set image.repository=infra-maps \
  --set image.tag=v1.0.0 \
  --set image.pullPolicy=Never \
  --set scrapers.prometheus.url=http://prometheus.monitoring:9090
```

---

## Variables d'Environnement — Référence Complète

| Variable | Défaut | Obligatoire | Description |
|----------|--------|-------------|-------------|
| `PORT` | `8080` | Non | Port HTTP du serveur |
| `PROMETHEUS_URL` | — | **Oui** | URL Prometheus. Ex: `http://prometheus:9090` |
| `K8S_ENABLED` | `true` | Non | Activer le scraper K8s |
| `K8S_IN_CLUSTER` | `true` | Non | `true` = ServiceAccount K8s, `false` = kubeconfig |
| `K8S_KUBECONFIG` | `~/.kube/config` | Non | Chemin kubeconfig (si IN_CLUSTER=false) |
| `DOCKER_ENABLED` | `true` | Non | Activer le scraper Docker |
| `DOCKER_SOCKET` | `/var/run/docker.sock` | Non | Chemin du socket Docker |
| `LOKI_URL` | `` | Non | URL Loki. Vide = désactivé. |
| `LOKI_MAX_LINES` | `100` | Non | Limite logs par requête proxy |
| `SCRAPE_INTERVAL` | `30s` | Non | Intervalle scraping (format Go duration) |
| `CACHE_TTL` | `30s` | Non | TTL cache mémoire |
| `AUTH_ENABLED` | `true` | Non | `false` pour dev local uniquement |
| `LOG_LEVEL` | `info` | Non | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | Non | `json` pour prod, `text` pour dev |

---

## CI/CD — GitHub Actions

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main, develop]
  pull_request:

jobs:
  test-backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - name: Test
        run: cd infra-maps-api && go test ./... -race -coverprofile=coverage.out
      - name: Build
        run: cd infra-maps-api && go build ./cmd/server/...

  test-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
        with: { version: 8 }
      - uses: actions/setup-node@v4
        with: { node-version: '20', cache: 'pnpm', cache-dependency-path: infra-maps-ui/pnpm-lock.yaml }
      - run: cd infra-maps-ui && pnpm install --frozen-lockfile
      - run: cd infra-maps-ui && pnpm test --run
      - run: cd infra-maps-ui && pnpm build

  docker:
    needs: [test-backend, test-frontend]
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - uses: docker/login-action@v3
        with:
          registry: ghcr.io
          username: ${{ github.actor }}
          password: ${{ secrets.GITHUB_TOKEN }}
      - uses: docker/build-push-action@v5
        with:
          push: true
          tags: ghcr.io/${{ github.repository }}:latest,ghcr.io/${{ github.repository }}:${{ github.sha }}
```
