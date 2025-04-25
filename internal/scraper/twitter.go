package scraper

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Infoblox-CTO/review-scraper/internal/config"
	"github.com/Infoblox-CTO/review-scraper/pkg/models"
	"github.com/google/uuid"
)

// TwitterScraper implements the Scraper interface for Twitter/X
type TwitterScraper struct {
	config     config.TwitterScraperConfig
	rateLimits config.RateLimitConfig
	proxies    config.ProxyConfig
	client     *http.Client
	userAgents []string
	enabled    bool
}

// NewTwitterScraper creates a new Twitter scraper
func NewTwitterScraper(cfg config.TwitterScraperConfig, rates config.RateLimitConfig,
	proxies config.ProxyConfig) *TwitterScraper {

	// Create client with default timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Common user agents for rotation
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36",
	}

	return &TwitterScraper{
		config:     cfg,
		rateLimits: rates,
		proxies:    proxies,
		client:     client,
		userAgents: userAgents,
		enabled:    cfg.Enabled,
	}
}

// Name returns the name of this scraper
func (s *TwitterScraper) Name() string {
	return "Twitter"
}

// IsEnabled returns whether this scraper is enabled
func (s *TwitterScraper) IsEnabled() bool {
	return s.enabled
}

// Scrape retrieves reviews/comments from Twitter
func (s *TwitterScraper) Scrape(ctx context.Context) ([]models.Review, error) {
	var reviews []models.Review

	// Construct search queries for each keyword
	for _, keyword := range s.config.Keywords {
		// Respect context cancellation
		if ctx.Err() != nil {
			return reviews, ctx.Err()
		}

		// Search tweets containing the keyword
		tweets, err := s.searchTweets(ctx, keyword)
		if err != nil {
			return reviews, fmt.Errorf("error searching tweets for keyword '%s': %w", keyword, err)
		}

		// Filter tweets by exclude words
		tweets = s.filterTweets(tweets)

		// Convert tweets to reviews
		for _, tweet := range tweets {
			review := s.convertTweetToReview(tweet)
			reviews = append(reviews, review)
		}

		// Respect rate limits
		if len(s.config.Keywords) > 1 && s.rateLimits.PauseBetweenRequests {
			select {
			case <-time.After(s.rateLimits.PauseDuration):
				// Continue after pause
			case <-ctx.Done():
				return reviews, ctx.Err()
			}
		}
	}

	return reviews, nil
}

// Tweet represents a Twitter post
type Tweet struct {
	ID        string `json:"id_str"`
	Text      string `json:"full_text"`
	CreatedAt string `json:"created_at"` // Twitter timestamp format: "Wed Oct 10 20:19:24 +0000 2018"
	User      struct {
		ScreenName string `json:"screen_name"`
	} `json:"user"`
	Entities struct {
		Hashtags []struct {
			Text string `json:"text"`
		} `json:"hashtags"`
		URLs []struct {
			ExpandedURL string `json:"expanded_url"`
		} `json:"urls"`
	} `json:"entities"`
	RetweetCount    int    `json:"retweet_count"`
	FavoriteCount   int    `json:"favorite_count"`
	QuotedStatusID  string `json:"quoted_status_id_str,omitempty"`
	InReplyToID     string `json:"in_reply_to_status_id_str,omitempty"`
	InReplyToUserID string `json:"in_reply_to_user_id_str,omitempty"`
}

// TweetSearchResponse represents a Twitter API response
type TweetSearchResponse struct {
	Statuses []Tweet `json:"statuses"`
	Metadata struct {
		NextResultsURL string `json:"next_results,omitempty"`
	} `json:"search_metadata"`
}

// searchTweets searches for tweets containing the given keyword
func (s *TwitterScraper) searchTweets(ctx context.Context, keyword string) ([]Tweet, error) {
	// Construct the Twitter API URL for tweet search
	baseURL := "https://api.twitter.com/1.1/search/tweets.json"

	// URL encode the query parameters
	query := url.Values{}
	query.Add("q", keyword)
	query.Add("count", fmt.Sprintf("%d", s.config.MaxResults))
	query.Add("tweet_mode", "extended") // Get full text instead of truncated
	query.Add("result_type", "recent")  // Get most recent tweets
	query.Add("lang", "en")             // Limit to English tweets

	// Build the URL with query parameters
	apiURL := fmt.Sprintf("%s?%s", baseURL, query.Encode())

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set necessary headers for OAuth 1.0a
	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	nonce := uuid.New().String()

	// Add OAuth headers
	s.addOAuthHeaders(req, "GET", baseURL, query, timestamp, nonce)

	// Set user agent if rotation is enabled
	if s.rateLimits.RandomizeUserAgents && len(s.userAgents) > 0 {
		userAgent := s.userAgents[time.Now().UnixNano()%int64(len(s.userAgents))]
		req.Header.Set("User-Agent", userAgent)
	}

	// Send the request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Twitter API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var searchResp TweetSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return searchResp.Statuses, nil
}

