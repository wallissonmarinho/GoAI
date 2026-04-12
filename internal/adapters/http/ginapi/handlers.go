package ginapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/wallissonmarinho/GoAI/internal/core/domain"
	"github.com/wallissonmarinho/GoAI/internal/core/ports"
)

// Handler is the driving HTTP adapter (depends on ports.AuditService).
type Handler struct {
	Audit ports.AuditService
	// GeminiKeyCheck runs a lightweight list-models probe per configured key (nil disables the route).
	GeminiKeyCheck func(context.Context) []domain.GeminiKeyCheckResult
}

func (h *Handler) getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *Handler) getDiagGeminiKeys(c *gin.Context) {
	if h == nil || h.GeminiKeyCheck == nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": "gemini_key_check_disabled"})
		return
	}
	results := h.GeminiKeyCheck(c.Request.Context())
	okN := 0
	for _, r := range results {
		if r.OK {
			okN++
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"keys_total":   len(results),
		"keys_ok":      okN,
		"keys_checked": results,
	})
}

func (h *Handler) postAuditSeries(c *gin.Context) {
	if h == nil || h.Audit == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service_unavailable"})
		return
	}
	var req domain.SeriesAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_json"})
		return
	}
	out, err := h.Audit.AuditSeries(c.Request.Context(), req)
	if err != nil {
		writeAuditErr(c, err)
		return
	}
	raw, _ := json.Marshal(out)
	c.JSON(http.StatusOK, gin.H{
		"prompt_version":     h.Audit.PromptVersion(),
		"thetvdb_series_id":  out.TheTVDBSeriesID,
		"thetvdb_series_url": out.TheTVDBSeriesURL,
		"mal_id":             out.MalID,
		"anidb_aid":          out.AniDBAID,
		"anilist_id":         out.AniListID,
		"tmdb_tv_id":         out.TMDBTVID,
		"release_season":     out.ReleaseSeason,
		"release_episode":    out.ReleaseEpisode,
		"release_is_special": out.ReleaseIsSpecial,
		"confidence":         out.Confidence,
		"notes":              out.Notes,
		"thetvdb_name":       out.TheTVDBName,
		"thetvdb_slug":       out.TheTVDBSlug,
		"raw_response_json":  string(raw),
	})
}

func (h *Handler) postAuditRelease(c *gin.Context) {
	if h == nil || h.Audit == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "service_unavailable"})
		return
	}
	var req domain.ReleaseAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "invalid_json"})
		return
	}
	out, err := h.Audit.AuditRelease(c.Request.Context(), req)
	if err != nil {
		writeAuditErr(c, err)
		return
	}
	raw, _ := json.Marshal(out)
	c.JSON(http.StatusOK, gin.H{
		"prompt_version":    h.Audit.PromptVersion(),
		"season":            out.Season,
		"episode":           out.Episode,
		"is_special":        out.IsSpecial,
		"confidence":        out.Confidence,
		"notes":             out.Notes,
		"raw_response_json": string(raw),
	})
}

func writeAuditErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrLLMNoCapacity):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "gemini_keys_exhausted"})
	case errors.Is(err, domain.ErrLLMQuotaOrRate):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "gemini_quota"})
	default:
		if strings.Contains(err.Error(), "parse") {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
