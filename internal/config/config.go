package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Config represents the main application configuration
type Config struct {
	// General settings
	ScrapingInterval time.Duration `json:"scrapingInterval"`

	// Component-specific configurations
	Scrapers ScrapersConfig `json:"scrapers"`
	Analyzer AnalyzerConfig `json:"analyzer"`
	Router   RouterConfig   `json:"router"`
	Notifier NotifierConfig `json:"notifier"`
	API      APIConfig      `json:"api"`
}

// ScrapersConfig contains settings for all scrapers
type ScrapersConfig struct {
	Twitter       TwitterScraperConfig      `json:"twitter"`
	Reddit        RedditScraperConfig       `json:"reddit"`
	AppStore      AppStoreScraperConfig     `json:"appStore"`
	GooglePlay    GooglePlayScraperConfig   `json:"googlePlay"`
	G2            G2ScraperConfig           `json:"g2"`
	Trustpilot    TrustpilotScraperConfig   `json:"trustpilot"`
	CustomSites   []CustomSiteScraperConfig `json:"customSites"`
	RateLimits    RateLimitConfig           `json:"rateLimits"`
	ProxySettings ProxyConfig               `json:"proxySettings"`
}

// TwitterConfig is an alias for TwitterScraperConfig for backward compatibility
type TwitterConfig TwitterScraperConfig

// TwitterScraperConfig contains Twitter-specific scraper settings
type TwitterScraperConfig struct {
	Enabled      bool     `json:"enabled"`
	APIKey       string   `json:"apiKey"`
	APISecret    string   `json:"apiSecret"`
	AccessToken  string   `json:"accessToken"`
	AccessSecret string   `json:"accessSecret"`
	Keywords     []string `json:"keywords"`
	ExcludeWords []string `json:"excludeWords"`
	MaxResults   int      `json:"maxResults"`
}

// RedditScraperConfig contains Reddit-specific scraper settings
type RedditScraperConfig struct {
	Enabled      bool     `json:"enabled"`
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	Subreddits   []string `json:"subreddits"`
	Keywords     []string `json:"keywords"`
	TimeFrame    string   `json:"timeFrame"`
}

// AppStoreScraperConfig contains App Store scraper settings
type AppStoreScraperConfig struct {
	Enabled   bool     `json:"enabled"`
	AppIDs    []string `json:"appIds"`
	Countries []string `json:"countries"`
	MaxPages  int      `json:"maxPages"`
}

// GooglePlayScraperConfig contains Google Play Store scraper settings
type GooglePlayScraperConfig struct {
	Enabled   bool     `json:"enabled"`
	AppIDs    []string `json:"appIds"`
	Countries []string `json:"countries"`
	MaxPages  int      `json:"maxPages"`
}

// G2ScraperConfig contains G2 review site scraper settings
type G2ScraperConfig struct {
	Enabled   bool   `json:"enabled"`
	ProductID string `json:"product_id"`
	MaxPages  int    `json:"maxPages"`
}

// TrustpilotScraperConfig contains Trustpilot scraper settings
type TrustpilotScraperConfig struct {
	Enabled    bool   `json:"enabled"`
	BusinessID string `json:"business_id"`
	MaxPages   int    `json:"maxPages"`
}

// CustomSiteScraperConfig contains settings for custom website scrapers
type CustomSiteScraperConfig struct {
	Enabled      bool     `json:"enabled"`
	Name         string   `json:"name"`
	URL          string   `json:"url"`
	ReviewURLs   []string `json:"reviewUrls"`
	ReviewXPaths []string `json:"reviewXPaths"`
	DateXPath    string   `json:"dateXPath"`
	AuthorXPath  string   `json:"authorXPath"`
	RatingXPath  string   `json:"ratingXPath"`
}

