package api

import (
	"github.com/aristidewafo/infra-maps-api/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

// Register branche toutes les routes documentées dans API_CONTRACT.md.
// Aucun endpoint non documenté.
func Register(r *gin.Engine, tree *handlers.TreeHandler, health *handlers.HealthHandler, metrics *handlers.MetricsHandler, logs *handlers.LogsHandler) {
	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", health.Health)
		v1.GET("/tree", tree.GetTree)
		v1.GET("/tree/:nodeId", tree.GetSubtree)
		v1.GET("/nodes/:nodeId", tree.GetNode)
		v1.GET("/nodes/:nodeId/metrics", metrics.GetMetrics)
		v1.GET("/nodes/:nodeId/logs", logs.GetLogs)
		v1.GET("/connections", tree.GetConnections)
		v1.GET("/tags", tree.GetTags)
	}
}
