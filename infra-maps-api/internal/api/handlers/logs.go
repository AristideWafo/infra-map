package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/gin-gonic/gin"
)

// LogsProvider fournit les logs récents d'un nœud (Loki).
type LogsProvider interface {
	Enabled() bool
	Logs(ctx context.Context, nodeName string, limit int, from time.Time, level string) ([]models.LogEntry, error)
}

type LogsHandler struct {
	Provider LogsProvider
	Tree     *TreeHandler
}

func NewLogsHandler(p LogsProvider, tree *TreeHandler) *LogsHandler {
	return &LogsHandler{Provider: p, Tree: tree}
}

// GetLogs — GET /nodes/:nodeId/logs (API_CONTRACT.md).
func (h *LogsHandler) GetLogs(c *gin.Context) {
	if h.Provider == nil || !h.Provider.Enabled() {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loki is not configured", "code": "ERR_LOKI_DISABLED"})
		return
	}

	nodeID := c.Param("nodeId")

	// Loki filtre par nom de pod, pas par UID : résoudre le nom via l'arbre.
	nodeName := nodeID
	if tree, ok := h.Tree.getCachedTree(c); ok {
		if node := findNode(tree, nodeID); node != nil && node.Name != "" {
			nodeName = node.Name
		}
	} else {
		return // getCachedTree a déjà écrit la réponse 503
	}

	limit := 100
	if raw := c.Query("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v <= 0 || v > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be 1-500", "code": "ERR_INVALID_PARAM"})
			return
		}
		limit = v
	}
	from, ok := unixParam(c, "from", time.Now().Add(-15*time.Minute))
	if !ok {
		return
	}

	entries, err := h.Provider.Logs(c.Request.Context(), nodeName, limit, from, c.Query("level"))
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error(), "code": "ERR_SCRAPER_UNAVAILABLE"})
		return
	}
	if entries == nil {
		entries = []models.LogEntry{}
	}
	c.JSON(http.StatusOK, gin.H{"nodeId": nodeID, "entries": entries})
}
