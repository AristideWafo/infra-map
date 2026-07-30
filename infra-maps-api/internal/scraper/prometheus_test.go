package scraper

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakePromAPI struct {
	targets    promv1.TargetsResult
	targetsErr error
	// vectors par requête PromQL
	vectors  map[string]model.Vector
	matrices  map[string]model.Matrix
	rangeErr  error
	alerts    promv1.AlertsResult
	alertsErr error
}

func (f *fakePromAPI) Targets(_ context.Context) (promv1.TargetsResult, error) {
	return f.targets, f.targetsErr
}

func (f *fakePromAPI) Query(_ context.Context, q string, _ time.Time, _ ...promv1.Option) (model.Value, promv1.Warnings, error) {
	if v, ok := f.vectors[q]; ok {
		return v, nil, nil
	}
	return model.Vector{}, nil, nil
}

func (f *fakePromAPI) QueryRange(_ context.Context, q string, _ promv1.Range, _ ...promv1.Option) (model.Value, promv1.Warnings, error) {
	if f.rangeErr != nil {
		return nil, nil, f.rangeErr
	}
	if v, ok := f.matrices[q]; ok {
		return v, nil, nil
	}
	return model.Matrix{}, nil, nil
}

func sample(instance string, value float64) *model.Sample {
	return &model.Sample{
		Metric: model.Metric{"instance": model.LabelValue(instance)},
		Value:  model.SampleValue(value),
	}
}

func activeTarget(instance, dc, health string) promv1.ActiveTarget {
	labels := model.LabelSet{"instance": model.LabelValue(instance)}
	if dc != "" {
		labels["datacenter"] = model.LabelValue(dc)
	}
	return promv1.ActiveTarget{Labels: labels, Health: promv1.HealthStatus(health)}
}

func TestPrometheus_ScrapeBuildsVMsAndIslands(t *testing.T) {
	const gb = 1024 * 1024 * 1024
	fake := &fakePromAPI{
		targets: promv1.TargetsResult{Active: []promv1.ActiveTarget{
			activeTarget("192.168.1.10:9100", "dc1", "up"),
			activeTarget("192.168.1.11:9100", "dc1", "up"),
		}},
		vectors: map[string]model.Vector{
			queryCPU:      {sample("192.168.1.10:9100", 42.5), sample("192.168.1.11:9100", 90.0)},
			queryMemTotal: {sample("192.168.1.10:9100", 8*gb), sample("192.168.1.11:9100", 8*gb)},
			queryMemAvail: {sample("192.168.1.10:9100", 4*gb), sample("192.168.1.11:9100", 1*gb)},
			queryDisk:     {sample("192.168.1.10:9100", 38.0)},
		},
	}
	p := newPrometheusWithAPI(fake)

	nodes, err := p.Scrape(context.Background())
	require.NoError(t, err)
	require.Len(t, nodes, 3) // 1 island + 2 VMs

	island := nodes[0]
	assert.Equal(t, "vm-island-dc1", island.ID)
	assert.Equal(t, models.NodeTypeVMIsland, island.Type)

	vm1 := nodes[1]
	assert.Equal(t, "192.168.1.10:9100", vm1.ID)
	assert.Equal(t, "vm-island-dc1", vm1.ParentID)
	require.NotNil(t, vm1.CPU)
	assert.InDelta(t, 42.5, *vm1.CPU, 0.01)
	// Mémoire en MB : 8 GB total, 4 GB used
	assert.InDelta(t, 8192, *vm1.MemoryTotal, 0.01)
	assert.InDelta(t, 4096, *vm1.Memory, 0.01)
	assert.Equal(t, models.StatusHealthy, vm1.Status)
	require.NotNil(t, vm1.Disk)
	assert.InDelta(t, 38.0, *vm1.Disk, 0.01)

	// VM2 : CPU 90% → critical
	assert.Equal(t, models.StatusCritical, nodes[2].Status)
}

func TestPrometheus_DownTargetIsCritical(t *testing.T) {
	fake := &fakePromAPI{
		targets: promv1.TargetsResult{Active: []promv1.ActiveTarget{
			activeTarget("10.0.0.1:9100", "", "down"),
		}},
	}
	p := newPrometheusWithAPI(fake)

	nodes, err := p.Scrape(context.Background())
	require.NoError(t, err)
	require.Len(t, nodes, 2)
	assert.Equal(t, "vm-island-default", nodes[0].ID)
	assert.Equal(t, models.StatusCritical, nodes[1].Status)
}

func TestPrometheus_MissingMetricsGiveUnknown(t *testing.T) {
	fake := &fakePromAPI{
		targets: promv1.TargetsResult{Active: []promv1.ActiveTarget{
			activeTarget("10.0.0.1:9100", "dc1", "up"),
		}},
		// aucun vecteur : requêtes vides
	}
	p := newPrometheusWithAPI(fake)

	nodes, err := p.Scrape(context.Background())
	require.NoError(t, err)
	assert.Equal(t, models.StatusUnknown, nodes[1].Status)
}

