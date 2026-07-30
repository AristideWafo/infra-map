package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/cache"
	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/aristidewafo/infra-maps-api/internal/scraper"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeScraper struct {
	name      string
	healthErr error
}

func (f *fakeScraper) Name() string  { return f.name }
func (f *fakeScraper) Health() error { return f.healthErr }
func (f *fakeScraper) Scrape(_ context.Context) ([]*models.UnifiedNode, error) {
	return nil, nil
}
func (f *fakeScraper) Connections(_ context.Context) ([]*models.Connection, error) {
	return nil, nil
}

func doHealth(t *testing.T, scrapers []scraper.Scraper, c *cache.Memory) (*httptest.ResponseRecorder, models.HealthStatus) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/v1/health", NewHealthHandler(scrapers, c).Health)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	r.ServeHTTP(w, req)

	var body models.HealthStatus
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	return w, body
}

func TestHealth_AllScrapersOK(t *testing.T) {
	c := cache.New(30 * time.Second)
	c.Set(scraper.CacheKeyTree, &models.UnifiedNode{ID: "root"})

	w, body := doHealth(t, []scraper.Scraper{&fakeScraper{name: "prometheus"}}, c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", body.Status)
	assert.Equal(t, "ok", body.Scrapers["prometheus"].Status)
}

func TestHealth_DegradedScraper(t *testing.T) {
	c := cache.New(30 * time.Second)
	w, body := doHealth(t, []scraper.Scraper{
		&fakeScraper{name: "prometheus", healthErr: errors.New("connection timeout after 5s")},
		&fakeScraper{name: "kubernetes"},
	}, c)

	// Toujours 200 : dégradé n'est pas une erreur HTTP
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "degraded", body.Status)
	assert.Equal(t, "error", body.Scrapers["prometheus"].Status)
	assert.Equal(t, "connection timeout after 5s", body.Scrapers["prometheus"].Message)
	assert.Equal(t, "ok", body.Scrapers["kubernetes"].Status)
}

func TestHealth_EmptyCacheAge(t *testing.T) {
	c := cache.New(30 * time.Second)
	_, body := doHealth(t, nil, c)
	assert.Equal(t, -1, body.CacheAge)
}
