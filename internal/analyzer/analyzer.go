package analyzer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Infoblox-CTO/review-scraper/internal/config"
	"github.com/Infoblox-CTO/review-scraper/pkg/models"
)

// Analyzer processes review text to determine sentiment and intent
type Analyzer struct {
	config          config.AnalyzerConfig
	httpClient      *http.Client
	keywordMap      map[string]bool
	infobloxTerms   map[string]string   // Maps Infoblox terms to their categories
	productKeywords map[string][]string // Maps product categories to relevant keywords
	cache           map[string]models.AnalysisResult
	cacheMutex      sync.RWMutex
	categoryMap     map[string]string // Maps keywords to categories
}

// New creates a new analyzer with the provided configuration
func New(cfg config.AnalyzerConfig) *Analyzer {
	// Create a map for faster keyword lookups
	keywordMap := make(map[string]bool, len(cfg.Keywords))
	for _, keyword := range cfg.Keywords {
		keywordMap[strings.ToLower(keyword)] = true
	}

	// Define some default category mappings if needed
	defaultCategories := map[string]string{
		"bug":        "bug_report",
		"crash":      "bug_report",
		"error":      "bug_report",
		"broken":     "bug_report",
		"freeze":     "bug_report",
		"slow":       "performance",
		"laggy":      "performance",
		"hang":       "performance",
		"feature":    "feature_request",
		"missing":    "feature_request",
		"wish":       "feature_request",
		"delivery":   "logistics",
		"shipping":   "logistics",
		"payment":    "billing",
		"charge":     "billing",
		"refund":     "billing",
		"support":    "customer_service",
		"service":    "customer_service",
		"rude":       "customer_service",
		"interface":  "ui_ux",
		"confusing":  "ui_ux",
		"difficult":  "ui_ux",
		"dns":        "network_services",
		"dhcp":       "network_services",
		"ipam":       "network_services",
		"ddi":        "network_services",
		"nios":       "nios_platform",
		"bloxone":    "bloxone_platform",
		"threat":     "security",
		"secure":     "security",
		"protection": "security",
		"ddos":       "security",
		"cloud":      "cloud_services",
		"automation": "network_automation",
		"netmri":     "network_automation",
	}

	// Create Infoblox-specific term mappings
	infobloxTerms := map[string]string{
		"ddi":            "core_product",
		"dns":            "core_product",
		"dhcp":           "core_product",
		"ipam":           "core_product",
		"bloxone":        "cloud_product",
		"nios":           "on_prem_product",
		"universal ddi":  "core_product",
		"threat defense": "security_product",
		"advanced dns":   "security_product",
		"grid":           "infrastructure",
		"netmri":         "automation",
		"cloud network":  "cloud_product",
		"ip address":     "core_product",
		"outbound dns":   "security_product",
		"dns firewall":   "security_product",
		"lookup":         "core_product",
		"zone":           "core_product",
		"record":         "core_product",
		"domain":         "core_product",
		"bind":           "core_product",
		"lease":          "core_product",
		"subnet":         "core_product",
		"network":        "core_product",
		"vista":          "company",
		"warburg":        "company",
	}

	// Define product keywords for better categorization
	productKeywords := map[string][]string{
		"bloxone": {
			"bloxone", "cloud ddi", "cloud-native", "saas", "branch", "distributed",
		},
		"nios": {
			"nios", "grid", "on-premise", "on-prem", "appliance", "virtual appliance", "vm",
		},
		"threat_defense": {
			"threat defense", "security", "protect", "threat", "malware", "ransomware",
			"exfiltration", "dns security", "secure dns", "threat intelligence",
		},
		"dns": {
			"dns", "domain", "lookup", "zone", "record", "cname", "mx", "soa", "ns", "ptr",
			"a record", "aaaa", "bind", "recursive", "resolver",
		},
		"dhcp": {
			"dhcp", "lease", "ip assignment", "dynamic", "scope", "ip address allocation",
		},
		"ipam": {
			"ipam", "ip address management", "subnet", "allocation", "address space",
			"network management", "ip management",
		},
	}

	return &Analyzer{
		config:          cfg,
		httpClient:      &http.Client{Timeout: 30 * time.Second},
		keywordMap:      keywordMap,
		infobloxTerms:   infobloxTerms,
		productKeywords: productKeywords,
		cache:           make(map[string]models.AnalysisResult),
		cacheMutex:      sync.RWMutex{},
		categoryMap:     defaultCategories,
	}
}

