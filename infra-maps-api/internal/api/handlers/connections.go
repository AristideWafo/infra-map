package handlers

import (
	"net/http"
	"strconv"

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

	total := len(filtered)
	page, limit, ok := paginationParams(c)
	if !ok {
		return
	}
	start := (page - 1) * limit
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	c.Header("Cache-Control", "max-age=25")
	c.JSON(http.StatusOK, gin.H{
		"connections": filtered[start:end],
		"total":       total,
		"page":        page,
		"limit":       limit,
	})
}

// paginationParams applique les conventions du contrat : page>=1,
// limit défaut 50, max 200. Écrit un 400 ERR_INVALID_PARAM si invalide.
func paginationParams(c *gin.Context) (page, limit int, ok bool) {
	page, limit = 1, 50
	if raw := c.Query("page"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "page must be >= 1", "code": "ERR_INVALID_PARAM"})
			return 0, 0, false
		}
		page = v
	}
	if raw := c.Query("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 200 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be 1-200", "code": "ERR_INVALID_PARAM"})
			return 0, 0, false
		}
		limit = v
	}
	return page, limit, true
}
