package handlers

import (
	"net/http"

	"github.com/aristidewafo/infra-maps-api/internal/cache"
	"github.com/aristidewafo/infra-maps-api/internal/models"
	"github.com/aristidewafo/infra-maps-api/internal/scraper"
	"github.com/gin-gonic/gin"
)

const Version = "0.1.0"

// HealthHandler expose l'état des scrapers et l'âge du cache (API_CONTRACT.md).
// Répond toujours 200 : un scraper en erreur = statut "degraded", pas un 5xx.
type HealthHandler struct {
	Scrapers []scraper.Scraper
	Cache    *cache.Memory
}

func NewHealthHandler(scrapers []scraper.Scraper, c *cache.Memory) *HealthHandler {
	return &HealthHandler{Scrapers: scrapers, Cache: c}
}

func (h *HealthHandler) Health(c *gin.Context) {
	status := models.HealthStatus{
		Status:   "ok",
		Scrapers: make(map[string]models.ScraperHealth, len(h.Scrapers)),
		CacheAge: h.Cache.Age(scraper.CacheKeyTree),
	}

	for _, s := range h.Scrapers {
		if err := s.Health(); err != nil {
			status.Scrapers[s.Name()] = models.ScraperHealth{Status: "error", Message: err.Error()}
			status.Status = "degraded"
			continue
		}
		status.Scrapers[s.Name()] = models.ScraperHealth{Status: "ok"}
	}

	c.JSON(http.StatusOK, status)
}
