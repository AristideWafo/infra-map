# Backend Development Guide — Go

> Ce guide définit les patterns, conventions et anti-patterns pour le backend InfraMaps.
> Lire `docs/DATA_MODELS.md` et `docs/API_CONTRACT.md` avant de coder.

---

## Structure du Projet

```
infra-maps-api/
├── cmd/
│   └── server/
│       └── main.go          # Wiring (injection de dépendances), démarrage serveur
├── internal/
│   ├── api/
│   │   ├── routes.go        # Enregistrement des routes Gin
│   │   ├── middleware/
│   │   │   ├── auth.go      # Middleware auth Bearer token
│   │   │   └── logger.go    # Middleware logs structurés
│   │   └── handlers/
│   │       ├── tree.go
│   │       ├── nodes.go
│   │       ├── metrics.go
│   │       ├── logs.go
│   │       ├── connections.go
│   │       └── websocket.go
│   ├── scraper/
│   │   ├── interface.go     # Scraper interface — source de vérité
│   │   ├── orchestrator.go  # Orchestration goroutines
│   │   ├── prometheus.go    # Métriques + discovery VMs
│   │   ├── kubernetes.go    # Topologie K8s
│   │   └── docker.go        # Containers standalone + réseaux
│   ├── layout/
│   │   └── engine.go        # Layout déterministe 2D (+ 3D Phase 4)
│   ├── resolver/
│   │   └── connections.go   # Discovery connexions réseau
│   ├── cache/
│   │   └── memory.go        # Cache TTL en mémoire
│   ├── models/
│   │   └── unified.go       # Types canoniques (voir DATA_MODELS.md)
│   └── config/
│       └── config.go        # Config struct + lecture env vars
├── Dockerfile
├── docker-compose.dev.yml
└── go.mod
```

---

## Pattern Fondamental — Interface Scraper

> Toute source de données implémente cette interface. Jamais de dépendance directe à une implémentation.

```go
// internal/scraper/interface.go

package scraper

import (
    "context"
    "infra-maps/internal/models"
)

// Scraper est le contrat que toute source de données doit respecter.
// Chaque implémentation doit être testable sans vraie connexion réseau.
type Scraper interface {
    // Name retourne l'identifiant de la source (utilisé dans les logs et health checks)
    Name() string

    // Scrape collecte les données et retourne des UnifiedNodes normalisés.
    // Doit respecter le context (timeout, annulation).
    Scrape(ctx context.Context) ([]*models.UnifiedNode, error)

    // Health retourne nil si la source est accessible, une erreur sinon.
    Health() error

    // Connections retourne les connexions réseau découvertes (peut retourner nil).
    Connections(ctx context.Context) ([]*models.Connection, error)
}

// Noop est une implémentation vide pour les tests et les sources désactivées.
type Noop struct{ name string }
func NewNoop(name string) *Noop { return &Noop{name: name} }
func (n *Noop) Name() string { return n.name }
func (n *Noop) Scrape(ctx context.Context) ([]*models.UnifiedNode, error) { return nil, nil }
func (n *Noop) Health() error { return nil }
func (n *Noop) Connections(ctx context.Context) ([]*models.Connection, error) { return nil, nil }
```

---

## Pattern Orchestrateur

```go
// internal/scraper/orchestrator.go

type Orchestrator struct {
    scrapers []Scraper
    cache    *cache.Memory
    layout   *layout.Engine
    interval time.Duration
    log      *slog.Logger
}

func (o *Orchestrator) Run(ctx context.Context) {
    // Scraping initial immédiat au démarrage
    o.scrapeAll(ctx)

    ticker := time.NewTicker(o.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            o.scrapeAll(ctx)
        case <-ctx.Done():
            o.log.Info("orchestrator stopped")
            return
        }
    }
}

func (o *Orchestrator) scrapeAll(ctx context.Context) {
    scrapeCtx, cancel := context.WithTimeout(ctx, 25*time.Second) // < interval
    defer cancel()

    type result struct {
        nodes []*models.UnifiedNode
        conns []*models.Connection
        err   error
        name  string
    }

    results := make(chan result, len(o.scrapers))

    for _, s := range o.scrapers {
        go func(sc Scraper) {
            nodes, err := sc.Scrape(scrapeCtx)
            conns, _ := sc.Connections(scrapeCtx) // Erreur connexions non fatale
            results <- result{nodes: nodes, conns: conns, err: err, name: sc.Name()}
        }(s)
    }

    var allNodes []*models.UnifiedNode
    var allConns []*models.Connection
    for range o.scrapers {
        r := <-results
        if r.err != nil {
            o.log.Warn("scraper error", "scraper", r.name, "error", r.err)
            continue // Erreur d'un scraper = dégradé, pas bloquant
        }
        allNodes = append(allNodes, r.nodes...)
        allConns = append(allConns, r.conns...)
    }

    tree := buildTree(allNodes)
    o.layout.Apply(tree)           // Calculer positions
    o.cache.Set("tree", tree)
    o.cache.Set("connections", allConns)
}
```

---

## Pattern Cache

```go
// internal/cache/memory.go

type Memory struct {
    mu    sync.RWMutex
    items map[string]*item
    ttl   time.Duration
}

type item struct {
    value     interface{}
    expiresAt time.Time
}

func (m *Memory) Set(key string, value interface{}) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.items[key] = &item{value: value, expiresAt: time.Now().Add(m.ttl)}
}

func (m *Memory) Get(key string) (interface{}, bool) {
    m.mu.RLock()
    defer m.mu.RUnlock()
    it, ok := m.items[key]
    if !ok || time.Now().After(it.expiresAt) {
        return nil, false
    }
    return it.value, true
}

func (m *Memory) Age(key string) int {
    m.mu.RLock()
    defer m.mu.RUnlock()
    it, ok := m.items[key]
    if !ok { return -1 }
    return int(m.ttl.Seconds()) - int(time.Until(it.expiresAt).Seconds())
}
```

