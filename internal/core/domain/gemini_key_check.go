package domain

// GeminiKeyCheckResult is one entry from GET /v1/diag/gemini-keys (no raw API key is ever returned).
type GeminiKeyCheckResult struct {
	Index      int    `json:"index"`
	OK         bool   `json:"ok"`
	Status     string `json:"status"`
	HTTPStatus int    `json:"http_status,omitempty"`
	Detail     string `json:"detail,omitempty"`
}
