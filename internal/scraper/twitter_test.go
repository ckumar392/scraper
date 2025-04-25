package scraper

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	cfg := config.TwitterConfig{
		Keywords:     []string{"infoblox", "bloxone"},
		ExcludeWords: []string{"spam"},
		MaxResults:   100,
	}

	scraper := NewTwitterScraper(cfg)

	assert.NotNil(t, scraper)
	assert.Equal(t, cfg, scraper.config)
}

func TestFetchReviews(t *testing.T) {
	// Setup
	cfg := config.TwitterConfig{
		Keywords:     []string{"infoblox", "bloxone", "ddi"},
		ExcludeWords: []string{"spam", "ad"},
		MaxResults:   100,
	}

	mockClient := new(MockTwitterClient)
	scraper := NewTwitterScraper(cfg)
	scraper.client = mockClient

	// Test data - Infoblox specific tweets
	tweets := []models.Tweet{
		{
			ID:        "1234567890",
			Text:      "Just implemented Infoblox BloxOne in our network. Great product for DNS management!",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UserName:  "networkadmin",
			UserID:    "user123",
		},
		{
			ID:        "0987654321",
			Text:      "Having issues with NIOS DHCP scopes. @Infoblox support please help!",
			CreatedAt: time.Now().Add(-48 * time.Hour),
			UserName:  "networkengineer",
			UserID:    "user456",
		},
		{
			ID:        "5678901234",
			Text:      "The Infoblox Grid technology is revolutionary for enterprise DNS.",
			CreatedAt: time.Now().Add(-72 * time.Hour),
			UserName:  "techCTO",
			UserID:    "user789",
		},
		{
			ID:        "1122334455",
			Text:      "Spam post not related to Infoblox - should be filtered out",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UserName:  "spammer",
			UserID:    "spam123",
		},
		{
			ID:        "5566778899",
			Text:      "Free ad for something - also mentions infoblox - should be filtered",
			CreatedAt: time.Now().Add(-24 * time.Hour),
			UserName:  "advertiser",
			UserID:    "ad456",
		},
	}

	// Setup mock expectations
	for _, keyword := range cfg.Keywords {
		query := "\"" + keyword + "\" -is:retweet"
		mockClient.On("SearchTweets", query, cfg.MaxResults).Return(tweets, nil)
	}

	// Execute
	ctx := context.Background()
	reviews, err := scraper.FetchReviews(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, reviews)

	// Should have 3 valid reviews (excluding the spam and ad ones)
	validReviews := filterOutExcludedContent(reviews, cfg.ExcludeWords)
	assert.Equal(t, 3, len(validReviews))

	// Check that all reviews have proper fields
	for _, review := range validReviews {
		assert.NotEmpty(t, review.ID)
		assert.NotEmpty(t, review.Content)
		assert.NotEmpty(t, review.Source)
		assert.Equal(t, "Twitter", review.Platform)
		assert.NotEmpty(t, review.AuthorID)
		assert.NotEmpty(t, review.AuthorName)
		assert.NotZero(t, review.CreatedAt)

		// Check that reviews contain relevant Infoblox-specific keywords
		assert.True(t,
			contains(review.Content, "infoblox", "bloxone", "nios", "ddi", "dns", "dhcp", "grid"),
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
	// Setup
	cfg := config.TwitterConfig{}
	scraper := NewTwitterScraper(cfg)

	// Test data - Infoblox specific tweet
	tweet := models.Tweet{
		ID:        "1234567890",
		Text:      "Really impressed with @Infoblox BloxOne Threat Defense. Detected DNS exfiltration attempts instantly.",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UserName:  "securityexpert",
		UserID:    "user123",
	}

	// Execute
	review := scraper.parseReviewFromTweet(tweet)

	// Assert
	assert.Equal(t, tweet.ID, review.ExternalID)
	assert.Equal(t, tweet.Text, review.Content)
	assert.Equal(t, "Twitter", review.Platform)
	assert.Equal(t, tweet.UserName, review.AuthorName)
	assert.Equal(t, tweet.UserID, review.AuthorID)

	// Test for truncated text preservation
	longTweet := models.Tweet{
		ID:        "0987654321",
		Text:      strings.Repeat("Very long feedback about Infoblox NIOS configuration. ", 10),
		CreatedAt: time.Now(),
		UserName:  "networkengineer",
		UserID:    "user456",
	}

	longReview := scraper.parseReviewFromTweet(longTweet)
	assert.Equal(t, longTweet.Text, longReview.Content)
}