---

## Pattern Handler

> Les handlers ne contiennent aucune logique métier. Ils lisent le cache et sérialisent.

```go
// internal/api/handlers/tree.go

type TreeHandler struct {
    cache  *cache.Memory
    logger *slog.Logger
}

func (h *TreeHandler) GetTree(c *gin.Context) {
    val, ok := h.cache.Get("tree")
    if !ok {
        c.JSON(http.StatusServiceUnavailable, gin.H{
            "error": "Cache not ready, scrapers still initializing",
            "code":  "ERR_CACHE_EMPTY",
        })
        return
    }

    tree, ok := val.(*models.UnifiedNode)
    if !ok {
        h.logger.Error("cache type assertion failed for 'tree'")
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error", "code": "ERR_INTERNAL"})
        return
    }

    // Appliquer les filtres query params
    filtered := applyFilters(tree, c.QueryMap())

    c.Header("Cache-Control", "max-age=25")
    c.JSON(http.StatusOK, filtered)
}
```

---

## Conventions Go

### Nommage

```
Packages     : snake_case, singulier                (scraper, cache, layout)
Types        : PascalCase                           (UnifiedNode, Orchestrator)
Fonctions    : camelCase privé, PascalCase public   (buildTree, NewOrchestrator)
Variables    : camelCase court                      (ctx, err, nodes)
Constantes   : PascalCase ou SCREAMING_SNAKE_CASE  (StatusHealthy, MAX_RETRIES)
```

### Gestion des erreurs

```go
// ✅ Correct — wrap avec contexte
nodes, err := scraper.Scrape(ctx)
if err != nil {
    return nil, fmt.Errorf("prometheus scraper: %w", err)
}

// ✅ Correct — erreur sentinelle typée
var ErrSourceUnavailable = errors.New("source unavailable")
if errors.Is(err, ErrSourceUnavailable) { ... }

// ❌ Interdit en production
if err != nil { panic(err) }

// ❌ Interdit — erreur silencieuse
nodes, _ := scraper.Scrape(ctx)
```

### Logging (slog)

```go
// ✅ Logs structurés avec slog
slog.Info("scrape complete", "scraper", "prometheus", "nodes", len(nodes), "duration_ms", d.Milliseconds())
slog.Warn("scraper degraded", "scraper", "kubernetes", "error", err)
slog.Error("cache write failed", "key", "tree", "error", err)

// ❌ Interdit
fmt.Printf("scraping done: %d nodes\n", len(nodes))
log.Println("error:", err)
```

### Context

```go
// ✅ Timeout sur tout appel externe
ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
defer cancel()
result, err := externalCall(ctx, ...)

// ❌ context.Background() dans les scrapers
// (utiliser le context passé en paramètre)
```

---

## Tests

### Table-driven tests (pattern Go standard)

```go
func TestStatusCalculation(t *testing.T) {
    tests := []struct {
        name     string
        cpu      float64
        memPct   float64
        restarts int
        want     models.NodeStatus
    }{
        {"healthy — all ok", 50, 60, 0, models.StatusHealthy},
        {"warning — high cpu", 75, 60, 0, models.StatusWarning},
        {"critical — very high cpu", 90, 60, 0, models.StatusCritical},
        {"warning — restarts", 40, 50, 3, models.StatusWarning},
        {"critical — high memory", 30, 92, 0, models.StatusCritical},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := CalculateStatus(tt.cpu, tt.memPct, tt.restarts)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Tester un scraper avec mock

```go
// Mock minimal — implémente l'interface Scraper
type mockScraper struct {
    nodes []*models.UnifiedNode
    err   error
}
func (m *mockScraper) Name() string { return "mock" }
func (m *mockScraper) Scrape(_ context.Context) ([]*models.UnifiedNode, error) {
    return m.nodes, m.err
}
func (m *mockScraper) Health() error { return nil }
func (m *mockScraper) Connections(_ context.Context) ([]*models.Connection, error) {
    return nil, nil
}

func TestOrchestrator_ScraperError(t *testing.T) {
    mock := &mockScraper{err: errors.New("connection refused")}
    cache := cache.New(30 * time.Second)
    orch := NewOrchestrator([]Scraper{mock}, cache, layout.NewEngine())

    // Un scraper en erreur ne doit pas bloquer le cache
    orch.scrapeAll(context.Background())
    _, ok := cache.Get("tree")
    assert.True(t, ok) // Cache mis à jour quand même (avec les autres scrapers)
}
```

### Test du Layout Engine (idempotence)

```go
func TestLayoutEngine_Idempotent(t *testing.T) {
    engine := layout.NewEngine()
    tree := buildTestTree() // arbre fixe

    result1 := engine.Apply(deepCopy(tree))
    result2 := engine.Apply(deepCopy(tree))

    assert.Equal(t, result1[0].X, result2[0].X)
    assert.Equal(t, result1[0].Y, result2[0].Y)
}
```

---

## Anti-patterns — Ne jamais faire

```go
// ❌ Global state mutable
var globalCache = map[string]interface{}{}

// ❌ init() pour la logique métier
func init() { connectToDatabase() }

// ❌ Goroutine sans context ni erreur handling
go func() { scraper.Scrape(context.Background()) }()

// ❌ Pointer receiver sur valeur non modifiée (convention Go)
func (s SomeStruct) Method() {} // OK si pas de mutation
func (s *SomeStruct) Method() {} // OK si mutation

// ❌ Base de données — absolument interdit
import "database/sql"
import "github.com/mattn/go-sqlite3"
```
