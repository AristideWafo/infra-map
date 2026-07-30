package scraper

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/cache"
	"github.com/aristidewafo/infra-maps-api/internal/layout"
	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockScraper struct {
	name  string
	nodes []*models.UnifiedNode
	err   error
}

func (m *mockScraper) Name() string  { return m.name }
func (m *mockScraper) Health() error { return nil }
func (m *mockScraper) Scrape(_ context.Context) ([]*models.UnifiedNode, error) {
	return m.nodes, m.err
}
func (m *mockScraper) Connections(_ context.Context) ([]*models.Connection, error) {
	return nil, nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newOrch(scrapers ...Scraper) (*Orchestrator, *cache.Memory) {
	c := cache.New(30 * time.Second)
	return NewOrchestrator(scrapers, c, layout.NewEngine(), 30*time.Second, testLogger()), c
}

func TestScrapeAll_PopulatesCache(t *testing.T) {
	orch, c := newOrch(&mockScraper{name: "ok", nodes: []*models.UnifiedNode{
		{ID: "vm-1", Status: models.StatusHealthy},
	}})

	orch.ScrapeAll(context.Background())

	val, ok := c.Get(CacheKeyTree)
	require.True(t, ok)
	tree := val.(*models.UnifiedNode)
	assert.Equal(t, "root", tree.ID)
	require.Len(t, tree.Children, 1)
	assert.Equal(t, "vm-1", tree.Children[0].ID)
}

func TestScrapeAll_ScraperErrorIsNotFatal(t *testing.T) {
	orch, c := newOrch(
		&mockScraper{name: "down", err: errors.New("connection refused")},
		&mockScraper{name: "up", nodes: []*models.UnifiedNode{{ID: "vm-1"}}},
	)

	orch.ScrapeAll(context.Background())

	val, ok := c.Get(CacheKeyTree)
	require.True(t, ok)
	assert.Len(t, val.(*models.UnifiedNode).Children, 1)
}

func TestScrapeAll_AppliesLayout(t *testing.T) {
	orch, c := newOrch(&mockScraper{name: "ok", nodes: []*models.UnifiedNode{
		{ID: "a"}, {ID: "b"},
	}})

	orch.ScrapeAll(context.Background())

	val, _ := c.Get(CacheKeyTree)
	tree := val.(*models.UnifiedNode)
	// "b" est en colonne 1 : position non nulle → layout appliqué
	assert.Greater(t, tree.Children[1].X, tree.Children[0].X)
}

func TestBuildTree_OrphanAttachedToRoot(t *testing.T) {
	tree := BuildTree([]*models.UnifiedNode{
		{ID: "pod-1", ParentID: "missing-parent", Status: models.StatusHealthy},
	})
	require.Len(t, tree.Children, 1)
	assert.Equal(t, "pod-1", tree.Children[0].ID)
}

func TestBuildTree_HierarchyAndStatusPropagation(t *testing.T) {
	tree := BuildTree([]*models.UnifiedNode{
		{ID: "cluster", Status: models.StatusHealthy},
		{ID: "node-1", ParentID: "cluster", Status: models.StatusHealthy},
		{ID: "pod-1", ParentID: "node-1", Status: models.StatusCritical},
	})

	require.Len(t, tree.Children, 1)
	cluster := tree.Children[0]
	assert.Equal(t, models.StatusCritical, tree.Status)
	assert.Equal(t, models.StatusCritical, cluster.Status)
}

func TestMockPrometheus_ScrapeIsRealCodePath(t *testing.T) {
	m := NewMockPrometheus()
	nodes, err := m.Scrape(context.Background())
	require.NoError(t, err)
	assert.NotEmpty(t, nodes)

	// Chaque nœud feuille a un statut calculé, pas hardcodé n'importe comment
	for _, n := range nodes {
		assert.NotEmpty(t, n.ID)
		assert.NotEmpty(t, n.Status)
		assert.Equal(t, "prometheus", n.Source)
	}
}

func TestMockPrometheus_RespectsContext(t *testing.T) {
	m := NewMockPrometheus()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := m.Scrape(ctx)
	assert.Error(t, err)
}