// addOAuthHeaders adds OAuth 1.0a authentication headers to the request
func (s *TwitterScraper) addOAuthHeaders(req *http.Request, method, baseURL string, params url.Values, timestamp, nonce string) {
	// Add OAuth parameters
	oauthParams := url.Values{}
	oauthParams.Add("oauth_consumer_key", s.config.APIKey)
	oauthParams.Add("oauth_nonce", nonce)
	oauthParams.Add("oauth_signature_method", "HMAC-SHA1")
	oauthParams.Add("oauth_timestamp", timestamp)
	oauthParams.Add("oauth_token", s.config.AccessToken)
	oauthParams.Add("oauth_version", "1.0")

	// Create signature base string
	allParams := url.Values{}
	for k, v := range oauthParams {
		allParams[k] = v
	}
	for k, v := range params {
		allParams[k] = v
	}

	// Sort parameters alphabetically
	parameterString := allParams.Encode()

	// Create signature base string
	signatureBaseString := fmt.Sprintf("%s&%s&%s",
		method,
		url.QueryEscape(baseURL),
		url.QueryEscape(parameterString),
	)

	// Create signing key
	signingKey := fmt.Sprintf("%s&%s",
		url.QueryEscape(s.config.APISecret),
		url.QueryEscape(s.config.AccessSecret),
	)

	// Calculate HMAC-SHA1 signature
	h := hmac.New(sha1.New, []byte(signingKey))
	h.Write([]byte(signatureBaseString))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// Add signature to OAuth parameters
	oauthParams.Add("oauth_signature", signature)

	// Construct Authorization header
	var headerParts []string
	for k, v := range oauthParams {
		headerParts = append(headerParts, fmt.Sprintf("%s=\"%s\"", k, url.QueryEscape(v[0])))
	}

	// Set Authorization header
	authHeader := fmt.Sprintf("OAuth %s", strings.Join(headerParts, ", "))
	req.Header.Set("Authorization", authHeader)
}

// filterTweets removes tweets that contain excluded words
func (s *TwitterScraper) filterTweets(tweets []Tweet) []Tweet {
	if len(s.config.ExcludeWords) == 0 {
		return tweets
	}

	filtered := make([]Tweet, 0, len(tweets))

	for _, tweet := range tweets {
		exclude := false
		lowerText := strings.ToLower(tweet.Text)

		for _, word := range s.config.ExcludeWords {
			if strings.Contains(lowerText, strings.ToLower(word)) {
				exclude = true
				break
			}
		}

		if !exclude {
			filtered = append(filtered, tweet)
		}
	}

	return filtered
}

// parseTweetTime parses Twitter's timestamp format
func parseTweetTime(timestamp string) (time.Time, error) {
	// Twitter timestamp format: "Wed Oct 10 20:19:24 +0000 2018"
	layout := "Mon Jan 02 15:04:05 -0700 2006"
	return time.Parse(layout, timestamp)
}

// convertTweetToReview converts a Tweet to a Review
func (s *TwitterScraper) convertTweetToReview(tweet Tweet) models.Review {
	// Parse tweet creation time
	createdAt, err := parseTweetTime(tweet.CreatedAt)
	if err != nil {
		createdAt = time.Now() // Fallback to current time if parsing fails
	}

	// Extract sentiment indicators (e.g., hashtags)
	hashtags := make([]string, 0, len(tweet.Entities.Hashtags))
	for _, tag := range tweet.Entities.Hashtags {
		hashtags = append(hashtags, tag.Text)
	}

	// Construct tweet URL
	tweetURL := fmt.Sprintf("https://twitter.com/%s/status/%s", tweet.User.ScreenName, tweet.ID)

	// Create metadata
	metadata := map[string]interface{}{
		"retweet_count":  tweet.RetweetCount,
		"favorite_count": tweet.FavoriteCount,
		"hashtags":       hashtags,
	}

	if tweet.QuotedStatusID != "" {
		metadata["quoted_status_id"] = tweet.QuotedStatusID
	}

	if tweet.InReplyToID != "" {
		metadata["in_reply_to_id"] = tweet.InReplyToID
		metadata["in_reply_to_user_id"] = tweet.InReplyToUserID
	}

	// Generate a rating based on engagement (crude approximation)
	// In a real implementation, this would use a more sophisticated approach
	engagementScore := float64(tweet.FavoriteCount+tweet.RetweetCount) / 10.0
	if engagementScore > 5.0 {
		engagementScore = 5.0
	}
	rating := 5.0 - engagementScore // Invert so high engagement = potentially negative review

	// Clean text - remove URLs
	text := tweet.Text
	urlPattern := regexp.MustCompile(`https?://\S+`)
	text = urlPattern.ReplaceAllString(text, "")

	return models.Review{
		ID:          fmt.Sprintf("twitter-%s", tweet.ID),
		Source:      "twitter",
		SourceID:    tweet.ID,
		Content:     strings.TrimSpace(text),
		Author:      tweet.User.ScreenName,
		Rating:      &rating,
		URL:         tweetURL,
		CreatedAt:   createdAt,
		RetrievedAt: time.Now(),
		Metadata:    metadata,
	}
}
