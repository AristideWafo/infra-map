# Data Models — Source de Vérité

> Ce document définit les types canoniques. En cas de divergence entre Go et TypeScript, **Go fait foi**.
> Toute modification de type doit être répercutée dans les deux langages.

---

## Règles Globales

- **Mémoire** : toujours en **MB** dans l'API. Jamais en bytes, jamais en GB.
- **CPU** : pourcentage float64, valeurs 0.0–100.0.
- **IDs** : string libre. Convention par source :
  - VM → hostname ou `instance` label Prometheus (ex: `192.168.1.10:9100`)
  - K8s node/pod/service → UID Kubernetes (ex: `550e8400-e29b-41d4-a716-446655440000`)
  - Docker container → ContainerID[:12] (ex: `abc123def456`)
  - Nœuds virtuels → string préfixée (ex: `cluster-prod`, `vm-island-dc1`)
- **Timestamps** : RFC3339 dans l'API JSON, `time.Time` en Go.
- **Champs optionnels** : `omitempty` en Go, `| undefined` en TypeScript. Ne pas envoyer `null`.

---

## Go — Structs Canoniques

### `models/unified.go`

```go
package models

import "time"

// NodeType identifie le rôle d'un nœud dans la hiérarchie
type NodeType string

const (
    NodeTypeRoot              NodeType = "root"           // Racine de l'arbre
    NodeTypeVMIsland          NodeType = "vm-island"      // Groupe de VMs (par datacenter/env)
    NodeTypeDockerComposeGroup NodeType = "compose-group" // Groupe de containers Docker Compose
    NodeTypeCluster           NodeType = "cluster"        // Cluster Kubernetes
    NodeTypeNode              NodeType = "node"            // Nœud Kubernetes (worker)
    NodeTypePod               NodeType = "pod"             // Pod Kubernetes
    NodeTypeVM                NodeType = "vm"              // VM bare metal (via node_exporter)
    NodeTypeContainer         NodeType = "container"       // Container Docker standalone
)

// NodeStatus représente l'état de santé calculé d'un nœud
type NodeStatus string

const (
    StatusHealthy  NodeStatus = "healthy"  // CPU <70% ET mémoire <80% ET 0 restart
    StatusWarning  NodeStatus = "warning"  // CPU 70-85% OU mémoire 80-90% OU restarts >0
    StatusCritical NodeStatus = "critical" // CPU >85% OU mémoire >90% OU pod CrashLoopBackOff
    StatusUnknown  NodeStatus = "unknown"  // Source indisponible ou données manquantes
)

// UnifiedNode est la structure normalisée pour toutes les sources
type UnifiedNode struct {
    // Identité
    ID     string   `json:"id"`
    Name   string   `json:"name"`
    Type   NodeType `json:"type"`
    Status NodeStatus `json:"status"`
    Tags   []string `json:"tags,omitempty"`

    // Métriques (en MB pour la mémoire, % pour CPU/Disk)
    CPU         *float64 `json:"cpu,omitempty"`         // % 0–100
    Memory      *float64 `json:"memory,omitempty"`      // MB utilisés
    MemoryTotal *float64 `json:"memoryTotal,omitempty"` // MB total
    Disk        *float64 `json:"disk,omitempty"`         // % 0–100
    Pods        *int     `json:"pods,omitempty"`         // Nombre de pods (K8s nodes)
    Restarts    *int     `json:"restarts,omitempty"`     // Redémarrages (pods/containers)

    // Position 2D/3D (calculée par Layout Engine, stable)
    X float64 `json:"x"`
    Y float64 `json:"y"`
    Z float64 `json:"z"` // 0 en mode Grid 2D

    // Hiérarchie
    ParentID string          `json:"parentId,omitempty"`
    Children []*UnifiedNode  `json:"children,omitempty"`

    // Provenance
    Source string `json:"source"` // "prometheus", "kubernetes", "docker", "prometheus-targets"

    // Métadonnées brutes (labels K8s, labels Prometheus, labels Docker)
    RawLabels map[string]string `json:"rawLabels,omitempty"`

    // Timestamps
    LastSeen  time.Time `json:"lastSeen"`
    CreatedAt time.Time `json:"createdAt,omitempty"` // Disponible pour K8s

    // Namespace (K8s uniquement)
    Namespace string `json:"namespace,omitempty"`
}

// Connection représente un lien réseau entre deux nœuds
type Connection struct {
    ID       string  `json:"id"`       // "{fromId}-{toId}"
    FromID   string  `json:"fromId"`   // ID du nœud source
    ToID     string  `json:"toId"`     // ID du nœud destination
    Traffic  float64 `json:"traffic"`  // MB/s — 0 sans service mesh
    Latency  float64 `json:"latency"`  // ms — 0 sans service mesh
    Errors   float64 `json:"errors"`   // erreurs/s — 0 sans service mesh
    Protocol string  `json:"protocol"` // "tcp", "http", "grpc"
    Type     string  `json:"type"`     // "service", "ingress", "network-policy", "docker-network"
}

// LogEntry représente une entrée de log (depuis Loki)
type LogEntry struct {
    Timestamp time.Time `json:"timestamp"`
    Level     string    `json:"level"`     // "info", "warn", "error", "debug"
    Message   string    `json:"message"`
    Pod       string    `json:"pod,omitempty"`
    Container string    `json:"container,omitempty"`
    Node      string    `json:"node,omitempty"`
}

// Alert représente une alerte Prometheus en cours
type Alert struct {
    ID       string    `json:"id"`
    NodeID   string    `json:"nodeId"`   // ID du nœud concerné
    Name     string    `json:"name"`     // Nom de l'alerte Prometheus
    Severity string    `json:"severity"` // "warning", "critical"
    Message  string    `json:"message"`
    FiredAt  time.Time `json:"firedAt"`
}

// HealthStatus état de santé des scrapers
type HealthStatus struct {
    Status   string                    `json:"status"` // "ok", "degraded", "error"
    Scrapers map[string]ScraperHealth  `json:"scrapers"`
    CacheAge int                       `json:"cacheAgeSeconds"`
}

type ScraperHealth struct {
    Status  string `json:"status"`  // "ok", "disabled", "error"
    Message string `json:"message,omitempty"`
}
```

