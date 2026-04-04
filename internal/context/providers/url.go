// Package providers provides URL-based context
package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// URLProvider provides context from URLs
type URLProvider struct {
	client  *http.Client
	logger  *logrus.Logger
	maxSize int64
	timeout time.Duration
}

// NewURLProvider creates a new URL provider
func NewURLProvider() *URLProvider {
	return &URLProvider{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger:  logrus.New(),
		maxSize: 1024 * 1024, // 1MB
		timeout: 30 * time.Second,
	}
}

// Name returns the provider name
func (u *URLProvider) Name() string {
	return "url"
}

// Description returns the provider description
func (u *URLProvider) Description() string {
	return "Provides context from URLs"
}

// Resolve resolves URL context
func (u *URLProvider) Resolve(ctx context.Context, query string) ([]ContextItem, error) {
	// Validate URL
	parsedURL, err := url.Parse(query)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	
	// Only allow http/https
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("unsupported URL scheme: %s", parsedURL.Scheme)
	}
	
	// Fetch content
	item, err := u.fetch(ctx, parsedURL.String())
	if err != nil {
		return nil, err
	}
	
	return []ContextItem{*item}, nil
}

// fetch fetches content from a URL
func (u *URLProvider) fetch(ctx context.Context, urlStr string) (*ContextItem, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return nil, err
	}
	
	// Set headers
	req.Header.Set("User-Agent", "HelixAgent/1.0")
	
	resp, err := u.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !isSupportedContentType(contentType) {
		return nil, fmt.Errorf("unsupported content type: %s", contentType)
	}
	
	// Read content with size limit
	reader := io.LimitReader(resp.Body, u.maxSize)
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("read content: %w", err)
	}
	
	// Check if we hit the limit
	if int64(len(content)) >= u.maxSize {
		return &ContextItem{
			Name:        u.extractName(urlStr),
			Description: fmt.Sprintf("Content truncated (exceeds %d bytes)", u.maxSize),
			Content:     string(content),
			Source:      urlStr,
			Timestamp:   time.Now(),
		}, nil
	}
	
	// Clean content if HTML
	contentStr := string(content)
	if strings.Contains(contentType, "text/html") {
		contentStr = stripHTML(contentStr)
	}
	
	return &ContextItem{
		Name:        u.extractName(urlStr),
		Description: fmt.Sprintf("%d bytes, %s", len(content), contentType),
		Content:     contentStr,
		Source:      urlStr,
		Timestamp:   time.Now(),
	}, nil
}

// extractName extracts a name from URL
func (u *URLProvider) extractName(urlStr string) string {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}
	
	// Use last path component or host
	path := strings.Trim(parsed.Path, "/")
	if path == "" {
		return parsed.Host
	}
	
	parts := strings.Split(path, "/")
	return parts[len(parts)-1]
}

// isSupportedContentType checks if content type is supported
func isSupportedContentType(contentType string) bool {
	supported := []string{
		"text/",
		"application/json",
		"application/xml",
		"application/javascript",
		"application/typescript",
	}
	
	for _, prefix := range supported {
		if strings.Contains(contentType, prefix) {
			return true
		}
	}
	
	return false
}

// stripHTML removes HTML tags (basic)
func stripHTML(html string) string {
	var result strings.Builder
	inTag := false
	
	for _, r := range html {
		switch r {
		case '<':
			inTag = true
		case '>':
			inTag = false
		default:
			if !inTag {
				result.WriteRune(r)
			}
		}
	}
	
	return strings.TrimSpace(result.String())
}

// WithTimeout sets the request timeout
func (u *URLProvider) WithTimeout(timeout time.Duration) *URLProvider {
	u.timeout = timeout
	u.client.Timeout = timeout
	return u
}

// WithMaxSize sets the max content size
func (u *URLProvider) WithMaxSize(size int64) *URLProvider {
	u.maxSize = size
	return u
}

// Fetch fetches content from a URL (public method)
func (u *URLProvider) Fetch(ctx context.Context, urlStr string) (*ContextItem, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}
	
	return u.fetch(ctx, parsedURL.String())
}

// IsValidURL checks if a string is a valid URL
func IsValidURL(str string) bool {
	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	
	return u.Scheme == "http" || u.Scheme == "https"
}

// ExtractURLs extracts URLs from text
func ExtractURLs(text string) []string {
	var urls []string
	words := strings.Fields(text)
	
	for _, word := range words {
		if IsValidURL(word) {
			urls = append(urls, word)
		}
	}
	
	return urls
}
