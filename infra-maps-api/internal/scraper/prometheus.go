package scraper

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// promAPI est le sous-ensemble de l'API Prometheus v1 utilisé — mockable en tests.
type promAPI interface {
	Query(ctx context.Context, query string, ts time.Time, opts ...promv1.Option) (model.Value, promv1.Warnings, error)
	QueryRange(ctx context.Context, query string, r promv1.Range, opts ...promv1.Option) (model.Value, promv1.Warnings, error)
	Targets(ctx context.Context) (promv1.TargetsResult, error)
	Alerts(ctx context.Context) (promv1.AlertsResult, error)
}

// Requêtes PromQL (DATA_MODELS.md — mapping Prometheus → VM).
const (
	queryCPU      = `100 - avg by (instance) (rate(node_cpu_seconds_total{mode="idle"}[5m])) * 100`
	queryMemTotal = `node_memory_MemTotal_bytes`
	queryMemAvail = `node_memory_MemAvailable_bytes`
	queryDisk     = `100 - (node_filesystem_avail_bytes{mountpoint="/"} / node_filesystem_size_bytes{mountpoint="/"}) * 100`
)

// Prometheus découvre les VMs via /api/v1/targets (node_exporter) et
// enrichit CPU/RAM/Disk via PromQL. Groupe par label "datacenter" en vm-islands.
type Prometheus struct {
	api promAPI
	now func() time.Time

	mu      sync.RWMutex
	lastErr error
}

func NewPrometheus(url string) (*Prometheus, error) {
	client, err := api.NewClient(api.Config{Address: url})
	if err != nil {
		return nil, fmt.Errorf("prometheus client: %w", err)
	}
	return &Prometheus{api: promv1.NewAPI(client), now: time.Now}, nil
}

// newPrometheusWithAPI est le constructeur de test (API injectée).
func newPrometheusWithAPI(a promAPI) *Prometheus {
	return &Prometheus{api: a, now: time.Now}
}

func (p *Prometheus) Name() string { return "prometheus" }

// Health reflète le résultat du dernier scrape : pas d'appel réseau bloquant
// dans le health check.
func (p *Prometheus) Health() error {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.lastErr
}

func (p *Prometheus) setErr(err error) {
	p.mu.Lock()
	p.lastErr = err
	p.mu.Unlock()
}

func (p *Prometheus) Scrape(ctx context.Context) ([]*models.UnifiedNode, error) {
	targets, err := p.api.Targets(ctx)
	if err != nil {
		err = fmt.Errorf("prometheus targets: %w", err)
		p.setErr(err)
		return nil, err
	}

	cpu := p.vectorByInstance(ctx, queryCPU)
	memTotal := p.vectorByInstance(ctx, queryMemTotal)
	memAvail := p.vectorByInstance(ctx, queryMemAvail)
	disk := p.vectorByInstance(ctx, queryDisk)

	seen := p.now()
	islands := map[string]*models.UnifiedNode{}
	var nodes []*models.UnifiedNode

	for _, t := range targets.Active {
		instance := string(t.Labels["instance"])
		if instance == "" {
			continue
		}
		dc := string(t.Labels["datacenter"])
		if dc == "" {
			dc = "default"
		}

		islandID := "vm-island-" + dc
		if _, ok := islands[islandID]; !ok {
			island := &models.UnifiedNode{
				ID: islandID, Name: "Datacenter " + dc, Type: models.NodeTypeVMIsland,
				Status: models.StatusHealthy, Source: "prometheus-targets",
				Tags: []string{dc}, LastSeen: seen,
			}
			islands[islandID] = island
			nodes = append(nodes, island)
		}

		vm := &models.UnifiedNode{
			ID: instance, Name: instance, Type: models.NodeTypeVM,
			ParentID: islandID, Source: "prometheus-targets",
			LastSeen: seen, Status: models.StatusUnknown,
			RawLabels: labelsToMap(t.Labels),
		}
		if env := string(t.Labels["environment"]); env != "" {
			vm.Tags = append(vm.Tags, env)
		}
		if t.Health != promv1.HealthGood {
			vm.Status = models.StatusCritical
			nodes = append(nodes, vm)
			continue
		}

		var cpuPct, memPct float64 = -1, -1
		if v, ok := cpu[instance]; ok {
			vm.CPU = &v
			cpuPct = v
		}
		if total, ok := memTotal[instance]; ok {
			totalMB := total / 1024 / 1024
			vm.MemoryTotal = &totalMB
			if avail, ok := memAvail[instance]; ok {
				usedMB := (total - avail) / 1024 / 1024
				vm.Memory = &usedMB
				if totalMB > 0 {
					memPct = usedMB / totalMB * 100
				}
			}
		}
		if v, ok := disk[instance]; ok {
			vm.Disk = &v
		}

		if cpuPct >= 0 && memPct >= 0 {
			vm.Status = models.CalculateStatus(cpuPct, memPct, 0)
		}
		nodes = append(nodes, vm)
	}

	p.setErr(nil)
	return nodes, nil
}

// Connections : Prometheus seul ne voit pas les connexions réseau.
func (p *Prometheus) Connections(ctx context.Context) ([]*models.Connection, error) {
	return nil, nil
}

// vectorByInstance exécute une requête instantanée et indexe le résultat par
// label instance. Une erreur de requête rend une map vide : métrique absente,
// pas de scrape raté (dégradation partielle).
func (p *Prometheus) vectorByInstance(ctx context.Context, query string) map[string]float64 {
	out := map[string]float64{}
	val, _, err := p.api.Query(ctx, query, p.now())
	if err != nil {
		return out
	}
	vec, ok := val.(model.Vector)
	if !ok {
		return out
	}
	for _, s := range vec {
		out[string(s.Metric["instance"])] = float64(s.Value)
	}
	return out
}

