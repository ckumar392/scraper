package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// InputReview represents the format of reviews in scraped_data.json
type InputReview struct {
	ID            int      `json:"id"`
	ReviewID      int      `json:"reviewID"`
	Author        string   `json:"author"`
	Platform      string   `json:"platform"`
	Title         string   `json:"title"`
	Postcontent   string   `json:"Postcontent"`
	ReplyContents string   `json:"replyContents"`
	Timestamp     string   `json:"timestamp"`
	Tags          []string `json:"tags"`
	Rating        int      `json:"rating"`
}

// EnrichedReview adds AI-generated fields to the input review
type EnrichedReview struct {
	ID            int      `json:"id"`
	ReviewID      int      `json:"reviewID"`
	Author        string   `json:"author"`
	Platform      string   `json:"platform"`
	Title         string   `json:"title"`
	Postcontent   string   `json:"Postcontent"`
	ReplyContents string   `json:"replyContents"`
	Timestamp     string   `json:"timestamp"`
	Tags          []string `json:"tags"`
	Rating        int      `json:"rating"`
	Sentiment     string   `json:"sentiment"`
	Department    string   `json:"department"`
	Product       string   `json:"product"`
	NeedsAction   bool     `json:"needsAction"`
}

// OpenAIRequest represents the request structure for the OpenAI API
type OpenAIRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

// AzureOpenAIRequest represents the request structure for the Azure OpenAI API
type AzureOpenAIRequest struct {
	Messages    []ChatMessage `json:"messages"`
	Temperature float64       `json:"temperature"`
}

// ChatMessage represents a message in the OpenAI chat completion API
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIResponse represents the response from the OpenAI API
type OpenAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

// AIAnalysisResult represents the structured output from the AI
type AIAnalysisResult struct {
	Sentiment   string `json:"sentiment"`
	Department  string `json:"department"`
	Product     string `json:"product"`
	NeedsAction bool   `json:"needsAction"`
}

// InfobloxCategories contains mappings of keywords to their categories for analysis
type InfobloxCategories struct {
	defaultCategories map[string]string
	infobloxTerms     map[string]string
	productKeywords   map[string][]string
}

// Initialize Infoblox-specific knowledge base for enhanced AI analysis
var infobloxCategories = InfobloxCategories{
	defaultCategories: map[string]string{
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
	},
	infobloxTerms: map[string]string{
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
	},
	productKeywords: map[string][]string{
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
	},
}

// Department mappings for consistent department assignment
var departmentMappings = map[string]string{
	"bug_report":         "Engineering",
	"performance":        "Engineering",
	"feature_request":    "Product",
	"ui_ux":              "Product",
	"customer_service":   "Support",
	"billing":            "Sales",
	"logistics":          "Sales",
	"network_services":   "Engineering",
	"security":           "Engineering",
	"cloud_services":     "Engineering",
	"network_automation": "Engineering",
	"nios_platform":      "Engineering",
	"bloxone_platform":   "Engineering",
	"general":            "General",
}

// Product mappings based on product keywords
var productNameMappings = map[string]string{
	"bloxone":          "BloxOne Platform",
	"nios":             "NIOS",
	"threat_defense":   "BloxOne Threat Defense",
	"dns":              "BloxOne DNS",
	"dhcp":             "BloxOne DHCP",
	"ipam":             "BloxOne IPAM",
	"cloud_network":    "BloxOne Cloud Network Automation",
	"core_product":     "BloxOne DDI",
	"security_product": "BloxOne Threat Defense",
	"network_services": "BloxOne DDI",
}

