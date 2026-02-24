// Package tools provides tool integration for debate agents.
// git_tool.go manages git worktrees for debate version control,
// enabling isolated code snapshots and diffs per debate session.
package tools

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// GitTool manages git worktrees for debate sessions.
// Each debate session gets an isolated worktree where code snapshots
// can be committed and compared without affecting the main repository.
type GitTool struct {
	workDir     string
	worktreeDir string // base directory for worktrees
	mu          sync.Mutex
}

// GitToolConfig configures the git tool.
type GitToolConfig struct {
	RepoDir     string `json:"repo_dir"`
	WorktreeDir string `json:"worktree_dir"` // defaults to <repo_dir>/.debate-worktrees
}

// WorktreeInfo describes a debate worktree.
type WorktreeInfo struct {
	SessionID string    `json:"session_id"`
	Path      string    `json:"path"`
	Branch    string    `json:"branch"`
	CreatedAt time.Time `json:"created_at"`
}

// SnapshotResult captures the result of a code snapshot commit.
type SnapshotResult struct {
	CommitHash string    `json:"commit_hash"`
	Branch     string    `json:"branch"`
	Timestamp  time.Time `json:"timestamp"`
	Message    string    `json:"message"`
}

// debateBranchPrefix is the prefix used for all debate worktree branches.
const debateBranchPrefix = "debate/"

// NewGitTool creates a new GitTool instance.
// It validates that the repo directory exists and is a git repository,
// and creates the worktree base directory if it does not exist.
func NewGitTool(config GitToolConfig) (*GitTool, error) {
	if config.RepoDir == "" {
		return nil, fmt.Errorf("repo_dir is required")
	}

	// Resolve to absolute path.
	absRepoDir, err := filepath.Abs(config.RepoDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve repo dir: %w", err)
	}

	// Verify the directory exists.
	info, err := os.Stat(absRepoDir)
	if err != nil {
		return nil, fmt.Errorf("repo dir does not exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("repo dir is not a directory: %s", absRepoDir)
	}

	// Verify it is a git repository by checking for .git.
	gitDir := filepath.Join(absRepoDir, ".git")
	if _, err := os.Stat(gitDir); err != nil {
		return nil, fmt.Errorf(
			"not a git repository (no .git found): %s: %w", absRepoDir, err,
		)
	}

	// Determine worktree base directory.
	worktreeDir := config.WorktreeDir
	if worktreeDir == "" {
		worktreeDir = filepath.Join(absRepoDir, ".debate-worktrees")
	} else {
		worktreeDir, err = filepath.Abs(worktreeDir)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve worktree dir: %w", err)
		}
	}

	// Create the worktree base directory if it does not exist.
	if err := os.MkdirAll(worktreeDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create worktree dir: %w", err)
	}

	return &GitTool{
		workDir:     absRepoDir,
		worktreeDir: worktreeDir,
	}, nil
}

// CreateWorktree creates an isolated git worktree for a debate session.
// It runs `git worktree add <path> -b debate/<sessionID>` to create a
// new branch and worktree directory dedicated to the session.
func (g *GitTool) CreateWorktree(
	ctx context.Context, sessionID string,
) (*WorktreeInfo, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	branch := debateBranchPrefix + sessionID
	wtPath := filepath.Join(g.worktreeDir, sessionID)

	// Check if the worktree path already exists.
	if _, err := os.Stat(wtPath); err == nil {
		return nil, fmt.Errorf(
			"worktree already exists for session %s: %s", sessionID, wtPath,
		)
	}

	// Create the worktree with a new branch.
	_, err := g.runGit(ctx, g.workDir, "worktree", "add", wtPath, "-b", branch)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create worktree for session %s: %w", sessionID, err,
		)
	}

	return &WorktreeInfo{
		SessionID: sessionID,
		Path:      wtPath,
		Branch:    branch,
		CreatedAt: time.Now(),
	}, nil
}

