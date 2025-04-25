package router

import (
	"sort"
	"sync"

	"github.com/Infoblox-CTO/review-scraper/internal/config"
	"github.com/Infoblox-CTO/review-scraper/pkg/models"
)

// Router routes analyzed reviews to the appropriate department
type Router struct {
	config       config.RouterConfig
	departments  map[string]models.Department
	mappingCache map[string]string
	mu           sync.RWMutex
}

// New creates a new department router with the provided configuration
func New(cfg config.RouterConfig) *Router {
	// Initialize the departments map with default departments
	departments := make(map[string]models.Department)

	// Default mappings if none are provided in config
	if len(cfg.Mappings) == 0 {
		defaultMappings := []config.DepartmentMapping{
			// Infoblox-specific mappings
			{Category: "product_issue", Department: "engineering", Priority: 10},
			{Category: "technical_support", Department: "support", Priority: 8},
			{Category: "performance", Department: "engineering", Priority: 8},
			{Category: "security", Department: "security", Priority: 10},
			{Category: "feature_request", Department: "product", Priority: 5},
			{Category: "ui_ux", Department: "design", Priority: 7},
			{Category: "billing_licensing", Department: "finance", Priority: 9},
			{Category: "documentation", Department: "documentation", Priority: 6},
			{Category: "deployment", Department: "professional_services", Priority: 8},
			{Category: "cloud_integration", Department: "cloud_team", Priority: 8},
			{Category: "automation", Department: "automation_team", Priority: 7},
			{Category: "upgrade_issue", Department: "engineering", Priority: 9},
			{Category: "general_complaint", Department: "support", Priority: 6},
		}
		cfg.Mappings = defaultMappings
	}

	// Create a cache for quick lookups
	mappingCache := make(map[string]string)
	for _, mapping := range cfg.Mappings {
		mappingCache[mapping.Category] = mapping.Department
	}

	// Define Infoblox-specific departments with contact info
	departments["engineering"] = models.Department{
		ID:          "engineering",
		Name:        "Infoblox Engineering",
		ContactInfo: "engineering@infoblox.com",
		Categories:  []string{"product_issue", "performance", "upgrade_issue"},
	}

	departments["security"] = models.Department{
		ID:          "security",
		Name:        "Security Team",
		ContactInfo: "security@infoblox.com",
		Categories:  []string{"security"},
	}

	departments["product"] = models.Department{
		ID:          "product",
		Name:        "Product Management",
		ContactInfo: "product@infoblox.com",
		Categories:  []string{"feature_request"},
	}

	departments["design"] = models.Department{
		ID:          "design",
		Name:        "UX Design",
		ContactInfo: "design@infoblox.com",
		Categories:  []string{"ui_ux"},
	}

	departments["finance"] = models.Department{
		ID:          "finance",
		Name:        "Customer Success & Licensing",
		ContactInfo: "licensing@infoblox.com",
		Categories:  []string{"billing_licensing"},
	}

	departments["support"] = models.Department{
		ID:          "support",
		Name:        "Technical Support",
		ContactInfo: "support@infoblox.com",
		Categories:  []string{"technical_support", "general_complaint"},
	}

	departments["documentation"] = models.Department{
		ID:          "documentation",
		Name:        "Documentation Team",
		ContactInfo: "docs@infoblox.com",
		Categories:  []string{"documentation"},
	}

	departments["professional_services"] = models.Department{
		ID:          "professional_services",
		Name:        "Professional Services",
		ContactInfo: "services@infoblox.com",
		Categories:  []string{"deployment"},
	}

	departments["cloud_team"] = models.Department{
		ID:          "cloud_team",
		Name:        "BloxOne Cloud Team",
		ContactInfo: "cloud@infoblox.com",
		Categories:  []string{"cloud_integration"},
	}

	departments["automation_team"] = models.Department{
		ID:          "automation_team",
		Name:        "Network Automation Team",
		ContactInfo: "automation@infoblox.com",
		Categories:  []string{"automation"},
	}

	return &Router{
		config:       cfg,
		departments:  departments,
		mappingCache: mappingCache,
		mu:           sync.RWMutex{},
	}
}

// Route determines the appropriate department for a review based on analysis
func (r *Router) Route(analysis models.AnalysisResult) models.Department {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// If we have an explicit category-to-department mapping, use that first
	if departmentID, exists := r.mappingCache[analysis.IntentCategory]; exists {
		if dept, exists := r.departments[departmentID]; exists {
			return dept
		}
	}

	// If no direct mapping exists, use the categoryScores to find the best department
	type categoryScore struct {
		category string
		score    float64
	}

	var scores []categoryScore
	for category, score := range analysis.CategoryScores {
		scores = append(scores, categoryScore{category, score})
	}

	// Sort by score in descending order
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Try to find a department for each category, starting with highest score
	for _, cs := range scores {
		if departmentID, exists := r.mappingCache[cs.category]; exists {
			if dept, exists := r.departments[departmentID]; exists {
				return dept
			}
		}
	}

	// If no mapping found, return the default department
	if dept, exists := r.departments[r.config.DefaultDepartment]; exists {
		return dept
	}

	// Fallback to support if default department doesn't exist
	if dept, exists := r.departments["support"]; exists {
		return dept
	}

	// Ultimate fallback
	return models.Department{
		ID:          "support",
		Name:        "Customer Support",
		ContactInfo: "support@company.com",
		Categories:  []string{"general_complaint"},
	}
}

// GetDepartment returns a department by ID
func (r *Router) GetDepartment(departmentID string) (models.Department, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	dept, exists := r.departments[departmentID]
	return dept, exists
}

// GetAllDepartments returns all departments
func (r *Router) GetAllDepartments() []models.Department {
	r.mu.RLock()
	defer r.mu.RUnlock()

	departments := make([]models.Department, 0, len(r.departments))
	for _, dept := range r.departments {
		departments = append(departments, dept)
	}
	return departments
}

// AddDepartment adds a new department
func (r *Router) AddDepartment(department models.Department) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.departments[department.ID] = department

	// Update mappings for the department's categories
	for _, category := range department.Categories {
		r.mappingCache[category] = department.ID
	}
}

// UpdateMapping updates a category-to-department mapping
func (r *Router) UpdateMapping(category, departmentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.mappingCache[category] = departmentID
}
