// Package servers provides MCP server adapters for various services.
package servers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// GitHubAdapterConfig holds configuration for GitHub MCP adapter
type GitHubAdapterConfig struct {
	// Token is the GitHub personal access token
	Token string `json:"token,omitempty"`
	// BaseURL is the GitHub API base URL (for Enterprise)
	BaseURL string `json:"base_url,omitempty"`
	// Timeout is the request timeout
	Timeout time.Duration `json:"timeout,omitempty"`
	// RateLimit enables rate limit handling
	RateLimit bool `json:"rate_limit"`
	// MaxRetries is the maximum number of retries
	MaxRetries int `json:"max_retries,omitempty"`
}

// DefaultGitHubAdapterConfig returns default configuration
func DefaultGitHubAdapterConfig() GitHubAdapterConfig {
	return GitHubAdapterConfig{
		BaseURL:    "https://api.github.com",
		Timeout:    30 * time.Second,
		RateLimit:  true,
		MaxRetries: 3,
	}
}

// GitHubAdapter implements MCP adapter for GitHub operations
type GitHubAdapter struct {
	config      GitHubAdapterConfig
	client      *http.Client
	initialized bool
	mu          sync.RWMutex
	logger      *logrus.Logger
}

// NewGitHubAdapter creates a new GitHub MCP adapter
func NewGitHubAdapter(config GitHubAdapterConfig, logger *logrus.Logger) *GitHubAdapter {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	if config.BaseURL == "" {
		config.BaseURL = "https://api.github.com"
	}
	if config.Timeout <= 0 {
		config.Timeout = 30 * time.Second
	}
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}

	return &GitHubAdapter{
		config: config,
		logger: logger,
	}
}

// Initialize initializes the GitHub adapter
func (g *GitHubAdapter) Initialize(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.client = &http.Client{
		Timeout: g.config.Timeout,
	}

	// Test authentication if token is provided
	if g.config.Token != "" {
		req, err := http.NewRequestWithContext(ctx, "GET", g.config.BaseURL+"/user", nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+g.config.Token)
		req.Header.Set("Accept", "application/vnd.github+json")

		resp, err := g.client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			return fmt.Errorf("invalid GitHub token")
		}
	}

	g.initialized = true
	g.logger.Info("GitHub adapter initialized")
	return nil
}

// Health returns health status
func (g *GitHubAdapter) Health(ctx context.Context) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return fmt.Errorf("GitHub adapter not initialized")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", g.config.BaseURL+"/rate_limit", nil)
	if err != nil {
		return err
	}
	if g.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.config.Token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := g.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API health check failed: %d", resp.StatusCode)
	}

	return nil
}

// Close closes the adapter
func (g *GitHubAdapter) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.initialized = false
	return nil
}