func TestPrometheus_TargetsErrorPropagatesToHealth(t *testing.T) {
	fake := &fakePromAPI{targetsErr: errors.New("connection refused")}
	p := newPrometheusWithAPI(fake)

	_, err := p.Scrape(context.Background())
	require.Error(t, err)
	assert.Error(t, p.Health())

	// Un scrape réussi remet Health à nil
	fake.targetsErr = nil
	_, err = p.Scrape(context.Background())
	require.NoError(t, err)
	assert.NoError(t, p.Health())
}

func TestPrometheus_RangeMetrics(t *testing.T) {
	pair := func(ts int64, v float64) model.SamplePair {
		return model.SamplePair{Timestamp: model.Time(ts * 1000), Value: model.SampleValue(v)}
	}
	instance := "192.168.1.10:9100"
	cpuQuery := `100 - avg(rate(node_cpu_seconds_total{mode="idle",instance="` + instance + `"}[5m])) * 100`
	fake := &fakePromAPI{matrices: map[string]model.Matrix{
		cpuQuery: {&model.SampleStream{Values: []model.SamplePair{pair(100, 42.0), pair(160, 45.0)}}},
	}}
	p := newPrometheusWithAPI(fake)

	series, err := p.RangeMetrics(context.Background(), instance,
		time.Unix(100, 0), time.Unix(200, 0), time.Minute, "cpu")
	require.NoError(t, err)
	require.Len(t, series["cpu"], 2)
	assert.InDelta(t, 42.0, series["cpu"][0].Value, 0.01)
	// metric="cpu" : pas d'autres séries
	assert.NotContains(t, series, "memory")
}

func TestPrometheus_RangeMetricsError(t *testing.T) {
	fake := &fakePromAPI{rangeErr: errors.New("timeout")}
	p := newPrometheusWithAPI(fake)

	_, err := p.RangeMetrics(context.Background(), "x",
		time.Unix(0, 0), time.Unix(60, 0), time.Minute, "cpu")
	assert.Error(t, err)
}

func (f *fakePromAPI) Alerts(_ context.Context) (promv1.AlertsResult, error) {
	return f.alerts, f.alertsErr
}

func TestPrometheus_AlertsNormalization(t *testing.T) {
	fake := &fakePromAPI{alerts: promv1.AlertsResult{Alerts: []promv1.Alert{
		{
			State:  promv1.AlertStateFiring,
			Labels: model.LabelSet{"alertname": "PodCrashLooping", "pod": "api-1", "severity": "critical"},
			Annotations: model.LabelSet{"summary": "Pod api-1 restarting"},
		},
		{State: promv1.AlertStatePending, Labels: model.LabelSet{"alertname": "Ignored"}},
	}}}
	p := newPrometheusWithAPI(fake)

	alerts, err := p.Alerts(context.Background())
	require.NoError(t, err)
	require.Len(t, alerts, 1) // pending exclu

	a := alerts[0]
	assert.Equal(t, "alert-PodCrashLooping-api-1", a.ID)
	assert.Equal(t, "api-1", a.NodeID)
	assert.Equal(t, "critical", a.Severity)
	assert.Equal(t, "Pod api-1 restarting", a.Message)
}

func TestPrometheus_EnrichPods(t *testing.T) {
	fake := &fakePromAPI{vectors: map[string]model.Vector{
		queryPodCPU: {podSample("api-1", 42.0)},
		queryPodMem: {podSample("api-1", 256.0)},
	}}
	p := newPrometheusWithAPI(fake)

	restarts := 0
	memTotal := 512.0
	pod := &models.UnifiedNode{
		Name: "api-1", Type: models.NodeTypePod,
		Status: models.StatusHealthy, Restarts: &restarts, MemoryTotal: &memTotal,
	}
	vm := &models.UnifiedNode{Name: "vm-1", Type: models.NodeTypeVM}

	p.Enrich(context.Background(), []*models.UnifiedNode{pod, vm})

	require.NotNil(t, pod.CPU)
	assert.InDelta(t, 42.0, *pod.CPU, 0.01)
	assert.InDelta(t, 256.0, *pod.Memory, 0.01)
	assert.Equal(t, models.StatusHealthy, pod.Status)
	assert.Nil(t, vm.CPU) // VM non touchée
}

func TestPrometheus_EnrichDoesNotDowngradeCritical(t *testing.T) {
	fake := &fakePromAPI{vectors: map[string]model.Vector{
		queryPodCPU: {podSample("bad-pod", 10.0)},
		queryPodMem: {podSample("bad-pod", 64.0)},
	}}
	p := newPrometheusWithAPI(fake)

	pod := &models.UnifiedNode{Name: "bad-pod", Type: models.NodeTypePod, Status: models.StatusCritical}
	p.Enrich(context.Background(), []*models.UnifiedNode{pod})

	// CrashLoopBackOff posé par K8s ne doit pas être écrasé par des métriques saines
	assert.Equal(t, models.StatusCritical, pod.Status)
}

func podSample(pod string, value float64) *model.Sample {
	return &model.Sample{
		Metric: model.Metric{"pod": model.LabelValue(pod)},
		Value:  model.SampleValue(value),
	}
}
