package ginapi

import (
	"github.com/gin-gonic/gin"
)

// NewRouter builds Gin engine: GET /healthz (public), /v1/* (Bearer).
func NewRouter(internalAPIKey string, h *Handler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	e := gin.New()
	e.Use(gin.Recovery())
	e.GET("/healthz", h.getHealth)

	v1 := e.Group("/v1")
	v1.Use(BearerAuth(internalAPIKey))
	v1.POST("/audit/series", h.postAuditSeries)
	v1.POST("/audit/release", h.postAuditRelease)
	v1.GET("/diag/gemini-keys", h.getDiagGeminiKeys)
	return e
}
