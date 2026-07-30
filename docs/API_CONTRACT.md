# API Contract — InfraMaps

> Toute modification de ce contrat doit être rétrocompatible jusqu'à la prochaine version majeure.
> Les agents ne doivent pas ajouter d'endpoints non documentés ici.

---

## Conventions Générales

| Convention | Valeur |
|-----------|--------|
| Base URL | `/api/v1` |
| Format | JSON (`Content-Type: application/json`) |
| Pagination | `?page=1&limit=50` (défaut: page=1, limit=50, max limit=200) |
| Authentification | `Authorization: Bearer <k8s-sa-token>` — désactivée si `AUTH_ENABLED=false` |
| Cache-Control | `Cache-Control: max-age=25` (légèrement inférieur au TTL scraper de 30s) |
| Timezone | UTC partout |
| Mémoire | MB dans tous les champs |

### Format des Erreurs

```json
{
  "error": "description lisible",
  "code": "ERR_SCRAPER_UNAVAILABLE",
  "details": {}
}
```

### Codes d'Erreur

| Code | HTTP | Description |
|------|------|-------------|
| `ERR_NODE_NOT_FOUND` | 404 | NodeID inexistant dans le cache |
| `ERR_CACHE_EMPTY` | 503 | Cache non encore populé (démarrage) |
| `ERR_SCRAPER_UNAVAILABLE` | 503 | Source indisponible, données périmées |
| `ERR_PROMETHEUS_TIMEOUT` | 504 | Timeout proxy Prometheus |
| `ERR_LOKI_DISABLED` | 404 | Loki non configuré |
| `ERR_INVALID_PARAM` | 400 | Paramètre invalide |

---

## REST Endpoints

---

### `GET /api/v1/tree`

Retourne l'arbre complet de l'infrastructure normalisé.

**Query params :**

| Param | Type | Défaut | Description |
|-------|------|--------|-------------|
| `maxDepth` | int | illimité | Profondeur max de l'arbre retourné |
| `namespace` | string | — | Filtrer les pods par namespace K8s |
| `tag` | string | — | Filtrer les nœuds ayant ce tag |
| `status` | string | — | Filtrer par statut (`healthy`, `warning`, `critical`) |
| `source` | string | — | Filtrer par source (`kubernetes`, `docker`, `prometheus-targets`) |

**Réponse 200 :**

```json
{
  "id": "root",
  "name": "Infrastructure",
  "type": "root",
  "status": "warning",
  "x": 0, "y": 0, "z": 0,
  "source": "internal",
  "lastSeen": "2026-07-21T10:00:00Z",
  "children": [
    {
      "id": "cluster-prod",
      "name": "cluster-production",
      "type": "cluster",
      "status": "healthy",
      "cpu": 42.5,
      "memory": 8192,
      "memoryTotal": 16384,
      "pods": 47,
      "x": 0, "y": 0, "z": 0,
      "source": "kubernetes",
      "lastSeen": "2026-07-21T10:00:00Z",
      "children": [ ... ]
    },
    {
      "id": "vm-island-dc1",
      "name": "Datacenter Paris",
      "type": "vm-island",
      "status": "warning",
      "x": 700, "y": 0, "z": 0,
      "source": "prometheus-targets",
      "lastSeen": "2026-07-21T10:00:00Z",
      "children": [ ... ]
    }
  ]
}
```

**Réponse 503 (cache vide) :**
```json
{ "error": "Cache not ready, scrapers still initializing", "code": "ERR_CACHE_EMPTY" }
```

---

### `GET /api/v1/tree/{nodeId}`

Retourne le sous-arbre depuis un nœud spécifique (navigation drill-down).

**Path params :** `nodeId` — ID du nœud racine du sous-arbre.

**Query params :** mêmes filtres que `/api/v1/tree` + `maxDepth`.

**Réponse 200 :** `UnifiedNode` avec ses `children`.

**Réponse 404 :**
```json
{ "error": "Node 'pod-xyz' not found in cache", "code": "ERR_NODE_NOT_FOUND" }
```

---

### `GET /api/v1/nodes/{nodeId}`

Retourne les détails d'un nœud sans ses enfants.

**Réponse 200 :** `UnifiedNode` sans champ `children`.

---

### `GET /api/v1/nodes/{nodeId}/metrics`

Proxy vers Prometheus — retourne les métriques historiques d'un nœud.

**Query params :**

| Param | Type | Défaut | Description |
|-------|------|--------|-------------|
| `from` | unix timestamp | now - 1h | Début de la plage |
| `to` | unix timestamp | now | Fin de la plage |
| `step` | string | `60s` | Résolution (`30s`, `60s`, `5m`) |
| `metric` | string | all | Filtrer: `cpu`, `memory`, `disk`, `restarts` |

**Réponse 200 :**

```json
{
  "nodeId": "worker-01",
  "from": "2026-07-21T09:00:00Z",
  "to": "2026-07-21T10:00:00Z",
  "step": "60s",
  "series": {
    "cpu": [
      { "timestamp": "2026-07-21T09:00:00Z", "value": 42.1 },
      { "timestamp": "2026-07-21T09:01:00Z", "value": 38.7 }
    ],
    "memory": [ ... ],
    "disk": [ ... ]
  }
}
```