func main() {
	log.Println("Starting Review Enrichment Process...")

	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: Could not load .env file:", err)
		log.Println("Continuing with system environment variables only.")
	} else {
		log.Println("Successfully loaded environment variables from .env file")
	}

	// Add command-line flag for offline mode
	offlinePtr := flag.Bool("offline", false, "Use offline analysis mode instead of AI API")
	inputFilePtr := flag.String("input", "scraped_data.json", "Path to input JSON file")
	outputFilePtr := flag.String("output", "enriched_reviews.json", "Path to output JSON file")
	flag.Parse()

	// Define file paths
	inputFilePath := *inputFilePtr
	outputFilePath := *outputFilePtr

	// Check if input file exists
	if _, err := os.Stat(inputFilePath); os.IsNotExist(err) {
		// Try to find the file relative to the executable
		execPath, err := os.Executable()
		if err == nil {
			possiblePath := filepath.Join(filepath.Dir(execPath), inputFilePath)
			if _, err := os.Stat(possiblePath); err == nil {
				inputFilePath = possiblePath
			}
		}
	}

	// Read the scraped data
	inputData, err := os.ReadFile(inputFilePath)
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	var inputReviews []InputReview
	if err := json.Unmarshal(inputData, &inputReviews); err != nil {
		log.Fatalf("Error parsing input JSON: %v", err)
	}

	log.Printf("Processing %d reviews...", len(inputReviews))

	// Check if we're using offline mode
	useOfflineMode := *offlinePtr

	// If not explicitly offline, check for API keys in order of preference
	if !useOfflineMode {
		// First, check for Azure OpenAI credentials
		azureApiKey := os.Getenv("AZURE_OPENAI_API_KEY")
		azureEndpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
		azureDeployment := os.Getenv("AZURE_OPENAI_DEPLOYMENT")

		// If not found, check for standard OpenAI API key
		openaiApiKey := os.Getenv("OPENAI_API_KEY")

		if azureApiKey != "" && azureEndpoint != "" {
			if azureDeployment == "" {
				azureDeployment = "gpt-4.1-mini" // Set default if not specified
				log.Println("AZURE_OPENAI_DEPLOYMENT not set, using default: gpt-4.1-mini")
			}
			log.Printf("Using Azure OpenAI API for analysis with deployment %s", azureDeployment)
		} else if openaiApiKey != "" {
			log.Println("Using OpenAI API for analysis.")
		} else {
			log.Println("No API keys found. Using offline analysis mode.")
			useOfflineMode = true
		}
	} else {
		log.Println("Offline analysis mode selected. Using local analysis algorithms.")
	}

	enrichedReviews := make([]EnrichedReview, 0, len(inputReviews))

	// Process each review
	for i, inputReview := range inputReviews {
		var analysisResult AIAnalysisResult
		var err error

		if !useOfflineMode {
			// First try Azure OpenAI if credentials are available
			azureApiKey := os.Getenv("AZURE_OPENAI_API_KEY")
			azureEndpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
			azureDeployment := os.Getenv("AZURE_OPENAI_DEPLOYMENT")

			if azureDeployment == "" {
				azureDeployment = "gpt-4.1-mini" // Use default if not set
			}

			if azureApiKey != "" && azureEndpoint != "" {
				// Try to analyze with Azure OpenAI
				analysisResult, err = analyzeWithAzureOpenAI(azureApiKey, azureEndpoint, azureDeployment, inputReview)
				if err != nil {
					log.Printf("Error analyzing review %d with Azure OpenAI: %v", inputReview.ID, err)

					// Fall back to standard OpenAI if available
					openaiApiKey := os.Getenv("OPENAI_API_KEY")
					if openaiApiKey != "" {
						log.Println("Falling back to standard OpenAI API")
						analysisResult, err = analyzeWithOpenAI(openaiApiKey, inputReview)
					}
				}
			} else {
				// Try standard OpenAI
				openaiApiKey := os.Getenv("OPENAI_API_KEY")
				if openaiApiKey != "" {
					analysisResult, err = analyzeWithOpenAI(openaiApiKey, inputReview)
				}
			}

			// If we still have an error, fall back to offline mode
			if err != nil {
				log.Printf("Error analyzing review %d: %v", inputReview.ID, err)
				log.Println("Switching to offline analysis mode for this review")
				analysisResult = analyzeOffline(inputReview)
			}
		} else {
			// Use offline analysis
			analysisResult = analyzeOffline(inputReview)
		}

		// Create enriched review
		enrichedReview := EnrichedReview{
			ID:            inputReview.ID,
			ReviewID:      inputReview.ReviewID,
			Author:        inputReview.Author,
			Platform:      inputReview.Platform,
			Title:         inputReview.Title,
			Postcontent:   inputReview.Postcontent,
			ReplyContents: inputReview.ReplyContents,
			Timestamp:     inputReview.Timestamp,
			Tags:          inputReview.Tags,
			Rating:        inputReview.Rating,
			Sentiment:     analysisResult.Sentiment,
			Department:    analysisResult.Department,
			Product:       analysisResult.Product,
			NeedsAction:   analysisResult.NeedsAction,
		}

		enrichedReviews = append(enrichedReviews, enrichedReview)
		log.Printf("Processed review %d: Sentiment=%s, Department=%s, Product=%s, NeedsAction=%v",
			inputReview.ID, analysisResult.Sentiment, analysisResult.Department, analysisResult.Product, analysisResult.NeedsAction)

		// Add small delay to avoid rate limiting if using online mode
		if !useOfflineMode && i < len(inputReviews)-1 {
			time.Sleep(200 * time.Millisecond)
		}
	}

	log.Printf("Processed %d reviews successfully", len(enrichedReviews))

	// Write enriched reviews to output file
	outputData, err := json.MarshalIndent(enrichedReviews, "", "  ")
	if err != nil {
		log.Fatalf("Error creating output JSON: %v", err)
	}

	if err := os.WriteFile(outputFilePath, outputData, 0644); err != nil {
		log.Fatalf("Error writing output file: %v", err)
	}

	log.Printf("Enriched reviews saved to %s", outputFilePath)
}

