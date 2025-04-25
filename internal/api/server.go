package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Infoblox-CTO/review-scraper/internal/analyzer"
	"github.com/Infoblox-CTO/review-scraper/internal/config"
	"github.com/Infoblox-CTO/review-scraper/internal/notifier"
	"github.com/Infoblox-CTO/review-scraper/internal/router"
	"github.com/Infoblox-CTO/review-scraper/internal/scraper"
	"github.com/Infoblox-CTO/review-scraper/pkg/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Server represents the API server
type Server struct {
	config         config.APIConfig
	router         *chi.Mux
	httpServer     *http.Server
	scraperManager *scraper.Manager
	analyzer       *analyzer.Analyzer
	deptRouter     *router.Router
	notifier       *notifier.Notifier
	recentReviews  []models.Review
	reviewsMutex   sync.RWMutex
}

// NewServer creates a new API server
func NewServer(cfg config.APIConfig, scraperManager *scraper.Manager,
	analyzer *analyzer.Analyzer, deptRouter *router.Router, notifier *notifier.Notifier) *Server {

	s := &Server{
		config:         cfg,
		scraperManager: scraperManager,
		analyzer:       analyzer,
		deptRouter:     deptRouter,
		notifier:       notifier,
		recentReviews:  make([]models.Review, 0, 100), // Keep last 100 reviews
		reviewsMutex:   sync.RWMutex{},
	}

	s.setupRouter()

	return s
}

// setupRouter configures the HTTP router
func (s *Server) setupRouter() {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Authentication middleware
	r.Use(s.authMiddleware)

	// API routes
	r.Route("/api/v1", func(r chi.Router) {
		// Health check endpoint
		r.Get("/health", s.handleHealthCheck)

		// Reviews endpoints
		r.Route("/reviews", func(r chi.Router) {
			r.Get("/", s.handleGetReviews)
			r.Get("/{id}", s.handleGetReview)
		})

		// Departments endpoints
		r.Route("/departments", func(r chi.Router) {
			r.Get("/", s.handleGetDepartments)
			r.Get("/{id}", s.handleGetDepartment)
			r.Get("/{id}/reviews", s.handleGetDepartmentReviews)
		})

		// Dashboard endpoints
		r.Route("/dashboard", func(r chi.Router) {
			r.Get("/metrics", s.handleGetDashboardMetrics)
			r.Get("/stats", s.handleGetSystemStats)
		})

		// Scraping endpoints
		r.Route("/scraping", func(r chi.Router) {
			r.Post("/run", s.handleRunScraping)
			r.Get("/stats", s.handleGetScrapingStats)
		})

		// Analysis endpoints
		r.Route("/analyze", func(r chi.Router) {
			r.Post("/", s.handleAnalyzeText)
		})

		// Config endpoints
		r.Route("/config", func(r chi.Router) {
			r.Get("/{component}", s.handleGetConfig)
			r.Put("/{component}", s.handleUpdateConfig)
		})
	})

	// Swagger UI if enabled
	if s.config.EnableSwagger {
		r.Get("/swagger/*", s.handleSwagger)
	}

	s.router = r
}

// Start starts the API server
func (s *Server) Start() error {
	serverAddr := fmt.Sprintf(":%d", s.config.Port)

	s.httpServer = &http.Server{
		Addr:    serverAddr,
		Handler: s.router,
	}

	log.Printf("Starting API server at %s", serverAddr)

	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the API server
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Stopping API server...")
	return s.httpServer.Shutdown(ctx)
}

// AddRecentReview adds a review to the recent reviews list
func (s *Server) AddRecentReview(review models.Review) {
	s.reviewsMutex.Lock()
	defer s.reviewsMutex.Unlock()

	// Add to front of slice
	s.recentReviews = append([]models.Review{review}, s.recentReviews...)

	// Keep only the last 100 reviews
	if len(s.recentReviews) > 100 {
		s.recentReviews = s.recentReviews[:100]
	}
}

// authMiddleware handles authentication
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for health check and swagger
		if r.URL.Path == "/api/v1/health" || strings.HasPrefix(r.URL.Path, "/swagger") {
			next.ServeHTTP(w, r)
			return
		}

		// Check for token authentication
		token := r.Header.Get("Authorization")
		if token == "" {
			// Also check query param for token (useful for dashboard iframe embedding)
			token = r.URL.Query().Get("token")
		} else {
			// Remove "Bearer " prefix if present
			token = strings.TrimPrefix(token, "Bearer ")
		}

		if token == "" || token != s.config.AuthToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// respond sends a JSON response
func (s *Server) respond(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}
}

// respondError sends an error response
func (s *Server) respondError(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	response := models.APIResponse{
		Success: false,
		Error:   message,
	}
	s.respond(w, r, statusCode, response)
}

// Route handlers

// handleHealthCheck handles the health check endpoint
func (s *Server) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	s.respond(w, r, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Service is healthy",
		Data: map[string]interface{}{
			"version":   "1.0.0",
			"timestamp": time.Now(),
		},
	})
}

// handleGetReviews gets recent reviews
func (s *Server) handleGetReviews(w http.ResponseWriter, r *http.Request) {
	s.reviewsMutex.RLock()
	defer s.reviewsMutex.RUnlock()

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default limit

	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	// Apply limit
	reviews := s.recentReviews
	if len(reviews) > limit {
		reviews = reviews[:limit]
	}

	s.respond(w, r, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    reviews,
	})
}

