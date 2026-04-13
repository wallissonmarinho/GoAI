package app

import (
	"context"
	"net/http"
	"strings"
)

type httpURLExistenceChecker struct {
	hc *http.Client
	ua string
}

func newHTTPURLExistenceChecker(hc *http.Client, userAgent string) *httpURLExistenceChecker {
	return &httpURLExistenceChecker{
		hc: hc,
		ua: strings.TrimSpace(userAgent),
	}
}

func (c *httpURLExistenceChecker) Exists(ctx context.Context, rawURL string) (bool, error) {
	ok, code, err := c.do(ctx, http.MethodHead, rawURL)
	if err != nil {
		return false, err
	}
	if code == http.StatusMethodNotAllowed || code == http.StatusNotImplemented {
		ok, code, err = c.do(ctx, http.MethodGet, rawURL)
		if err != nil {
			return false, err
		}
	}
	if code == http.StatusNotFound {
		return false, nil
	}
	return ok, nil
}

func (c *httpURLExistenceChecker) do(ctx context.Context, method, rawURL string) (bool, int, error) {
	req, err := http.NewRequestWithContext(ctx, method, rawURL, nil)
	if err != nil {
		return false, 0, err
	}
	if c.ua != "" {
		req.Header.Set("User-Agent", c.ua)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400, resp.StatusCode, nil
}