// CommitSnapshot writes code to a file in the worktree and commits it.
// This creates a versioned snapshot of code produced during a debate round.
func (g *GitTool) CommitSnapshot(
	ctx context.Context,
	worktreeDir string,
	code string,
	filename string,
	message string,
) (*SnapshotResult, error) {
	if worktreeDir == "" {
		return nil, fmt.Errorf("worktree directory is required")
	}
	if code == "" {
		return nil, fmt.Errorf("code content is required")
	}
	if filename == "" {
		return nil, fmt.Errorf("filename is required")
	}
	if message == "" {
		return nil, fmt.Errorf("commit message is required")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Verify the worktree directory exists.
	if _, err := os.Stat(worktreeDir); err != nil {
		return nil, fmt.Errorf("worktree directory does not exist: %w", err)
	}

	// Ensure parent directories for the file exist within the worktree.
	filePath := filepath.Join(worktreeDir, filename)
	fileDir := filepath.Dir(filePath)
	if err := os.MkdirAll(fileDir, 0o755); err != nil {
		return nil, fmt.Errorf(
			"failed to create directories for %s: %w", filename, err,
		)
	}

	// Write the code to the file.
	if err := os.WriteFile(filePath, []byte(code), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write file %s: %w", filename, err)
	}

	// Stage the file.
	if _, err := g.runGit(ctx, worktreeDir, "add", filename); err != nil {
		return nil, fmt.Errorf("failed to stage file %s: %w", filename, err)
	}

	// Commit the snapshot.
	if _, err := g.runGit(
		ctx, worktreeDir, "commit", "-m", message,
	); err != nil {
		return nil, fmt.Errorf("failed to commit snapshot: %w", err)
	}

	// Get the commit hash.
	commitHash, err := g.runGit(ctx, worktreeDir, "rev-parse", "HEAD")
	if err != nil {
		return nil, fmt.Errorf("failed to get commit hash: %w", err)
	}

	// Get the current branch name.
	branch, err := g.runGit(
		ctx, worktreeDir, "rev-parse", "--abbrev-ref", "HEAD",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch name: %w", err)
	}

	return &SnapshotResult{
		CommitHash: commitHash,
		Branch:     branch,
		Timestamp:  time.Now(),
		Message:    message,
	}, nil
}

// CreateDiff generates a diff between two git references within a worktree.
// The references can be commit hashes, branch names, or tags.
func (g *GitTool) CreateDiff(
	ctx context.Context, worktreeDir string, fromRef, toRef string,
) (string, error) {
	if worktreeDir == "" {
		return "", fmt.Errorf("worktree directory is required")
	}
	if fromRef == "" {
		return "", fmt.Errorf("fromRef is required")
	}
	if toRef == "" {
		return "", fmt.Errorf("toRef is required")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	diff, err := g.runGit(ctx, worktreeDir, "diff", fromRef, toRef)
	if err != nil {
		return "", fmt.Errorf(
			"failed to create diff between %s and %s: %w", fromRef, toRef, err,
		)
	}

	return diff, nil
}

// ListWorktrees lists all debate worktrees associated with this repository.
// It parses the output of `git worktree list` and filters for debate branches.
func (g *GitTool) ListWorktrees(ctx context.Context) ([]*WorktreeInfo, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	output, err := g.runGit(ctx, g.workDir, "worktree", "list")
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	var worktrees []*WorktreeInfo
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse the worktree list output.
		// Format: <path> <commit> [<branch>]
		wt := g.parseWorktreeLine(line)
		if wt == nil {
			continue
		}

		// Only include debate worktrees.
		if !strings.HasPrefix(wt.Branch, debateBranchPrefix) {
			continue
		}

		worktrees = append(worktrees, wt)
	}

	return worktrees, nil
}

// parseWorktreeLine parses a single line from `git worktree list` output.
// Expected format: <path> <commit> [<branch>]
func (g *GitTool) parseWorktreeLine(line string) *WorktreeInfo {
	// Find the branch in brackets.
	branchStart := strings.Index(line, "[")
	branchEnd := strings.Index(line, "]")
	if branchStart < 0 || branchEnd < 0 || branchEnd <= branchStart {
		return nil
	}

	branch := line[branchStart+1 : branchEnd]
	// The path is the first field before the commit hash.
	fields := strings.Fields(line[:branchStart])
	if len(fields) < 1 {
		return nil
	}

	path := fields[0]

	// Extract session ID from branch name.
	sessionID := ""
	if strings.HasPrefix(branch, debateBranchPrefix) {
		sessionID = strings.TrimPrefix(branch, debateBranchPrefix)
	}

	// Try to get creation time from the worktree directory.
	createdAt := time.Time{}
	if info, err := os.Stat(path); err == nil {
		createdAt = info.ModTime()
	}

	return &WorktreeInfo{
		SessionID: sessionID,
		Path:      path,
		Branch:    branch,
		CreatedAt: createdAt,
	}
}

// Cleanup removes a specific debate worktree and its associated branch.
// It runs `git worktree remove` followed by `git branch -D` to clean up.
func (g *GitTool) Cleanup(ctx context.Context, worktreeDir string) error {
	if worktreeDir == "" {
		return fmt.Errorf("worktree directory is required")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	return g.cleanupWorktree(ctx, worktreeDir)
}

// cleanupWorktree performs the actual worktree cleanup without locking.
// Must be called with g.mu held.
func (g *GitTool) cleanupWorktree(
	ctx context.Context, worktreeDir string,
) error {
	// Get the branch name before removing the worktree.
	branch, err := g.runGit(
		ctx, worktreeDir, "rev-parse", "--abbrev-ref", "HEAD",
	)
	if err != nil {
		// If we cannot get the branch, still try to remove the worktree.
		branch = ""
	}

	// Remove the worktree.
	_, err = g.runGit(ctx, g.workDir, "worktree", "remove", worktreeDir, "--force")
	if err != nil {
		// If git worktree remove fails, try manual cleanup.
		if removeErr := os.RemoveAll(worktreeDir); removeErr != nil {
			return fmt.Errorf(
				"failed to remove worktree %s: git: %w, manual: %v",
				worktreeDir, err, removeErr,
			)
		}
		// Prune stale worktree references after manual removal.
		if _, pruneErr := g.runGit(
			ctx, g.workDir, "worktree", "prune",
		); pruneErr != nil {
			return fmt.Errorf(
				"failed to prune worktrees after manual removal: %w", pruneErr,
			)
		}
	}

	// Delete the associated branch if it was a debate branch.
	if branch != "" && strings.HasPrefix(branch, debateBranchPrefix) {
		if _, err := g.runGit(
			ctx, g.workDir, "branch", "-D", branch,
		); err != nil {
			// Branch deletion failure is non-fatal; the worktree is already
			// removed. Log-worthy but not an error to return.
			return fmt.Errorf(
				"worktree removed but failed to delete branch %s: %w",
				branch, err,
			)
		}
	}

	return nil
}

// CleanupAll removes all debate worktrees and their associated branches.
// It discovers all debate worktrees via ListWorktrees and cleans each one.
func (g *GitTool) CleanupAll(ctx context.Context) error {
	// List worktrees without holding the lock (ListWorktrees acquires it).
	worktrees, err := g.ListWorktrees(ctx)
	if err != nil {
		return fmt.Errorf("failed to list worktrees for cleanup: %w", err)
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	var errs []string
	for _, wt := range worktrees {
		if err := g.cleanupWorktree(ctx, wt.Path); err != nil {
			errs = append(errs, fmt.Sprintf(
				"session %s: %v", wt.SessionID, err,
			))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(
			"failed to cleanup %d worktree(s): %s",
			len(errs), strings.Join(errs, "; "),
		)
	}

	return nil
}

// runGit executes a git command in the specified directory and returns
// the trimmed stdout output. Returns an error with stderr details on failure.
func (g *GitTool) runGit(
	ctx context.Context, dir string, args ...string,
) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf(
			"git %s failed: %v: %s",
			strings.Join(args, " "), err, stderr.String(),
		)
	}

	return strings.TrimSpace(stdout.String()), nil
}
