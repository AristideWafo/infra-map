package handlers

import (
	"fmt"
	"encoding/json"
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

func fixtureTree() *models.UnifiedNode {
	return &models.UnifiedNode{
		ID: "root", Name: "Infrastructure", Type: models.NodeTypeRoot, Status: models.StatusWarning,
		Children: []*models.UnifiedNode{
			{ID: "cluster-prod", Type: models.NodeTypeCluster, Status: models.StatusWarning, Tags: []string{"prod"},
				Children: []*models.UnifiedNode{
					{ID: "pod-api-1", Type: models.NodeTypePod, Status: models.StatusHealthy, Namespace: "default", Tags: []string{"api"}},
					{ID: "pod-job-1", Type: models.NodeTypePod, Status: models.StatusWarning, Namespace: "jobs", Tags: []string{"worker"}},
				}},
			{ID: "vm-island-dc1", Type: models.NodeTypeVMIsland, Status: models.StatusHealthy, Tags: []string{"dc1"}},
		},
	}
}

func setupRouter(t *testing.T, tree *models.UnifiedNode) (*gin.Engine, *cache.Memory) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c := cache.New(30 * time.Second)
	if tree != nil {
		c.Set(scraper.CacheKeyTree, tree)
	}
	h := NewTreeHandler(c)
	r := gin.New()
	r.GET("/api/v1/tree", h.GetTree)
	r.GET("/api/v1/tree/:nodeId", h.GetSubtree)
	r.GET("/api/v1/nodes/:nodeId", h.GetNode)
	r.GET("/api/v1/tags", h.GetTags)
	r.GET("/api/v1/connections", h.GetConnections)
	return r, c
}

func get(r *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	r.ServeHTTP(w, req)
	return w
}

func decodeNode(t *testing.T, w *httptest.ResponseRecorder) models.UnifiedNode {
	t.Helper()
	var n models.UnifiedNode
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &n))
	return n
}

func TestGetTree_OK(t *testing.T) {
	r, _ := setupRouter(t, fixtureTree())
	w := get(r, "/api/v1/tree")

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "max-age=25", w.Header().Get("Cache-Control"))
	n := decodeNode(t, w)
	assert.Equal(t, "root", n.ID)
	assert.Len(t, n.Children, 2)
}

func TestGetTree_EmptyCache503(t *testing.T) {
	r, _ := setupRouter(t, nil)
	w := get(r, "/api/v1/tree")

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_CACHE_EMPTY")
}

func TestGetTree_FilterNamespace(t *testing.T) {
	r, _ := setupRouter(t, fixtureTree())
	n := decodeNode(t, get(r, "/api/v1/tree?namespace=jobs"))

	// Seul le chemin cluster-prod → pod-job-1 survit
	require.Len(t, n.Children, 1)
	require.Len(t, n.Children[0].Children, 1)
	assert.Equal(t, "pod-job-1", n.Children[0].Children[0].ID)
}

func TestGetTree_FilterStatus(t *testing.T) {
	r, _ := setupRouter(t, fixtureTree())
	n := decodeNode(t, get(r, "/api/v1/tree?status=warning"))

	ids := []string{}
	for _, c := range n.Children {
		ids = append(ids, c.ID)
	}
	assert.NotContains(t, ids, "vm-island-dc1")
}

func TestGetTree_MaxDepth(t *testing.T) {
	r, _ := setupRouter(t, fixtureTree())
	n := decodeNode(t, get(r, "/api/v1/tree?maxDepth=1"))

	require.Len(t, n.Children, 2)
	for _, c := range n.Children {
		assert.Empty(t, c.Children)
	}
}

func TestGetTree_InvalidMaxDepth400(t *testing.T) {
	r, _ := setupRouter(t, fixtureTree())
	w := get(r, "/api/v1/tree?maxDepth=abc")
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_INVALID_PARAM")
}

