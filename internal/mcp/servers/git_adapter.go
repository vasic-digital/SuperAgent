// Package servers provides MCP server adapters for various services.
package servers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// GitAdapterConfig holds configuration for Git MCP adapter
type GitAdapterConfig struct {
	// DefaultRepoPath is the default repository path if not specified
	DefaultRepoPath string `json:"default_repo_path,omitempty"`
	// AllowedPaths are paths where Git operations are allowed
	AllowedPaths []string `json:"allowed_paths,omitempty"`
	// DeniedPaths are paths that are explicitly denied
	DeniedPaths []string `json:"denied_paths,omitempty"`
	// AllowPush enables push operations
	AllowPush bool `json:"allow_push"`
	// AllowForce enables force operations (force push, hard reset)
	AllowForce bool `json:"allow_force"`
	// AllowRemoteOperations enables fetch/pull/push
	AllowRemoteOperations bool `json:"allow_remote_operations"`
	// GitPath is the path to git executable
	GitPath string `json:"git_path,omitempty"`
	// CommandTimeout is the timeout for git commands
	CommandTimeout time.Duration `json:"command_timeout,omitempty"`
}

// DefaultGitAdapterConfig returns default configuration
func DefaultGitAdapterConfig() GitAdapterConfig {
	homeDir, _ := os.UserHomeDir()
	return GitAdapterConfig{
		DefaultRepoPath:       "",
		AllowedPaths:          []string{homeDir},
		DeniedPaths:           []string{},
		AllowPush:             false,
		AllowForce:            false,
		AllowRemoteOperations: true,
		GitPath:               "git",
		CommandTimeout:        120 * time.Second,
	}
}

// GitAdapter implements MCP adapter for Git operations
type GitAdapter struct {
	config      GitAdapterConfig
	initialized bool
	gitVersion  string
	mu          sync.RWMutex
	logger      *logrus.Logger
}

// NewGitAdapter creates a new Git MCP adapter
func NewGitAdapter(config GitAdapterConfig, logger *logrus.Logger) *GitAdapter {
	if logger == nil {
		logger = logrus.New()
		logger.SetLevel(logrus.WarnLevel)
	}

	if config.CommandTimeout == 0 {
		config.CommandTimeout = 120 * time.Second
	}

	if config.GitPath == "" {
		config.GitPath = "git"
	}

	return &GitAdapter{
		config: config,
		logger: logger,
	}
}

// Initialize initializes the Git adapter
func (g *GitAdapter) Initialize(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	// Check git is available
	cmd := exec.CommandContext(ctx, g.config.GitPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("git not available: %w", err)
	}

	g.gitVersion = strings.TrimSpace(string(output))
	g.initialized = true
	g.logger.WithField("version", g.gitVersion).Info("Git adapter initialized")

	return nil
}

// Health returns health status
func (g *GitAdapter) Health(ctx context.Context) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return fmt.Errorf("git adapter not initialized")
	}

	// Quick version check
	cmd := exec.CommandContext(ctx, g.config.GitPath, "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git health check failed: %w", err)
	}

	return nil
}

// Close closes the adapter
func (g *GitAdapter) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.initialized = false
	return nil
}

// isPathAllowed checks if a path is allowed for Git operations
func (g *GitAdapter) isPathAllowed(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Check denied paths first
	for _, denied := range g.config.DeniedPaths {
		if strings.HasPrefix(absPath, denied) {
			return false
		}
	}

	// If no allowed paths specified, allow all
	if len(g.config.AllowedPaths) == 0 {
		return true
	}

	// Check allowed paths
	for _, allowed := range g.config.AllowedPaths {
		if strings.HasPrefix(absPath, allowed) {
			return true
		}
	}

	return false
}

