package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/Infoblox-CTO/review-scraper/internal/config"
	"github.com/Infoblox-CTO/review-scraper/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupMockTwitterServer creates a test server that returns predefined responses
func setupMockTwitterServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request includes Infoblox-related keywords
		query := r.URL.Query().Get("q")
		var mockResponse string

		// We'll respond with different JSON based on the keyword
		if query == "infoblox" || query == "bloxone" || query == "nios" || query == "ddi" {
			mockResponse = `{
				"statuses": [
					{
						"id_str": "1234567890",
						"full_text": "I'm really impressed with Infoblox's DDI solution. Great management interface and reliable performance! #networkautomation",
						"created_at": "Wed Apr 24 10:30:00 +0000 2025",
						"user": {
							"screen_name": "NetworkAdmin123"
						},
						"entities": {
							"hashtags": [
								{ "text": "networkautomation" }
							],
							"urls": []
						},
						"retweet_count": 5,
						"favorite_count": 10
					},
					{
						"id_str": "1234567891",
						"full_text": "Having issues with the BloxOne Cloud dashboard today. Anyone else experiencing slowness? @Infoblox #ddi #dns",
						"created_at": "Wed Apr 23 14:20:00 +0000 2025",
						"user": {
							"screen_name": "CloudEngineer42"
						},
						"entities": {
							"hashtags": [
								{ "text": "ddi" },
								{ "text": "dns" }
							],
							"urls": []
						},
						"retweet_count": 2,
						"favorite_count": 1
					},
					{
						"id_str": "1234567892",
						"full_text": "Wish NIOS had better integration with our SIEM tools. Great product otherwise for DNS security. #infoblox #security",
						"created_at": "Wed Apr 22 09:15:00 +0000 2025",
						"user": {
							"screen_name": "SecurityPro99"
						},
						"entities": {
							"hashtags": [
								{ "text": "infoblox" },
								{ "text": "security" }
							],
							"urls": []
						},
						"retweet_count": 8,
						"favorite_count": 15
					}
				],
				"search_metadata": {}
			}`
		} else if query == "dns" || query == "dhcp" || query == "ipam" {
			mockResponse = `{
				"statuses": [
					{
						"id_str": "2234567890",
						"full_text": "Infoblox DNS services are rock solid. Haven't had any outages in months. #happycustomer",
						"created_at": "Wed Apr 24 11:30:00 +0000 2025",
						"user": {
							"screen_name": "ITManager76"
						},
						"entities": {
							"hashtags": [
								{ "text": "happycustomer" }
							],
							"urls": []
						},
						"retweet_count": 3,
						"favorite_count": 7
					},
					{
						"id_str": "2234567891",
						"full_text": "DHCP failover configuration on Infoblox is so much easier than our previous solution. #networking",
						"created_at": "Wed Apr 23 15:45:00 +0000 2025",
						"user": {
							"screen_name": "NetworkNinja"
						},
						"entities": {
							"hashtags": [
								{ "text": "networking" }
							],
							"urls": []
						},
						"retweet_count": 4,
						"favorite_count": 9
					}
				],
				"search_metadata": {}
			}`
		} else {
			mockResponse = `{
				"statuses": [],
				"search_metadata": {}
			}`
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	}))
}

// Mock Twitter API client for testing
type MockTwitterClient struct {
	mock.Mock
}

func (m *MockTwitterClient) SearchTweets(query string, maxResults int) ([]models.Tweet, error) {
	args := m.Called(query, maxResults)
	return args.Get(0).([]models.Tweet), args.Error(1)
}

// Helper function to check if a string contains any of the provided substrings
func contains(s string, substrings ...string) bool {
	s = strings.ToLower(s)
	for _, sub := range substrings {
		if strings.Contains(s, strings.ToLower(sub)) {
			return true
		}
	}
	return false
}

