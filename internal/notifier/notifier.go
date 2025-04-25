package notifier

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"strings"
	"sync"
	"time"

	"github.com/Infoblox-CTO/review-scraper/internal/config"
	"github.com/Infoblox-CTO/review-scraper/pkg/models"
	"github.com/google/uuid"
)

// Notifier handles sending notifications about negative reviews to departments
type Notifier struct {
	config      config.NotifierConfig
	httpClient  *http.Client
	notifCache  map[string]models.Notification
	cacheMutex  sync.RWMutex
	db          *sql.DB
	dbConnected bool
}

// New creates a new notifier with the provided configuration
func New(cfg config.NotifierConfig) *Notifier {
	n := &Notifier{
		config:     cfg,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		notifCache: make(map[string]models.Notification),
	}

	// Connect to database if enabled
	if cfg.Databases.Enabled {
		n.connectToDatabase()
	}

	return n
}

// connectToDatabase establishes a connection to the configured database
func (n *Notifier) connectToDatabase() {
	// This is a placeholder - in a real implementation, this would establish a database connection
	n.dbConnected = false
}

// Notify sends a notification about a negative review to the appropriate department
func (n *Notifier) Notify(ctx context.Context, department models.Department, review models.Review, analysis models.AnalysisResult) error {
	// Create a notification object
	notification := models.Notification{
		ID:         uuid.New().String(),
		Review:     review,
		Analysis:   analysis,
		Department: department,
		SentAt:     time.Now(),
		Status:     "sent",
	}

	// Store notification in cache and/or database
	n.cacheNotification(notification)

	// Send notifications through all enabled channels
	var errs []error

	// Email notification
	if n.config.Email.Enabled {
		if err := n.sendEmailNotification(notification); err != nil {
			errs = append(errs, fmt.Errorf("email notification error: %w", err))
		}
	}

	// Slack notification
	if n.config.Slack.Enabled {
		if err := n.sendSlackNotification(notification); err != nil {
			errs = append(errs, fmt.Errorf("slack notification error: %w", err))
		}
	}

	// Update dashboard if enabled
	if n.config.Dashboard.Enabled {
		n.updateDashboard(notification)
	}

	// If there were errors, return a combined error
	if len(errs) > 0 {
		var errMsgs []string
		for _, err := range errs {
			errMsgs = append(errMsgs, err.Error())
		}
		return fmt.Errorf("notification errors: %s", strings.Join(errMsgs, "; "))
	}

	return nil
}

// cacheNotification stores a notification in the cache and/or database
func (n *Notifier) cacheNotification(notification models.Notification) {
	// Store in memory cache
	n.cacheMutex.Lock()
	n.notifCache[notification.ID] = notification
	n.cacheMutex.Unlock()

	// Store in database if connected
	if n.dbConnected {
		n.storeNotificationInDB(notification)
	}
}

// storeNotificationInDB saves a notification to the database
func (n *Notifier) storeNotificationInDB(notification models.Notification) {
	// This is a placeholder - in a real implementation, this would store data in a database
}

// sendEmailNotification sends an email notification about a review
func (n *Notifier) sendEmailNotification(notification models.Notification) error {
	// Skip if SMTP settings are not configured
	if n.config.Email.SMTPServer == "" || n.config.Email.SMTPPort == 0 {
		return fmt.Errorf("SMTP not configured")
	}

	// Get recipient email for department
	toEmail, exists := n.config.Email.DeptAddresses[notification.Department.ID]
	if !exists {
		toEmail = notification.Department.ContactInfo
	}

	if !strings.Contains(toEmail, "@") {
		return fmt.Errorf("invalid email address for department: %s", notification.Department.ID)
	}

	// Determine severity level for email subject
	severityLevel := "Medium"
	if notification.Analysis.SentimentScore < -0.7 {
		severityLevel = "High"
	} else if notification.Analysis.SentimentScore > -0.3 {
		severityLevel = "Low"
	}

	// Create email subject
	subject := fmt.Sprintf("[%s Priority] Negative Customer Feedback - %s",
		severityLevel, notification.Analysis.IntentCategory)

	// Format review creation time
	reviewTime := notification.Review.CreatedAt.Format("Jan 2, 2006 at 15:04")

	// Create email body
	body := fmt.Sprintf(`
Dear %s Team,

A negative customer review has been detected that requires your attention.

Review Details:
---------------
Source: %s
Time: %s
Author: %s
Content: "%s"
URL: %s
Sentiment Score: %.2f
Category: %s

Analysis:
---------
%s

Please review this feedback and take appropriate action.

This is an automated message from the Customer Feedback Analysis System.
`,
		notification.Department.Name,
		notification.Review.Source,
		reviewTime,
		notification.Review.Author,
		notification.Review.Content,
		notification.Review.URL,
		notification.Analysis.SentimentScore,
		notification.Analysis.IntentCategory,
		formatAnalysisDetails(notification.Analysis),
	)

	// Set up email authentication
	auth := smtp.PlainAuth("",
		n.config.Email.Username,
		n.config.Email.Password,
		n.config.Email.SMTPServer,
	)

	// Set up email headers
	headers := make(map[string]string)
	headers["From"] = n.config.Email.FromAddress
	headers["To"] = toEmail
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=\"utf-8\""

	// Construct message with headers
	message := ""
	for key, value := range headers {
		message += fmt.Sprintf("%s: %s\r\n", key, value)
	}
	message += "\r\n" + body

	// Send the email
	err := smtp.SendMail(
		fmt.Sprintf("%s:%d", n.config.Email.SMTPServer, n.config.Email.SMTPPort),
		auth,
		n.config.Email.FromAddress,
		[]string{toEmail},
		[]byte(message),
	)

	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Email notification sent to %s for review %s", toEmail, notification.Review.ID)
	return nil
}

