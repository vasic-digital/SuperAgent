# Implementation Specification: Checkpoint System

**Document ID:** IMPL-004  
**Feature:** Workspace Checkpoints  
**Priority:** HIGH  
**Phase:** 1  
**Estimated Effort:** 2 weeks  
**Source:** Cline

---

## Overview

Implement a checkpoint system that captures workspace state before/after operations, enabling one-click restore and diff visualization.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                      Checkpoint System                               │
├─────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │
│  │  Checkpoint  │  │   Storage    │  │     Diff     │              │
│  │   Manager    │  │   Backend    │  │   Engine     │              │
│  │              │  │              │  │              │              │
│  │ - Create     │  │ - Local      │  │ - Generate   │              │
│  │ - Restore    │  │ - S3         │  │ - Visualize  │              │
│  │ - List       │  │ - Git        │  │ - Compare    │              │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘              │
│         │                 │                 │                       │
│         └─────────────────┴─────────────────┘                       │
│                           │                                         │
│                           ▼                                         │
│              ┌──────────────────────┐                               │
│              │   Checkpoint Store   │                               │
│              └──────────────────────┘                               │
└─────────────────────────────────────────────────────────────────────┘
```

## Components

### 1. Checkpoint Manager (`internal/checkpoints/manager.go`)

```go
package checkpoints

// Manager handles checkpoint lifecycle
type Manager struct {
    store     CheckpointStore
    storage   StorageBackend
    differ    *DiffEngine
    config    Config
}

type Config struct {
    AutoCheckpoint      bool
    AutoCheckpointOn    []string // "file_write", "bash", "git_commit"
    MaxCheckpoints      int
    StorageBackend      string // "local", "s3", "git"
    RetentionDays       int
}

// Checkpoint represents a workspace snapshot
type Checkpoint struct {
    ID          string
    Name        string
    Description string
    CreatedAt   time.Time
    CreatedBy   string
    
    // Git state
    GitRef      string // commit hash or branch
    GitBranch   string
    
    // File snapshot
    Files       []FileSnapshot
    
    // Metadata
    Tags        []string
    Size        int64
}

type FileSnapshot struct {
    Path        string
    Content     []byte
    Mode        os.FileMode
    ModTime     time.Time
    Hash        string // SHA256
}

// Operations
func (m *Manager) Create(ctx context.Context, opts CreateOptions) (*Checkpoint, error)
func (m *Manager) Restore(ctx context.Context, checkpointID string) error
func (m *Manager) Delete(ctx context.Context, checkpointID string) error
func (m *Manager) List(ctx context.Context, filter ListFilter) ([]*Checkpoint, error)
func (m *Manager) Diff(ctx context.Context, checkpointID string) (*DiffResult, error)
func (m *Manager) Compare(ctx context.Context, fromID, toID string) (*DiffResult, error)
```

### 2. Storage Backends

```go
package checkpoints

// StorageBackend interface
type StorageBackend interface {
    Save(ctx context.Context, checkpoint *Checkpoint) error
    Load(ctx context.Context, checkpointID string) (*Checkpoint, error)
    Delete(ctx context.Context, checkpointID string) error
    List(ctx context.Context) ([]string, error)
}

// LocalStorage stores checkpoints on filesystem
type LocalStorage struct {
    basePath string
}

func (s *LocalStorage) Save(ctx context.Context, cp *Checkpoint) error {
    // Create checkpoint directory
    // Write metadata as JSON
    // Write files to tar archive
}

// S3Storage stores checkpoints in S3
type S3Storage struct {
    client *s3.Client
    bucket string
    prefix string
}

// GitStorage uses git for checkpoint storage
type GitStorage struct {
    repo *git.Repository
}

func (s *GitStorage) Save(ctx context.Context, cp *Checkpoint) error {
    // Create git stash or commit
    // Tag with checkpoint ID
}
```

### 3. Diff Engine (`internal/checkpoints/diff.go`)

```go
package checkpoints

// DiffEngine generates diffs
type DiffEngine struct {
    // diff library
}

type DiffResult struct {
    FilesChanged   int
    Insertions     int
    Deletions      int
    FileDiffs      []FileDiff
}

type FileDiff struct {
    Path       string
    Status     string // added, modified, deleted, renamed
    OldContent string
    NewContent string
    Diff       string // unified diff format
}

