package ginapi

import (
	"github.com/gin-gonic/gin"
)

// RoutesHandler is the HTTP contract expected by the router.
type RoutesHandler interface {
	getHealth(c *gin.Context)
	postAuditSeries(c *gin.Context)
	postAuditRelease(c *gin.Context)
	getDiagGeminiKeys(c *gin.Context)
}

// NewRouter builds Gin engine: GET /healthz (public), /v1/* (Bearer).
func NewRouter(internalAPIKey string, h RoutesHandler) *gin.Engine {
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