// RateLimitConfig contains settings for rate limiting
type RateLimitConfig struct {
	RequestsPerMinute    int           `json:"requestsPerMinute"`
	PauseAfterRequests   int           `json:"pauseAfterRequests"`
	PauseDuration        time.Duration `json:"pauseDuration"`
	PauseBetweenRequests bool          `json:"pauseBetweenRequests"`
	RandomizeUserAgents  bool          `json:"randomizeUserAgents"`
	RandomizePauseTimes  bool          `json:"randomizePauseTimes"`
}

// ProxyConfig contains proxy settings for scrapers
type ProxyConfig struct {
	Enabled  bool     `json:"enabled"`
	URL      string   `json:"url"`
	URLs     []string `json:"urls"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	Rotate   bool     `json:"rotate"`
}

// AnalyzerConfig contains settings for the sentiment and intent analyzer
type AnalyzerConfig struct {
	Mode               string   `json:"mode"` // local, openai, google, aws, or azure
	ModelEndpoint      string   `json:"modelEndpoint"`
	APIKey             string   `json:"apiKey"`
	NegativeThreshold  float64  `json:"negativeThreshold"`
	RelevanceThreshold float64  `json:"relevanceThreshold"`
	Keywords           []string `json:"keywords"`
	IntentCategories   []string `json:"intentCategories"`
}

// RouterConfig contains settings for the department router
type RouterConfig struct {
	Mappings          []DepartmentMapping `json:"mappings"`
	DefaultDepartment string              `json:"defaultDepartment"`
}

// DepartmentMapping maps a category to a department
type DepartmentMapping struct {
	Category   string `json:"category"`
	Department string `json:"department"`
	Priority   int    `json:"priority"`
}

// NotifierConfig contains settings for notifications
type NotifierConfig struct {
	Email     EmailConfig     `json:"email"`
	Slack     SlackConfig     `json:"slack"`
	Dashboard DashboardConfig `json:"dashboard"`
	Databases DatabaseConfig  `json:"databases"`
}

// EmailConfig contains email notification settings
type EmailConfig struct {
	Enabled       bool              `json:"enabled"`
	SMTPServer    string            `json:"smtpServer"`
	SMTPPort      int               `json:"smtpPort"`
	Username      string            `json:"username"`
	Password      string            `json:"password"`
	FromAddress   string            `json:"fromAddress"`
	DeptAddresses map[string]string `json:"departmentAddresses"`
}

// SlackConfig contains Slack notification settings
type SlackConfig struct {
	Enabled      bool              `json:"enabled"`
	WebhookURL   string            `json:"webhookUrl"`
	DeptChannels map[string]string `json:"departmentChannels"`
}

// DashboardConfig contains dashboard settings
type DashboardConfig struct {
	Enabled        bool          `json:"enabled"`
	UpdateInterval time.Duration `json:"updateInterval"`
	Port           int           `json:"port"`
}

// DatabaseConfig contains database settings for storing reviews
type DatabaseConfig struct {
	Enabled  bool   `json:"enabled"`
	Type     string `json:"type"` // mysql, postgres, mongodb
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	DBName   string `json:"dbName"`
}

// APIConfig contains REST API server settings
type APIConfig struct {
	Port            int           `json:"port"`
	EnableSwagger   bool          `json:"enableSwagger"`
	AuthToken       string        `json:"authToken"`
	RateLimit       int           `json:"rateLimit"`
	RateLimitWindow time.Duration `json:"rateLimitWindow"`
}

// Load reads the application configuration from a file
func Load() (*Config, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults if not specified
	if cfg.ScrapingInterval == 0 {
		cfg.ScrapingInterval = 1 * time.Hour // Default to every hour
	}

	return &cfg, nil
}

// getConfigPath determines the configuration file path
func getConfigPath() string {
	// Check if a path is specified via environment variable
	if path := os.Getenv("REVIEW_SCRAPER_CONFIG"); path != "" {
		return path
	}

	// Default to configs/config.json in the current directory
	return filepath.Join("configs", "config.json")
}