func (e *DiffEngine) GenerateDiff(from, to *Checkpoint) (*DiffResult, error)
func (e *DiffEngine) GenerateDiffFromWorking(cp *Checkpoint) (*DiffResult, error)
func (e *DiffEngine) GenerateUnifiedDiff(oldPath, newPath string, oldContent, newContent []byte) string
```

## API

```go
// internal/handlers/checkpoints_handler.go

func (h *CheckpointsHandler) Create(c *gin.Context) {
    var req CreateRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Error: err.Error()})
        return
    }
    
    checkpoint, err := h.manager.Create(c.Request.Context(), CreateOptions{
        Name:        req.Name,
        Description: req.Description,
        Tags:        req.Tags,
    })
    
    if err != nil {
        c.JSON(500, ErrorResponse{Error: err.Error()})
        return
    }
    
    c.JSON(201, checkpoint)
}

func (h *CheckpointsHandler) Restore(c *gin.Context) {
    checkpointID := c.Param("id")
    
    if err := h.manager.Restore(c.Request.Context(), checkpointID); err != nil {
        c.JSON(500, ErrorResponse{Error: err.Error()})
        return
    }
    
    c.JSON(200, gin.H{"restored": true})
}

func (h *CheckpointsHandler) Diff(c *gin.Context) {
    checkpointID := c.Param("id")
    
    result, err := h.manager.Diff(c.Request.Context(), checkpointID)
    if err != nil {
        c.JSON(500, ErrorResponse{Error: err.Error()})
        return
    }
    
    c.JSON(200, result)
}
```

## MCP Tools

```go
var CheckpointTools = []ToolDefinition{
    {
        Name: "checkpoint_create",
        Description: "Create a workspace checkpoint",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "name": map[string]interface{}{
                    "type": "string",
                    "description": "Checkpoint name",
                },
                "description": map[string]interface{}{
                    "type": "string",
                },
                "tags": map[string]interface{}{
                    "type": "array",
                    "items": map[string]interface{}{"type": "string"},
                },
            },
            "required": []string{"name"},
        },
    },
    {
        Name: "checkpoint_restore",
        Description: "Restore workspace to checkpoint",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "checkpoint_id": map[string]interface{}{
                    "type": "string",
                },
            },
            "required": []string{"checkpoint_id"},
        },
    },
    {
        Name: "checkpoint_list",
        Description: "List all checkpoints",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "tag": map[string]interface{}{
                    "type": "string",
                },
            },
        },
    },
    {
        Name: "checkpoint_diff",
        Description: "Show diff between checkpoint and current state",
        InputSchema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "checkpoint_id": map[string]interface{}{
                    "type": "string",
                },
            },
            "required": []string{"checkpoint_id"},
        },
    },
}
```

## Configuration

```yaml
# configs/checkpoints.yaml
checkpoints:
  enabled: true
  
  auto_checkpoint:
    enabled: true
    on_actions:
      - "write_file"
      - "edit_file"
      - "bash"
      - "git_commit"
    exclude_patterns:
      - "*.log"
      - "tmp/**"
      
  storage:
    backend: "local"  # or "s3", "git"
    
    local:
      path: "~/.helixagent/checkpoints"
      
    s3:
      bucket: "helixagent-checkpoints"
      region: "us-east-1"
      prefix: "checkpoints/"
      
  retention:
    max_checkpoints: 50
    max_age_days: 30
    auto_cleanup: true
    
  diff:
    context_lines: 3
    max_file_size: 1048576  # 1MB
```

## CLI Commands

```bash
# Create checkpoint
helixagent checkpoint create --name "before-refactor" --tag "safety"

# List checkpoints
helixagent checkpoint list
helixagent checkpoint list --tag "safety"

# Show diff
helixagent checkpoint diff <checkpoint-id>

# Restore checkpoint
helixagent checkpoint restore <checkpoint-id>

# Delete checkpoint
helixagent checkpoint delete <checkpoint-id>

# Compare two checkpoints
helixagent checkpoint compare <id1> <id2>
```

## Implementation Timeline

**Week 1: Core System**
- [ ] Checkpoint data model
- [ ] Storage backends
- [ ] Create/restore operations
- [ ] File snapshotting

**Week 2: Diff & Integration**
- [ ] Diff engine
- [ ] MCP tools
- [ ] API endpoints
- [ ] Auto-checkpoint integration
- [ ] Testing

## Testing

```go
func TestCheckpointManager_Create(t *testing.T) {}
func TestCheckpointManager_Restore(t *testing.T) {}
func TestDiffEngine_GenerateDiff(t *testing.T) {}
```