// runGitCommand executes a git command in the specified directory
func (g *GitAdapter) runGitCommand(ctx context.Context, dir string, args ...string) (string, error) {
	if !g.isPathAllowed(dir) {
		return "", fmt.Errorf("path not allowed: %s", dir)
	}

	ctx, cancel := context.WithTimeout(ctx, g.config.CommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, g.config.GitPath, args...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command timed out")
		}
		return "", fmt.Errorf("git command failed: %s: %w", stderr.String(), err)
	}

	return stdout.String(), nil
}

// GitStatus represents the status of a repository
type GitStatus struct {
	Branch        string            `json:"branch"`
	Ahead         int               `json:"ahead"`
	Behind        int               `json:"behind"`
	Staged        []string          `json:"staged"`
	Modified      []string          `json:"modified"`
	Untracked     []string          `json:"untracked"`
	Deleted       []string          `json:"deleted"`
	Renamed       map[string]string `json:"renamed"`
	HasConflicts  bool              `json:"has_conflicts"`
	IsClean       bool              `json:"is_clean"`
	RemoteTracking string           `json:"remote_tracking,omitempty"`
}

// Status returns the status of a repository
func (g *GitAdapter) Status(ctx context.Context, repoPath string) (*GitStatus, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if repoPath == "" {
		repoPath = g.config.DefaultRepoPath
	}

	status := &GitStatus{
		Staged:    []string{},
		Modified:  []string{},
		Untracked: []string{},
		Deleted:   []string{},
		Renamed:   make(map[string]string),
	}

	// Get branch info - handle case where repo has no commits
	branch, err := g.runGitCommand(ctx, repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		// Try to get branch from symbolic-ref for repos with no commits
		branch, err = g.runGitCommand(ctx, repoPath, "symbolic-ref", "--short", "HEAD")
		if err != nil {
			// Default to "master" if we can't determine
			status.Branch = "master"
		} else {
			status.Branch = strings.TrimSpace(branch)
		}
	} else {
		status.Branch = strings.TrimSpace(branch)
	}

	// Get tracking branch (only works if there are commits)
	tracking, err := g.runGitCommand(ctx, repoPath, "rev-parse", "--abbrev-ref", "@{upstream}")
	if err == nil {
		status.RemoteTracking = strings.TrimSpace(tracking)

		// Get ahead/behind counts
		revList, err := g.runGitCommand(ctx, repoPath, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
		if err == nil {
			parts := strings.Fields(revList)
			if len(parts) == 2 {
				fmt.Sscanf(parts[0], "%d", &status.Ahead)
				fmt.Sscanf(parts[1], "%d", &status.Behind)
			}
		}
	}

	// Get porcelain status
	output, err := g.runGitCommand(ctx, repoPath, "status", "--porcelain", "-z")
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	// Parse porcelain output (null-separated)
	entries := strings.Split(output, "\x00")
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		if len(entry) < 3 {
			continue
		}

		index := entry[0]
		worktree := entry[1]
		file := entry[3:]

		// Handle renames (R status has two entries)
		if index == 'R' || worktree == 'R' {
			if i+1 < len(entries) {
				status.Renamed[entries[i+1]] = file
				i++
			}
		}

		// Staged changes
		if index == 'A' || index == 'M' || index == 'D' || index == 'R' {
			status.Staged = append(status.Staged, file)
		}

		// Working tree changes
		if worktree == 'M' {
			status.Modified = append(status.Modified, file)
		} else if worktree == 'D' {
			status.Deleted = append(status.Deleted, file)
		}

		// Untracked files
		if index == '?' && worktree == '?' {
			status.Untracked = append(status.Untracked, file)
		}

		// Conflicts
		if index == 'U' || worktree == 'U' {
			status.HasConflicts = true
		}
	}

	status.IsClean = len(status.Staged) == 0 && len(status.Modified) == 0 &&
		len(status.Untracked) == 0 && len(status.Deleted) == 0 && !status.HasConflicts

	return status, nil
}

