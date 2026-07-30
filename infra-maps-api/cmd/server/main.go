package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/api"
	"github.com/aristidewafo/infra-maps-api/internal/api/handlers"
	"github.com/aristidewafo/infra-maps-api/internal/cache"
	"github.com/aristidewafo/infra-maps-api/internal/layout"
	"github.com/aristidewafo/infra-maps-api/internal/scraper"
	"github.com/gin-gonic/gin"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	port := envOr("PORT", "8080")
	interval := durationEnvOr("SCRAPE_INTERVAL", 30*time.Second, log)
	ttl := durationEnvOr("CACHE_TTL", 30*time.Second, log)

	c := cache.New(ttl)
	engine := layout.NewEngine()
	scrapers := []scraper.Scraper{scraper.NewMockPrometheus()}
	orch := scraper.NewOrchestrator(scrapers, c, engine, interval, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go orch.Run(ctx)

	r := gin.Default()
	r.Use(corsDev())
	api.Register(r, handlers.NewTreeHandler(c), handlers.NewHealthHandler(scrapers, c))

	log.Info("starting infra-maps-api", "port", port, "scrape_interval", interval.String())
	if err := r.Run(":" + port); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

// corsDev autorise le front Vite en dev local. Restreint via CORS_ORIGIN.
func corsDev() gin.HandlerFunc {
	origin := envOr("CORS_ORIGIN", "*")
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func durationEnvOr(key string, def time.Duration, log *slog.Logger) time.Duration {
	raw := os.Getenv(key)
	if raw == "" {
		return def
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		log.Warn("invalid duration env, using default", "key", key, "value", raw, "default", def.String())
		return def
	}
	return d
}