// doRequest performs an HTTP request to the GitHub API
func (g *GitHubAdapter) doRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	url := g.config.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	if g.config.Token != "" {
		req.Header.Set("Authorization", "Bearer "+g.config.Token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	return g.client.Do(req)
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Name      string `json:"name,omitempty"`
	Email     string `json:"email,omitempty"`
	Bio       string `json:"bio,omitempty"`
	Company   string `json:"company,omitempty"`
	Location  string `json:"location,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	HTMLURL   string `json:"html_url,omitempty"`
	Followers int    `json:"followers,omitempty"`
	Following int    `json:"following,omitempty"`
	PublicRepos int  `json:"public_repos,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// GetUser gets user information
func (g *GitHubAdapter) GetUser(ctx context.Context, username string) (*GitHubUser, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	path := "/user"
	if username != "" {
		path = "/users/" + url.PathEscape(username)
	}

	resp, err := g.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("user not found: %s", username)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user: status %d", resp.StatusCode)
	}

	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

// GitHubRepository represents a GitHub repository
type GitHubRepository struct {
	ID              int64  `json:"id"`
	Name            string `json:"name"`
	FullName        string `json:"full_name"`
	Description     string `json:"description,omitempty"`
	Private         bool   `json:"private"`
	Fork            bool   `json:"fork"`
	HTMLURL         string `json:"html_url"`
	CloneURL        string `json:"clone_url"`
	Language        string `json:"language,omitempty"`
	DefaultBranch   string `json:"default_branch"`
	StargazersCount int    `json:"stargazers_count"`
	ForksCount      int    `json:"forks_count"`
	OpenIssuesCount int    `json:"open_issues_count"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
	PushedAt        string `json:"pushed_at,omitempty"`
}

// ListRepositories lists repositories for a user or organization
func (g *GitHubAdapter) ListRepositories(ctx context.Context, owner string, repoType string) ([]GitHubRepository, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	path := "/user/repos"
	if owner != "" {
		path = "/users/" + url.PathEscape(owner) + "/repos"
	}
	if repoType != "" {
		path += "?type=" + url.QueryEscape(repoType)
	}

	resp, err := g.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list repositories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list repositories: status %d", resp.StatusCode)
	}

	var repos []GitHubRepository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, fmt.Errorf("failed to decode repositories: %w", err)
	}

	return repos, nil
}

// GetRepository gets repository details
func (g *GitHubAdapter) GetRepository(ctx context.Context, owner, repo string) (*GitHubRepository, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	path := "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repo)
	resp, err := g.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repository not found: %s/%s", owner, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get repository: status %d", resp.StatusCode)
	}

	var repository GitHubRepository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, fmt.Errorf("failed to decode repository: %w", err)
	}

	return &repository, nil
}

// GitHubIssue represents a GitHub issue
type GitHubIssue struct {
	ID        int64        `json:"id"`
	Number    int          `json:"number"`
	Title     string       `json:"title"`
	State     string       `json:"state"`
	Body      string       `json:"body,omitempty"`
	HTMLURL   string       `json:"html_url"`
	User      *GitHubUser  `json:"user,omitempty"`
	Labels    []GitHubLabel `json:"labels,omitempty"`
	Assignees []GitHubUser `json:"assignees,omitempty"`
	CreatedAt string       `json:"created_at"`
	UpdatedAt string       `json:"updated_at"`
	ClosedAt  string       `json:"closed_at,omitempty"`
}

// GitHubLabel represents a GitHub label
type GitHubLabel struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// ListIssues lists issues for a repository
func (g *GitHubAdapter) ListIssues(ctx context.Context, owner, repo, state string) ([]GitHubIssue, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	path := "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repo) + "/issues"
	if state != "" {
		path += "?state=" + url.QueryEscape(state)
	}

	resp, err := g.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list issues: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list issues: status %d", resp.StatusCode)
	}

	var issues []GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, fmt.Errorf("failed to decode issues: %w", err)
	}

	return issues, nil
}

// GetIssue gets a specific issue
func (g *GitHubAdapter) GetIssue(ctx context.Context, owner, repo string, number int) (*GitHubIssue, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	path := fmt.Sprintf("/repos/%s/%s/issues/%d", url.PathEscape(owner), url.PathEscape(repo), number)
	resp, err := g.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("issue not found: %s/%s#%d", owner, repo, number)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get issue: status %d", resp.StatusCode)
	}

	var issue GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode issue: %w", err)
	}

	return &issue, nil
}

// GitHubPullRequest represents a GitHub pull request
type GitHubPullRequest struct {
	ID        int64       `json:"id"`
	Number    int         `json:"number"`
	Title     string      `json:"title"`
	State     string      `json:"state"`
	Body      string      `json:"body,omitempty"`
	HTMLURL   string      `json:"html_url"`
	User      *GitHubUser `json:"user,omitempty"`
	Head      *GitHubRef  `json:"head,omitempty"`
	Base      *GitHubRef  `json:"base,omitempty"`
	Merged    bool        `json:"merged"`
	Mergeable *bool       `json:"mergeable,omitempty"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
	ClosedAt  string      `json:"closed_at,omitempty"`
	MergedAt  string      `json:"merged_at,omitempty"`
}

// GitHubRef represents a Git reference
type GitHubRef struct {
	Ref  string `json:"ref"`
	SHA  string `json:"sha"`
	Repo *GitHubRepository `json:"repo,omitempty"`
}

// ListPullRequests lists pull requests for a repository
func (g *GitHubAdapter) ListPullRequests(ctx context.Context, owner, repo, state string) ([]GitHubPullRequest, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	path := "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repo) + "/pulls"
	if state != "" {
		path += "?state=" + url.QueryEscape(state)
	}

	resp, err := g.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list pull requests: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to list pull requests: status %d", resp.StatusCode)
	}

	var prs []GitHubPullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, fmt.Errorf("failed to decode pull requests: %w", err)
	}

	return prs, nil
}

