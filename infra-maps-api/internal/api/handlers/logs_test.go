package handlers

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type fakeLogsProvider struct {
	enabled bool
	entries []models.LogEntry
	gotName string
}

func (f *fakeLogsProvider) Enabled() bool { return f.enabled }
func (f *fakeLogsProvider) Logs(_ context.Context, nodeName string, _ int, _ time.Time, _ string) ([]models.LogEntry, error) {
	f.gotName = nodeName
	return f.entries, nil
}

func TestGetLogs_DisabledIs404(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := setupRouter(t, fixtureTree())
	r := gin.New()
	r.GET("/api/v1/nodes/:nodeId/logs", NewLogsHandler(nil, NewTreeHandler(c)).GetLogs)

	w := get(r, "/api/v1/nodes/pod-api-1/logs")
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_LOKI_DISABLED")
}

func TestGetLogs_ResolvesNodeNameFromTree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := setupRouter(t, fixtureTree())
	fake := &fakeLogsProvider{enabled: true, entries: []models.LogEntry{}}
	r := gin.New()
	r.GET("/api/v1/nodes/:nodeId/logs", NewLogsHandler(fake, NewTreeHandler(c)).GetLogs)

	w := get(r, "/api/v1/nodes/pod-api-1/logs")
	assert.Equal(t, http.StatusOK, w.Code)
	// fixtureTree n'a pas de Name sur pod-api-1 → fallback sur l'ID
	assert.Equal(t, "pod-api-1", fake.gotName)
}

func TestGetLogs_InvalidLimit400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, c := setupRouter(t, fixtureTree())
	fake := &fakeLogsProvider{enabled: true}
	r := gin.New()
	r.GET("/api/v1/nodes/:nodeId/logs", NewLogsHandler(fake, NewTreeHandler(c)).GetLogs)

	assert.Equal(t, http.StatusBadRequest, get(r, "/api/v1/nodes/pod-api-1/logs?limit=9999").Code)
}
