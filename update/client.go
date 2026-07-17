package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var (
	releasesURL = "https://api.github.com/repos/hxsggsz/kanba/releases/latest"
	httpClient  = &http.Client{Timeout: 3 * time.Second}
)

type releaseResponse struct {
	TagName string `json:"tag_name"`
}

func fetchLatestTag(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releasesURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "kanba-update-checker")
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("github api returned status %d", resp.StatusCode)
	}

	var rel releaseResponse
	if err := json.NewDecoder(resp.Body).Decode(&rel); err != nil {
		return "", err
	}
	if rel.TagName == "" {
		return "", fmt.Errorf("release response missing tag_name")
	}

	return rel.TagName, nil
}