// GetPullRequest gets a specific pull request
func (g *GitHubAdapter) GetPullRequest(ctx context.Context, owner, repo string, number int) (*GitHubPullRequest, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	path := fmt.Sprintf("/repos/%s/%s/pulls/%d", url.PathEscape(owner), url.PathEscape(repo), number)
	resp, err := g.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("pull request not found: %s/%s#%d", owner, repo, number)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get pull request: status %d", resp.StatusCode)
	}

	var pr GitHubPullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("failed to decode pull request: %w", err)
	}

	return &pr, nil
}

// GitHubContent represents file content from GitHub
type GitHubContent struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int    `json:"size"`
	Content     string `json:"content,omitempty"`
	Encoding    string `json:"encoding,omitempty"`
	HTMLURL     string `json:"html_url"`
	DownloadURL string `json:"download_url,omitempty"`
}

// GetContent gets file or directory content from a repository
func (g *GitHubAdapter) GetContent(ctx context.Context, owner, repo, path, ref string) ([]GitHubContent, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if owner == "" || repo == "" {
		return nil, fmt.Errorf("owner and repo are required")
	}

	apiPath := "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repo) + "/contents"
	if path != "" {
		apiPath += "/" + path
	}
	if ref != "" {
		apiPath += "?ref=" + url.QueryEscape(ref)
	}

	resp, err := g.doRequest(ctx, "GET", apiPath, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("content not found: %s/%s/%s", owner, repo, path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get content: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Try to parse as array (directory)
	var contents []GitHubContent
	if err := json.Unmarshal(body, &contents); err == nil {
		return contents, nil
	}

	// Try to parse as single file
	var content GitHubContent
	if err := json.Unmarshal(body, &content); err != nil {
		return nil, fmt.Errorf("failed to decode content: %w", err)
	}

	return []GitHubContent{content}, nil
}

// GitHubSearchResult represents search results
type GitHubSearchResult struct {
	TotalCount int                 `json:"total_count"`
	Items      []GitHubRepository  `json:"items,omitempty"`
}

// GitHubCodeSearchResult represents code search results
type GitHubCodeSearchResult struct {
	TotalCount int                `json:"total_count"`
	Items      []GitHubCodeMatch  `json:"items,omitempty"`
}

// GitHubCodeMatch represents a code search match
type GitHubCodeMatch struct {
	Name       string           `json:"name"`
	Path       string           `json:"path"`
	SHA        string           `json:"sha"`
	HTMLURL    string           `json:"html_url"`
	Repository *GitHubRepository `json:"repository,omitempty"`
}

// SearchRepositories searches for repositories
func (g *GitHubAdapter) SearchRepositories(ctx context.Context, query string, sort string, order string) (*GitHubSearchResult, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	path := "/search/repositories?q=" + url.QueryEscape(query)
	if sort != "" {
		path += "&sort=" + url.QueryEscape(sort)
	}
	if order != "" {
		path += "&order=" + url.QueryEscape(order)
	}

	resp, err := g.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search repositories: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to search repositories: status %d", resp.StatusCode)
	}

	var result GitHubSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	return &result, nil
}

// SearchCode searches for code
func (g *GitHubAdapter) SearchCode(ctx context.Context, query string) (*GitHubCodeSearchResult, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	path := "/search/code?q=" + url.QueryEscape(query)
	resp, err := g.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to search code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to search code: status %d", resp.StatusCode)
	}

	var result GitHubCodeSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	return &result, nil
}