**Réponse 504 :** Timeout Prometheus.

---

### `GET /api/v1/nodes/{nodeId}/logs`

Proxy vers Loki — retourne les logs récents d'un pod/container.

**Query params :**

| Param | Type | Défaut | Description |
|-------|------|--------|-------------|
| `limit` | int | 100 | Max lignes (max: 500) |
| `from` | unix timestamp | now - 15min | Début |
| `level` | string | all | Filtrer: `error`, `warn`, `info` |

**Réponse 200 :**

```json
{
  "nodeId": "pod-api-1",
  "entries": [
    {
      "timestamp": "2026-07-21T09:59:45Z",
      "level": "error",
      "message": "Connection refused to postgres:5432",
      "pod": "api-frontend-1",
      "container": "api"
    }
  ]
}
```

**Réponse 404 :** Loki non configuré (`ERR_LOKI_DISABLED`).

---

### `GET /api/v1/connections`

Retourne les connexions réseau découvertes.

**Query params :**

| Param | Type | Défaut | Description |
|-------|------|--------|-------------|
| `fromId` | string | — | Filtrer par nœud source |
| `toId` | string | — | Filtrer par nœud destination |
| `type` | string | — | `service`, `ingress`, `docker-network` |

**Réponse 200 :**

```json
{
  "connections": [
    {
      "id": "svc-api-pod-api-1",
      "fromId": "svc-api",
      "toId": "pod-api-1",
      "traffic": 0,
      "latency": 0,
      "errors": 0,
      "protocol": "tcp",
      "type": "service"
    }
  ],
  "total": 12
}
```

---

### `GET /api/v1/tags`

Retourne la liste des tags disponibles pour les filtres UI.

**Réponse 200 :**

```json
{
  "tags": ["prod", "staging", "db", "api", "monitoring"]
}
```

---

### `GET /api/v1/health`

Health check — état des scrapers et du cache.

**Réponse 200 (normal) :**

```json
{
  "status": "ok",
  "scrapers": {
    "prometheus": { "status": "ok" },
    "kubernetes": { "status": "ok" },
    "docker": { "status": "disabled" },
    "loki": { "status": "disabled" }
  },
  "cacheAgeSeconds": 12
}
```

**Réponse 200 (dégradé — données périmées mais service up) :**

```json
{
  "status": "degraded",
  "scrapers": {
    "prometheus": { "status": "error", "message": "connection timeout after 5s" },
    "kubernetes": { "status": "ok" }
  },
  "cacheAgeSeconds": 87
}
```

---

## WebSocket

### `WS /api/v1/ws/alerts`

Stream d'alertes en temps réel. Le client se connecte et reçoit les alertes Prometheus dès qu'elles arrivent.

**Connexion :**
```
ws://infra-maps:8080/api/v1/ws/alerts
Authorization: Bearer <token>  (header HTTP lors du handshake)
```

**Message — Alerte déclenchée :**

```json
{
  "type": "alert_fired",
  "payload": {
    "id": "alert-pod-crash-abc123",
    "nodeId": "pod-api-1",
    "name": "PodCrashLooping",
    "severity": "critical",
    "message": "Pod api-frontend-1 has been restarting 5 times in 10 minutes",
    "firedAt": "2026-07-21T09:59:00Z"
  }
}
```

**Message — Alerte résolue :**

```json
{
  "type": "alert_resolved",
  "payload": {
    "id": "alert-pod-crash-abc123",
    "nodeId": "pod-api-1",
    "resolvedAt": "2026-07-21T10:05:00Z"
  }
}
```

**Message — Ping (keepalive toutes les 30s) :**
```json
{ "type": "ping" }
```

**Comportement en cas d'erreur :** Le client doit implémenter une reconnexion avec backoff exponentiel (1s, 2s, 4s, 8s, max 30s).

---

## Variables d'Environnement Backend

| Variable | Défaut | Description |
|----------|--------|-------------|
| `PORT` | `8080` | Port HTTP |
| `PROMETHEUS_URL` | `http://prometheus:9090` | URL Prometheus (obligatoire) |
| `K8S_ENABLED` | `true` | Activer scraper K8s |
| `K8S_IN_CLUSTER` | `true` | `true` = ServiceAccount, `false` = kubeconfig |
| `K8S_KUBECONFIG` | `~/.kube/config` | Chemin kubeconfig (si IN_CLUSTER=false) |
| `DOCKER_ENABLED` | `true` | Activer scraper Docker |
| `DOCKER_SOCKET` | `/var/run/docker.sock` | Chemin socket Docker |
| `LOKI_URL` | `` | URL Loki (vide = désactivé) |
| `LOKI_MAX_LINES` | `100` | Limite logs par requête |
| `SCRAPE_INTERVAL` | `30s` | Intervalle de scraping |
| `CACHE_TTL` | `30s` | TTL cache mémoire |
| `AUTH_ENABLED` | `true` | Désactiver pour dev local |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
