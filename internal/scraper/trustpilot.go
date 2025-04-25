package scraper

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Infoblox-CTO/review-scraper/internal/config"
	"github.com/Infoblox-CTO/review-scraper/pkg/models"
	"github.com/PuerkitoBio/goquery"
)

// TrustpilotScraper implements the Scraper interface for Trustpilot
type TrustpilotScraper struct {
	config     config.TrustpilotScraperConfig
	rateLimits config.RateLimitConfig
	proxies    config.ProxyConfig
	client     *http.Client
	userAgents []string
	enabled    bool
}

// NewTrustpilotScraper creates a new Trustpilot scraper
func NewTrustpilotScraper(cfg config.TrustpilotScraperConfig, rates config.RateLimitConfig,
	proxies config.ProxyConfig) *TrustpilotScraper {

	// Create client with default timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Setup proxy if enabled
	if proxies.Enabled && proxies.URL != "" {
		proxyURL := proxies.URL
		if proxies.Username != "" && proxies.Password != "" {
			// Add auth credentials to proxy URL if provided
			proxyURL = fmt.Sprintf("http://%s:%s@%s",
				proxies.Username,
				proxies.Password,
				strings.TrimPrefix(proxies.URL, "http://"))
		}

		transport := &http.Transport{
			Proxy: http.ProxyURL(MustParseURL(proxyURL)),
		}
		client.Transport = transport
	}

	// Common user agents for rotation
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/14.1.1 Safari/605.1.15",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:89.0) Gecko/20100101 Firefox/89.0",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/92.0.4515.107 Safari/537.36",
	}

	return &TrustpilotScraper{
		config:     cfg,
		rateLimits: rates,
		proxies:    proxies,
		client:     client,
		userAgents: userAgents,
		enabled:    cfg.Enabled,
	}
}

// Name returns the name of this scraper
func (s *TrustpilotScraper) Name() string {
	return "Trustpilot"
}

// IsEnabled returns whether this scraper is enabled
func (s *TrustpilotScraper) IsEnabled() bool {
	return s.enabled
}

// MustParseURL is a helper that panics if URL parsing fails
func MustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}

// Scrape retrieves reviews from Trustpilot
func (s *TrustpilotScraper) Scrape(ctx context.Context) ([]models.Review, error) {
	var allReviews []models.Review

	// Set the maximum number of pages to scrape
	maxPages := s.config.MaxPages
	if maxPages <= 0 {
		maxPages = 5 // Default to 5 pages
	}

	// Scrape each page of reviews
	for page := 1; page <= maxPages; page++ {
		// Respect context cancellation
		if ctx.Err() != nil {
			return allReviews, ctx.Err()
		}

		// Scrape a page of reviews
		reviews, hasMorePages, err := s.scrapePage(ctx, page)
		if err != nil {
			return allReviews, fmt.Errorf("error scraping Trustpilot page %d: %w", page, err)
		}

		// Add reviews to the collection
		allReviews = append(allReviews, reviews...)

		// Stop if there are no more pages
		if !hasMorePages {
			break
		}

		// Respect rate limits
		if page < maxPages && s.rateLimits.PauseBetweenRequests {
			select {
			case <-time.After(time.Millisecond * time.Duration(s.rateLimits.PauseDuration)):
				// Continue after pause
			case <-ctx.Done():
				return allReviews, ctx.Err()
			}
		}
	}

	return allReviews, nil
}

