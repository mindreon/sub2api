package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/media"
)

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

func doJSON(ctx context.Context, client httpDoer, method, url, apiKey string, body any, out any) error {
	if client == nil {
		client = http.DefaultClient
	}
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", media.ErrUpstreamRequest, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("%w: read response: %v", media.ErrUpstreamRequest, err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: upstream http %d: %s", media.ErrUpstreamRequest, resp.StatusCode, truncateBody(raw))
	}
	if out == nil {
		return nil
	}
	if len(raw) == 0 {
		return nil
	}
	if err := json.Unmarshal(raw, out); err != nil {
		return fmt.Errorf("%w: decode response: %v", media.ErrUpstreamRequest, err)
	}
	return nil
}

func truncateBody(raw []byte) string {
	const limit = 512
	s := string(raw)
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "..."
}
