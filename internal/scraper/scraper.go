package scraper

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Infoblox-CTO/review-scraper/internal/config"
	"github.com/Infoblox-CTO/review-scraper/pkg/models"
)

// Scraper is the interface that all platform-specific scrapers must implement
type Scraper interface {
	// Name returns the name of the scraper (e.g., "Twitter", "Reddit")
	Name() string

	// Scrape retrieves reviews from the source
	Scrape(ctx context.Context) ([]models.Review, error)

	// IsEnabled checks if this scraper is enabled in config
	IsEnabled() bool
}

// Manager manages all scraper instances and coordinates scraping operations
type Manager struct {
	scrapers []Scraper
	config   config.ScrapersConfig
}

// NewManager creates a new scraper manager with the provided configuration
func NewManager(cfg config.ScrapersConfig) *Manager {
	m := &Manager{
		config: cfg,
	}

	// Initialize all scrapers
	m.initializeScrapers()

	return m
}

// initializeScrapers sets up all available scrapers based on configuration
func (m *Manager) initializeScrapers() {
	// Initialize Twitter scraper if enabled
	if m.config.Twitter.Enabled {
		m.scrapers = append(m.scrapers, NewTwitterScraper(m.config.Twitter, m.config.RateLimits, m.config.ProxySettings))
	}

	// Initialize Reddit scraper if enabled
	if m.config.Reddit.Enabled {
		m.scrapers = append(m.scrapers, NewRedditScraper(m.config.Reddit, m.config.RateLimits, m.config.ProxySettings))
	}

	// Initialize App Store scraper if enabled
	if m.config.AppStore.Enabled {
		m.scrapers = append(m.scrapers, NewAppStoreScraper(m.config.AppStore, m.config.RateLimits, m.config.ProxySettings))
	}

	// Initialize Google Play scraper if enabled
	if m.config.GooglePlay.Enabled {
		m.scrapers = append(m.scrapers, NewGooglePlayScraper(m.config.GooglePlay, m.config.RateLimits, m.config.ProxySettings))
	}

	// Initialize custom site scrapers
	for _, customCfg := range m.config.CustomSites {
		if customCfg.Enabled {
			m.scrapers = append(m.scrapers, NewCustomSiteScraper(customCfg, m.config.RateLimits, m.config.ProxySettings))
		}
	}
}

// ScrapeAll runs all enabled scrapers in parallel and aggregates their results
func (m *Manager) ScrapeAll(ctx context.Context) ([]models.Review, error) {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results []models.Review
		errs    []error
	)

	// Create a context with timeout to prevent hanging scrapers
	ctx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	// Start each scraper in its own goroutine
	for _, s := range m.scrapers {
		if !s.IsEnabled() {
			continue
		}

		wg.Add(1)
		go func(scraper Scraper) {
			defer wg.Done()

			reviews, err := scraper.Scrape(ctx)

			// Lock while updating shared data
			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				errs = append(errs, fmt.Errorf("%s scraper error: %w", scraper.Name(), err))
				return
			}

			results = append(results, reviews...)
		}(s)
	}

	// Wait for all scrapers to complete
	wg.Wait()

	// If no reviews were found but errors occurred, return an error
	if len(results) == 0 && len(errs) > 0 {
		// Combine all error messages
		var combinedErr error
		for _, err := range errs {
			if combinedErr == nil {
				combinedErr = err
			} else {
				combinedErr = fmt.Errorf("%v; %w", combinedErr, err)
			}
		}
		return nil, combinedErr
	}

	return results, nil
}

// GetScrapers returns all enabled scrapers
func (m *Manager) GetScrapers() []Scraper {
	var enabledScrapers []Scraper
	for _, s := range m.scrapers {
		if s.IsEnabled() {
			enabledScrapers = append(enabledScrapers, s)
		}
	}
	return enabledScrapers
}

// GetStats returns statistics about all scrapers
func (m *Manager) GetStats() []models.ScraperStats {
	// In a real implementation, this would track and return actual stats
	// This is a placeholder implementation
	stats := make([]models.ScraperStats, 0, len(m.scrapers))
	for _, s := range m.scrapers {
		if s.IsEnabled() {
			stats = append(stats, models.ScraperStats{
				Source:         s.Name(),
				StartTime:      time.Now().Add(-1 * time.Hour),
				EndTime:        time.Now(),
				ReviewsScraped: 0,
			})
		}
	}
	return stats
}

// NewRedditScraper creates a new Reddit scraper
func NewRedditScraper(cfg config.RedditScraperConfig, rates config.RateLimitConfig, proxies config.ProxyConfig) Scraper {
	// TODO: Implement actual Reddit scraper
	// This is a stub implementation to make the code compile
	return &stubScraper{
		name:    "Reddit",
		enabled: cfg.Enabled,
	}
}

// NewAppStoreScraper creates a new App Store scraper
func NewAppStoreScraper(cfg config.AppStoreScraperConfig, rates config.RateLimitConfig, proxies config.ProxyConfig) Scraper {
	// TODO: Implement actual App Store scraper
	// This is a stub implementation to make the code compile
	return &stubScraper{
		name:    "AppStore",
		enabled: cfg.Enabled,
	}
}

// NewGooglePlayScraper creates a new Google Play Store scraper
func NewGooglePlayScraper(cfg config.GooglePlayScraperConfig, rates config.RateLimitConfig, proxies config.ProxyConfig) Scraper {
	// TODO: Implement actual Google Play scraper
	// This is a stub implementation to make the code compile
	return &stubScraper{
		name:    "GooglePlay",
		enabled: cfg.Enabled,
	}
}

// NewCustomSiteScraper creates a new custom site scraper
func NewCustomSiteScraper(cfg config.CustomSiteScraperConfig, rates config.RateLimitConfig, proxies config.ProxyConfig) Scraper {
	// TODO: Implement actual custom site scraper
	// This is a stub implementation to make the code compile
	return &stubScraper{
		name:    cfg.Name,
		enabled: cfg.Enabled,
	}
}

// stubScraper is a simple stub implementation of the Scraper interface
// It's used to provide placeholder implementations for scrapers that aren't fully implemented yet
type stubScraper struct {
	name    string
	enabled bool
}

// Name returns the name of this scraper
func (s *stubScraper) Name() string {
	return s.name
}

// IsEnabled returns whether this scraper is enabled
func (s *stubScraper) IsEnabled() bool {
	return s.enabled
}

// Scrape is a stub implementation that returns an empty set of reviews
func (s *stubScraper) Scrape(ctx context.Context) ([]models.Review, error) {
	return []models.Review{}, nil
}