// analyzeWithOpenAI uses the OpenAI API to analyze a review
func analyzeWithOpenAI(apiKey string, review InputReview) (AIAnalysisResult, error) {
	url := "https://api.openai.com/v1/chat/completions"

	// Construct the prompt for OpenAI with Infoblox-specific knowledge
	prompt := fmt.Sprintf(`
As an Infoblox product review analyzer, analyze the following review and provide a structured response.

INFOBLOX PRODUCT CONTEXT:
- Core Products: DDI (DNS, DHCP, IPAM)
- Cloud products: BloxOne Platform (cloud-native DDI solutions)
- On-prem products: NIOS (traditional appliance-based DDI)
- Security products: BloxOne Threat Defense, DNS firewall, Advanced DNS Protection
- Network Automation: NetMRI, Cloud Network Automation

REVIEW INFORMATION:
Title: %s
Content: %s
Platform: %s
Rating: %d (out of 5)
Tags: %s

DEPARTMENT MAPPING:
- Product team: Handles feature requests, UI/UX issues
- Engineering: Handles bugs, performance issues, technical problems
- Support: Handles customer service issues, documentation
- Sales: Handles billing, pricing, licensing issues
- General: General feedback that doesn't fit elsewhere

ANALYSIS TASK:
1. Sentiment: Classify as exactly one of ["Positive", "Neutral", "Negative"] based on review content and rating if there is good review must map to positive
2. Department: Assign to exactly one of ["Product", "Engineering", "Support", "Sales", "General"] based on review content
3. Product: Identify which specific Infoblox product is being discussed (BloxOne Platform, NIOS, BloxOne Threat Defense, BloxOne DNS, BloxOne DHCP, BloxOne IPAM, etc.)
4. NeedsAction: Set to true if this is a high-priority issue that needs immediate attention (e.g., negative review with serious issues, security concern, etc.), otherwise false

Respond with a JSON object containing ONLY these four fields:
{
  "sentiment": "Positive/Neutral/Negative",
  "department": "Department name",
  "product": "Product name",
  "needsAction": true/false
}
`, review.Title, review.Postcontent, review.Platform, review.Rating, strings.Join(review.Tags, ", "))

	// Create the request body
	requestBody := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are an expert Infoblox product review analyzer. You classify reviews by sentiment, department, product, and prioritize actions needed.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.2, // Lower temperature for more consistent results
	}

	// Convert the request to JSON
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return AIAnalysisResult{}, fmt.Errorf("error marshalling request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestJSON))
	if err != nil {
		return AIAnalysisResult{}, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return AIAnalysisResult{}, fmt.Errorf("error sending request to OpenAI: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return AIAnalysisResult{}, fmt.Errorf("error reading response: %v", err)
	}

	// Check if the response status is not OK
	if resp.StatusCode != http.StatusOK {
		return AIAnalysisResult{}, fmt.Errorf("OpenAI API error: %s", string(respBody))
	}

	// Parse the response
	var openAIResp OpenAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return AIAnalysisResult{}, fmt.Errorf("error parsing OpenAI response: %v", err)
	}

	// Check if there's content in the response
	if len(openAIResp.Choices) == 0 {
		return AIAnalysisResult{}, fmt.Errorf("no content in OpenAI response")
	}

	// Extract the AI's analysis from the response content
	analysisJSON := openAIResp.Choices[0].Message.Content

	// Parse the AI's analysis
	var analysisResult AIAnalysisResult
	if err := json.Unmarshal([]byte(analysisJSON), &analysisResult); err != nil {
		return AIAnalysisResult{}, fmt.Errorf("error parsing AI analysis: %v\nResponse was: %s", err, analysisJSON)
	}

	return analysisResult, nil
}

