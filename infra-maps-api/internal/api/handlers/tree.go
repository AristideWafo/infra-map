package handlers

import (
	"net/http"
	"strconv"

	"github.com/aristidewafo/infra-maps-api/internal/cache"
	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/aristidewafo/infra-maps-api/internal/scraper"
	"github.com/gin-gonic/gin"
)

// TreeHandler sert /tree et /tree/:nodeId depuis le cache. Aucune logique
// métier ici : lecture cache, filtrage, sérialisation.
type TreeHandler struct {
	Cache *cache.Memory
}

func NewTreeHandler(c *cache.Memory) *TreeHandler { return &TreeHandler{Cache: c} }

func (h *TreeHandler) getCachedTree(c *gin.Context) (*models.UnifiedNode, bool) {
	val, ok := h.Cache.Get(scraper.CacheKeyTree)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Cache not ready, scrapers still initializing",
			"code":  "ERR_CACHE_EMPTY",
		})
		return nil, false
	}
	tree, ok := val.(*models.UnifiedNode)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error", "code": "ERR_INTERNAL"})
		return nil, false
	}
	return tree, true
}

func (h *TreeHandler) GetTree(c *gin.Context) {
	tree, ok := h.getCachedTree(c)
	if !ok {
		return
	}
	h.respond(c, tree)
}

func (h *TreeHandler) GetSubtree(c *gin.Context) {
	tree, ok := h.getCachedTree(c)
	if !ok {
		return
	}
	nodeID := c.Param("nodeId")
	sub := findNode(tree, nodeID)
	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Node '" + nodeID + "' not found in cache",
			"code":  "ERR_NODE_NOT_FOUND",
		})
		return
	}
	h.respond(c, sub)
}

func (h *TreeHandler) respond(c *gin.Context, tree *models.UnifiedNode) {
	filters := treeFilters{
		namespace: c.Query("namespace"),
		status:    c.Query("status"),
		tag:       c.Query("tag"),
		source:    c.Query("source"),
		maxDepth:  -1,
	}
	if raw := c.Query("maxDepth"); raw != "" {
		d, err := strconv.Atoi(raw)
		if err != nil || d < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "maxDepth must be a positive integer", "code": "ERR_INVALID_PARAM"})
			return
		}
		filters.maxDepth = d
	}

	c.Header("Cache-Control", "max-age=25")
	c.JSON(http.StatusOK, applyFilters(tree, filters, 0))
}

type treeFilters struct {
	namespace string
	status    string
	tag       string
	source    string
	maxDepth  int
}

func (f treeFilters) empty() bool {
	return f.namespace == "" && f.status == "" && f.tag == "" && f.source == ""
}

// applyFilters retourne une copie élaguée de l'arbre : un nœud est conservé
// s'il correspond aux filtres ou si l'un de ses descendants correspond.
// L'arbre en cache n'est jamais muté.
func applyFilters(n *models.UnifiedNode, f treeFilters, depth int) *models.UnifiedNode {
	copied := *n
	copied.Children = nil

	if f.maxDepth >= 0 && depth >= f.maxDepth {
		return &copied
	}

	for _, child := range n.Children {
		if fc := applyFilters(child, f, depth+1); fc != nil {
			copied.Children = append(copied.Children, fc)
		}
	}

	// La racine et les nœuds ayant des descendants retenus sont conservés.
	if depth == 0 || len(copied.Children) > 0 || matches(n, f) {
		return &copied
	}
	return nil
}

func matches(n *models.UnifiedNode, f treeFilters) bool {
	if f.empty() {
		return true
	}
	if f.namespace != "" && n.Namespace != f.namespace {
		return false
	}
	if f.status != "" && string(n.Status) != f.status {
		return false
	}
	if f.source != "" && n.Source != f.source {
		return false
	}
	if f.tag != "" {
		found := false
		for _, t := range n.Tags {
			if t == f.tag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func findNode(n *models.UnifiedNode, id string) *models.UnifiedNode {
	if n.ID == id {
		return n
	}
	for _, c := range n.Children {
		if found := findNode(c, id); found != nil {
			return found
		}
	}
	return nil
}
