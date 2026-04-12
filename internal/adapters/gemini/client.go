package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wallissonmarinho/GoAI/internal/core/domain"
	"github.com/wallissonmarinho/GoAI/internal/core/ports"
)

const defaultBaseURL = "https://generativelanguage.googleapis.com"

// Client calls Gemini generateContent over HTTPS (implements ports.TextCompletion).
type Client struct {
	HTTP      *http.Client
	BaseURL   string
	Model     string
	KeyPool   *KeyPool
	Timeout   time.Duration
	MaxBody   int64
	UserAgent string
}

// GenerateText runs generateContent with JSON MIME type and returns the first text part.
func (c *Client) GenerateText(ctx context.Context, userPrompt string) (string, error) {
	if c == nil || c.KeyPool == nil {
		return "", errors.New("gemini: nil client or pool")
	}
	model := strings.TrimSpace(c.Model)
	if model == "" {
		model = "gemini-2.0-flash"
	}
	base := strings.TrimSuffix(strings.TrimSpace(c.BaseURL), "/")
	if base == "" {
		base = defaultBaseURL
	}
	hc := c.HTTP
	if hc == nil {
		hc = http.DefaultClient
	}
	timeout := c.Timeout
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	maxBody := c.MaxBody
	if maxBody <= 0 {
		maxBody = 4 << 20
	}
	ua := strings.TrimSpace(c.UserAgent)
	if ua == "" {
		ua = "GoAI/1.0"
	}

	body := generateContentRequest{
		Contents: []contentBlock{
			{Role: "user", Parts: []part{{Text: userPrompt}}},
		},
		GenerationConfig: genConfig{
			Temperature:      0.2,
			ResponseMIMEType: "application/json",
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	apiPath := fmt.Sprintf("%s/v1beta/models/%s:generateContent", base, url.PathEscape(model))
	var outText string
	now := time.Now().UTC()
	err = c.KeyPool.WithKey(now, func(apiKey string) error {
		u := apiPath + "?key=" + url.QueryEscape(apiKey)
		reqCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, u, bytes.NewReader(raw))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")
		req.Header.Set("User-Agent", ua)

		resp, err := hc.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		b, err := io.ReadAll(io.LimitReader(resp.Body, maxBody+1))
		if err != nil {
			return err
		}
		if int64(len(b)) > maxBody {
			return errors.New("gemini: response too large")
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			return domain.ErrLLMQuotaOrRate
		}
		if resp.StatusCode == http.StatusServiceUnavailable {
			return domain.ErrLLMQuotaOrRate
		}

		var parsed generateContentResponse
		_ = json.Unmarshal(b, &parsed)
		if parsed.Error != nil {
			msg := strings.ToLower(parsed.Error.Message + " " + parsed.Error.Status)
			if strings.Contains(msg, "resource exhausted") || strings.Contains(msg, "quota") || strings.Contains(msg, "rate") {
				return domain.ErrLLMQuotaOrRate
			}
			return fmt.Errorf("gemini api error: %s", parsed.Error.Message)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if resp.StatusCode == 429 {
				return domain.ErrLLMQuotaOrRate
			}
			return fmt.Errorf("gemini: http %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
		}

		if len(parsed.Candidates) == 0 {
			return errors.New("gemini: empty candidates")
		}
		parts := parsed.Candidates[0].Content.Parts
		if len(parts) == 0 {
			return errors.New("gemini: empty parts")
		}
		outText = parts[0].Text
		if strings.TrimSpace(outText) == "" {
			return errors.New("gemini: empty text")
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return outText, nil
}

var _ ports.TextCompletion = (*Client)(nil)
