package app

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wallissonmarinho/GoAI/internal/adapters/gemini"
	ginapi "github.com/wallissonmarinho/GoAI/internal/adapters/http/ginapi"
	"github.com/wallissonmarinho/GoAI/internal/core/domain"
	"github.com/wallissonmarinho/GoAI/internal/core/ports"
	"github.com/wallissonmarinho/GoAI/internal/core/services"
)

// Config holds composition-root options for the HTTP server.
type Config struct {
	InternalAPIKey string
	GeminiKeysCSV  string
	GeminiModel    string
	GeminiBaseURL  string
	KeyCooldown    time.Duration
	UserAgent      string
}

// Wire returns the Gin engine with hex layers wired: ginapi → services → ports.TextCompletion → gemini.
func Wire(cfg Config) *gin.Engine {
	pool := gemini.NewKeyPool(cfg.GeminiKeysCSV, cfg.KeyCooldown)
	gc := &gemini.Client{
		Model:     cfg.GeminiModel,
		BaseURL:   cfg.GeminiBaseURL,
		KeyPool:   pool,
		Timeout:   90 * time.Second,
		UserAgent: strings.TrimSpace(cfg.UserAgent),
	}
	var llm ports.TextCompletion = gc
	urlCheckerClient := &http.Client{Timeout: 10 * time.Second}
	audit := services.NewAuditServiceWithURLChecker(llm, newHTTPURLExistenceChecker(urlCheckerClient, cfg.UserAgent))
	probeClient := &http.Client{Timeout: 45 * time.Second}
	h := &ginapi.Handler{
		Audit: audit,
		GeminiKeyCheck: func(ctx context.Context) []domain.GeminiKeyCheckResult {
			return gemini.CheckKeys(ctx, probeClient, cfg.GeminiBaseURL, cfg.UserAgent, pool.SnapshotKeys())
		},
	}
	return ginapi.NewRouter(cfg.InternalAPIKey, h)
}
