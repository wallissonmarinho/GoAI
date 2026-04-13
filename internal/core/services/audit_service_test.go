package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/wallissonmarinho/GoAI/internal/core/domain"
)

type fakeLLM struct {
	text string
	err  error
}

func (f fakeLLM) GenerateText(_ context.Context, _ string) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.text, nil
}

type fakeURLChecker struct {
	ok  bool
	err error
}

func (f fakeURLChecker) Exists(_ context.Context, _ string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return f.ok, nil
}

func TestAuditSeries_ClearsTVDBWhenURLDoesNotExist(t *testing.T) {
	svc := NewAuditServiceWithURLChecker(fakeLLM{
		text: `{"thetvdb_series_id":454178,"mal_id":58824,"anidb_aid":18584,"anilist_id":176413,"tmdb_tv_id":257790,"release_season":1,"release_episode":1,"release_is_special":false,"confidence":0.95,"notes":"mapped","thetvdb_name":"I Became Friends with the Second Cutest Girl in Class","thetvdb_slug":"i-became-friends-with-the-second-cutest-girl-in-class"}`,
	}, fakeURLChecker{ok: false})

	out, err := svc.AuditSeries(context.Background(), domain.SeriesAuditRequest{SeriesName: "test"})
	if err != nil {
		t.Fatalf("AuditSeries() error = %v", err)
	}
	if out.TheTVDBSeriesID != 0 {
		t.Fatalf("TheTVDBSeriesID = %d, want 0", out.TheTVDBSeriesID)
	}
	if out.TheTVDBSeriesURL != "" || out.TheTVDBSlug != "" || out.TheTVDBName != "" {
		t.Fatalf("expected tvdb URL fields cleared, got url=%q slug=%q name=%q", out.TheTVDBSeriesURL, out.TheTVDBSlug, out.TheTVDBName)
	}
	if out.Confidence > 0.6 {
		t.Fatalf("Confidence = %v, want <= 0.6", out.Confidence)
	}
	if !strings.Contains(out.Notes, "cleared tvdb mapping") {
		t.Fatalf("notes %q does not mention tvdb clear", out.Notes)
	}
}

func TestAuditSeries_KeepsTVDBWhenValidationErrors(t *testing.T) {
	svc := NewAuditServiceWithURLChecker(fakeLLM{
		text: `{"thetvdb_series_id":74796,"mal_id":21,"anidb_aid":69,"anilist_id":21,"tmdb_tv_id":37854,"release_season":0,"release_episode":63,"release_is_special":true,"confidence":0.95,"notes":"mapped","thetvdb_name":"One Piece","thetvdb_slug":"one-piece"}`,
	}, fakeURLChecker{err: errors.New("timeout")})

	out, err := svc.AuditSeries(context.Background(), domain.SeriesAuditRequest{SeriesName: "One Piece"})
	if err != nil {
		t.Fatalf("AuditSeries() error = %v", err)
	}
	if out.TheTVDBSeriesID != 74796 {
		t.Fatalf("TheTVDBSeriesID = %d, want 74796", out.TheTVDBSeriesID)
	}
	if out.TheTVDBSeriesURL == "" {
		t.Fatalf("expected TheTVDBSeriesURL to remain populated")
	}
}