// GitLog represents a commit log entry
type GitLog struct {
	Hash       string    `json:"hash"`
	ShortHash  string    `json:"short_hash"`
	Author     string    `json:"author"`
	AuthorEmail string   `json:"author_email"`
	Date       time.Time `json:"date"`
	Subject    string    `json:"subject"`
	Body       string    `json:"body,omitempty"`
}

// Log returns commit history
func (g *GitAdapter) Log(ctx context.Context, repoPath string, limit int, since string, author string) ([]GitLog, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if repoPath == "" {
		repoPath = g.config.DefaultRepoPath
	}

	if limit <= 0 {
		limit = 50
	}

	args := []string{
		"log",
		fmt.Sprintf("-n%d", limit),
		"--format=%H|%h|%an|%ae|%aI|%s|%b%x00",
	}

	if since != "" {
		args = append(args, "--since="+since)
	}
	if author != "" {
		args = append(args, "--author="+author)
	}

	output, err := g.runGitCommand(ctx, repoPath, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}

	var logs []GitLog
	entries := strings.Split(output, "\x00")
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		parts := strings.SplitN(entry, "|", 7)
		if len(parts) < 6 {
			continue
		}

		date, _ := time.Parse(time.RFC3339, parts[4])
		log := GitLog{
			Hash:        parts[0],
			ShortHash:   parts[1],
			Author:      parts[2],
			AuthorEmail: parts[3],
			Date:        date,
			Subject:     parts[5],
		}
		if len(parts) > 6 {
			log.Body = strings.TrimSpace(parts[6])
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GitDiff represents a diff
type GitDiff struct {
	Files   []GitDiffFile `json:"files"`
	Summary string        `json:"summary"`
}

// GitDiffFile represents a file in a diff
type GitDiffFile struct {
	Path      string `json:"path"`
	OldPath   string `json:"old_path,omitempty"`
	Status    string `json:"status"`
	Additions int    `json:"additions"`
	Deletions int    `json:"deletions"`
	Patch     string `json:"patch,omitempty"`
}

// Diff returns diff between refs or working tree
func (g *GitAdapter) Diff(ctx context.Context, repoPath string, ref1, ref2 string, paths []string, staged bool) (*GitDiff, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if repoPath == "" {
		repoPath = g.config.DefaultRepoPath
	}

	args := []string{"diff", "--numstat"}
	if staged {
		args = append(args, "--staged")
	}
	if ref1 != "" {
		args = append(args, ref1)
	}
	if ref2 != "" {
		args = append(args, ref2)
	}
	if len(paths) > 0 {
		args = append(args, "--")
		args = append(args, paths...)
	}

	output, err := g.runGitCommand(ctx, repoPath, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff: %w", err)
	}

	diff := &GitDiff{
		Files: []GitDiffFile{},
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		var additions, deletions int
		if parts[0] != "-" {
			fmt.Sscanf(parts[0], "%d", &additions)
		}
		if parts[1] != "-" {
			fmt.Sscanf(parts[1], "%d", &deletions)
		}

		file := GitDiffFile{
			Path:      parts[2],
			Additions: additions,
			Deletions: deletions,
			Status:    "modified",
		}

		// Check if it's a rename
		if len(parts) > 3 && strings.Contains(line, "=>") {
			file.OldPath = parts[2]
			file.Path = parts[len(parts)-1]
			file.Status = "renamed"
		}

		diff.Files = append(diff.Files, file)
	}

	// Get summary
	args = []string{"diff", "--stat"}
	if staged {
		args = append(args, "--staged")
	}
	if ref1 != "" {
		args = append(args, ref1)
	}
	if ref2 != "" {
		args = append(args, ref2)
	}

	summary, err := g.runGitCommand(ctx, repoPath, args...)
	if err == nil {
		diff.Summary = strings.TrimSpace(summary)
	}

	return diff, nil
}

// Add stages files for commit
func (g *GitAdapter) Add(ctx context.Context, repoPath string, paths []string, all bool) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return fmt.Errorf("adapter not initialized")
	}

	if repoPath == "" {
		repoPath = g.config.DefaultRepoPath
	}

	args := []string{"add"}
	if all {
		args = append(args, "-A")
	} else if len(paths) > 0 {
		args = append(args, paths...)
	} else {
		return fmt.Errorf("no paths specified and all flag not set")
	}

	_, err := g.runGitCommand(ctx, repoPath, args...)
	return err
}