// analyzeWithAzureOpenAI uses the Azure OpenAI API to analyze a review
func analyzeWithAzureOpenAI(apiKey string, endpoint string, deploymentName string, review InputReview) (AIAnalysisResult, error) {
	// Construct the URL for Azure OpenAI API
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2025-01-01-preview", endpoint, deploymentName)

	// Construct the prompt for OpenAI with Infoblox-specific knowledge
	prompt := fmt.Sprintf(`
As an Infoblox product review analyzer, analyze the following review and provide a structured response.

INFOBLOX PRODUCT CONTEXT:
- Core Products: DDI (DNS, DHCP, IPAM)
- Cloud products: BloxOne Platform (cloud-native DDI solutions)
- On-prem products: NIOS (traditional appliance-based DDI)
- Security products: BloxOne Threat Defense, DNS firewall, Advanced DNS Protection
- Network Automation: NetMRI, Cloud Network Automation

REVIEW INFORMATION:
Title: %s
Content: %s
Platform: %s
Rating: %d (out of 5)
Tags: %s

DEPARTMENT MAPPING:
- Product team: Handles feature requests, UI/UX issues
- Engineering: Handles bugs, performance issues, technical problems
- Support: Handles customer service issues, documentation
- Sales: Handles billing, pricing, licensing issues
- General: General feedback that doesn't fit elsewhere

ANALYSIS TASK:
1. Sentiment: Classify as exactly one of ["Positive", "Neutral", "Negative"] based on review content and rating
2. Department: Assign to exactly one of ["Product", "Engineering", "Support", "Sales", "General"] based on review content
3. Product: Identify which specific Infoblox product is being discussed (BloxOne Platform, NIOS, BloxOne Threat Defense, BloxOne DNS, BloxOne DHCP, BloxOne IPAM, etc.)
4. NeedsAction: Set to true if this is a high-priority issue that needs immediate attention (e.g., negative review with serious issues, security concern, etc.), otherwise false

Respond with a JSON object containing ONLY these four fields:
{
  "sentiment": "Positive/Neutral/Negative",
  "department": "Department name",
  "product": "Product name",
  "needsAction": true/false
}
`, review.Title, review.Postcontent, review.Platform, review.Rating, strings.Join(review.Tags, ", "))

	// Create the request body
	requestBody := AzureOpenAIRequest{
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are an expert Infoblox product review analyzer. You classify reviews by sentiment, department, product, and prioritize actions needed.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.2, // Lower temperature for more consistent results
	}

	// Convert the request to JSON
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return AIAnalysisResult{}, fmt.Errorf("error marshalling request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestJSON))
	if err != nil {
		return AIAnalysisResult{}, fmt.Errorf("error creating request: %v", err)
	}

	// Set headers for Azure OpenAI
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	// Send the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return AIAnalysisResult{}, fmt.Errorf("error sending request to Azure OpenAI: %v", err)
	}
	defer resp.Body.Close()

	// Read the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return AIAnalysisResult{}, fmt.Errorf("error reading response: %v", err)
	}

	// Check if the response status is not OK
	if resp.StatusCode != http.StatusOK {
		return AIAnalysisResult{}, fmt.Errorf("Azure OpenAI API error: %s", string(respBody))
	}

	// Parse the response (same structure as OpenAI response)
	var openAIResp OpenAIResponse
	if err := json.Unmarshal(respBody, &openAIResp); err != nil {
		return AIAnalysisResult{}, fmt.Errorf("error parsing Azure OpenAI response: %v", err)
	}

	// Check if there's content in the response
	if len(openAIResp.Choices) == 0 {
		return AIAnalysisResult{}, fmt.Errorf("no content in Azure OpenAI response")
	}

	// Extract the AI's analysis from the response content
	analysisJSON := openAIResp.Choices[0].Message.Content

	// Parse the AI's analysis
	var analysisResult AIAnalysisResult
	if err := json.Unmarshal([]byte(analysisJSON), &analysisResult); err != nil {
		return AIAnalysisResult{}, fmt.Errorf("error parsing AI analysis: %v\nResponse was: %s", err, analysisJSON)
	}

	return analysisResult, nil
}

