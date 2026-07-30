package scraper

import (
	"context"
	"log/slog"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/cache"
	"github.com/aristidewafo/infra-maps-api/internal/layout"
	"github.com/aristidewafo/infra-maps-api/internal/models"
)

// Enricher complète les nœuds d'autres sources avec des métriques
// supplémentaires (ex: Prometheus enrichit les pods K8s). Best-effort.
type Enricher interface {
	Enrich(ctx context.Context, nodes []*models.UnifiedNode)
}

// Clés de cache écrites par l'orchestrateur, lues par les handlers.
const (
	CacheKeyTree        = "tree"
	CacheKeyConnections = "connections"
)

// Orchestrator fait tourner tous les scrapers en parallèle à intervalle fixe,
// assemble l'arbre, applique le layout et publie dans le cache.
type Orchestrator struct {
	scrapers []Scraper
	cache    *cache.Memory
	layout   *layout.Engine
	interval time.Duration
	timeout  time.Duration
	log      *slog.Logger
}

// SetTimeout fixe le timeout d'un cycle de scrape (défaut : interval - 5s).
func (o *Orchestrator) SetTimeout(d time.Duration) {
	if d > 0 {
		o.timeout = d
	}
}

func NewOrchestrator(scrapers []Scraper, c *cache.Memory, l *layout.Engine, interval time.Duration, log *slog.Logger) *Orchestrator {
	return &Orchestrator{scrapers: scrapers, cache: c, layout: l, interval: interval, log: log}
}

// Scrapers expose la liste pour le health check.
func (o *Orchestrator) Scrapers() []Scraper { return o.scrapers }

// Run boucle jusqu'à annulation du context. Premier scrape immédiat.
func (o *Orchestrator) Run(ctx context.Context) {
	o.ScrapeAll(ctx)

	ticker := time.NewTicker(o.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			o.ScrapeAll(ctx)
		case <-ctx.Done():
			o.log.Info("orchestrator stopped")
			return
		}
	}
}

// ScrapeAll exécute un cycle complet avec un timeout inférieur à l'intervalle.
// L'erreur d'un scraper est loguée et ignorée : mode dégradé, jamais bloquant.
func (o *Orchestrator) ScrapeAll(ctx context.Context) {
	timeout := o.timeout
	if timeout <= 0 {
		timeout = o.interval - 5*time.Second
	}
	if timeout <= 0 {
		timeout = o.interval
	}
	scrapeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()

	type result struct {
		name  string
		nodes []*models.UnifiedNode
		conns []*models.Connection
		err   error
	}

	results := make(chan result, len(o.scrapers))
	for _, s := range o.scrapers {
		go func(sc Scraper) {
			nodes, err := sc.Scrape(scrapeCtx)
			conns, _ := sc.Connections(scrapeCtx) // erreur connexions non fatale
			results <- result{name: sc.Name(), nodes: nodes, conns: conns, err: err}
		}(s)
	}

	var allNodes []*models.UnifiedNode
	allConns := []*models.Connection{}
	for range o.scrapers {
		r := <-results
		if r.err != nil {
			o.log.Warn("scraper error", "scraper", r.name, "error", r.err)
			continue
		}
		allNodes = append(allNodes, r.nodes...)
		allConns = append(allConns, r.conns...)
	}

	for _, s := range o.scrapers {
		if e, ok := s.(Enricher); ok {
			e.Enrich(scrapeCtx, allNodes)
		}
	}

	tree := BuildTree(allNodes)
	o.layout.Apply(tree)
	o.cache.Set(CacheKeyTree, tree)
	o.cache.Set(CacheKeyConnections, allConns)

	o.log.Info("scrape complete", "nodes", len(allNodes), "connections", len(allConns), "duration_ms", time.Since(start).Milliseconds())
}
