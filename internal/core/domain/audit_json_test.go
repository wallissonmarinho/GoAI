package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseSeriesAuditResponse_fence(t *testing.T) {
	raw := "```json\n{\"thetvdb_series_id\":42,\"mal_id\":1,\"anidb_aid\":2,\"anilist_id\":3,\"tmdb_tv_id\":4,\"release_season\":4,\"release_episode\":2,\"release_is_special\":false,\"confidence\":0.9,\"notes\":\"ok\",\"thetvdb_slug\":\"ascendance-of-a-bookworm\"}\n```"
	out, err := ParseSeriesAuditResponse(raw)
	require.NoError(t, err)
	require.Equal(t, 42, out.TheTVDBSeriesID)
	require.Equal(t, 1, out.MalID)
	require.Equal(t, 2, out.AniDBAID)
	require.Equal(t, 3, out.AniListID)
	require.Equal(t, 4, out.TMDBTVID)
	require.Equal(t, 4, out.ReleaseSeason)
	require.Equal(t, 2, out.ReleaseEpisode)
	require.False(t, out.ReleaseIsSpecial)
	require.InDelta(t, 0.9, out.Confidence, 0.001)
}

func TestParseReleaseAuditResponse_plain(t *testing.T) {
	out, err := ParseReleaseAuditResponse(`{"season":2,"episode":3,"is_special":false,"confidence":0.5}`)
	require.NoError(t, err)
	require.Equal(t, 2, out.Season)
	require.Equal(t, 3, out.Episode)
	require.False(t, out.IsSpecial)
}
