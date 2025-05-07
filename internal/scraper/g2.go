package scraper

import (
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "net/http"
    "os"
    "path/filepath"
    "time"
)

// Review represents a standardized review structure
type Review struct {
    ID            int      `json:"id"`
    ReviewID      int      `json:"reviewID"`
    Author        string   `json:"author"`
    Platform      string   `json:"platform"`
    Title         string   `json:"title"`
    PostContent   string   `json:"Postcontent"`
    ReplyContents string   `json:"replyContents"`
    Timestamp     string   `json:"timestamp"`
    Tags          []string `json:"tags"`
    Rating        int      `json:"rating"`
}

// G2Response represents the expected response structure from G2 API
// Note: This might need adjustment based on the actual API response
type G2Response struct {
    Reviews []struct {
        ID            string   `json:"id"`
        ReviewerName  string   `json:"reviewerName"`
        Title         string   `json:"title"`
        Content       string   `json:"content"`
        VendorReply   string   `json:"vendorReply"`
        ReviewDate    string   `json:"reviewDate"`
        Tags          []string `json:"tags"`
        Rating        int      `json:"rating"`
        // Add other fields as needed
    } `json:"reviews"`
}

// G2Client handles API requests to G2
type G2Client struct {
    APIKey string
    Host   string
}

// NewG2Client creates a new G2 API client
func NewG2Client(apiKey string) *G2Client {
    return &G2Client{
        APIKey: apiKey,
        Host:   "g2-products-reviews-users2.p.rapidapi.com",
    }
}

// FetchReviews fetches reviews for a specific product from G2
func (c *G2Client) FetchReviews(product string, starRating, page int) ([]Review, error) {
    url := fmt.Sprintf("https://%s/product/%s/reviews?sortOrder=most_recent&page=%d&starRating=%d", 
        c.Host, product, page, starRating)

    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("error creating request: %w", err)
    }

    req.Header.Add("X-RapidAPI-Key", c.APIKey)
    req.Header.Add("X-RapidAPI-Host", c.Host)

    res, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error making request: %w", err)
    }
    defer res.Body.Close()

    if res.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("API returned non-200 status: %d", res.StatusCode)
    }

    body, err := io.ReadAll(res.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading response body: %w", err)
    }

    var g2Response G2Response
    if err := json.Unmarshal(body, &g2Response); err != nil {
        // Try to log the response for debugging
        fmt.Printf("Failed to parse response: %s\n", string(body))
        return nil, fmt.Errorf("error parsing JSON response: %w", err)
    }

    // Transform G2 response to our standardized Review format
    reviews := make([]Review, 0, len(g2Response.Reviews))
    for i, r := range g2Response.Reviews {
        reviewID := 0
        fmt.Sscanf(r.ID, "%d", &reviewID)
        
        review := Review{
            ID:            i + 1,
            ReviewID:      reviewID,
            Author:        r.ReviewerName,
            Platform:      "G2",
            Title:         r.Title,
            PostContent:   r.Content,
            ReplyContents: r.VendorReply,
            Timestamp:     r.ReviewDate,
            Tags:          r.Tags,
            Rating:        r.Rating,
        }
        reviews = append(reviews, review)
    }

    return reviews, nil
}

// SaveReviewsToFile saves reviews to a JSON file
func SaveReviewsToFile(reviews []Review, product string) (string, error) {
    // Create output directory if it doesn't exist
    outputDir := "output"
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        return "", fmt.Errorf("error creating output directory: %w", err)
    }

    // Generate filename with timestamp
    timestamp := time.Now().Format("20060102_150405")
    filename := fmt.Sprintf("%s_reviews_%s.json", product, timestamp)
    filePath := filepath.Join(outputDir, filename)

    // Create file
    file, err := os.Create(filePath)
    if err != nil {
        return "", fmt.Errorf("error creating file: %w", err)
    }
    defer file.Close()

    // Write reviews to file with pretty printing
    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    if err := encoder.Encode(reviews); err != nil {
        return "", fmt.Errorf("error encoding reviews to JSON: %w", err)
    }

    return filePath, nil
}

func main() {
    // Parse command line flags
    apiKey := flag.String("apikey", "", "RapidAPI key (required)")
    product := flag.String("product", "", "Product name (bloxone-ddi, infoblox-nios, or bloxone-threat-defense)")
    rating := flag.Int("rating", 0, "Filter by star rating (0-5, 0 means all ratings)")
    page := flag.Int("page", 1, "Page number")
    flag.Parse()

    // Validate API key
    if *apiKey == "" {
        apiKeyEnv := os.Getenv("RAPID_API_KEY")
        if apiKeyEnv == "" {
            fmt.Println("Error: API key is required. Set it with -apikey flag or RAPID_API_KEY environment variable")
            os.Exit(1)
        }
        *apiKey = apiKeyEnv
    }

    // Validate product name
    validProducts := map[string]bool{
        "bloxone-ddi":            true,
        "infoblox-nios":          true,
        "bloxone-threat-defense": true,
    }

    if *product == "" {
        fmt.Println("Error: Product name is required. Choose from: bloxone-ddi, infoblox-nios, bloxone-threat-defense")
        os.Exit(1)
    }

    if !validProducts[*product] {
        fmt.Printf("Error: Invalid product name. Choose from: bloxone-ddi, infoblox-nios, bloxone-threat-defense\n")
        os.Exit(1)
    }

    // Create G2 client
    client := NewG2Client(*apiKey)

    // Fetch reviews
    fmt.Printf("Fetching reviews for %s (page %d, rating %d)...\n", *product, *page, *rating)
    reviews, err := client.FetchReviews(*product, *rating, *page)
    if err != nil {
        fmt.Printf("Error fetching reviews: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Successfully fetched %d reviews\n", len(reviews))

    // Save reviews to file
    filePath, err := SaveReviewsToFile(reviews, *product)
    if err != nil {
        fmt.Printf("Error saving reviews: %v\n", err)
        os.Exit(1)
    }

    fmt.Printf("Reviews saved to %s\n", filePath)
}