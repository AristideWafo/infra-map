package config

import (
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func discard() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func TestLoad_Defaults(t *testing.T) {
	cfg := Load(discard())

	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, 30*time.Second, cfg.ScrapeInterval)
	// Sans PROMETHEUS_URL, le mock est forcé
	assert.True(t, cfg.MockEnabled)
}

func TestLoad_EnvOverrides(t *testing.T) {
	t.Setenv("PROMETHEUS_URL", "http://prom:9090")
	t.Setenv("SCRAPE_INTERVAL", "10s")
	t.Setenv("K8S_ENABLED", "false")

	cfg := Load(discard())

	assert.Equal(t, "http://prom:9090", cfg.PrometheusURL)
	assert.Equal(t, 10*time.Second, cfg.ScrapeInterval)
	assert.False(t, cfg.K8sEnabled)
	assert.False(t, cfg.MockEnabled)
}

func TestLoad_InvalidValuesFallBack(t *testing.T) {
	t.Setenv("SCRAPE_INTERVAL", "not-a-duration")
	t.Setenv("K8S_ENABLED", "not-a-bool")

	cfg := Load(discard())

	assert.Equal(t, 30*time.Second, cfg.ScrapeInterval)
	assert.True(t, cfg.K8sEnabled)
}
