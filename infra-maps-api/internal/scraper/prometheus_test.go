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
	vectors map[string]model.Vector
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
