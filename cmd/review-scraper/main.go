package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Infoblox-CTO/review-scraper/internal/analyzer"
	"github.com/Infoblox-CTO/review-scraper/internal/api"
	"github.com/Infoblox-CTO/review-scraper/internal/config"
	"github.com/Infoblox-CTO/review-scraper/internal/notifier"
	"github.com/Infoblox-CTO/review-scraper/internal/router"
	"github.com/Infoblox-CTO/review-scraper/internal/scraper"
)

func main() {
	log.Println("Starting Review Scraper System...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize components
	scraperManager := scraper.NewManager(cfg.Scrapers)
	analyzer := analyzer.New(cfg.Analyzer)
	router := router.New(cfg.Router)
	notifier := notifier.New(cfg.Notifier)

	// Start the API server
	apiServer := api.NewServer(cfg.API, scraperManager, analyzer, router, notifier)
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Printf("API server error: %v", err)
			cancel()
		}
	}()

	// Start the scraping pipeline
	go func() {
		ticker := time.NewTicker(cfg.ScrapingInterval)
		defer ticker.Stop()

		// Run immediately upon startup
		runPipeline(ctx, scraperManager, analyzer, router, notifier)

		for {
			select {
			case <-ticker.C:
				runPipeline(ctx, scraperManager, analyzer, router, notifier)
			case <-ctx.Done():
				log.Println("Scraping pipeline stopped")
				return
			}
		}
	}()

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

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Received shutdown signal. Stopping services...")

	// Allow graceful shutdown (e.g., finish current scraping jobs)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Stop API server
	if err := apiServer.Stop(shutdownCtx); err != nil {
		log.Printf("Error stopping API server: %v", err)
	}

	log.Println("Review Scraper System stopped")
}

// runPipeline executes the complete data processing pipeline
func runPipeline(ctx context.Context, scraperManager *scraper.Manager, analyzer *analyzer.Analyzer,
	router *router.Router, notifier *notifier.Notifier) {

	log.Println("Starting scraping pipeline...")

	// Gather reviews from all sources
	reviews, err := scraperManager.ScrapeAll(ctx)
	if err != nil {
		log.Printf("Error during scraping: %v", err)
		return
	}

	log.Printf("Scraped %d reviews", len(reviews))

	for _, review := range reviews {
		// Analyze sentiment and intent
		analysisResult, err := analyzer.Analyze(ctx, review)
		if err != nil {
			log.Printf("Error analyzing review: %v", err)
			continue
		}

		// Skip if not negative or not relevant
		if !analysisResult.IsNegative || !analysisResult.IsRelevant {
			continue
		}

		// Route to appropriate department
		department := router.Route(analysisResult)

		// Send notification
		if err := notifier.Notify(ctx, department, review, analysisResult); err != nil {
			log.Printf("Error sending notification: %v", err)
		}
	}

	log.Println("Scraping pipeline completed")
}
