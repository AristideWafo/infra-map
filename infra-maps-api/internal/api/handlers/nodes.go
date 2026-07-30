package handlers

import (
	"net/http"
	"sort"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/gin-gonic/gin"
)

// GetNode retourne les détails d'un nœud sans ses enfants.
func (h *TreeHandler) GetNode(c *gin.Context) {
	tree, ok := h.getCachedTree(c)
	if !ok {
		return
	}
	nodeID := c.Param("nodeId")
	node := findNode(tree, nodeID)
	if node == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Node '" + nodeID + "' not found in cache",
			"code":  "ERR_NODE_NOT_FOUND",
		})
		return
	}
	detail := *node
	detail.Children = nil
	c.Header("Cache-Control", "max-age=25")
	c.JSON(http.StatusOK, &detail)
}

// GetTags retourne la liste triée et dédupliquée des tags de l'arbre.
func (h *TreeHandler) GetTags(c *gin.Context) {
	tree, ok := h.getCachedTree(c)
	if !ok {
		return
	}
	seen := map[string]struct{}{}
	collectTags(tree, seen)

	tags := make([]string, 0, len(seen))
	for t := range seen {
		tags = append(tags, t)
	}
	sort.Strings(tags)

	c.Header("Cache-Control", "max-age=25")
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

func collectTags(n *models.UnifiedNode, seen map[string]struct{}) {
	for _, t := range n.Tags {
		seen[t] = struct{}{}
	}
	for _, c := range n.Children {
		collectTags(c, seen)
	}
}
