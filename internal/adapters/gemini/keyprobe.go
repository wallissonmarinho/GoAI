package gemini

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/wallissonmarinho/GoAI/internal/core/domain"
)

// CheckKeys calls the Gemini REST list-models endpoint once per key (lightweight; 429/503 ⇒ quota/rate).
func CheckKeys(ctx context.Context, hc *http.Client, baseURL, userAgent string, keys []string) []domain.GeminiKeyCheckResult {
	if len(keys) == 0 {
		return nil
	}
	base := strings.TrimSuffix(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = defaultBaseURL
	}
	if hc == nil {
		hc = http.DefaultClient
	}
	ua := strings.TrimSpace(userAgent)
	if ua == "" {
		ua = "GoAI/1.0"
	}

	out := make([]domain.GeminiKeyCheckResult, 0, len(keys))
	for i, apiKey := range keys {
		out = append(out, probeListModels(ctx, hc, base, ua, i, apiKey))
	}
	return out
}

type listModelsResponse struct {
	Models []json.RawMessage `json:"models"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error"`
}

func probeListModels(ctx context.Context, hc *http.Client, base, ua string, index int, apiKey string) domain.GeminiKeyCheckResult {
	r := domain.GeminiKeyCheckResult{Index: index}
	u := base + "/v1beta/models?pageSize=1&key=" + url.QueryEscape(apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		r.Status = "request_error"
		r.Detail = err.Error()
		return r
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", ua)

	resp, err := hc.Do(req)
	if err != nil {
		r.Status = "network_error"
		r.Detail = err.Error()
		return r
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		r.Status = "read_error"
		r.Detail = err.Error()
		return r
	}
	r.HTTPStatus = resp.StatusCode

	switch resp.StatusCode {
	case http.StatusOK:
		var parsed listModelsResponse
		_ = json.Unmarshal(b, &parsed)
		if parsed.Error != nil {
			r.Status = classifyAPIError(parsed.Error.Message, parsed.Error.Status)
			r.Detail = trimDetail(parsed.Error.Message, 500)
			return r
		}
		r.OK = true
		r.Status = "ok"
		return r
	case http.StatusTooManyRequests, http.StatusServiceUnavailable:
		r.Status = "quota_or_rate"
		r.Detail = trimDetail(string(b), 500)
		return r
	case http.StatusUnauthorized, http.StatusForbidden:
		r.Status = "auth_error"
		r.Detail = trimDetail(string(b), 500)
		return r
	default:
		var parsed listModelsResponse
		_ = json.Unmarshal(b, &parsed)
		if parsed.Error != nil {
			msg := strings.ToLower(parsed.Error.Message + " " + parsed.Error.Status)
			if strings.Contains(msg, "quota") || strings.Contains(msg, "resource exhausted") || strings.Contains(msg, "rate") {
				r.Status = "quota_or_rate"
				r.Detail = trimDetail(parsed.Error.Message, 500)
				return r
			}
		}
		r.Status = "http_error"
		r.Detail = trimDetail(string(b), 500)
		return r
	}
}

func classifyAPIError(message, status string) string {
	msg := strings.ToLower(message + " " + status)
	if strings.Contains(msg, "quota") || strings.Contains(msg, "resource exhausted") || strings.Contains(msg, "rate") {
		return "quota_or_rate"
	}
	if strings.Contains(msg, "permission") || strings.Contains(msg, "api key") || strings.Contains(msg, "invalid") {
		return "auth_error"
	}
	return "api_error"
}

func trimDetail(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