---

## TypeScript — Interfaces Miroir

### `src/types/infra.ts`

```typescript
// === Enums ===

export type NodeType =
  | 'root'
  | 'vm-island'
  | 'compose-group'
  | 'cluster'
  | 'node'
  | 'pod'
  | 'vm'
  | 'container'

export type NodeStatus = 'healthy' | 'warning' | 'critical' | 'unknown'

// === Nœuds ===

export interface UnifiedNode {
  // Identité
  id: string
  name: string
  type: NodeType
  status: NodeStatus
  tags?: string[]

  // Métriques (MB pour mémoire, % pour CPU/Disk)
  cpu?: number
  memory?: number
  memoryTotal?: number
  disk?: number
  pods?: number
  restarts?: number

  // Position (stable entre les refreshes)
  x: number
  y: number
  z: number

  // Hiérarchie
  parentId?: string
  children?: UnifiedNode[]

  // Provenance
  source: string
  rawLabels?: Record<string, string>

  // Timestamps (ISO 8601)
  lastSeen: string
  createdAt?: string

  // K8s
  namespace?: string
}

// === Connexions ===

export interface Connection {
  id: string
  fromId: string
  toId: string
  traffic: number  // MB/s
  latency: number  // ms
  errors: number   // erreurs/s
  protocol: string
  type: string
}

// === Logs ===

export interface LogEntry {
  timestamp: string // ISO 8601
  level: 'info' | 'warn' | 'error' | 'debug'
  message: string
  pod?: string
  container?: string
  node?: string
}

// === Alertes ===

export interface Alert {
  id: string
  nodeId: string
  name: string
  severity: 'warning' | 'critical'
  message: string
  firedAt: string // ISO 8601
}

// === Santé API ===

export interface HealthStatus {
  status: 'ok' | 'degraded' | 'error'
  scrapers: Record<string, { status: 'ok' | 'disabled' | 'error'; message?: string }>
  cacheAgeSeconds: number
}

// === Réponses API ===

export interface ApiError {
  error: string
  code: string
}

export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  limit: number
}
```

---

## Règles de Calcul du Statut

Le statut d'un `UnifiedNode` est calculé côté backend à partir des métriques brutes :

```
StatusHealthy  → CPU < 70%  ET Memory% < 80%  ET Restarts == 0
StatusWarning  → CPU 70-85% OU Memory% 80-90% OU Restarts > 0
StatusCritical → CPU > 85%  OU Memory% > 90%  OU pod en CrashLoopBackOff
StatusUnknown  → Source inaccessible OU métriques absentes depuis > 2 cycles (>60s)
```

Le statut d'un nœud parent (cluster, node, vm-island) est le **statut le plus sévère** de ses enfants.

---

## Tables de Mapping Source → UnifiedNode

### Prometheus → VM (via targets)

| Prometheus | UnifiedNode | Transformation |
|-----------|------------|----------------|
| `instance` label | `id` + `name` | direct |
| `datacenter` label | `parentId` (VMIsland ID) | grouper par datacenter |
| `environment` label | `tags[]` | ajouter au tableau |
| `node_cpu_seconds_total{mode!="idle"}` | `cpu` | `100 - avg(idle%) * 100` |
| `node_memory_MemAvailable_bytes` / `node_memory_MemTotal_bytes` | `memory` + `memoryTotal` | en MB : `value / 1024 / 1024` |
| `node_filesystem_avail_bytes{mountpoint="/"}` | `disk` | `100 - (avail/size)*100` |

### K8s API → Nœuds

| Champ K8s | UnifiedNode | Transformation |
|-----------|-------------|----------------|
| `node.UID` | `id` | string() |
| `node.Name` | `name` | direct |
| constante | `type` | `"node"` |
| `node.Status.Conditions` | `status` | Ready=True→healthy, etc. |
| `node.Labels` | `tags[]` + `rawLabels` | filtrer les labels significatifs |
| Via Prometheus `kube_node_*` | `cpu`, `memory`, `disk` | requêtes PromQL |

### Docker API → Container

| Champ Docker | UnifiedNode | Transformation |
|-------------|-------------|----------------|
| `container.ID[:12]` | `id` | direct |
| `container.Names[0]` (sans `/`) | `name` | `strings.TrimPrefix(name, "/")` |
| constante | `type` | `"container"` |
| `container.State` | `status` | "running"→healthy, "exited"→critical |
| `com.docker.compose.project` label | `parentId` (ComposeGroup ID) | grouper |
| Via `ContainerStats` API | `cpu`, `memory` | calcul depuis stats brutes |
