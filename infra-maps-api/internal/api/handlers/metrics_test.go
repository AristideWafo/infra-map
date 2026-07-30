package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeMetricsProvider struct {
	series map[string][]models.MetricPoint
	err    error
	got    struct {
		nodeID string
		metric string
		step   time.Duration
	}
}

func (f *fakeMetricsProvider) RangeMetrics(_ context.Context, nodeID string, _, _ time.Time, step time.Duration, metric string) (map[string][]models.MetricPoint, error) {
	f.got.nodeID, f.got.metric, f.got.step = nodeID, metric, step
	return f.series, f.err
}

func metricsRouter(p MetricsProvider) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/nodes/:nodeId/metrics", NewMetricsHandler(p).GetMetrics)
	return r
}

func TestGetMetrics_OK(t *testing.T) {
	fake := &fakeMetricsProvider{series: map[string][]models.MetricPoint{
		"cpu": {{Timestamp: time.Unix(100, 0).UTC(), Value: 42.1}},
	}}
	r := metricsRouter(fake)

	w := get(r, "/api/v1/nodes/worker-01/metrics?metric=cpu&step=30s")
	assert.Equal(t, http.StatusOK, w.Code)

	var body struct {
		NodeID string                              `json:"nodeId"`
		Step   string                              `json:"step"`
		Series map[string][]models.MetricPoint     `json:"series"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "worker-01", body.NodeID)
	assert.Equal(t, "30s", body.Step)
	assert.Len(t, body.Series["cpu"], 1)
	assert.Equal(t, "cpu", fake.got.metric)
	assert.Equal(t, 30*time.Second, fake.got.step)
}

func TestGetMetrics_ProviderError504(t *testing.T) {
	r := metricsRouter(&fakeMetricsProvider{err: errors.New("timeout")})
	w := get(r, "/api/v1/nodes/x/metrics")
	assert.Equal(t, http.StatusGatewayTimeout, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_PROMETHEUS_TIMEOUT")
}

func TestGetMetrics_InvalidParams400(t *testing.T) {
	r := metricsRouter(&fakeMetricsProvider{})
	assert.Equal(t, http.StatusBadRequest, get(r, "/api/v1/nodes/x/metrics?from=abc").Code)
	assert.Equal(t, http.StatusBadRequest, get(r, "/api/v1/nodes/x/metrics?step=nope").Code)
}
