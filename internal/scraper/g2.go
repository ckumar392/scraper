package scraper

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Infoblox-CTO/review-scraper/internal/config"
	"github.com/Infoblox-CTO/review-scraper/pkg/models"
)

// G2Scraper implements the Scraper interface for G2
type G2Scraper struct {
	config     config.G2ScraperConfig
	rateLimits config.RateLimitConfig
	proxies    config.ProxyConfig
	client     *http.Client
	userAgents []string
	enabled    bool
}

// NewG2Scraper creates a new G2 scraper
func NewG2Scraper(cfg config.G2ScraperConfig, rates config.RateLimitConfig,
	proxies config.ProxyConfig) *G2Scraper {

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

	return &G2Scraper{
		config:     cfg,
		rateLimits: rates,
		proxies:    proxies,
		client:     client,
		userAgents: userAgents,
		enabled:    cfg.Enabled,
	}
}

// Name returns the name of this scraper
func (s *G2Scraper) Name() string {
	return "G2"
}

// IsEnabled returns whether this scraper is enabled
func (s *G2Scraper) IsEnabled() bool {
	return s.enabled
}

// G2Review represents a review from G2's API
type G2Review struct {
	ID              string `json:"id"`
	Title           string `json:"headline"`
	Content         string `json:"review_text"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	Stars           int    `json:"stars"`
	Pros            string `json:"pros"`
	Cons            string `json:"cons"`
	UseCase         string `json:"use_case"`
	Recommendations string `json:"recommendations"`
	ReviewerInfo    struct {
		Name           string `json:"reviewer_name"`
		JobTitle       string `json:"job_title"`
		Company        string `json:"company"`
		Industry       string `json:"industry"`
		CompanySize    string `json:"company_size"`
		IsVerified     bool   `json:"is_verified"`
		TimeUsed       string `json:"time_used"`
		Relationship   string `json:"relationship"`
		DeploymentType string `json:"deployment_type"`
	} `json:"reviewer_info"`
	VendorResponse struct {
		Content   string `json:"response_text"`
		CreatedAt string `json:"created_at"`
		Author    string `json:"author"`
	} `json:"vendor_response,omitempty"`
}

// G2Response represents a page of reviews from G2's API
type G2Response struct {
	Reviews    []G2Review `json:"reviews"`
	Pagination struct {
		CurrentPage int `json:"current_page"`
		TotalPages  int `json:"total_pages"`
		TotalCount  int `json:"total_count"`
	} `json:"pagination"`
}

// Scrape retrieves reviews from G2
func (s *G2Scraper) Scrape(ctx context.Context) ([]models.Review, error) {
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
			return allReviews, fmt.Errorf("error scraping G2 page %d: %w", page, err)
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
			case <-time.After(s.rateLimits.PauseDuration):
				// Continue after pause
			case <-ctx.Done():
				return allReviews, ctx.Err()
			}
		}
	}

	return allReviews, nil
}

// scrapePage retrieves reviews from a single page on G2
func (s *G2Scraper) scrapePage(ctx context.Context, page int) ([]models.Review, bool, error) {
	// In a real implementation, we would use G2's API if available
	// Otherwise, we need to scrape the HTML page
	// Construct URL for the page of reviews
	url := fmt.Sprintf("https://www.g2.com/products/%s/reviews?page=%d", s.config.ProductID, page)

	// Create request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("error creating request: %w", err)
	}

	// Set user agent if rotation is enabled
	if s.rateLimits.RandomizeUserAgents && len(s.userAgents) > 0 {
		userAgent := s.userAgents[time.Now().UnixNano()%int64(len(s.userAgents))]
		req.Header.Set("User-Agent", userAgent)
	}

	// Send request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("G2 returned non-OK status: %d", resp.StatusCode)
	}

	// Parse HTML response to extract reviews
	reviews, hasMorePages, err := s.parseHTMLResponse(resp, page)
	if err != nil {
		return nil, false, fmt.Errorf("error parsing HTML response: %w", err)
	}

	return reviews, hasMorePages, nil
}

// parseHTMLResponse extracts reviews from the HTML response
func (s *G2Scraper) parseHTMLResponse(resp *http.Response, page int) ([]models.Review, bool, error) {
	var reviews []models.Review

	// Read the entire body
	// In a real implementation, you would use an HTML parser like goquery
	// This is a simplified implementation for educational purposes

	// Note: In a real implementation, you would use proper HTML parsing with goquery
	// For this example, we'll simulate finding reviews in the HTML

	// Create sample reviews for demonstration
	now := time.Now()

	// Simulate reviews per page (fewer as page increases)
	reviewsPerPage := 10 - page
	if reviewsPerPage < 3 {
		reviewsPerPage = 3
	}

	for i := 0; i < reviewsPerPage; i++ {
		reviewID := fmt.Sprintf("g2-%s-%d-%d", s.config.ProductID, page, i)
		rating := float64(3 + (i % 3)) // Ratings between 3-5

		// Mix of positive and detailed reviews
		var content, pros, cons string

		switch i % 4 {
		case 0:
			content = "The DNS and DHCP management capabilities are excellent. Easy to configure and maintain."
			pros = "Reliable DNS resolution, good IPAM features"
			cons = "UI could be more intuitive, steep learning curve for beginners"
		case 1:
			content = "We've been using this for our enterprise network for 3 years. Generally stable but occasionally has issues with high load."
			pros = "Scalability, integration with other systems"
			cons = "Performance can be slow with large networks, expensive licensing"
		case 2:
			content = "Great for managing complex network environments. The automation capabilities save us hours each week."
			pros = "Automation features, good API"
			cons = "Documentation could be better, support response times vary"
		case 3:
			content = "Solid product for DDI management. We migrated from a competitor and found the process smoother than expected."
			pros = "Feature-rich, reliable"
			cons = "Updates sometimes introduce new bugs, reporting could be improved"
		}

		review := models.Review{
			ID:          reviewID,
			Source:      "g2",
			SourceID:    reviewID,
			Content:     content,
			Title:       fmt.Sprintf("%d-Star Review of Infoblox DDI Solution", int(rating)),
			Author:      fmt.Sprintf("Network Admin %d", (page-1)*10+i),
			Rating:      &rating,
			URL:         fmt.Sprintf("https://www.g2.com/products/%s/reviews#review-%s", s.config.ProductID, reviewID),
			CreatedAt:   now.Add(-time.Duration((page-1)*240+(i*24)) * time.Hour), // Each page is older
			RetrievedAt: time.Now(),
			Metadata: map[string]interface{}{
				"platform":            "G2",
				"product_id":          s.config.ProductID,
				"pros":                pros,
				"cons":                cons,
				"reviewer_title":      fmt.Sprintf("IT Director %d", i%3+1),
				"company_size":        fmt.Sprintf("%d-500 employees", (i%5+1)*100),
				"industry":            []string{"Technology", "Healthcare", "Finance", "Education"}[i%4],
				"time_used":           fmt.Sprintf("%d+ years", i%3+1),
				"has_vendor_response": i%3 == 0,
				"vendor_response": func() string {
					if i%3 == 0 {
						return "Thank you for your review. We appreciate your feedback and are working on improving our UI."
					}
					return ""
				}(),
			},
		}

		reviews = append(reviews, review)
	}

	// In a real implementation, we would determine if there are more pages by looking for pagination elements
	// For this example, we'll assume there are more pages if page < maxPages and we still have a reasonable number of reviews
	hasMorePages := page < s.config.MaxPages && reviewsPerPage >= 3

	return reviews, hasMorePages, nil
}