// GitCommitResult represents the result of a commit
type GitCommitResult struct {
	Hash      string `json:"hash"`
	ShortHash string `json:"short_hash"`
	Subject   string `json:"subject"`
	Branch    string `json:"branch"`
}

// Commit creates a commit
func (g *GitAdapter) Commit(ctx context.Context, repoPath, message string, amend bool) (*GitCommitResult, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if repoPath == "" {
		repoPath = g.config.DefaultRepoPath
	}

	if message == "" && !amend {
		return nil, fmt.Errorf("commit message required")
	}

	args := []string{"commit"}
	if amend {
		if !g.config.AllowForce {
			return nil, fmt.Errorf("amend not allowed (force operations disabled)")
		}
		args = append(args, "--amend")
	}
	if message != "" {
		args = append(args, "-m", message)
	}

	_, err := g.runGitCommand(ctx, repoPath, args...)
	if err != nil {
		return nil, fmt.Errorf("commit failed: %w", err)
	}

	// Get commit info
	hash, err := g.runGitCommand(ctx, repoPath, "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get commit hash: %w", err)
	}

	shortHash, _ := g.runGitCommand(ctx, repoPath, "rev-parse", "--short", "HEAD")
	branch, _ := g.runGitCommand(ctx, repoPath, "rev-parse", "--abbrev-ref", "HEAD")

	return &GitCommitResult{
		Hash:      strings.TrimSpace(hash),
		ShortHash: strings.TrimSpace(shortHash),
		Subject:   message,
		Branch:    strings.TrimSpace(branch),
	}, nil
}

// Branch creates, lists, or deletes branches
func (g *GitAdapter) Branch(ctx context.Context, repoPath, name string, create, delete bool) ([]string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if repoPath == "" {
		repoPath = g.config.DefaultRepoPath
	}

	if create && name != "" {
		_, err := g.runGitCommand(ctx, repoPath, "branch", name)
		if err != nil {
			return nil, fmt.Errorf("failed to create branch: %w", err)
		}
		return []string{name}, nil
	}

	if delete && name != "" {
		_, err := g.runGitCommand(ctx, repoPath, "branch", "-d", name)
		if err != nil {
			return nil, fmt.Errorf("failed to delete branch: %w", err)
		}
		return []string{name}, nil
	}

	// List branches
	output, err := g.runGitCommand(ctx, repoPath, "branch", "--list", "--format=%(refname:short)")
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	branches := []string{}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			branches = append(branches, line)
		}
	}

	return branches, nil
}

// Checkout switches branches or restores files
func (g *GitAdapter) Checkout(ctx context.Context, repoPath, ref string, create bool, paths []string) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return fmt.Errorf("adapter not initialized")
	}

	if repoPath == "" {
		repoPath = g.config.DefaultRepoPath
	}

	args := []string{"checkout"}
	if create {
		args = append(args, "-b")
	}
	if ref != "" {
		args = append(args, ref)
	}
	if len(paths) > 0 {
		args = append(args, "--")
		args = append(args, paths...)
	}

	_, err := g.runGitCommand(ctx, repoPath, args...)
	return err
}