// handleGetReview gets a specific review
func (s *Server) handleGetReview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	s.reviewsMutex.RLock()
	defer s.reviewsMutex.RUnlock()

	for _, review := range s.recentReviews {
		if review.ID == id {
			s.respond(w, r, http.StatusOK, models.APIResponse{
				Success: true,
				Data:    review,
			})
			return
		}
	}

	s.respondError(w, r, http.StatusNotFound, "Review not found")
}

// handleGetDepartments gets all departments
func (s *Server) handleGetDepartments(w http.ResponseWriter, r *http.Request) {
	departments := s.deptRouter.GetAllDepartments()

	s.respond(w, r, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    departments,
	})
}

// handleGetDepartment gets a specific department
func (s *Server) handleGetDepartment(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	dept, exists := s.deptRouter.GetDepartment(id)
	if !exists {
		s.respondError(w, r, http.StatusNotFound, "Department not found")
		return
	}

	s.respond(w, r, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    dept,
	})
}

// handleGetDepartmentReviews gets reviews for a specific department
func (s *Server) handleGetDepartmentReviews(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder - in a real implementation, this would retrieve reviews from a database
	s.respondError(w, r, http.StatusNotImplemented, "Feature not yet implemented")
}

// handleGetDashboardMetrics gets metrics for the dashboard
func (s *Server) handleGetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	// This is a placeholder - in a real implementation, this would generate actual metrics
	s.respond(w, r, http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.DashboardMetrics{
			TotalReviews:     100,
			NegativeReviews:  25,
			AverageSentiment: -0.3,
			TopCategories: map[string]int{
				"bug_report":       10,
				"customer_service": 8,
				"performance":      5,
			},
			DepartmentWorkloads: map[string]int{
				"engineering": 15,
				"support":     12,
				"operations":  5,
			},
			SourceBreakdown: map[string]int{
				"twitter":    40,
				"reddit":     30,
				"app_store":  20,
				"play_store": 10,
			},
		},
	})
}

// handleGetSystemStats gets system statistics
func (s *Server) handleGetSystemStats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"scraper":  s.scraperManager.GetStats(),
		"analyzer": s.analyzer.GetStats(),
		"notifier": s.notifier.GetStats(),
	}

	s.respond(w, r, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    stats,
	})
}

// handleRunScraping triggers a scraping run
func (s *Server) handleRunScraping(w http.ResponseWriter, r *http.Request) {
	// Start scraping in background
	go func() {
		ctx := context.Background()
		reviews, err := s.scraperManager.ScrapeAll(ctx)
		if err != nil {
			log.Printf("Error during manual scraping: %v", err)
			return
		}

		log.Printf("Manual scraping completed: %d reviews found", len(reviews))

		// Process the reviews
		for _, review := range reviews {
			// Add to recent reviews
			s.AddRecentReview(review)

			// Analyze sentiment and intent
			analysisResult, err := s.analyzer.Analyze(ctx, review)
			if err != nil {
				log.Printf("Error analyzing review: %v", err)
				continue
			}

			// Skip if not negative or not relevant
			if !analysisResult.IsNegative || !analysisResult.IsRelevant {
				continue
			}

			// Route to appropriate department
			department := s.deptRouter.Route(analysisResult)

			// Send notification
			if err := s.notifier.Notify(ctx, department, review, analysisResult); err != nil {
				log.Printf("Error sending notification: %v", err)
			}
		}
	}()

	s.respond(w, r, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Scraping job started",
	})
}

// handleGetScrapingStats gets scraping statistics
func (s *Server) handleGetScrapingStats(w http.ResponseWriter, r *http.Request) {
	stats := s.scraperManager.GetStats()

	s.respond(w, r, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    stats,
	})
}

// handleAnalyzeText analyzes custom text
type AnalyzeRequest struct {
	Text   string `json:"text"`
	Source string `json:"source"`
	Author string `json:"author"`
}

func (s *Server) handleAnalyzeText(w http.ResponseWriter, r *http.Request) {
	var req AnalyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, r, http.StatusBadRequest, "Invalid request format")
		return
	}

	if req.Text == "" {
		s.respondError(w, r, http.StatusBadRequest, "Text is required")
		return
	}

	// Create a review from the request
	review := models.Review{
		ID:          fmt.Sprintf("manual-%d", time.Now().UnixNano()),
		Source:      req.Source,
		SourceID:    fmt.Sprintf("manual-%d", time.Now().UnixNano()),
		Content:     req.Text,
		Author:      req.Author,
		CreatedAt:   time.Now(),
		RetrievedAt: time.Now(),
	}

	// Analyze the review
	analysisResult, err := s.analyzer.Analyze(r.Context(), review)
	if err != nil {
		s.respondError(w, r, http.StatusInternalServerError, fmt.Sprintf("Analysis error: %v", err))
		return
	}

	// Return the analysis result
	s.respond(w, r, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    analysisResult,
	})
}

// handleGetConfig gets configuration for a component
func (s *Server) handleGetConfig(w http.ResponseWriter, r *http.Request) {
	// This would typically retrieve redacted configuration (no secrets)
	s.respondError(w, r, http.StatusNotImplemented, "Feature not yet implemented")
}

// handleUpdateConfig updates configuration for a component
func (s *Server) handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	// This would typically update configuration elements
	s.respondError(w, r, http.StatusNotImplemented, "Feature not yet implemented")
}

// handleSwagger serves Swagger UI documentation
func (s *Server) handleSwagger(w http.ResponseWriter, r *http.Request) {
	// This would serve Swagger UI files in a real implementation
	s.respondError(w, r, http.StatusNotImplemented, "Swagger UI not yet implemented")
}
