package ginapi

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// BearerAuth returns 401 when Authorization does not match "Bearer <expected>".
func BearerAuth(expected string) gin.HandlerFunc {
	want := strings.TrimSpace(expected)
	return func(c *gin.Context) {
		if want == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "server_misconfigured_auth"})
			c.Abort()
			return
		}
		h := strings.TrimSpace(c.GetHeader("Authorization"))
		const p = "Bearer "
		if !strings.HasPrefix(h, p) || strings.TrimSpace(strings.TrimPrefix(h, p)) != want {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}
