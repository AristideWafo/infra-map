package handlers

import (
	"net/http"

	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/aristidewafo/infra-maps-api/internal/scraper"
	"github.com/gin-gonic/gin"
)

// GetConnections retourne les connexions découvertes, filtrables par
// fromId / toId / type (API_CONTRACT.md).
func (h *TreeHandler) GetConnections(c *gin.Context) {
	val, ok := h.Cache.Get(scraper.CacheKeyConnections)
	if !ok {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Cache not ready, scrapers still initializing",
			"code":  "ERR_CACHE_EMPTY",
		})
		return
	}
	conns, ok := val.([]*models.Connection)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error", "code": "ERR_INTERNAL"})
		return
	}

	fromID, toID, typ := c.Query("fromId"), c.Query("toId"), c.Query("type")
	filtered := make([]*models.Connection, 0, len(conns))
	for _, cn := range conns {
		if fromID != "" && cn.FromID != fromID {
			continue
		}
		if toID != "" && cn.ToID != toID {
			continue
		}
		if typ != "" && cn.Type != typ {
			continue
		}
		filtered = append(filtered, cn)
	}

	c.Header("Cache-Control", "max-age=25")
	c.JSON(http.StatusOK, gin.H{"connections": filtered, "total": len(filtered)})
}