// formatAnalysisDetails creates a formatted string with detailed analysis information
func formatAnalysisDetails(analysis models.AnalysisResult) string {
	var details strings.Builder

	details.WriteString("Keywords: ")
	if len(analysis.Keywords) > 0 {
		details.WriteString(strings.Join(analysis.Keywords, ", "))
	} else {
		details.WriteString("None detected")
	}
	details.WriteString("\n")

	details.WriteString("Entities: ")
	if len(analysis.Entities) > 0 {
		var entityTexts []string
		for _, entity := range analysis.Entities {
			entityTexts = append(entityTexts, fmt.Sprintf("%s (%s)", entity.Text, entity.Type))
		}
		details.WriteString(strings.Join(entityTexts, ", "))
	} else {
		details.WriteString("None detected")
	}
	details.WriteString("\n")

	details.WriteString("Confidence: ")
	details.WriteString(fmt.Sprintf("%.2f", analysis.Confidence))
	details.WriteString("\n")

	return details.String()
}

// SlackMessage represents a formatted Slack message
type SlackMessage struct {
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Attachment represents a Slack message attachment
type Attachment struct {
	Color      string  `json:"color"`
	Title      string  `json:"title"`
	TitleLink  string  `json:"title_link,omitempty"`
	Text       string  `json:"text"`
	Fields     []Field `json:"fields,omitempty"`
	Footer     string  `json:"footer,omitempty"`
	FooterIcon string  `json:"footer_icon,omitempty"`
	Timestamp  int64   `json:"ts,omitempty"`
}

// Field represents a field in a Slack attachment
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// sendSlackNotification sends a notification to Slack
func (n *Notifier) sendSlackNotification(notification models.Notification) error {
	// Skip if webhook is not configured
	if n.config.Slack.WebhookURL == "" {
		return fmt.Errorf("Slack webhook URL not configured")
	}

	// Get the channel for this department
	webhookURL := n.config.Slack.WebhookURL
	if channelURL, exists := n.config.Slack.DeptChannels[notification.Department.ID]; exists && channelURL != "" {
		webhookURL = channelURL
	}

	// Determine color based on sentiment (red for very negative, orange for somewhat negative)
	var color string
	if notification.Analysis.SentimentScore < -0.7 {
		color = "#FF0000" // Red
	} else if notification.Analysis.SentimentScore < -0.4 {
		color = "#FFA500" // Orange
	} else {
		color = "#FFCC00" // Yellow
	}

	// Create Slack message
	message := SlackMessage{
		Text: fmt.Sprintf("Negative Customer Feedback for %s Team", notification.Department.Name),
		Attachments: []Attachment{
			{
				Color:     color,
				Title:     fmt.Sprintf("Customer Review from %s", notification.Review.Source),
				TitleLink: notification.Review.URL,
				Text:      notification.Review.Content,
				Fields: []Field{
					{
						Title: "Author",
						Value: notification.Review.Author,
						Short: true,
					},
					{
						Title: "Sentiment",
						Value: fmt.Sprintf("%.2f", notification.Analysis.SentimentScore),
						Short: true,
					},
					{
						Title: "Category",
						Value: notification.Analysis.IntentCategory,
						Short: true,
					},
					{
						Title: "Confidence",
						Value: fmt.Sprintf("%.2f", notification.Analysis.Confidence),
						Short: true,
					},
					{
						Title: "Keywords",
						Value: strings.Join(notification.Analysis.Keywords, ", "),
						Short: false,
					},
				},
				Footer:    "Customer Feedback Analysis System",
				Timestamp: notification.Review.CreatedAt.Unix(),
			},
		},
	}

	// Convert message to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Slack message: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest(http.MethodPost, webhookURL, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := n.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack API returned non-OK status: %d", resp.StatusCode)
	}

	log.Printf("Slack notification sent to %s channel for review %s",
		notification.Department.Name, notification.Review.ID)
	return nil
}

// updateDashboard sends data to the dashboard system
func (n *Notifier) updateDashboard(notification models.Notification) {
	// This is a placeholder - in a real implementation, this would update a dashboard system
	// Could publish to a message queue, update a database, or notify a websocket server
	log.Printf("Dashboard updated with notification %s", notification.ID)
}

// GetStats returns statistics about the notifier
func (n *Notifier) GetStats() map[string]interface{} {
	n.cacheMutex.RLock()
	defer n.cacheMutex.RUnlock()

	stats := map[string]interface{}{
		"notifications_cached": len(n.notifCache),
		"email_enabled":        n.config.Email.Enabled,
		"slack_enabled":        n.config.Slack.Enabled,
		"dashboard_enabled":    n.config.Dashboard.Enabled,
		"database_enabled":     n.config.Databases.Enabled,
		"database_connected":   n.dbConnected,
	}

	return stats
}