// Analyze processes a review to extract sentiment and intent
func (a *Analyzer) Analyze(ctx context.Context, review models.Review) (models.AnalysisResult, error) {
	// Generate a cache key based on the review content
	cacheKey := fmt.Sprintf("%x", review.Content)

	// Check if we already analyzed this review
	a.cacheMutex.RLock()
	if cachedResult, found := a.cache[cacheKey]; found {
		a.cacheMutex.RUnlock()
		return cachedResult, nil
	}
	a.cacheMutex.RUnlock()

	// Analyze based on the configured mode
	var result models.AnalysisResult
	var err error

	switch a.config.Mode {
	case "openai":
		result, err = a.analyzeWithOpenAI(ctx, review)
	case "google":
		result, err = a.analyzeWithGoogle(ctx, review)
	case "aws":
		result, err = a.analyzeWithAWS(ctx, review)
	case "azure":
		result, err = a.analyzeWithAzure(ctx, review)
	case "local":
		result, err = a.analyzeLocal(review)
	default:
		result, err = a.analyzeLocal(review)
	}

	if err != nil {
		return models.AnalysisResult{}, err
	}

	// Add the review ID to the result
	result.ReviewID = review.ID

	// Check if the result meets the thresholds for negativity and relevance
	result.IsNegative = result.SentimentScore <= a.config.NegativeThreshold
	result.IsRelevant = result.Confidence >= a.config.RelevanceThreshold

	// Cache the result for future queries
	a.cacheMutex.Lock()
	a.cache[cacheKey] = result
	a.cacheMutex.Unlock()

	return result, nil
}