// analyzeOffline performs a local analysis of the review without using external APIs
func analyzeOffline(review InputReview) AIAnalysisResult {
	result := AIAnalysisResult{
		Sentiment:   determineSentiment(review),
		Department:  determineDepartment(review),
		Product:     determineProduct(review),
		NeedsAction: determineNeedsAction(review),
	}

	return result
}

// determineSentiment classifies the review as positive, neutral, or negative based on rating and content
func determineSentiment(review InputReview) string {
	// First check the rating if available
	if review.Rating > 0 {
		if review.Rating >= 4 {
			return "Positive"
		} else if review.Rating <= 2 {
			return "Negative"
		} else {
			// Rating is 3, need to analyze content for better determination
		}
	}

	// Analyze content for sentiment
	text := strings.ToLower(review.Title + " " + review.Postcontent)

	// Define simple positive and negative word lists
	positiveWords := []string{
		"good", "great", "excellent", "amazing", "awesome", "helpful", "love",
		"best", "easy", "useful", "fantastic", "wonderful", "perfect", "solved",
		"works", "like", "recommend", "satisfied", "impressed",
	}

	negativeWords := []string{
		"bad", "poor", "terrible", "awful", "horrible", "issue", "problem",
		"bug", "crash", "error", "difficult", "confusing", "useless", "worst",
		"hate", "disappointing", "failed", "doesn't work", "doesn't", "don't",
		"not", "never", "slow", "issues", "problems", "bugs", "errors", "expensive",
	}

	// Count positive and negative words
	positiveCount := 0
	negativeCount := 0

	for _, word := range positiveWords {
		positiveCount += strings.Count(text, word)
	}

	for _, word := range negativeWords {
		negativeCount += strings.Count(text, word)
	}

	// Check for negations that flip sentiment
	negations := []string{"not", "don't", "doesn't", "didn't", "never", "no"}
	for _, negation := range negations {
		// This is a simple approach - for each negation, find nearby positive words and reduce the positive count
		matches := regexp.MustCompile(negation+`\s+(\w+\s+){0,3}(`+strings.Join(positiveWords, "|")+`)`).FindAllString(text, -1)
		positiveCount -= len(matches)
		negativeCount += len(matches)
	}

	// Determine sentiment based on counts
	if positiveCount > negativeCount {
		return "Positive"
	} else if negativeCount > positiveCount {
		return "Negative"
	} else {
		// If tied or no strong indicators, check rating again or default to Neutral
		if review.Rating == 4 || review.Rating == 5 {
			return "Positive"
		} else if review.Rating == 1 || review.Rating == 2 {
			return "Negative"
		} else {
			return "Neutral"
		}
	}
}

// determineDepartment assigns the review to a department based on content analysis
func determineDepartment(review InputReview) string {
	text := strings.ToLower(review.Title + " " + review.Postcontent)

	// Department keyword maps
	departmentKeywords := map[string][]string{
		"Product": {
			"feature", "missing", "wish", "roadmap", "functionality", "ui", "ux",
			"interface", "design", "workflow", "user", "experience", "implement",
		},
		"Engineering": {
			"bug", "error", "crash", "broken", "fix", "issue", "performance", "slow",
			"not working", "doesn't work", "failed", "technical", "latency", "problem",
		},
		"Support": {
			"support", "help", "documentation", "guide", "customer service", "response",
			"ticket", "case", "resolution", "assistance", "helped", "responsive",
		},
		"Sales": {
			"price", "cost", "expensive", "pricing", "license", "subscription",
			"renewal", "contract", "quote", "discount", "offer", "deal", "sales",
		},
	}

	// Count hits for each department
	departmentCounts := map[string]int{
		"Product":     0,
		"Engineering": 0,
		"Support":     0,
		"Sales":       0,
		"General":     0,
	}

	// Tally keywords for each department
	for dept, keywords := range departmentKeywords {
		for _, keyword := range keywords {
			departmentCounts[dept] += strings.Count(text, keyword)
		}
	}

	// Find department with highest count
	maxCount := 0
	maxDept := "General" // default

	for dept, count := range departmentCounts {
		if count > maxCount {
			maxCount = count
			maxDept = dept
		}
	}

	// If we found a clear department, return it
	if maxCount > 0 {
		return maxDept
	}

	// If no department keywords found, use other signals
	if review.Rating <= 2 {
		return "Support" // Low ratings often need customer support intervention
	}

	return "General"
}

