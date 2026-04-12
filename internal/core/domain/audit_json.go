package domain

import (
	"encoding/json"
	"strings"
)

func unwrapJSONFence(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimPrefix(s, "json")
		s = strings.TrimSpace(s)
		if i := strings.LastIndex(s, "```"); i >= 0 {
			s = strings.TrimSpace(s[:i])
		}
	}
	return strings.TrimSpace(s)
}

// ParseSeriesAuditResponse decodes JSON from LLM text.
func ParseSeriesAuditResponse(text string) (SeriesAuditResponse, error) {
	var out SeriesAuditResponse
	raw := unwrapJSONFence(text)
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return out, err
	}
	return out, nil
}

// ParseReleaseAuditResponse decodes JSON from LLM text.
func ParseReleaseAuditResponse(text string) (ReleaseAuditResponse, error) {
	var out ReleaseAuditResponse
	raw := unwrapJSONFence(text)
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return out, err
	}
	return out, nil
}
