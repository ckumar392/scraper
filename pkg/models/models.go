package models

import (
	"time"
)

// Review represents a user review or comment from any source
type Review struct {
	ID          string                 `json:"id"`
	Source      string                 `json:"source"`      // e.g., "twitter", "reddit", "app_store"
	SourceID    string                 `json:"sourceId"`    // Original ID from the source
	Content     string                 `json:"content"`     // The review text
	Title       string                 `json:"title"`       // The title or headline of the review
	Author      string                 `json:"author"`      // Username or identifier of the reviewer
	Rating      *float64               `json:"rating"`      // Star rating if available (1-5)
	URL         string                 `json:"url"`         // Link to the original review
	CreatedAt   time.Time              `json:"createdAt"`   // When the review was posted
	RetrievedAt time.Time              `json:"retrievedAt"` // When we scraped the review
	Metadata    map[string]interface{} `json:"metadata"`    // Additional platform-specific data
}

// Tweet represents a tweet from Twitter/X platform
type Tweet struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"createdAt"`
	UserName  string    `json:"userName"`
	UserID    string    `json:"userId"`
	URLs      []string  `json:"urls,omitempty"`
	Hashtags  []string  `json:"hashtags,omitempty"`
	Likes     int       `json:"likes,omitempty"`
	Retweets  int       `json:"retweets,omitempty"`
}

// AnalysisResult represents the output of sentiment and intent analysis
type AnalysisResult struct {
	ReviewID       string             `json:"reviewId"`
	SentimentScore float64            `json:"sentimentScore"` // -1 to 1, where -1 is very negative
	IsNegative     bool               `json:"isNegative"`     // True if the sentiment is negative
	IsRelevant     bool               `json:"isRelevant"`     // True if the review is relevant to our product
	IntentCategory string             `json:"intentCategory"` // e.g., "bug_report", "feature_request"
	Confidence     float64            `json:"confidence"`     // Confidence level of the analysis
	Keywords       []string           `json:"keywords"`       // Extracted keywords
	Entities       []Entity           `json:"entities"`       // Extracted entities
	CategoryScores map[string]float64 `json:"categoryScores"` // Scores for each intent category
}

// Entity represents a named entity extracted from the review
type Entity struct {
	Text     string `json:"text"`
	Type     string `json:"type"`     // e.g., "PRODUCT", "FEATURE", "PERSON"
	Position int    `json:"position"` // Position in the text
}

// Department represents a team or department within the organization
type Department struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	ContactInfo string   `json:"contactInfo"` // Email or Slack channel
	Categories  []string `json:"categories"`  // Categories this department handles
}

// Notification represents a message sent to a department
type Notification struct {
	ID           string         `json:"id"`
	Review       Review         `json:"review"`
	Analysis     AnalysisResult `json:"analysis"`
	Department   Department     `json:"department"`
	SentAt       time.Time      `json:"sentAt"`
	Status       string         `json:"status"`                 // "sent", "delivered", "read", "actioned"
	ResponseInfo string         `json:"responseInfo,omitempty"` // Action taken by the department
}

// ScraperStats contains statistics about a scraping run
type ScraperStats struct {
	Source           string    `json:"source"`
	StartTime        time.Time `json:"startTime"`
	EndTime          time.Time `json:"endTime"`
	ReviewsScraped   int       `json:"reviewsScraped"`
	ReviewsProcessed int       `json:"reviewsProcessed"`
	Errors           int       `json:"errors"`
	ErrorDetails     []string  `json:"errorDetails,omitempty"`
}

// DashboardMetrics represents aggregated metrics for dashboard display
type DashboardMetrics struct {
	TotalReviews        int                       `json:"totalReviews"`
	NegativeReviews     int                       `json:"negativeReviews"`
	AverageSentiment    float64                   `json:"averageSentiment"`
	TopCategories       map[string]int            `json:"topCategories"`
	DepartmentWorkloads map[string]int            `json:"departmentWorkloads"`
	RecentTrends        []DailyMetric             `json:"recentTrends"`
	SourceBreakdown     map[string]int            `json:"sourceBreakdown"`
	ResponseRates       map[string]ResponseMetric `json:"responseRates"`
}

// DailyMetric represents metrics for a single day
type DailyMetric struct {
	Date             time.Time `json:"date"`
	ReviewCount      int       `json:"reviewCount"`
	NegativeCount    int       `json:"negativeCount"`
	AverageSentiment float64   `json:"averageSentiment"`
}

// ResponseMetric contains response statistics for a department
type ResponseMetric struct {
	TotalReceived    int     `json:"totalReceived"`
	Responded        int     `json:"responded"`
	AverageTimeHours float64 `json:"averageTimeHours"`
}

// APIResponse is a standardized API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// --> scrape ==> classify(models) for --> notify(specifi), create tickets