// Push pushes commits to remote
func (g *GitAdapter) Push(ctx context.Context, repoPath, remote, branch string, force, setUpstream bool) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return fmt.Errorf("adapter not initialized")
	}

	if !g.config.AllowRemoteOperations {
		return fmt.Errorf("remote operations not allowed")
	}

	if !g.config.AllowPush {
		return fmt.Errorf("push not allowed")
	}

	if force && !g.config.AllowForce {
		return fmt.Errorf("force push not allowed")
	}

	if repoPath == "" {
		repoPath = g.config.DefaultRepoPath
	}

	if remote == "" {
		remote = "origin"
	}

	args := []string{"push"}
	if force {
		args = append(args, "--force")
	}
	if setUpstream {
		args = append(args, "-u")
	}
	args = append(args, remote)
	if branch != "" {
		args = append(args, branch)
	}

	_, err := g.runGitCommand(ctx, repoPath, args...)
	return err
}

// Pull fetches and merges from remote
func (g *GitAdapter) Pull(ctx context.Context, repoPath, remote, branch string, rebase bool) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return fmt.Errorf("adapter not initialized")
	}

	if !g.config.AllowRemoteOperations {
		return fmt.Errorf("remote operations not allowed")
	}

	if repoPath == "" {
		repoPath = g.config.DefaultRepoPath
	}

	args := []string{"pull"}
	if rebase {
		args = append(args, "--rebase")
	}
	if remote != "" {
		args = append(args, remote)
	}
	if branch != "" {
		args = append(args, branch)
	}

	_, err := g.runGitCommand(ctx, repoPath, args...)
	return err
}

// Fetch fetches from remote
func (g *GitAdapter) Fetch(ctx context.Context, repoPath, remote string, all, prune bool) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return fmt.Errorf("adapter not initialized")
	}

	if !g.config.AllowRemoteOperations {
		return fmt.Errorf("remote operations not allowed")
	}

	if repoPath == "" {
		repoPath = g.config.DefaultRepoPath
	}

	args := []string{"fetch"}
	if all {
		args = append(args, "--all")
	} else if remote != "" {
		args = append(args, remote)
	}
	if prune {
		args = append(args, "--prune")
	}

	_, err := g.runGitCommand(ctx, repoPath, args...)
	return err
}

// GitRemote represents a git remote
type GitRemote struct {
	Name     string `json:"name"`
	FetchURL string `json:"fetch_url"`
	PushURL  string `json:"push_url"`
}

// Remotes lists remote repositories
func (g *GitAdapter) Remotes(ctx context.Context, repoPath string) ([]GitRemote, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	if repoPath == "" {
		repoPath = g.config.DefaultRepoPath
	}

	output, err := g.runGitCommand(ctx, repoPath, "remote", "-v")
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}

	remotes := make(map[string]*GitRemote)
	lines := strings.Split(output, "\n")
	re := regexp.MustCompile(`^(\S+)\s+(\S+)\s+\((fetch|push)\)$`)

	for _, line := range lines {
		matches := re.FindStringSubmatch(line)
		if len(matches) < 4 {
			continue
		}

		name := matches[1]
		url := matches[2]
		urlType := matches[3]

		if _, exists := remotes[name]; !exists {
			remotes[name] = &GitRemote{Name: name}
		}

		if urlType == "fetch" {
			remotes[name].FetchURL = url
		} else {
			remotes[name].PushURL = url
		}
	}

	result := make([]GitRemote, 0, len(remotes))
	for _, remote := range remotes {
		result = append(result, *remote)
	}

	return result, nil
}

// Stash operations
func (g *GitAdapter) Stash(ctx context.Context, repoPath, message string, pop, list bool) (string, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return "", fmt.Errorf("adapter not initialized")
	}

	if repoPath == "" {
		repoPath = g.config.DefaultRepoPath
	}

	var args []string
	if list {
		args = []string{"stash", "list"}
	} else if pop {
		args = []string{"stash", "pop"}
	} else {
		args = []string{"stash", "push"}
		if message != "" {
			args = append(args, "-m", message)
		}
	}

	output, err := g.runGitCommand(ctx, repoPath, args...)
	return strings.TrimSpace(output), err
}

