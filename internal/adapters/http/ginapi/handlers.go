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

type errorResponse struct {
	Error string `json:"error"`
}

type healthResponse struct {
	OK bool `json:"ok"`
}

type diagGeminiKeysResponse struct {
	KeysTotal   int                           `json:"keys_total"`
	KeysOK      int                           `json:"keys_ok"`
	KeysChecked []domain.GeminiKeyCheckResult `json:"keys_checked"`
}

type seriesAuditResponse struct {
	PromptVersion    int     `json:"prompt_version"`
	TheTVDBSeriesID  int     `json:"thetvdb_series_id"`
	TheTVDBSeriesURL string  `json:"thetvdb_series_url"`
	MalID            int     `json:"mal_id"`
	AniDBAID         int     `json:"anidb_aid"`
	AniListID        int     `json:"anilist_id"`
	TMDBTVID         int     `json:"tmdb_tv_id"`
	ReleaseSeason    int     `json:"release_season"`
	ReleaseEpisode   int     `json:"release_episode"`
	ReleaseIsSpecial bool    `json:"release_is_special"`
	Confidence       float64 `json:"confidence"`
	Notes            string  `json:"notes,omitempty"`
	TheTVDBName      string  `json:"thetvdb_name,omitempty"`
	TheTVDBSlug      string  `json:"thetvdb_slug,omitempty"`
	RawResponseJSON  string  `json:"raw_response_json"`
}

type releaseAuditResponse struct {
	PromptVersion   int     `json:"prompt_version"`
	Season          int     `json:"season"`
	Episode         int     `json:"episode"`
	IsSpecial       bool    `json:"is_special"`
	Confidence      float64 `json:"confidence"`
	Notes           string  `json:"notes,omitempty"`
	RawResponseJSON string  `json:"raw_response_json"`
}

func (h *Handler) getHealth(c *gin.Context) {
	c.JSON(http.StatusOK, healthResponse{OK: true})
}

func (h *Handler) getDiagGeminiKeys(c *gin.Context) {
	if h == nil || h.GeminiKeyCheck == nil {
		c.JSON(http.StatusNotImplemented, errorResponse{Error: "gemini_key_check_disabled"})
		return
	}
	results := h.GeminiKeyCheck(c.Request.Context())
	okN := 0
	for _, r := range results {
		if r.OK {
			okN++
		}
	}
	c.JSON(http.StatusOK, diagGeminiKeysResponse{KeysTotal: len(results), KeysOK: okN, KeysChecked: results})
}

func (h *Handler) postAuditSeries(c *gin.Context) {
	if h == nil || h.Audit == nil {
		c.JSON(http.StatusServiceUnavailable, errorResponse{Error: "service_unavailable"})
		return
	}
	var req domain.SeriesAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, errorResponse{Error: "invalid_json"})
		return
	}
	out, err := h.Audit.AuditSeries(c.Request.Context(), req)
	if err != nil {
		writeAuditErr(c, err)
		return
	}
	raw, _ := json.Marshal(out)
	c.JSON(http.StatusOK, seriesAuditResponse{
		PromptVersion:    h.Audit.PromptVersion(),
		TheTVDBSeriesID:  out.TheTVDBSeriesID,
		TheTVDBSeriesURL: out.TheTVDBSeriesURL,
		MalID:            out.MalID,
		AniDBAID:         out.AniDBAID,
		AniListID:        out.AniListID,
		TMDBTVID:         out.TMDBTVID,
		ReleaseSeason:    out.ReleaseSeason,
		ReleaseEpisode:   out.ReleaseEpisode,
		ReleaseIsSpecial: out.ReleaseIsSpecial,
		Confidence:       out.Confidence,
		Notes:            out.Notes,
		TheTVDBName:      out.TheTVDBName,
		TheTVDBSlug:      out.TheTVDBSlug,
		RawResponseJSON:  string(raw),
	})
}

func (h *Handler) postAuditRelease(c *gin.Context) {
	if h == nil || h.Audit == nil {
		c.JSON(http.StatusServiceUnavailable, errorResponse{Error: "service_unavailable"})
		return
	}
	var req domain.ReleaseAuditRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, errorResponse{Error: "invalid_json"})
		return
	}
	out, err := h.Audit.AuditRelease(c.Request.Context(), req)
	if err != nil {
		writeAuditErr(c, err)
		return
	}
	raw, _ := json.Marshal(out)
	c.JSON(http.StatusOK, releaseAuditResponse{
		PromptVersion:   h.Audit.PromptVersion(),
		Season:          out.Season,
		Episode:         out.Episode,
		IsSpecial:       out.IsSpecial,
		Confidence:      out.Confidence,
		Notes:           out.Notes,
		RawResponseJSON: string(raw),
	})
}

func writeAuditErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrLLMNoCapacity):
		c.JSON(http.StatusServiceUnavailable, errorResponse{Error: "gemini_keys_exhausted"})
	case errors.Is(err, domain.ErrLLMQuotaOrRate):
		c.JSON(http.StatusServiceUnavailable, errorResponse{Error: "gemini_quota"})
	default:
		if strings.Contains(err.Error(), "parse") {
			c.JSON(http.StatusUnprocessableEntity, errorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, errorResponse{Error: err.Error()})
	}
}