func labelsToMap(ls model.LabelSet) map[string]string {
	m := make(map[string]string, len(ls))
	for k, v := range ls {
		m[string(k)] = string(v)
	}
	return m
}

// Requêtes range par métrique — %s = instance.
var rangeQueries = map[string]string{
	"cpu":    `100 - avg(rate(node_cpu_seconds_total{mode="idle",instance="%s"}[5m])) * 100`,
	"memory": `(node_memory_MemTotal_bytes{instance="%s"} - node_memory_MemAvailable_bytes{instance="%s"}) / 1024 / 1024`,
	"disk":   `100 - (node_filesystem_avail_bytes{mountpoint="/",instance="%s"} / node_filesystem_size_bytes{mountpoint="/",instance="%s"}) * 100`,
}

// RangeMetrics retourne les séries historiques d'un nœud (proxy /metrics).
// metric vide = toutes les métriques connues.
func (p *Prometheus) RangeMetrics(ctx context.Context, nodeID string, from, to time.Time, step time.Duration, metric string) (map[string][]models.MetricPoint, error) {
	out := map[string][]models.MetricPoint{}
	r := promv1.Range{Start: from, End: to, Step: step}

	for name, tmpl := range rangeQueries {
		if metric != "" && metric != name {
			continue
		}
		query := fmt.Sprintf(tmpl, repeatArg(tmpl, nodeID)...)
		val, _, err := p.api.QueryRange(ctx, query, r)
		if err != nil {
			return nil, fmt.Errorf("prometheus range query %s: %w", name, err)
		}
		matrix, ok := val.(model.Matrix)
		if !ok || len(matrix) == 0 {
			continue
		}
		series := make([]models.MetricPoint, 0, len(matrix[0].Values))
		for _, v := range matrix[0].Values {
			series = append(series, models.MetricPoint{Timestamp: v.Timestamp.Time(), Value: float64(v.Value)})
		}
		out[name] = series
	}
	return out, nil
}

// repeatArg fournit autant de fois nodeID que le template a de %s.
func repeatArg(tmpl, arg string) []interface{} {
	n := strings.Count(tmpl, "%s")
	args := make([]interface{}, n)
	for i := range args {
		args[i] = arg
	}
	return args
}

// Alerts retourne les alertes Prometheus actives, normalisées.
func (p *Prometheus) Alerts(ctx context.Context) ([]models.Alert, error) {
	res, err := p.api.Alerts(ctx)
	if err != nil {
		return nil, fmt.Errorf("prometheus alerts: %w", err)
	}
	var out []models.Alert
	for _, a := range res.Alerts {
		if a.State != promv1.AlertStateFiring {
			continue
		}
		name := string(a.Labels["alertname"])
		nodeID := string(a.Labels["pod"])
		if nodeID == "" {
			nodeID = string(a.Labels["instance"])
		}
		severity := string(a.Labels["severity"])
		if severity == "" {
			severity = "warning"
		}
		msg := string(a.Annotations["summary"])
		if msg == "" {
			msg = string(a.Annotations["description"])
		}
		out = append(out, models.Alert{
			ID:       "alert-" + name + "-" + nodeID,
			NodeID:   nodeID,
			Name:     name,
			Severity: severity,
			Message:  msg,
			FiredAt:  a.ActiveAt,
		})
	}
	return out, nil
}

// Requêtes cadvisor pour enrichir les pods K8s (topologie portée par client-go,
// métriques par Prometheus). CPU en % d'un cœur, mémoire working set en MB.
const (
	queryPodCPU = `sum by (pod) (rate(container_cpu_usage_seconds_total{container!=""}[5m])) * 100`
	queryPodMem = `sum by (pod) (container_memory_working_set_bytes{container!=""}) / 1024 / 1024`
)

// Enrich complète les pods K8s (matchés par nom) avec CPU/RAM depuis cadvisor,
// puis recalcule leur statut. Best-effort : requête en échec = pas d'enrichissement.
func (p *Prometheus) Enrich(ctx context.Context, nodes []*models.UnifiedNode) {
	cpu := p.vectorByLabel(ctx, queryPodCPU, "pod")
	mem := p.vectorByLabel(ctx, queryPodMem, "pod")
	if len(cpu) == 0 && len(mem) == 0 {
		return
	}

	for _, n := range nodes {
		if n.Type != models.NodeTypePod {
			continue
		}
		if v, ok := cpu[n.Name]; ok {
			c := v
			n.CPU = &c
		}
		if v, ok := mem[n.Name]; ok {
			m := v
			n.Memory = &m
		}
		if n.CPU != nil && n.Memory != nil {
			memPct := 0.0
			if n.MemoryTotal != nil && *n.MemoryTotal > 0 {
				memPct = *n.Memory / *n.MemoryTotal * 100
			}
			restarts := 0
			if n.Restarts != nil {
				restarts = *n.Restarts
			}
			// Ne pas rétrograder un statut critique posé par K8s (CrashLoopBackOff)
			if n.Status != models.StatusCritical {
				n.Status = models.CalculateStatus(*n.CPU, memPct, restarts)
			}
		}
	}
}

// vectorByLabel indexe le résultat d'une requête instantanée par un label donné.
func (p *Prometheus) vectorByLabel(ctx context.Context, query, label string) map[string]float64 {
	out := map[string]float64{}
	val, _, err := p.api.Query(ctx, query, p.now())
	if err != nil {
		return out
	}
	vec, ok := val.(model.Vector)
	if !ok {
		return out
	}
	for _, s := range vec {
		out[string(s.Metric[model.LabelName(label)])] = float64(s.Value)
	}
	return out
}