// analyzeLocal performs a basic sentiment and intent analysis without external APIs
func (a *Analyzer) analyzeLocal(review models.Review) (models.AnalysisResult, error) {
	content := strings.ToLower(review.Content)

	// Basic sentiment analysis using predefined word lists
	positiveWords := []string{
		"good", "great", "awesome", "excellent", "amazing", "love", "best", "fantastic",
		"perfect", "happy", "pleased", "satisfied", "wonderful", "helpful", "thank", "thanks",
		"reliable", "secure", "efficient", "intuitive",
	}

	negativeWords := []string{
		"bad", "poor", "terrible", "awful", "horrible", "worst", "hate", "disappointed",
		"frustrating", "useless", "broken", "annoying", "slow", "expensive", "waste",
		"difficult", "confusing", "crash", "bug", "error", "problem", "issue", "fail", "fails",
		"failed", "failing", "failure", "cannot", "can't", "won't", "doesn't", "didn't",
		"insecure", "vulnerability", "breach", "outage", "downtime",
	}

	// Count word occurrences
	var positiveCount, negativeCount int

	for _, word := range positiveWords {
		positiveCount += countOccurrences(content, word)
	}

	for _, word := range negativeWords {
		negativeCount += countOccurrences(content, word)
	}

	// Calculate sentiment score between -1 (very negative) and 1 (very positive)
	var sentimentScore float64
	totalWords := positiveCount + negativeCount
	if totalWords > 0 {
		sentimentScore = float64(positiveCount-negativeCount) / float64(totalWords)
	}

	// Check for exclamation marks and ALL CAPS, which might indicate stronger sentiment
	exclamationCount := strings.Count(review.Content, "!")
	capsPattern := regexp.MustCompile(`[A-Z]{3,}`)
	capsCount := len(capsPattern.FindAllString(review.Content, -1))

	// Adjust sentiment score based on these indicators
	if exclamationCount > 2 || capsCount > 2 {
		if sentimentScore < 0 {
			// Make more negative if already negative
			sentimentScore *= 1.2
			if sentimentScore < -1 {
				sentimentScore = -1
			}
		} else if sentimentScore > 0 {
			// Make more positive if already positive
			sentimentScore *= 1.2
			if sentimentScore > 1 {
				sentimentScore = 1
			}
		}
	}

	// Determine if the review contains relevant keywords
	var keywords []string
	for keyword := range a.keywordMap {
		if strings.Contains(content, keyword) {
			keywords = append(keywords, keyword)
		}
	}

	// Infoblox product specific detection
	infobloxProducts := []string{
		"infoblox", "bloxone", "nios", "ddi", "dns firewall", "threat defense",
		"netmri", "ip address management", "ipam", "dhcp", "advanced dns protection",
	}

	for _, product := range infobloxProducts {
		if strings.Contains(content, product) && !contains(keywords, product) {
			keywords = append(keywords, product)
		}
	}

	// Determine intent category based on keywords
	categoryScores := make(map[string]float64)
	for _, keyword := range keywords {
		if category, exists := a.categoryMap[keyword]; exists {
			categoryScores[category]++
		}
	}

	// Choose the category with the highest score
	var topCategory string
	var topScore float64
	for category, score := range categoryScores {
		if score > topScore {
			topCategory = category
			topScore = score
		}
	}

	// Set default category if none was found
	if topCategory == "" {
		topCategory = "general_complaint"
	}

	// Extract simple entities (products, features)
	var entities []models.Entity

	// Infoblox specific products and features to look for
	productPatterns := []string{
		"infoblox", "bloxone", "nios", "ddi", "dhcp", "dns", "ipam", "netmri",
		"threat defense", "dns firewall", "cloud network automation",
	}

	for _, pattern := range productPatterns {
		if idx := strings.Index(content, pattern); idx >= 0 {
			entities = append(entities, models.Entity{
				Text:     pattern,
				Type:     "PRODUCT",
				Position: idx,
			})
		}
	}

	// Add generic product/feature detection as fallback
	genericPatterns := []string{"app", "website", "service", "product", "interface", "platform", "system"}
	for _, pattern := range genericPatterns {
		if idx := strings.Index(content, pattern); idx >= 0 && !containsEntityWithText(entities, pattern) {
			entities = append(entities, models.Entity{
				Text:     pattern,
				Type:     "PRODUCT",
				Position: idx,
			})
		}
	}

	// Set confidence proportional to the number of keywords found
	confidence := 0.5
	if len(keywords) > 0 {
		confidence = float64(len(keywords)) / 10.0
		if confidence > 1.0 {
			confidence = 1.0
		} else if confidence < 0.3 {
			confidence = 0.3
		}
	}

	// If the content has a star rating, use that to adjust sentiment
	if review.Rating != nil {
		// Convert 5-star scale to -1 to 1 scale
		// 1 star = -1, 5 stars = 1
		ratingScore := (*review.Rating - 3) / 2

		// Blend the lexical and rating-based scores (60% lexical, 40% rating)
		sentimentScore = sentimentScore*0.6 + ratingScore*0.4
	}

	return models.AnalysisResult{
		SentimentScore: sentimentScore,
		IsNegative:     sentimentScore < 0,
		IsRelevant:     len(keywords) > 0,
		IntentCategory: topCategory,
		Confidence:     confidence,
		Keywords:       keywords,
		Entities:       entities,
		CategoryScores: categoryScores,
	}, nil
}

// contains checks if a string is present in a slice
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// containsEntityWithText checks if an entity with the given text exists in the slice
func containsEntityWithText(entities []models.Entity, text string) bool {
	for _, entity := range entities {
		if entity.Text == text {
			return true
		}
	}
	return false
}

// countOccurrences counts how many times a word appears in text
func countOccurrences(text, word string) int {
	return len(strings.Split(text, word)) - 1
}

// OpenAI request and response types
type OpenAIRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