// Clone clones a repository
func (g *GitAdapter) Clone(ctx context.Context, url, destPath string, shallow bool, depth int) error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return fmt.Errorf("adapter not initialized")
	}

	if !g.config.AllowRemoteOperations {
		return fmt.Errorf("remote operations not allowed")
	}

	if !g.isPathAllowed(destPath) {
		return fmt.Errorf("destination path not allowed: %s", destPath)
	}

	args := []string{"clone"}
	if shallow {
		if depth <= 0 {
			depth = 1
		}
		args = append(args, "--depth", fmt.Sprintf("%d", depth))
	}
	args = append(args, url, destPath)

	// Clone runs in parent directory
	parentDir := filepath.Dir(destPath)
	_, err := g.runGitCommand(ctx, parentDir, args...)
	return err
}

// GetMCPTools returns the list of MCP tools provided by this adapter
func (g *GitAdapter) GetMCPTools() []MCPTool {
	return []MCPTool{
		{
			Name:        "git_status",
			Description: "Get the status of a Git repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the git repository",
					},
				},
			},
		},
		{
			Name:        "git_log",
			Description: "Get commit history of a Git repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the git repository",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of commits to return",
						"default":     50,
					},
					"since": map[string]interface{}{
						"type":        "string",
						"description": "Show commits more recent than date (e.g., '2 weeks ago')",
					},
					"author": map[string]interface{}{
						"type":        "string",
						"description": "Filter by author name or email",
					},
				},
			},
		},
		{
			Name:        "git_diff",
			Description: "Get diff between refs or working tree changes",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the git repository",
					},
					"ref1": map[string]interface{}{
						"type":        "string",
						"description": "First ref (commit, branch, tag)",
					},
					"ref2": map[string]interface{}{
						"type":        "string",
						"description": "Second ref (commit, branch, tag)",
					},
					"staged": map[string]interface{}{
						"type":        "boolean",
						"description": "Show staged changes",
						"default":     false,
					},
					"paths": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Limit diff to specific paths",
					},
				},
			},
		},
		{
			Name:        "git_add",
			Description: "Stage files for commit",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the git repository",
					},
					"paths": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Files to stage",
					},
					"all": map[string]interface{}{
						"type":        "boolean",
						"description": "Stage all changes",
						"default":     false,
					},
				},
			},
		},
		{
			Name:        "git_commit",
			Description: "Create a commit with staged changes",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the git repository",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Commit message",
					},
					"amend": map[string]interface{}{
						"type":        "boolean",
						"description": "Amend the last commit",
						"default":     false,
					},
				},
				"required": []string{"message"},
			},
		},
		{
			Name:        "git_branch",
			Description: "Create, list, or delete branches",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the git repository",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Branch name",
					},
					"create": map[string]interface{}{
						"type":        "boolean",
						"description": "Create a new branch",
						"default":     false,
					},
					"delete": map[string]interface{}{
						"type":        "boolean",
						"description": "Delete a branch",
						"default":     false,
					},
				},
			},
		},
		{
			Name:        "git_checkout",
			Description: "Switch branches or restore files",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the git repository",
					},
					"ref": map[string]interface{}{
						"type":        "string",
						"description": "Branch, tag, or commit to checkout",
					},
					"create": map[string]interface{}{
						"type":        "boolean",
						"description": "Create a new branch",
						"default":     false,
					},
					"paths": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Files to restore",
					},
				},
			},
		},
		{
			Name:        "git_push",
			Description: "Push commits to remote repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the git repository",
					},
					"remote": map[string]interface{}{
						"type":        "string",
						"description": "Remote name",
						"default":     "origin",
					},
					"branch": map[string]interface{}{
						"type":        "string",
						"description": "Branch to push",
					},
					"force": map[string]interface{}{
						"type":        "boolean",
						"description": "Force push",
						"default":     false,
					},
					"set_upstream": map[string]interface{}{
						"type":        "boolean",
						"description": "Set upstream for the branch",
						"default":     false,
					},
				},
			},
		},
		{
			Name:        "git_pull",
			Description: "Pull changes from remote repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the git repository",
					},
					"remote": map[string]interface{}{
						"type":        "string",
						"description": "Remote name",
					},
					"branch": map[string]interface{}{
						"type":        "string",
						"description": "Branch to pull",
					},
					"rebase": map[string]interface{}{
						"type":        "boolean",
						"description": "Rebase instead of merge",
						"default":     false,
					},
				},
			},
		},
		{
			Name:        "git_fetch",
			Description: "Fetch changes from remote repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the git repository",
					},
					"remote": map[string]interface{}{
						"type":        "string",
						"description": "Remote name",
					},
					"all": map[string]interface{}{
						"type":        "boolean",
						"description": "Fetch from all remotes",
						"default":     false,
					},
					"prune": map[string]interface{}{
						"type":        "boolean",
						"description": "Prune deleted remote branches",
						"default":     false,
					},
				},
			},
		},
		{
			Name:        "git_remotes",
			Description: "List remote repositories",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the git repository",
					},
				},
			},
		},
		{
			Name:        "git_stash",
			Description: "Stash or restore working directory changes",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo_path": map[string]interface{}{
						"type":        "string",
						"description": "Path to the git repository",
					},
					"message": map[string]interface{}{
						"type":        "string",
						"description": "Stash message",
					},
					"pop": map[string]interface{}{
						"type":        "boolean",
						"description": "Pop the top stash",
						"default":     false,
					},
					"list": map[string]interface{}{
						"type":        "boolean",
						"description": "List all stashes",
						"default":     false,
					},
				},
			},
		},
		{
			Name:        "git_clone",
			Description: "Clone a repository",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"url": map[string]interface{}{
						"type":        "string",
						"description": "Repository URL to clone",
					},
					"dest_path": map[string]interface{}{
						"type":        "string",
						"description": "Destination path",
					},
					"shallow": map[string]interface{}{
						"type":        "boolean",
						"description": "Create a shallow clone",
						"default":     false,
					},
					"depth": map[string]interface{}{
						"type":        "integer",
						"description": "Depth for shallow clone",
						"default":     1,
					},
				},
				"required": []string{"url", "dest_path"},
			},
		},
	}
}

