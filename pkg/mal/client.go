package mal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client interfaces with the MyAnimeList API (via Jikan)
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new MAL API client using Jikan (unofficial MAL API)
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.jikan.moe/v4",
	}
}

// AnimeData represents anime information from MAL
type AnimeData struct {
	MalID    int    `json:"mal_id"`
	URL      string `json:"url"`
	Title    string `json:"title"`
	TitleEnglish string `json:"title_english"`
	Images   struct {
		JPG struct {
			ImageURL      string `json:"image_url"`
			SmallImageURL string `json:"small_image_url"`
			LargeImageURL string `json:"large_image_url"`
		} `json:"jpg"`
	} `json:"images"`
	Score      float64 `json:"score"`
	ScoredBy   int     `json:"scored_by"`
	Rank       int     `json:"rank"`
	Popularity int     `json:"popularity"`
	Members    int     `json:"members"`
	Favorites  int     `json:"favorites"`
	Status     string  `json:"status"`
	Airing     bool    `json:"airing"`
	Statistics *AnimeStatistics `json:"statistics,omitempty"`
}

// AnimeStatistics contains detailed viewing statistics
type AnimeStatistics struct {
	Watching    int `json:"watching"`
	Completed   int `json:"completed"`
	OnHold      int `json:"on_hold"`
	Dropped     int `json:"dropped"`
	PlanToWatch int `json:"plan_to_watch"`
	Total       int `json:"total"`
}

// AnimeResponse wraps the API response for a single anime
type AnimeResponse struct {
	Data AnimeData `json:"data"`
}

// AnimeListResponse wraps the API response for anime lists
type AnimeListResponse struct {
	Data       []AnimeData `json:"data"`
	Pagination struct {
		LastVisiblePage int  `json:"last_visible_page"`
		HasNextPage     bool `json:"has_next_page"`
	} `json:"pagination"`
}

// StatisticsResponse wraps the statistics API response
type StatisticsResponse struct {
	Data AnimeStatistics `json:"data"`
}

// GetAnime fetches details for a specific anime by MAL ID
func (c *Client) GetAnime(ctx context.Context, malID int) (*AnimeData, error) {
	url := fmt.Sprintf("%s/anime/%d/full", c.baseURL, malID)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited by MAL API, retry later")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result AnimeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result.Data, nil
}

// GetAnimeStatistics fetches statistics for a specific anime
func (c *Client) GetAnimeStatistics(ctx context.Context, malID int) (*AnimeStatistics, error) {
	url := fmt.Sprintf("%s/anime/%d/statistics", c.baseURL, malID)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited by MAL API, retry later")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result StatisticsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result.Data, nil
}

// GetTopAiring fetches the top currently airing anime
func (c *Client) GetTopAiring(ctx context.Context, limit int) ([]AnimeData, error) {
	url := fmt.Sprintf("%s/top/anime?filter=airing&limit=%d", c.baseURL, limit)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited by MAL API, retry later")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result AnimeListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Data, nil
}

// GetSeasonNow fetches anime from the current season
func (c *Client) GetSeasonNow(ctx context.Context, limit int) ([]AnimeData, error) {
	url := fmt.Sprintf("%s/seasons/now?limit=%d", c.baseURL, limit)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited by MAL API, retry later")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result AnimeListResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return result.Data, nil
}

// GetRecentRecommendations fetches recent anime recommendations (indicates user activity)
func (c *Client) GetRecentRecommendations(ctx context.Context) (int, error) {
	url := fmt.Sprintf("%s/recommendations/anime", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return 0, fmt.Errorf("rate limited by MAL API, retry later")
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		Pagination struct {
			Items struct {
				Count int `json:"count"`
				Total int `json:"total"`
			} `json:"items"`
		} `json:"pagination"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("decoding response: %w", err)
	}

	return result.Pagination.Items.Total, nil
}

// ActivityMetrics represents aggregated activity metrics
type ActivityMetrics struct {
	TotalActiveUsers    int
	TotalWatching       int
	TotalMembers        int
	AverageScore        float64
	TopAiringCount      int
	RecentActivityCount int
}

// GetOverallActivity calculates overall MAL activity metrics
func (c *Client) GetOverallActivity(ctx context.Context) (*ActivityMetrics, error) {
	metrics := &ActivityMetrics{}

	// Get top airing anime to calculate total engagement
	topAiring, err := c.GetTopAiring(ctx, 25)
	if err != nil {
		return nil, fmt.Errorf("fetching top airing: %w", err)
	}

	var totalScore float64
	var scoreCount int

	for _, anime := range topAiring {
		metrics.TotalMembers += anime.Members
		metrics.TotalActiveUsers += anime.Favorites // Favorites as proxy for active users
		if anime.Score > 0 {
			totalScore += anime.Score
			scoreCount++
		}
	}

	if scoreCount > 0 {
		metrics.AverageScore = totalScore / float64(scoreCount)
	}

	metrics.TopAiringCount = len(topAiring)

	// Estimate active users based on member engagement
	// This is a rough estimation - typically 1-5% of members are active
	metrics.TotalWatching = metrics.TotalMembers / 20

	return metrics, nil
}

