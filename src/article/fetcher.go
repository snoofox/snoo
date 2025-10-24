package article

import (
	"context"
	"fmt"
	"net/http"
	"snoo/src/debug"
	"strings"
	"time"

	"github.com/go-shiori/go-readability"
)

func Fetch(ctx context.Context, url string) (string, error) {
	if url == "" || (!strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://")) {
		return "", fmt.Errorf("invalid URL")
	}

	debug.Log("Article: Fetching content from %s", url)

	fetchCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(fetchCtx, "GET", url, nil)
	if err != nil {
		debug.Log("Article: Error creating request: %v", err)
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; snoo/1.0)")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		debug.Log("Article: Error fetching URL: %v", err)
		return "", fmt.Errorf("error fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		debug.Log("Article: HTTP status %d", resp.StatusCode)
		return "", fmt.Errorf("HTTP status %d", resp.StatusCode)
	}

	article, err := readability.FromReader(resp.Body, resp.Request.URL)
	if err != nil {
		debug.Log("Article: Error parsing article: %v", err)
		return "", fmt.Errorf("error parsing article: %w", err)
	}

	debug.Log("Article: Successfully extracted content (length: %d)", len(article.TextContent))
	return article.TextContent, nil
}