func TestNewTwitterScraper(t *testing.T) {
	// Create the required config objects
	twitterCfg := config.TwitterScraperConfig{
		Keywords:     []string{"infoblox", "bloxone"},
		ExcludeWords: []string{"spam"},
		MaxResults:   100,
		Enabled:      true,
	}
	rateLimitCfg := config.RateLimitConfig{
		PauseBetweenRequests: true,
		PauseDuration:        time.Second,
		RandomizeUserAgents:  true,
	}
	proxyCfg := config.ProxyConfig{
		Enabled: false,
	}

	scraper := NewTwitterScraper(twitterCfg, rateLimitCfg, proxyCfg)

	assert.NotNil(t, scraper)
	assert.Equal(t, twitterCfg, scraper.config)
	assert.Equal(t, rateLimitCfg, scraper.rateLimits)
	assert.Equal(t, proxyCfg, scraper.proxies)
}

func TestFetchReviews(t *testing.T) {
	// Setup mock server
	server := setupMockTwitterServer()
	defer server.Close()

	// Create configs needed for the scraper
	twitterCfg := config.TwitterScraperConfig{
		Keywords:     []string{"infoblox", "bloxone", "ddi"},
		ExcludeWords: []string{"spam", "ad"},
		MaxResults:   100,
		Enabled:      true,
		APIKey:       "test_key",
		APISecret:    "test_secret",
		AccessToken:  "test_token",
		AccessSecret: "test_token_secret",
	}

	rateLimitCfg := config.RateLimitConfig{
		PauseBetweenRequests: true,
		PauseDuration:        time.Millisecond * 100, // Short pause for tests
		RandomizeUserAgents:  true,
	}

	proxyCfg := config.ProxyConfig{
		Enabled: false,
	}

	// Create the scraper
	scraper := NewTwitterScraper(twitterCfg, rateLimitCfg, proxyCfg)

	// Override the client to use our test server
	scraper.client = &http.Client{
		Transport: &http.Transport{
			Proxy: func(req *http.Request) (*url.URL, error) {
				// Redirect all requests to our mock server
				return url.Parse(server.URL)
			},
		},
	}

	// Execute
	ctx := context.Background()
	reviews, err := scraper.Scrape(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, reviews)

	// Check that all reviews have proper fields
	for _, review := range reviews {
		assert.NotEmpty(t, review.ID)
		assert.NotEmpty(t, review.Content)
		assert.Equal(t, "twitter", review.Source)
		assert.NotEmpty(t, review.SourceID)
		assert.NotEmpty(t, review.Author)
		assert.NotZero(t, review.CreatedAt)

		// Check that reviews contain relevant Infoblox-specific keywords
		assert.True(t,
			contains(review.Content, "infoblox", "bloxone", "nios", "ddi", "dns", "dhcp"),
			"Review does not contain any Infoblox-specific terms: %s", review.Content)
	}
}

// Helper function to filter out reviews containing excluded words
func filterOutExcludedContent(reviews []models.Review, excludeWords []string) []models.Review {
	var filtered []models.Review

	for _, review := range reviews {
		if !contains(review.Content, excludeWords...) {
			filtered = append(filtered, review)
		}
	}

	return filtered
}

func TestParseReviewFromTweet(t *testing.T) {
	// Setup with proper config parameters
	twitterCfg := config.TwitterScraperConfig{
		Enabled: true,
	}
	rateLimitCfg := config.RateLimitConfig{}
	proxyCfg := config.ProxyConfig{}

	scraper := NewTwitterScraper(twitterCfg, rateLimitCfg, proxyCfg)

	// Test data - create a Tweet structure as defined in the twitter.go file
	tweet := Tweet{
		ID:        "1234567890",
		Text:      "Really impressed with @Infoblox BloxOne Threat Defense. Detected DNS exfiltration attempts instantly.",
		CreatedAt: "Wed Apr 24 10:30:00 +0000 2025",
		User: struct {
			ScreenName string `json:"screen_name"`
		}{
			ScreenName: "securityexpert",
		},
		RetweetCount:  5,
		FavoriteCount: 10,
	}

	// Execute - using the actual method from TwitterScraper
	review := scraper.convertTweetToReview(tweet)

	// Assert
	assert.Equal(t, "twitter-"+tweet.ID, review.ID)
	assert.Equal(t, tweet.ID, review.SourceID)
	assert.Equal(t, "twitter", review.Source)
	assert.Equal(t, tweet.User.ScreenName, review.Author)
	assert.True(t, strings.Contains(review.Content, "BloxOne Threat Defense"))
	assert.NotNil(t, review.Rating)
}
