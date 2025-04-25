# Review Scraper

A complete system for monitoring and scraping user reviews and social media comments related to products/services, identifying negative customer experiences, and routing them to appropriate internal support departments based on content classification.

## Features

- **Social Listening & Data Collection**
  - Scrape reviews and comments from Twitter, Reddit, App Store, Google Play, and custom websites
  - Scheduled periodic scraping with configurable intervals
  - Anti-ban mechanisms including rate limiting, proxy rotation, and user-agent randomization
  - Filtering to focus on brand-relevant content

- **Sentiment and Intent Analysis**
  - Multiple analysis modes (local, OpenAI, Google, AWS, Azure)
  - Negative sentiment detection with configurable thresholds
  - Issue classification (bug reports, feature requests, performance issues, etc.)
  - Keyword and entity extraction

- **Department Routing**
  - Intelligent routing based on issue classification
  - Customizable department mappings
  - Priority-based assignment

- **Notification System**
  - Email notifications with detailed analysis
  - Slack integration with formatted messages
  - Dashboard updates (optional)
  - Database storage (optional)

- **REST API**
  - Comprehensive endpoints for system management and monitoring
  - Authentication and rate limiting
  - Swagger documentation

## Architecture

The system follows a modular architecture with the following components:

- **Scraper**: Collects reviews from multiple sources
- **Analyzer**: Processes reviews to determine sentiment and intent
- **Router**: Routes negative reviews to appropriate departments
- **Notifier**: Sends notifications through configured channels
- **API Server**: Provides REST API endpoints for monitoring and management

## Requirements

- Go 1.16 or higher
- Access tokens for platforms you want to scrape (Twitter, Reddit, etc.)
- SMTP server for email notifications (optional)
- Slack webhooks for Slack notifications (optional)
- Database for persistent storage (optional)

## Installation

1. Clone the repository:

```
git clone https://github.com/Infoblox-CTO/review-scraper.git
cd review-scraper
```

2. Install dependencies:

```
go mod download
```

3. Copy the sample configuration file and edit it:

```
cp configs/config.sample.json configs/config.json
```

4. Edit the configuration file with your API keys, tokens, and settings.

5. Build the application:

```
go build -o review-scraper ./cmd/review-scraper
```

## Configuration

The system is configured through a JSON file located at `configs/config.json`. You can specify a different configuration file using the `REVIEW_SCRAPER_CONFIG` environment variable.

Key configuration sections:

- **Scrapers**: Configure data sources (Twitter, Reddit, etc.)
- **Analyzer**: Configure sentiment analysis and intent classification
- **Router**: Configure department mappings and routing rules
- **Notifier**: Configure notification channels (email, Slack, etc.)
- **API**: Configure REST API settings

See `configs/config.sample.json` for a complete configuration example with comments.

## Usage

### Running the Service

Run the service with:

```
./review-scraper
```

The service will:
- Start scraping reviews based on the configured interval
- Process reviews to detect sentiment and intent
- Route negative reviews to appropriate departments
- Send notifications through configured channels
- Start the REST API server

### REST API Endpoints

The API server provides the following endpoints:

#### Health Check

- `GET /api/v1/health`: Check service health

#### Reviews

- `GET /api/v1/reviews`: Get recent reviews
- `GET /api/v1/reviews/{id}`: Get a specific review

#### Departments

- `GET /api/v1/departments`: Get all departments
- `GET /api/v1/departments/{id}`: Get a specific department
- `GET /api/v1/departments/{id}/reviews`: Get reviews for a specific department

#### Dashboard

- `GET /api/v1/dashboard/metrics`: Get dashboard metrics
- `GET /api/v1/dashboard/stats`: Get system statistics

#### Scraping

- `POST /api/v1/scraping/run`: Manually trigger scraping
- `GET /api/v1/scraping/stats`: Get scraping statistics

#### Analysis

- `POST /api/v1/analyze`: Analyze custom text

#### Configuration

- `GET /api/v1/config/{component}`: Get configuration for a component
- `PUT /api/v1/config/{component}`: Update configuration for a component

### Authentication

API endpoints are secured with token authentication. Include the token in the `Authorization` header:

```
Authorization: Bearer YOUR_API_AUTH_TOKEN
```

Or as a query parameter:

```
?token=YOUR_API_AUTH_TOKEN
```

## Extending the System

### Adding a New Scraper

1. Create a new file in `internal/scraper/` for your scraper
2. Implement the `Scraper` interface defined in `internal/scraper/scraper.go`
3. Add your scraper to the `initializeScrapers` method in `internal/scraper/scraper.go`
4. Update the configuration structure in `internal/config/config.go`

### Adding a New Analysis Method

1. Add your analysis method to the `Analyze` function in `internal/analyzer/analyzer.go`
2. Add a new function for your analysis method (e.g., `analyzeWithNewMethod`)
3. Update the configuration structure in `internal/config/config.go`

### Adding a New Notification Channel

1. Update the `Notify` function in `internal/notifier/notifier.go`
2. Add a new function for your notification channel (e.g., `sendNewChannelNotification`)
3. Update the configuration structure in `internal/config/config.go`

## License

Copyright (c) 2025 Infoblox CTO Team. All rights reserved.
