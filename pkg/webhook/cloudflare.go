package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// CloudflareKVClient pushes activity data to Cloudflare Workers KV
type CloudflareKVClient struct {
	httpClient   *http.Client
	accountID    string
	namespaceID  string
	apiToken     string
}

// NewCloudflareKVClient creates a new Cloudflare KV client
func NewCloudflareKVClient(accountID, namespaceID, apiToken string) *CloudflareKVClient {
	return &CloudflareKVClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		accountID:   accountID,
		namespaceID: namespaceID,
		apiToken:    apiToken,
	}
}

// ActivityPayload is the data structure sent to Cloudflare
type ActivityPayload struct {
	MonitorName    string         `json:"monitorName"`
	AnimeID        int            `json:"animeId,omitempty"`
	AnimeName      string         `json:"animeName,omitempty"`
	ActivityLevel  string         `json:"activityLevel"`
	WeebcastStatus string         `json:"weebcastStatus"`
	Metrics        MetricsPayload `json:"metrics"`
	TrendingAnime  []TrendingItem `json:"trendingAnime,omitempty"`
	LastUpdated    time.Time      `json:"lastUpdated"`
}

// MetricsPayload contains activity metrics
type MetricsPayload struct {
	ActiveUsers   int     `json:"activeUsers"`
	WatchingCount int     `json:"watchingCount"`
	Members       int     `json:"members"`
	Score         float64 `json:"score"`
	Rank          int     `json:"rank"`
	Favorites     int     `json:"favorites"`
}

// TrendingItem represents a trending anime
type TrendingItem struct {
	ID            int     `json:"id"`
	Title         string  `json:"title"`
	Score         float64 `json:"score"`
	Members       int     `json:"members"`
	ActivityLevel string  `json:"activityLevel"`
	ImageURL      string  `json:"imageUrl"`
}

// PushActivity sends activity data to Cloudflare Workers KV
func (c *CloudflareKVClient) PushActivity(ctx context.Context, key string, payload *ActivityPayload) error {
	if c.apiToken == "" {
		return nil // Skip if not configured
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	url := fmt.Sprintf(
		"https://api.cloudflare.com/client/v4/accounts/%s/storage/kv/namespaces/%s/values/%s",
		c.accountID, c.namespaceID, key,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

// WebhookClient sends activity updates to a webhook URL
type WebhookClient struct {
	httpClient *http.Client
}

// NewWebhookClient creates a new webhook client
func NewWebhookClient() *WebhookClient {
	return &WebhookClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// SendNotification sends an activity notification to a webhook
func (c *WebhookClient) SendNotification(ctx context.Context, webhookURL string, payload *ActivityPayload) error {
	if webhookURL == "" {
		return nil
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status: %d", resp.StatusCode)
	}

	return nil
}


