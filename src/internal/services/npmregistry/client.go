package npmregistry

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	registryURL = "https://registry.npmjs.org"
	timeout     = 3 * time.Second
	userAgent   = "claudex-update-check"
)

type Client struct {
	httpClient *http.Client
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (c *Client) GetLatestVersion(packageName string) (string, error) {
	url := fmt.Sprintf("%s/%s", registryURL, packageName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("npm registry returned status %d", resp.StatusCode)
	}

	var info PackageInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}

	return info.DistTags.Latest, nil
}
