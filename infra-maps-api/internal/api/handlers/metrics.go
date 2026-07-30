package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/gin-gonic/gin"
)

// MetricsProvider fournit les séries historiques d'un nœud.
// Implémenté par le scraper Prometheus (réel ou mock).
type MetricsProvider interface {
	RangeMetrics(ctx context.Context, nodeID string, from, to time.Time, step time.Duration, metric string) (map[string][]models.MetricPoint, error)
}

type MetricsHandler struct {
	Provider MetricsProvider
}

func NewMetricsHandler(p MetricsProvider) *MetricsHandler { return &MetricsHandler{Provider: p} }

// GetMetrics — GET /nodes/:nodeId/metrics (API_CONTRACT.md).
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	nodeID := c.Param("nodeId")
	now := time.Now()

	from, ok := unixParam(c, "from", now.Add(-time.Hour))
	if !ok {
		return
	}
	to, ok := unixParam(c, "to", now)
	if !ok {
		return
	}
	step := 60 * time.Second
	if raw := c.Query("step"); raw != "" {
		d, err := time.ParseDuration(raw)
		if err != nil || d <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "step must be a positive duration", "code": "ERR_INVALID_PARAM"})
			return
		}
		step = d
	}

	series, err := h.Provider.RangeMetrics(c.Request.Context(), nodeID, from, to, step, c.Query("metric"))
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error(), "code": "ERR_PROMETHEUS_TIMEOUT"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nodeId": nodeID,
		"from":   from.UTC(),
		"to":     to.UTC(),
		"step":   step.String(),
		"series": series,
	})
}

func unixParam(c *gin.Context, name string, def time.Time) (time.Time, bool) {
	raw := c.Query(name)
	if raw == "" {
		return def, true
	}
	sec, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": name + " must be a unix timestamp", "code": "ERR_INVALID_PARAM"})
		return time.Time{}, false
	}
	return time.Unix(sec, 0), true
}
