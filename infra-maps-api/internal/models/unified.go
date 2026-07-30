package models

import "time"

// NodeType identifie le rôle d'un nœud dans la hiérarchie.
type NodeType string

const (
	NodeTypeRoot               NodeType = "root"
	NodeTypeVMIsland           NodeType = "vm-island"
	NodeTypeDockerComposeGroup NodeType = "compose-group"
	NodeTypeCluster            NodeType = "cluster"
	NodeTypeNode               NodeType = "node"
	NodeTypePod                NodeType = "pod"
	NodeTypeVM                 NodeType = "vm"
	NodeTypeContainer          NodeType = "container"
)

// NodeStatus représente l'état de santé calculé d'un nœud.
type NodeStatus string

const (
	StatusHealthy  NodeStatus = "healthy"
	StatusWarning  NodeStatus = "warning"
	StatusCritical NodeStatus = "critical"
	StatusUnknown  NodeStatus = "unknown"
)

// severity ordonne les statuts du moins au plus sévère pour la propagation parent.
var severity = map[NodeStatus]int{
	StatusHealthy:  0,
	StatusUnknown:  1,
	StatusWarning:  2,
	StatusCritical: 3,
}

// UnifiedNode est la structure normalisée pour toutes les sources.
type UnifiedNode struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Type   NodeType   `json:"type"`
	Status NodeStatus `json:"status"`
	Tags   []string   `json:"tags,omitempty"`

	CPU         *float64 `json:"cpu,omitempty"`
	Memory      *float64 `json:"memory,omitempty"`
	MemoryTotal *float64 `json:"memoryTotal,omitempty"`
	Disk        *float64 `json:"disk,omitempty"`
	Pods        *int     `json:"pods,omitempty"`
	Restarts    *int     `json:"restarts,omitempty"`

	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`

	ParentID string         `json:"parentId,omitempty"`
	Children []*UnifiedNode `json:"children,omitempty"`

	Source    string            `json:"source"`
	RawLabels map[string]string `json:"rawLabels,omitempty"`

	LastSeen  time.Time `json:"lastSeen"`
	CreatedAt time.Time `json:"createdAt,omitempty"`

	Namespace string `json:"namespace,omitempty"`
}

// Connection représente un lien réseau entre deux nœuds.
type Connection struct {
	ID       string  `json:"id"`
	FromID   string  `json:"fromId"`
	ToID     string  `json:"toId"`
	Traffic  float64 `json:"traffic"`
	Latency  float64 `json:"latency"`
	Errors   float64 `json:"errors"`
	Protocol string  `json:"protocol"`
	Type     string  `json:"type"`
}

// MetricPoint est un point de série temporelle (proxy Prometheus).
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// LogEntry représente une entrée de log (depuis Loki).
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Pod       string    `json:"pod,omitempty"`
	Container string    `json:"container,omitempty"`
	Node      string    `json:"node,omitempty"`
}

// Alert représente une alerte Prometheus en cours.
type Alert struct {
	ID       string    `json:"id"`
	NodeID   string    `json:"nodeId"`
	Name     string    `json:"name"`
	Severity string    `json:"severity"`
	Message  string    `json:"message"`
	FiredAt  time.Time `json:"firedAt"`
}

// HealthStatus est l'état de santé des scrapers exposé par /health.
type HealthStatus struct {
	Status   string                   `json:"status"`
	Scrapers map[string]ScraperHealth `json:"scrapers"`
	CacheAge int                      `json:"cacheAgeSeconds"`
}

type ScraperHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// CalculateStatus applique les seuils définis dans DATA_MODELS.md.
// memPct est le pourcentage de mémoire utilisée (0–100).
func CalculateStatus(cpu, memPct float64, restarts int) NodeStatus {
	switch {
	case cpu > 85 || memPct > 90:
		return StatusCritical
	case cpu >= 70 || memPct >= 80 || restarts > 0:
		return StatusWarning
	default:
		return StatusHealthy
	}
}

// PropagateStatus remonte le statut le plus sévère des enfants vers chaque parent.
func PropagateStatus(n *UnifiedNode) NodeStatus {
	if len(n.Children) == 0 {
		return n.Status
	}
	worst := StatusHealthy
	for _, c := range n.Children {
		s := PropagateStatus(c)
		if severity[s] > severity[worst] {
			worst = s
		}
	}
	n.Status = worst
	return worst
}
