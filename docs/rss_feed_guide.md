# RSS Feed Integration Guide

This document provides detailed information on configuring and using the RSS feed integration in the TaiPhanVan Blog Backend.

## Overview

The RSS feed integration allows the blog to automatically fetch and display news articles from multiple RSS feeds. This functionality works alongside the existing NewsAPI integration, giving you multiple sources for news content.

## Configuration

### Environment Variables

Configure RSS feeds through the following environment variables:

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `RSS_FEEDS` | Comma-separated list of RSS feeds in the format `NAME=URL=CATEGORY` | empty | `TechCrunch=https://techcrunch.com/feed/=technology,TheVerge=https://www.theverge.com/rss/index.xml=science` |
| `RSS_DEFAULT_LIMIT` | Default number of items to fetch per feed | 10 | `15` |
| `RSS_FETCH_INTERVAL` | How often to automatically fetch from feeds | 1h | `30m` |
| `RSS_ENABLE_AUTO_FETCH` | Whether to automatically fetch in the background | false | `true` |

### RSS Feed Format

Each RSS feed in the `RSS_FEEDS` environment variable must follow this format:

```bash
NAME=URL=CATEGORY
```

Where:

- `NAME`: A friendly name for the feed (e.g., "TechCrunch")
- `URL`: The full URL to the RSS feed (e.g., "<https://techcrunch.com/feed/>")
- `CATEGORY`: The category to assign to articles from this feed (e.g., "technology")

Multiple feeds must be separated by commas.

### Category Mapping

The system maps RSS feed categories to predefined categories in the application:

- "technology" or "tech" → Technology category
- "science" → Science category
- Any other value will be used as-is

If no category is specified for a feed, "technology" will be used as the default.

## Usage

### Automatic Fetching

When `RSS_ENABLE_AUTO_FETCH` is set to `true`, the application will:

1. Fetch articles from all configured RSS feeds immediately when the server starts
2. Continue to fetch at the interval specified by `RSS_FETCH_INTERVAL`

### Manual Fetching

Administrators can manually trigger RSS feed fetching using the admin API endpoint:

```bash
POST /api/admin/news/fetch-rss
```

Request body:

```json
{
  "limit": 15  // Optional: Number of articles to fetch per feed
}
```

The endpoint requires administrator authentication.

### Convenience Script

A shell script is provided to manually fetch RSS feeds:

```bash
./scripts/fetch_rss.sh
```

This script authenticates as an admin user and triggers the RSS feed fetch endpoint.

## How It Works

1. The system reads the configuration for RSS feeds from environment variables
2. For each feed, it fetches the RSS content and processes items into news articles
3. Each article is assigned:
   - A title from the RSS item title
   - Content from the RSS item content or description
   - A unique slug generated from the title
   - A category based on the feed configuration
   - A source name from the feed name
   - An external ID to prevent duplicates

4. Articles are saved to the database, skipping any that already exist

## Testing

Test the RSS feed integration using the provided script:

```bash
./scripts/api_tests/test_rss_feed_api.sh
```

This script:

1. Authenticates as an admin
2. Triggers the RSS feed fetch endpoint
3. Verifies that articles were properly imported and categorized

## Troubleshooting

Common issues:

1. **No feeds configured**: Ensure `RSS_FEEDS` environment variable is properly set with at least one feed
2. **Failed to fetch**: Check feed URLs are correct and publicly accessible
3. **No articles saved**: Articles may already exist in the database (they're identified by external ID)
4. **Category issues**: Verify category mapping is working correctly

Check server logs for detailed error messages related to RSS fetching.
