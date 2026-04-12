package services

import (
	"fmt"
	"strings"

	"github.com/wallissonmarinho/GoAI/internal/core/domain"
)

const seriesSystem = `You are an expert anime catalog assistant. Map the series to public database IDs when you can justify them from well-known listings (TheTVDB v4, MyAnimeList, AniDB, AniList, TMDB TV). Jikan is an API over MyAnimeList; the numeric id to return is the MAL anime id (mal_id).
Use public knowledge and the hints given. TheTVDB uses TV seasons; anime often maps season 1 to the main cour.
If torrent_title or torrent_link are present, also infer which episode release they describe: use TheTVDB-style aired season numbering (e.g. "S4", "Season 4", "4th Season", cours split as separate seasons when that is how TheTVDB lists them). Episode is the number after that season marker (e.g. "- 02", "E02", "EP02"). Mark release_is_special true for obvious OVA/ONA/special-only filenames.
Respond with ONLY a single JSON object, no markdown, no code fences, with these keys:
- thetvdb_series_id (integer, 0 if unknown or uncertain)
- mal_id (integer, MyAnimeList anime id, 0 if unknown)
- anidb_aid (integer, AniDB anime id, 0 if unknown)
- anilist_id (integer, AniList media id, 0 if unknown)
- tmdb_tv_id (integer, TMDB TV series id, 0 if unknown)
- release_season (integer, TheTVDB aired-order season this file refers to; 0 if no torrent fields or not inferable)
- release_episode (integer, episode number for that season; 0 if unknown)
- release_is_special (boolean)
- confidence (number from 0 to 1; lower if any non-zero id is uncertain)
- notes (short string, English; mention which ids are solid vs guessed)
- thetvdb_name (optional string, suggested official English or primary listing title for humans and API search)
- thetvdb_slug (optional string, URL slug as on thetvdb.com series pages, e.g. ascendance-of-a-bookworm; no domain or path prefix)

If an existing_* hint in the input is non-zero and matches the hints, you may return the same value with high confidence for that field.`

func buildSeriesPrompt(in domain.SeriesAuditRequest) string {
	var b strings.Builder
	b.WriteString(seriesSystem)
	b.WriteString("\n\nInput JSON:\n")
	displayName := strings.TrimSpace(in.SeriesName)
	if displayName == "" {
		displayName = strings.TrimSpace(in.TorrentTitle)
	}
	b.WriteString(fmt.Sprintf(`{"series_name":%q`, displayName))
	if strings.TrimSpace(in.TorrentTitle) != "" {
		b.WriteString(fmt.Sprintf(`,"torrent_title":%q`, strings.TrimSpace(in.TorrentTitle)))
	}
	if strings.TrimSpace(in.TorrentLink) != "" {
		b.WriteString(fmt.Sprintf(`,"torrent_link":%q`, strings.TrimSpace(in.TorrentLink)))
	}
	if in.SeriesID != "" {
		b.WriteString(fmt.Sprintf(`,"series_id":%q`, in.SeriesID))
	}
	if in.MalID > 0 {
		b.WriteString(fmt.Sprintf(`,"mal_id":%d`, in.MalID))
	}
	if strings.TrimSpace(in.ImdbID) != "" {
		b.WriteString(fmt.Sprintf(`,"imdb_id":%q`, strings.TrimSpace(in.ImdbID)))
	}
	if in.Year > 0 {
		b.WriteString(fmt.Sprintf(`,"year":%d`, in.Year))
	}
	if strings.TrimSpace(in.TitlePreferred) != "" {
		b.WriteString(fmt.Sprintf(`,"title_preferred":%q`, strings.TrimSpace(in.TitlePreferred)))
	}
	if strings.TrimSpace(in.TitleNative) != "" {
		b.WriteString(fmt.Sprintf(`,"title_native":%q`, strings.TrimSpace(in.TitleNative)))
	}
	if in.ExistingTVDBSeriesID > 0 {
		b.WriteString(fmt.Sprintf(`,"existing_tvdb_series_id":%d`, in.ExistingTVDBSeriesID))
	}
	if in.ExistingAniDBAID > 0 {
		b.WriteString(fmt.Sprintf(`,"existing_anidb_aid":%d`, in.ExistingAniDBAID))
	}
	if in.ExistingAniListID > 0 {
		b.WriteString(fmt.Sprintf(`,"existing_anilist_id":%d`, in.ExistingAniListID))
	}
	if in.ExistingTMDBTVID > 0 {
		b.WriteString(fmt.Sprintf(`,"existing_tmdb_tv_id":%d`, in.ExistingTMDBTVID))
	}
	b.WriteString("}\n")
	return b.String()
}

const releaseSystem = `You parse anime torrent release titles. Infer broadcast season number, episode number, and whether it is a special (OVA/ONA/special marked as S0 or SP in filename patterns).
Respond with ONLY a single JSON object, no markdown, with keys:
- season (integer, >=1 for regular cours; use 0 only if truly unknown)
- episode (integer, >=0; specials sometimes use 0)
- is_special (boolean)
- confidence (0 to 1)
- notes (short string, English)

Use current_season/current_episode/is_special from the parser as hints; correct them if the torrent title clearly implies otherwise.`

func buildReleasePrompt(in domain.ReleaseAuditRequest) string {
	var b strings.Builder
	b.WriteString(releaseSystem)
	b.WriteString("\n\n")
	b.WriteString(fmt.Sprintf("torrent_title: %q\n", in.TorrentTitle))
	if strings.TrimSpace(in.SeriesName) != "" {
		b.WriteString(fmt.Sprintf("series_name: %q\n", strings.TrimSpace(in.SeriesName)))
	}
	if strings.TrimSpace(in.SeriesID) != "" {
		b.WriteString(fmt.Sprintf("series_id: %q\n", strings.TrimSpace(in.SeriesID)))
	}
	b.WriteString(fmt.Sprintf("current_season: %d\ncurrent_episode: %d\nis_special: %v\n", in.CurrentSeason, in.CurrentEpisode, in.IsSpecial))
	return b.String()
}