func TestGetTree_DoesNotMutateCache(t *testing.T) {
	tree := fixtureTree()
	r, c := setupRouter(t, tree)
	get(r, "/api/v1/tree?namespace=jobs")

	val, _ := c.Get(scraper.CacheKeyTree)
	cached := val.(*models.UnifiedNode)
	assert.Len(t, cached.Children, 2)
	assert.Len(t, cached.Children[0].Children, 2)
}

func TestGetSubtree_OK(t *testing.T) {
	r, _ := setupRouter(t, fixtureTree())
	n := decodeNode(t, get(r, "/api/v1/tree/cluster-prod"))
	assert.Equal(t, "cluster-prod", n.ID)
	assert.Len(t, n.Children, 2)
}

func TestGetSubtree_NotFound404(t *testing.T) {
	r, _ := setupRouter(t, fixtureTree())
	w := get(r, "/api/v1/tree/pod-xyz")
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "ERR_NODE_NOT_FOUND")
}

func TestGetNode_StripsChildren(t *testing.T) {
	r, _ := setupRouter(t, fixtureTree())
	n := decodeNode(t, get(r, "/api/v1/nodes/cluster-prod"))
	assert.Equal(t, "cluster-prod", n.ID)
	assert.Empty(t, n.Children)
}

func TestGetTags_SortedUnique(t *testing.T) {
	r, _ := setupRouter(t, fixtureTree())
	w := get(r, "/api/v1/tags")

	var body struct {
		Tags []string `json:"tags"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, []string{"api", "dc1", "prod", "worker"}, body.Tags)
}

func TestGetConnections_FilterByFromID(t *testing.T) {
	r, c := setupRouter(t, fixtureTree())
	c.Set(scraper.CacheKeyConnections, []*models.Connection{
		{ID: "a-b", FromID: "a", ToID: "b", Type: "service"},
		{ID: "c-d", FromID: "c", ToID: "d", Type: "service"},
	})

	w := get(r, "/api/v1/connections?fromId=a")
	var body struct {
		Connections []*models.Connection `json:"connections"`
		Total       int                  `json:"total"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, 1, body.Total)
	assert.Equal(t, "a-b", body.Connections[0].ID)
}

func TestGetTree_ServesStaleDataAfterTTL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c := cache.New(1 * time.Nanosecond) // expire immédiatement
	c.Set(scraper.CacheKeyTree, fixtureTree())
	time.Sleep(time.Millisecond)

	h := NewTreeHandler(c)
	r := gin.New()
	r.GET("/api/v1/tree", h.GetTree)

	w := get(r, "/api/v1/tree")
	// Données périmées servies quand même, avec l'âge exposé
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Cache-Age-Seconds"))
}

func TestGetConnections_Pagination(t *testing.T) {
	r, c := setupRouter(t, fixtureTree())
	conns := make([]*models.Connection, 0, 120)
	for i := 0; i < 120; i++ {
		conns = append(conns, &models.Connection{ID: fmt.Sprintf("c-%03d", i), Type: "service"})
	}
	c.Set(scraper.CacheKeyConnections, conns)

	var body struct {
		Connections []*models.Connection `json:"connections"`
		Total       int                  `json:"total"`
		Page        int                  `json:"page"`
		Limit       int                  `json:"limit"`
	}

	// Défauts : page=1, limit=50
	require.NoError(t, json.Unmarshal(get(r, "/api/v1/connections").Body.Bytes(), &body))
	assert.Equal(t, 120, body.Total)
	assert.Len(t, body.Connections, 50)
	assert.Equal(t, 1, body.Page)

	// Page 3 : les 20 restants
	require.NoError(t, json.Unmarshal(get(r, "/api/v1/connections?page=3").Body.Bytes(), &body))
	assert.Len(t, body.Connections, 20)

	// Page au-delà : vide, pas d'erreur
	require.NoError(t, json.Unmarshal(get(r, "/api/v1/connections?page=9").Body.Bytes(), &body))
	assert.Empty(t, body.Connections)

	// Params invalides
	assert.Equal(t, http.StatusBadRequest, get(r, "/api/v1/connections?limit=999").Code)
	assert.Equal(t, http.StatusBadRequest, get(r, "/api/v1/connections?page=0").Code)
}
