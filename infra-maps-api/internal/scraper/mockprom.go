package scraper

import (
	"math"
	"context"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/models"
)

// MockPrometheus est le scraper Phase 0 : il exerce toute la chaîne réelle
// (interface Scraper → orchestrateur → buildTree → layout → cache → API)
// avec des données simulées mais plausibles. Remplacé par le vrai scraper
// Prometheus en Phase 1 sans toucher au reste de la chaîne.
type MockPrometheus struct {
	now func() time.Time
}

func NewMockPrometheus() *MockPrometheus {
	return &MockPrometheus{now: time.Now}
}

func (m *MockPrometheus) Name() string  { return "prometheus" }
func (m *MockPrometheus) Health() error { return nil }

func (m *MockPrometheus) Scrape(ctx context.Context) ([]*models.UnifiedNode, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	seen := m.now()

	f := func(v float64) *float64 { return &v }
	i := func(v int) *int { return &v }

	mk := func(id, name string, typ models.NodeType, parent string, cpu, memMB, memTotalMB float64, restarts int, tags []string, ns string) *models.UnifiedNode {
		n := &models.UnifiedNode{
			ID: id, Name: name, Type: typ, ParentID: parent,
			CPU: f(cpu), Memory: f(memMB), MemoryTotal: f(memTotalMB),
			Restarts: i(restarts), Tags: tags, Namespace: ns,
			Source: "prometheus", LastSeen: seen,
			Status: models.CalculateStatus(cpu, memMB/memTotalMB*100, restarts),
		}
		return n
	}

	nodes := []*models.UnifiedNode{
		// Cluster K8s simulé
		{ID: "cluster-prod", Name: "cluster-production", Type: models.NodeTypeCluster,
			Source: "prometheus", LastSeen: seen, Status: models.StatusHealthy,
			Tags: []string{"prod"}, Pods: i(6)},
		{ID: "node-worker-01", Name: "worker-01", Type: models.NodeTypeNode, ParentID: "cluster-prod",
			Source: "prometheus", LastSeen: seen, Tags: []string{"prod"},
			CPU: f(45.2), Memory: f(9830), MemoryTotal: f(16384), Disk: f(38),
			Pods: i(3), Status: models.CalculateStatus(45.2, 60, 0)},
		{ID: "node-worker-02", Name: "worker-02", Type: models.NodeTypeNode, ParentID: "cluster-prod",
			Source: "prometheus", LastSeen: seen, Tags: []string{"prod"},
			CPU: f(78.9), Memory: f(13763), MemoryTotal: f(16384), Disk: f(71),
			Pods: i(3), Status: models.CalculateStatus(78.9, 84, 0)},

		mk("pod-api-1", "api-frontend-1", models.NodeTypePod, "node-worker-01", 42.5, 256, 512, 0, []string{"api", "prod"}, "default"),
		mk("pod-api-2", "api-frontend-2", models.NodeTypePod, "node-worker-01", 38.1, 240, 512, 0, []string{"api", "prod"}, "default"),
		mk("pod-db-1", "postgres-main", models.NodeTypePod, "node-worker-01", 55.0, 1843, 2048, 0, []string{"db", "prod"}, "default"),
		mk("pod-cache-1", "redis-cache", models.NodeTypePod, "node-worker-02", 12.3, 128, 256, 0, []string{"cache", "prod"}, "default"),
		mk("pod-worker-1", "job-worker-1", models.NodeTypePod, "node-worker-02", 91.7, 480, 512, 4, []string{"worker", "prod"}, "jobs"),
		mk("pod-monitor-1", "prometheus-server", models.NodeTypePod, "node-worker-02", 33.0, 900, 2048, 0, []string{"monitoring"}, "monitoring"),

		// Island de VMs simulée (Prometheus targets / node_exporter)
		{ID: "vm-island-dc1", Name: "Datacenter Paris", Type: models.NodeTypeVMIsland,
			Source: "prometheus", LastSeen: seen, Status: models.StatusHealthy, Tags: []string{"dc1"}},
		mk("vm-192.168.1.10", "bastion-01", models.NodeTypeVM, "vm-island-dc1", 8.4, 1024, 4096, 0, []string{"dc1", "bastion"}, ""),
		mk("vm-192.168.1.11", "backup-01", models.NodeTypeVM, "vm-island-dc1", 22.0, 3072, 8192, 0, []string{"dc1", "backup"}, ""),
		mk("vm-192.168.1.12", "legacy-app-01", models.NodeTypeVM, "vm-island-dc1", 88.9, 7782, 8192, 0, []string{"dc1", "legacy"}, ""),
	}

	return nodes, nil
}

func (m *MockPrometheus) Connections(ctx context.Context) ([]*models.Connection, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []*models.Connection{
		{ID: "pod-api-1-pod-db-1", FromID: "pod-api-1", ToID: "pod-db-1", Protocol: "tcp", Type: "service"},
		{ID: "pod-api-2-pod-db-1", FromID: "pod-api-2", ToID: "pod-db-1", Protocol: "tcp", Type: "service"},
		{ID: "pod-api-1-pod-cache-1", FromID: "pod-api-1", ToID: "pod-cache-1", Protocol: "tcp", Type: "service"},
	}, nil
}

// RangeMetrics génère des séries synthétiques plausibles (dev sans Prometheus).
// Déterministe : mêmes bornes → mêmes valeurs.
func (m *MockPrometheus) RangeMetrics(ctx context.Context, nodeID string, from, to time.Time, step time.Duration, metric string) (map[string][]models.MetricPoint, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if step <= 0 {
		step = time.Minute
	}
	base := map[string]float64{"cpu": 45, "memory": 512, "disk": 40}
	amp := map[string]float64{"cpu": 25, "memory": 128, "disk": 5}

	out := map[string][]models.MetricPoint{}
	for name := range base {
		if metric != "" && metric != name {
			continue
		}
		var series []models.MetricPoint
		for ts := from; !ts.After(to); ts = ts.Add(step) {
			phase := float64(ts.Unix()%3600) / 3600 * 2 * math.Pi
			series = append(series, models.MetricPoint{
				Timestamp: ts,
				Value:     base[name] + amp[name]*math.Sin(phase),
			})
		}
		out[name] = series
	}
	return out, nil
}

// Alerts simule une alerte critique permanente sur le pod en surcharge.
func (m *MockPrometheus) Alerts(ctx context.Context) ([]models.Alert, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return []models.Alert{{
		ID: "alert-PodHighCPU-pod-worker-1", NodeID: "pod-worker-1",
		Name: "PodHighCPU", Severity: "critical",
		Message: "Pod job-worker-1 CPU above 90% for 10 minutes",
		FiredAt: m.now().Add(-10 * time.Minute),
	}}, nil
}