// determineProduct identifies which Infoblox product the review is about
func determineProduct(review InputReview) string {
	text := strings.ToLower(review.Title + " " + review.Postcontent)
	tags := []string{}
	for _, tag := range review.Tags {
		tags = append(tags, strings.ToLower(tag))
	}

	// Check for specific product mentions
	productMatches := map[string]int{
		"BloxOne Platform":                 0,
		"NIOS":                             0,
		"BloxOne Threat Defense":           0,
		"BloxOne DNS":                      0,
		"BloxOne DHCP":                     0,
		"BloxOne IPAM":                     0,
		"BloxOne Cloud Network Automation": 0,
		"BloxOne DDI":                      0,
	}

	// Match product keywords to specific products
	for product, keywords := range infobloxCategories.productKeywords {
		for _, keyword := range keywords {
			count := strings.Count(text, keyword)

			// Check tags too
			for _, tag := range tags {
				if strings.Contains(tag, keyword) {
					count++
				}
			}

			// Assign counts to proper product names
			switch product {
			case "bloxone":
				productMatches["BloxOne Platform"] += count
			case "nios":
				productMatches["NIOS"] += count
			case "threat_defense":
				productMatches["BloxOne Threat Defense"] += count
			case "dns":
				productMatches["BloxOne DNS"] += count
				productMatches["BloxOne DDI"] += count / 2 // Partial match for DDI
			case "dhcp":
				productMatches["BloxOne DHCP"] += count
				productMatches["BloxOne DDI"] += count / 2 // Partial match for DDI
			case "ipam":
				productMatches["BloxOne IPAM"] += count
				productMatches["BloxOne DDI"] += count / 2 // Partial match for DDI
			}
		}
	}

	// Find product with most mentions
	maxCount := 0
	maxProduct := "BloxOne Platform" // Default product if nothing specific mentioned

	for product, count := range productMatches {
		if count > maxCount {
			maxCount = count
			maxProduct = product
		}
	}

	// If DDI components are mentioned together, prioritize the overall DDI product
	ddiComponents := productMatches["BloxOne DNS"] + productMatches["BloxOne DHCP"] + productMatches["BloxOne IPAM"]
	if ddiComponents >= 2 && productMatches["BloxOne DDI"] > 0 {
		return "BloxOne DDI"
	}

	// Special case for security-related content
	securityTerms := []string{"security", "threat", "protection", "malware", "ransomware", "vulnerability"}
	securityCount := 0

	for _, term := range securityTerms {
		securityCount += strings.Count(text, term)
		for _, tag := range tags {
			if strings.Contains(tag, term) {
				securityCount++
			}
		}
	}

	if securityCount > 0 && strings.Contains(text, "dns") {
		return "BloxOne Threat Defense"
	}

	return maxProduct
}

// determineNeedsAction flags high-priority issues that need attention
func determineNeedsAction(review InputReview) bool {
	// Low ratings are high priority
	if review.Rating <= 2 {
		return true
	}

	text := strings.ToLower(review.Title + " " + review.Postcontent)

	// Check for urgent keywords
	urgentKeywords := []string{
		"urgent", "critical", "immediately", "security", "breach",
		"broken", "unusable", "crash", "down", "outage", "emergency",
		"compromised", "vulnerability", "attacked", "hacked",
	}

	for _, keyword := range urgentKeywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}

	// Check for negative sentiment combined with important product mentions
	if determineSentiment(review) == "Negative" {
		importantProducts := []string{"dns", "dhcp", "security", "threat"}
		for _, product := range importantProducts {
			if strings.Contains(text, product) {
				return true
			}
		}
	}

	return false
}