// CreateIssue creates a new issue
func (g *GitHubAdapter) CreateIssue(ctx context.Context, owner, repo, title, body string, labels []string) (*GitHubIssue, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if owner == "" || repo == "" || title == "" {
		return nil, fmt.Errorf("owner, repo, and title are required")
	}

	payload := map[string]interface{}{
		"title": title,
	}
	if body != "" {
		payload["body"] = body
	}
	if len(labels) > 0 {
		payload["labels"] = labels
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	path := "/repos/" + url.PathEscape(owner) + "/" + url.PathEscape(repo) + "/issues"
	resp, err := g.doRequest(ctx, "POST", path, strings.NewReader(string(payloadJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create issue: status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var issue GitHubIssue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode issue: %w", err)
	}

	return &issue, nil
}

// GetMCPTools returns the list of MCP tools provided by this adapter
func (g *GitHubAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "github_get_user",
			Description: "Get information about a GitHub user",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"username": map[string]interface{}{
						"type":        "string",
						"description": "Username (empty for authenticated user)",
					},
				},
			},
		},
		{
			Name:        "github_list_repos",
			Description: "List repositories for a user",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Repository owner (empty for authenticated user)",
					},
					"type": map[string]interface{}{
						"type":        "string",
						"description": "Repository type (all, owner, member)",
						"enum":        []string{"all", "owner", "member"},
					},
				},
			},
		},
		{
			Name:        "github_get_repo",
			Description: "Get repository details",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Repository owner",
					},
					"repo": map[string]interface{}{
						"type":        "string",
						"description": "Repository name",
					},
				},
				"required": []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_list_issues",
			Description: "List issues for a repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Repository owner",
					},
					"repo": map[string]interface{}{
						"type":        "string",
						"description": "Repository name",
					},
					"state": map[string]interface{}{
						"type":        "string",
						"description": "Issue state (open, closed, all)",
						"enum":        []string{"open", "closed", "all"},
						"default":     "open",
					},
				},
				"required": []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_get_issue",
			Description: "Get a specific issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Repository owner",
					},
					"repo": map[string]interface{}{
						"type":        "string",
						"description": "Repository name",
					},
					"number": map[string]interface{}{
						"type":        "integer",
						"description": "Issue number",
					},
				},
				"required": []string{"owner", "repo", "number"},
			},
		},
		{
			Name:        "github_list_prs",
			Description: "List pull requests for a repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Repository owner",
					},
					"repo": map[string]interface{}{
						"type":        "string",
						"description": "Repository name",
					},
					"state": map[string]interface{}{
						"type":        "string",
						"description": "PR state (open, closed, all)",
						"enum":        []string{"open", "closed", "all"},
						"default":     "open",
					},
				},
				"required": []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_get_pr",
			Description: "Get a specific pull request",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Repository owner",
					},
					"repo": map[string]interface{}{
						"type":        "string",
						"description": "Repository name",
					},
					"number": map[string]interface{}{
						"type":        "integer",
						"description": "PR number",
					},
				},
				"required": []string{"owner", "repo", "number"},
			},
		},
		{
			Name:        "github_get_content",
			Description: "Get file or directory content from a repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Repository owner",
					},
					"repo": map[string]interface{}{
						"type":        "string",
						"description": "Repository name",
					},
					"path": map[string]interface{}{
						"type":        "string",
						"description": "Path to file or directory",
					},
					"ref": map[string]interface{}{
						"type":        "string",
						"description": "Git reference (branch, tag, or commit SHA)",
					},
				},
				"required": []string{"owner", "repo"},
			},
		},
		{
			Name:        "github_search_repos",
			Description: "Search for repositories",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query",
					},
					"sort": map[string]interface{}{
						"type":        "string",
						"description": "Sort field (stars, forks, updated)",
						"enum":        []string{"stars", "forks", "updated"},
					},
					"order": map[string]interface{}{
						"type":        "string",
						"description": "Sort order (asc, desc)",
						"enum":        []string{"asc", "desc"},
						"default":     "desc",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "github_search_code",
			Description: "Search for code in repositories",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query (include repo: or user: for filtering)",
					},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "github_create_issue",
			Description: "Create a new issue",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"owner": map[string]interface{}{
						"type":        "string",
						"description": "Repository owner",
					},
					"repo": map[string]interface{}{
						"type":        "string",
						"description": "Repository name",
					},
					"title": map[string]interface{}{
						"type":        "string",
						"description": "Issue title",
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": "Issue body",
					},
					"labels": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Labels to add",
					},
				},
				"required": []string{"owner", "repo", "title"},
			},
		},
	}
}

