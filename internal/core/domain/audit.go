package domain

// AuditPromptVersion bump when changing system instructions (consumers may invalidate cache).
const AuditPromptVersion = 6

// SeriesAuditRequest is the POST /v1/audit/series JSON body.
// Provide series_name and/or torrent_title (at least one). torrent_link is optional context (path, group, season hints).
type SeriesAuditRequest struct {
	SeriesName           string `json:"series_name,omitempty"`
	TorrentTitle         string `json:"torrent_title,omitempty"`
	TorrentLink          string `json:"torrent_link,omitempty"`
	FeedPublishedAt      string `json:"feed_published_at,omitempty"`
	ParsedSeasonHint     int    `json:"parsed_season_hint,omitempty"`
	ParsedEpisodeHint    int    `json:"parsed_episode_hint,omitempty"`
	ParsedIsSpecialHint  bool   `json:"parsed_is_special_hint,omitempty"`
	SeriesID             string `json:"series_id,omitempty"`
	MalID                int    `json:"mal_id,omitempty"`
	ImdbID               string `json:"imdb_id,omitempty"`
	Year                 int    `json:"year,omitempty"`
	TitlePreferred       string `json:"title_preferred,omitempty"`
	TitleNative          string `json:"title_native,omitempty"`
	ExistingTVDBSeriesID int    `json:"existing_tvdb_series_id,omitempty"`
	ExistingAniDBAID     int    `json:"existing_anidb_aid,omitempty"`
	ExistingAniListID    int    `json:"existing_anilist_id,omitempty"`
	ExistingTMDBTVID     int    `json:"existing_tmdb_tv_id,omitempty"`
}

// SeriesAuditResponse is validated model output (0 = unknown / do not trust for an ID field).
// TheTVDBSeriesURL is filled server-side from thetvdb_slug when possible (not from the model).
type SeriesAuditResponse struct {
	TheTVDBSeriesID  int     `json:"thetvdb_series_id"`
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
	TheTVDBSeriesURL string  `json:"thetvdb_series_url,omitempty"`
}

// ReleaseAuditRequest is the POST /v1/audit/release JSON body.
type ReleaseAuditRequest struct {
	TorrentTitle   string `json:"torrent_title"`
	TorrentLink    string `json:"torrent_link,omitempty"`
	FeedPublishedAt string `json:"feed_published_at,omitempty"`
	SeriesName     string `json:"series_name,omitempty"`
	SeriesID       string `json:"series_id,omitempty"`
	CurrentSeason  int    `json:"current_season,omitempty"`
	CurrentEpisode int    `json:"current_episode,omitempty"`
	IsSpecial      bool   `json:"is_special,omitempty"`
}

// ReleaseAuditResponse is validated model output.
type ReleaseAuditResponse struct {
	Season     int     `json:"season"`
	Episode    int     `json:"episode"`
	IsSpecial  bool    `json:"is_special"`
	Confidence float64 `json:"confidence"`
	Notes      string  `json:"notes,omitempty"`
}
