package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/aristidewafo/infra-maps-api/internal/api"
	"github.com/aristidewafo/infra-maps-api/internal/api/handlers"
	"github.com/aristidewafo/infra-maps-api/internal/cache"
	"github.com/aristidewafo/infra-maps-api/internal/config"
	"github.com/aristidewafo/infra-maps-api/internal/layout"
	"github.com/aristidewafo/infra-maps-api/internal/scraper"
	"github.com/aristidewafo/infra-maps-api/internal/ws"
	"github.com/gin-gonic/gin"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load(log)

	c := cache.New(cfg.CacheTTL)
	engine := layout.NewEngine()
	scrapers, metricsProvider := buildScrapers(cfg, log)
	orch := scraper.NewOrchestrator(scrapers, c, engine, cfg.ScrapeInterval, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go orch.Run(ctx)

	var logsProvider handlers.LogsProvider
	if cfg.LokiURL != "" {
		logsProvider = scraper.NewLoki(cfg.LokiURL, 100)
	}

	hub := ws.NewHub(log)
	go hub.RunPing(ctx)
	if alertsProvider, ok := metricsProvider.(ws.AlertsProvider); ok {
		watcher := ws.NewAlertWatcher(alertsProvider, hub, cfg.ScrapeInterval, log)
		go watcher.Run(ctx)
	}

	treeHandler := handlers.NewTreeHandler(c)
	r := gin.Default()
	r.Use(cors(cfg.CORSOrigin))
	api.Register(r, treeHandler, handlers.NewHealthHandler(scrapers, c),
		handlers.NewMetricsHandler(metricsProvider), handlers.NewLogsHandler(logsProvider, treeHandler), hub)

	log.Info("starting infra-maps-api",
		"port", cfg.Port,
		"scrape_interval", cfg.ScrapeInterval.String(),
		"mock", cfg.MockEnabled,
	)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

// buildScrapers assemble les sources selon la config et retourne aussi le
// provider de métriques historiques (réel ou mock). Une source qui échoue à
// l'initialisation est remplacée par un Noop : le service démarre quand même.
func buildScrapers(cfg config.Config, log *slog.Logger) ([]scraper.Scraper, handlers.MetricsProvider) {
	if cfg.MockEnabled {
		mock := scraper.NewMockPrometheus()
		return []scraper.Scraper{mock}, mock
	}

	var scrapers []scraper.Scraper
	var metrics handlers.MetricsProvider
	prom, err := scraper.NewPrometheus(cfg.PrometheusURL)
	if err != nil {
		log.Error("prometheus scraper init failed", "error", err)
		scrapers = append(scrapers, scraper.NewNoop("prometheus"))
		metrics = scraper.NewMockPrometheus()
	} else {
		scrapers = append(scrapers, prom)
		metrics = prom
	}

	if cfg.K8sEnabled {
		k8s, err := scraper.NewKubernetes(cfg.K8sInCluster, cfg.K8sKubeconfig)
		if err != nil {
			log.Error("kubernetes scraper init failed", "error", err)
			scrapers = append(scrapers, scraper.NewNoop("kubernetes"))
		} else {
			scrapers = append(scrapers, k8s)
		}
	}

	if cfg.DockerEnabled {
		docker, err := scraper.NewDocker(cfg.DockerSocket)
		if err != nil {
			log.Error("docker scraper init failed", "error", err)
			scrapers = append(scrapers, scraper.NewNoop("docker"))
		} else {
			scrapers = append(scrapers, docker)
		}
	}
	return scrapers, metrics
}

func cors(origin string) gin.HandlerFunc {
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