// scrapePage retrieves reviews from a single page on Trustpilot
func (s *TrustpilotScraper) scrapePage(ctx context.Context, page int) ([]models.Review, bool, error) {
	// Construct URL for the page of reviews
	url := fmt.Sprintf("https://www.trustpilot.com/review/%s?page=%d", s.config.BusinessID, page)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("error creating request: %w", err)
	}

	// Set user agent rotation if enabled
	if len(s.userAgents) > 0 {
		userAgent := s.userAgents[time.Now().UnixNano()%int64(len(s.userAgents))]
		req.Header.Set("User-Agent", userAgent)
	}

	// Set common headers to look like a browser
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("Trustpilot returned non-OK status: %d", resp.StatusCode)
	}

	// Parse HTML response
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, false, fmt.Errorf("error parsing HTML: %w", err)
	}

	var reviews []models.Review
	retrievedTime := time.Now()

	// Store the business ID in a local variable for use in the callback
	businessID := s.config.BusinessID

	// Extract reviews from the page
	// Note: Selectors may need to be updated if Trustpilot changes their HTML structure
	doc.Find("article.review").Each(func(i int, s *goquery.Selection) {
		// Extract review ID
		reviewID, _ := s.Attr("id")
		if reviewID == "" {
			reviewID = fmt.Sprintf("trustpilot-%s-%d-%d", businessID, page, i)
		}

		// Extract rating (1-5 stars)
		ratingStr := ""
		s.Find("div.star-rating").Each(func(j int, stars *goquery.Selection) {
			imgAlt, exists := stars.Find("img").Attr("alt")
			if exists {
				ratingRe := regexp.MustCompile(`(\d+)`)
				matches := ratingRe.FindStringSubmatch(imgAlt)
				if len(matches) > 1 {
					ratingStr = matches[1]
				}
			}
		})

		var rating *float64
		if ratingStr != "" {
			ratingVal, err := strconv.ParseFloat(ratingStr, 64)
			if err == nil {
				rating = &ratingVal
			}
		}

		// Extract title
		title := strings.TrimSpace(s.Find("h2.review-content__title").Text())

		// Extract content
		content := strings.TrimSpace(s.Find("p.review-content__text").Text())

		// Extract author
		author := strings.TrimSpace(s.Find("div.consumer-information__name").Text())

		// Extract date
		dateStr := ""
		s.Find("div.review-content-header__dates").Each(func(j int, dates *goquery.Selection) {
			dateStr = strings.TrimSpace(dates.Text())
		})

		// Parse date (format can vary)
		// This is a simplistic approach; in real code, you'd want more robust date parsing
		var createdAt time.Time
		if dateStr != "" {
			// Try to parse the date in various formats
			formats := []string{
				"Jan 2, 2006",
				"January 2, 2006",
				"2 Jan 2006",
				"2006-01-02",
			}

			for _, format := range formats {
				parsed, err := time.Parse(format, dateStr)
				if err == nil {
					createdAt = parsed
					break
				}
			}
		}

		// If date parsing failed, use a fallback date
		if createdAt.IsZero() {
			createdAt = time.Now().AddDate(0, 0, -((page-1)*10 + i)) // Each review is a day older
		}

		// Extract URL
		url := fmt.Sprintf("https://www.trustpilot.com/reviews/%s#%s", businessID, reviewID)

		// Create review
		review := models.Review{
			ID:          reviewID,
			Source:      "trustpilot",
			SourceID:    reviewID,
			Content:     content,
			Title:       title,
			Author:      author,
			Rating:      rating,
			URL:         url,
			CreatedAt:   createdAt,
			RetrievedAt: retrievedTime,
			Metadata: map[string]interface{}{
				"platform":     "Trustpilot",
				"business_id":  businessID,
				"review_page":  page,
				"review_index": i,
			},
		}

		// Check for vendor response (Trustpilot allows companies to reply to reviews)
		vendorResponse := strings.TrimSpace(s.Find("div.brand-reply").Text())
		if vendorResponse != "" {
			review.Metadata["has_vendor_response"] = true
			review.Metadata["vendor_response"] = vendorResponse
		} else {
			review.Metadata["has_vendor_response"] = false
		}

		reviews = append(reviews, review)
	})

	// Check if there are more pages
	hasMorePages := false
	doc.Find("a.pagination-link--next").Each(func(i int, s *goquery.Selection) {
		hasMorePages = true
	})

	return reviews, hasMorePages, nil
}