// ExecuteTool executes an MCP tool
func (g *GitAdapter) ExecuteTool(ctx context.Context, toolName string, params map[string]interface{}) (interface{}, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if !g.initialized {
		return nil, fmt.Errorf("adapter not initialized")
	}

	repoPath := ""
	if path, ok := params["repo_path"].(string); ok {
		repoPath = path
	}
	if repoPath == "" {
		repoPath = g.config.DefaultRepoPath
	}

	switch toolName {
	case "git_status":
		// Unlock before calling Status which re-locks
		g.mu.RUnlock()
		result, err := g.Status(ctx, repoPath)
		g.mu.RLock()
		return result, err

	case "git_log":
		limit := 50
		if l, ok := params["limit"].(float64); ok {
			limit = int(l)
		}
		since, _ := params["since"].(string)
		author, _ := params["author"].(string)

		g.mu.RUnlock()
		result, err := g.Log(ctx, repoPath, limit, since, author)
		g.mu.RLock()
		return result, err

	case "git_diff":
		ref1, _ := params["ref1"].(string)
		ref2, _ := params["ref2"].(string)
		staged, _ := params["staged"].(bool)
		var paths []string
		if p, ok := params["paths"].([]interface{}); ok {
			for _, path := range p {
				if s, ok := path.(string); ok {
					paths = append(paths, s)
				}
			}
		}

		g.mu.RUnlock()
		result, err := g.Diff(ctx, repoPath, ref1, ref2, paths, staged)
		g.mu.RLock()
		return result, err

	case "git_add":
		all, _ := params["all"].(bool)
		var paths []string
		if p, ok := params["paths"].([]interface{}); ok {
			for _, path := range p {
				if s, ok := path.(string); ok {
					paths = append(paths, s)
				}
			}
		}

		g.mu.RUnlock()
		err := g.Add(ctx, repoPath, paths, all)
		g.mu.RLock()
		return map[string]interface{}{"success": err == nil}, err

	case "git_commit":
		message, _ := params["message"].(string)
		amend, _ := params["amend"].(bool)

		g.mu.RUnlock()
		result, err := g.Commit(ctx, repoPath, message, amend)
		g.mu.RLock()
		return result, err

	case "git_branch":
		name, _ := params["name"].(string)
		create, _ := params["create"].(bool)
		delete, _ := params["delete"].(bool)

		g.mu.RUnlock()
		result, err := g.Branch(ctx, repoPath, name, create, delete)
		g.mu.RLock()
		return map[string]interface{}{"branches": result}, err

	case "git_checkout":
		ref, _ := params["ref"].(string)
		create, _ := params["create"].(bool)
		var paths []string
		if p, ok := params["paths"].([]interface{}); ok {
			for _, path := range p {
				if s, ok := path.(string); ok {
					paths = append(paths, s)
				}
			}
		}

		g.mu.RUnlock()
		err := g.Checkout(ctx, repoPath, ref, create, paths)
		g.mu.RLock()
		return map[string]interface{}{"success": err == nil}, err

	case "git_push":
		remote, _ := params["remote"].(string)
		branch, _ := params["branch"].(string)
		force, _ := params["force"].(bool)
		setUpstream, _ := params["set_upstream"].(bool)

		g.mu.RUnlock()
		err := g.Push(ctx, repoPath, remote, branch, force, setUpstream)
		g.mu.RLock()
		return map[string]interface{}{"success": err == nil}, err

	case "git_pull":
		remote, _ := params["remote"].(string)
		branch, _ := params["branch"].(string)
		rebase, _ := params["rebase"].(bool)

		g.mu.RUnlock()
		err := g.Pull(ctx, repoPath, remote, branch, rebase)
		g.mu.RLock()
		return map[string]interface{}{"success": err == nil}, err

	case "git_fetch":
		remote, _ := params["remote"].(string)
		all, _ := params["all"].(bool)
		prune, _ := params["prune"].(bool)

		g.mu.RUnlock()
		err := g.Fetch(ctx, repoPath, remote, all, prune)
		g.mu.RLock()
		return map[string]interface{}{"success": err == nil}, err

	case "git_remotes":
		g.mu.RUnlock()
		result, err := g.Remotes(ctx, repoPath)
		g.mu.RLock()
		return map[string]interface{}{"remotes": result}, err

	case "git_stash":
		message, _ := params["message"].(string)
		pop, _ := params["pop"].(bool)
		list, _ := params["list"].(bool)

		g.mu.RUnlock()
		result, err := g.Stash(ctx, repoPath, message, pop, list)
		g.mu.RLock()
		return map[string]interface{}{"result": result}, err

	case "git_clone":
		url, _ := params["url"].(string)
		destPath, _ := params["dest_path"].(string)
		shallow, _ := params["shallow"].(bool)
		depth := 1
		if d, ok := params["depth"].(float64); ok {
			depth = int(d)
		}

		g.mu.RUnlock()
		err := g.Clone(ctx, url, destPath, shallow, depth)
		g.mu.RLock()
		return map[string]interface{}{"success": err == nil}, err

	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// GetCapabilities returns adapter capabilities
func (g *GitAdapter) GetCapabilities() map[string]interface{} {
	return map[string]interface{}{
		"name":                   "git",
		"version":                g.gitVersion,
		"allow_push":             g.config.AllowPush,
		"allow_force":            g.config.AllowForce,
		"allow_remote_operations": g.config.AllowRemoteOperations,
		"tools":                  len(g.GetMCPTools()),
	}
}

// MarshalJSON implements custom JSON marshaling
func (g *GitAdapter) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"initialized":  g.initialized,
		"git_version":  g.gitVersion,
		"capabilities": g.GetCapabilities(),
	})
}