// analyzeWithOpenAI uses OpenAI API for sentiment and intent analysis
func (a *Analyzer) analyzeWithOpenAI(ctx context.Context, review models.Review) (models.AnalysisResult, error) {
	if a.config.APIKey == "" {
		return models.AnalysisResult{}, errors.New("OpenAI API key is not configured")
	}

	// Create a prompt that asks for sentiment analysis and intent classification
	prompt := fmt.Sprintf(`
Analyze the following product review/comment for sentiment and intent.
Return the analysis as a JSON object with the following fields:
- sentimentScore: a number between -1 (very negative) and 1 (very positive)
- intentCategory: one of [bug_report, feature_request, performance, billing, logistics, customer_service, ui_ux, security, general_complaint]
- confidence: a number between 0 and 1 indicating confidence in your analysis
- keywords: an array of important keywords from the text
- entities: an array of detected entities like product names, features, etc.
- categoryScores: a dictionary mapping each category to a relevance score

Review: "%s"
`, review.Content)

	// Create the request
	openaiReq := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{
			{
				Role:    "system",
				Content: "You are a sentiment and intent analysis system. Your responses should be valid JSON only, no prose.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		MaxTokens: 500,
	}

	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return models.AnalysisResult{}, fmt.Errorf("error marshaling OpenAI request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		"https://api.openai.com/v1/chat/completions",
		strings.NewReader(string(reqBody)),
	)
	if err != nil {
		return models.AnalysisResult{}, fmt.Errorf("error creating OpenAI request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.config.APIKey)

	// Send the request
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return models.AnalysisResult{}, fmt.Errorf("error sending request to OpenAI: %w", err)
	}
	defer resp.Body.Close()

	// Check for API errors
	if resp.StatusCode != http.StatusOK {
		return models.AnalysisResult{}, fmt.Errorf("OpenAI API returned status code %d", resp.StatusCode)
	}

	// Parse the response
	var openaiResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return models.AnalysisResult{}, fmt.Errorf("error parsing OpenAI response: %w", err)
	}

	// Ensure we have at least one choice
	if len(openaiResp.Choices) == 0 {
		return models.AnalysisResult{}, errors.New("no choices returned from OpenAI")
	}

	// Parse the JSON content from the response
	var analysisResult models.AnalysisResult
	err = json.Unmarshal([]byte(openaiResp.Choices[0].Message.Content), &analysisResult)
	if err != nil {
		return models.AnalysisResult{}, fmt.Errorf("error unmarshaling analysis result: %w", err)
	}

	return analysisResult, nil
}

// analyzeWithGoogle uses Google Cloud Natural Language API for analysis
func (a *Analyzer) analyzeWithGoogle(ctx context.Context, review models.Review) (models.AnalysisResult, error) {
	// This is a placeholder - in a real implementation, this would use the Google Cloud Natural Language API
	return models.AnalysisResult{}, errors.New("Google Cloud analysis not implemented")
}

// analyzeWithAWS uses AWS Comprehend for analysis
func (a *Analyzer) analyzeWithAWS(ctx context.Context, review models.Review) (models.AnalysisResult, error) {
	// This is a placeholder - in a real implementation, this would use AWS Comprehend
	return models.AnalysisResult{}, errors.New("AWS Comprehend analysis not implemented")
}

// analyzeWithAzure uses Azure Text Analytics for analysis
func (a *Analyzer) analyzeWithAzure(ctx context.Context, review models.Review) (models.AnalysisResult, error) {
	// This is a placeholder - in a real implementation, this would use Azure Text Analytics
	return models.AnalysisResult{}, errors.New("Azure Text Analytics analysis not implemented")
}

// GetStats returns statistics about the analyzer
func (a *Analyzer) GetStats() map[string]interface{} {
	a.cacheMutex.RLock()
	defer a.cacheMutex.RUnlock()

	return map[string]interface{}{
		"cache_size":          len(a.cache),
		"mode":                a.config.Mode,
		"negative_threshold":  a.config.NegativeThreshold,
		"relevance_threshold": a.config.RelevanceThreshold,
		"keyword_count":       len(a.keywordMap),
		"category_count":      len(a.categoryMap),
	}
}
