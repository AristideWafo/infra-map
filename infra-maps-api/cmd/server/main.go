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
	"github.com/gin-gonic/gin"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load(log)

	c := cache.New(cfg.CacheTTL)
	engine := layout.NewEngine()
	scrapers := buildScrapers(cfg, log)
	orch := scraper.NewOrchestrator(scrapers, c, engine, cfg.ScrapeInterval, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go orch.Run(ctx)

	r := gin.Default()
	r.Use(cors(cfg.CORSOrigin))
	api.Register(r, handlers.NewTreeHandler(c), handlers.NewHealthHandler(scrapers, c))

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

// buildScrapers assemble les sources selon la config. Une source qui échoue à
// l'initialisation est remplacée par un Noop : le service démarre quand même.
func buildScrapers(cfg config.Config, log *slog.Logger) []scraper.Scraper {
	if cfg.MockEnabled {
		return []scraper.Scraper{scraper.NewMockPrometheus()}
	}

	var scrapers []scraper.Scraper
	prom, err := scraper.NewPrometheus(cfg.PrometheusURL)
	if err != nil {
		log.Error("prometheus scraper init failed", "error", err)
		scrapers = append(scrapers, scraper.NewNoop("prometheus"))
	} else {
		scrapers = append(scrapers, prom)
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
	return scrapers
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