// ExecuteTool executes an MCP tool
func (g *GitHubAdapter) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	g.mu.RLock()
	initialized := g.initialized
	g.mu.RUnlock()

	if !initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	switch toolName {
	case "github_get_user":
		username, _ := params["username"].(string)
		return g.GetUser(ctx, username)

	case "github_list_repos":
		owner, _ := params["owner"].(string)
		repoType, _ := params["type"].(string)
		return g.ListRepositories(ctx, owner, repoType)

	case "github_get_repo":
		owner, _ := params["owner"].(string)
		repo, _ := params["repo"].(string)
		return g.GetRepository(ctx, owner, repo)

	case "github_list_issues":
		owner, _ := params["owner"].(string)
		repo, _ := params["repo"].(string)
		state, _ := params["state"].(string)
		return g.ListIssues(ctx, owner, repo, state)

	case "github_get_issue":
		owner, _ := params["owner"].(string)
		repo, _ := params["repo"].(string)
		number, _ := params["number"].(float64)
		return g.GetIssue(ctx, owner, repo, int(number))

	case "github_list_prs":
		owner, _ := params["owner"].(string)
		repo, _ := params["repo"].(string)
		state, _ := params["state"].(string)
		return g.ListPullRequests(ctx, owner, repo, state)

	case "github_get_pr":
		owner, _ := params["owner"].(string)
		repo, _ := params["repo"].(string)
		number, _ := params["number"].(float64)
		return g.GetPullRequest(ctx, owner, repo, int(number))

	case "github_get_content":
		owner, _ := params["owner"].(string)
		repo, _ := params["repo"].(string)
		path, _ := params["path"].(string)
		ref, _ := params["ref"].(string)
		return g.GetContent(ctx, owner, repo, path, ref)

	case "github_search_repos":
		query, _ := params["query"].(string)
		sort, _ := params["sort"].(string)
		order, _ := params["order"].(string)
		return g.SearchRepositories(ctx, query, sort, order)

	case "github_search_code":
		query, _ := params["query"].(string)
		return g.SearchCode(ctx, query)

	case "github_create_issue":
		owner, _ := params["owner"].(string)
		repo, _ := params["repo"].(string)
		title, _ := params["title"].(string)
		body, _ := params["body"].(string)
		var labels []string
		if l, ok := params["labels"].([]interface{}); ok {
			for _, v := range l {
				if s, ok := v.(string); ok {
					labels = append(labels, s)
				}
			}
		}
		return g.CreateIssue(ctx, owner, repo, title, body, labels)

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// GetCapabilities returns adapter capabilities
func (g *GitHubAdapter) GetCapabilities() map[string]interface{} {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return map[string]interface{}{
		"name":         "github",
		"base_url":     g.config.BaseURL,
		"authenticated": g.config.Token != "",
		"tools":        len(g.GetMCPTools()),
		"initialized":  g.initialized,
	}
}

// MarshalJSON implements custom JSON marshaling
func (g *GitHubAdapter) MarshalJSON() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	return json.Marshal(map[string]interface{}{
		"initialized":  g.initialized,
		"capabilities": g.GetCapabilities(),
	})
}
