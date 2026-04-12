package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/wallissonmarinho/GoAI/internal/core/domain"
	"github.com/wallissonmarinho/GoAI/internal/core/ports"
)

// AuditService implements ports.AuditService (hex core: depends on TextCompletion port only).
type AuditService struct {
	llm ports.TextCompletion
}

// NewAuditService wires the audit use-case.
func NewAuditService(llm ports.TextCompletion) *AuditService {
	return &AuditService{llm: llm}
}

// PromptVersion returns the prompt schema version for downstream caches.
func (s *AuditService) PromptVersion() int {
	return domain.AuditPromptVersion
}

// AuditSeries maps a show to public catalog ids (TheTVDB, MAL, AniDB, AniList, TMDB) best effort via LLM.
// Optional torrent_title/torrent_link add release_season / release_episode inference in the same call.
func (s *AuditService) AuditSeries(ctx context.Context, in domain.SeriesAuditRequest) (domain.SeriesAuditResponse, error) {
	if s == nil || s.llm == nil {
		return domain.SeriesAuditResponse{}, errors.New("audit: nil service")
	}
	if strings.TrimSpace(in.SeriesName) == "" && strings.TrimSpace(in.TorrentTitle) == "" {
		return domain.SeriesAuditResponse{}, fmt.Errorf("audit: series_name or torrent_title required")
	}
	text, err := s.llm.GenerateText(ctx, buildSeriesPrompt(in))
	if err != nil {
		return domain.SeriesAuditResponse{}, err
	}
	out, err := domain.ParseSeriesAuditResponse(text)
	if err != nil {
		return domain.SeriesAuditResponse{}, fmt.Errorf("audit: parse series json: %w", err)
	}
	normalizeSeriesResponse(&out)
	return out, nil
}

// AuditRelease infers season/episode from a torrent title.
func (s *AuditService) AuditRelease(ctx context.Context, in domain.ReleaseAuditRequest) (domain.ReleaseAuditResponse, error) {
	if s == nil || s.llm == nil {
		return domain.ReleaseAuditResponse{}, errors.New("audit: nil service")
	}
	if strings.TrimSpace(in.TorrentTitle) == "" {
		return domain.ReleaseAuditResponse{}, fmt.Errorf("audit: torrent_title required")
	}
	text, err := s.llm.GenerateText(ctx, buildReleasePrompt(in))
	if err != nil {
		return domain.ReleaseAuditResponse{}, err
	}
	out, err := domain.ParseReleaseAuditResponse(text)
	if err != nil {
		return domain.ReleaseAuditResponse{}, fmt.Errorf("audit: parse release json: %w", err)
	}
	normalizeReleaseResponse(&out)
	return out, nil
}

func normalizeSeriesResponse(out *domain.SeriesAuditResponse) {
	if out.Confidence < 0 {
		out.Confidence = 0
	}
	if out.Confidence > 1 {
		out.Confidence = 1
	}
	if out.ReleaseSeason < 0 {
		out.ReleaseSeason = 0
	}
	if out.ReleaseEpisode < 0 {
		out.ReleaseEpisode = 0
	}
	slug := strings.Trim(strings.TrimSpace(out.TheTVDBSlug), "/")
	if slug != "" {
		out.TheTVDBSeriesURL = "https://www.thetvdb.com/series/" + slug
	} else {
		out.TheTVDBSeriesURL = ""
	}
}

func normalizeReleaseResponse(out *domain.ReleaseAuditResponse) {
	if out.Confidence < 0 {
		out.Confidence = 0
	}
	if out.Confidence > 1 {
		out.Confidence = 1
	}
	if out.Season < 0 {
		out.Season = 1
	}
	if out.Episode < 0 {
		out.Episode = 0
	}
}

var _ ports.AuditService = (*AuditService)(nil)
