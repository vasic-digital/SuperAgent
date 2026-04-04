// Package checkpoints provides workspace checkpoint management
package checkpoints

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
)

// Checkpoint represents a workspace snapshot
type Checkpoint struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
	CreatedBy   string
	GitRef      string
	GitBranch   string
	Files       []FileSnapshot
	Tags        []string
	Size        int64
}

// FileSnapshot represents a file at a point in time
type FileSnapshot struct {
	Path    string
	Hash    string
	Mode    os.FileMode
	ModTime time.Time
	Content []byte
}

// Manager manages checkpoints
type Manager struct {
	basePath string
	repo     *git.Repository
}

// NewManager creates a new checkpoint manager
func NewManager(basePath string) (*Manager, error) {
	repo, err := git.PlainOpen(basePath)
	if err != nil {
		// No git repo, that's ok
		repo = nil
	}

	// Create checkpoints directory
	checkpointsDir := filepath.Join(basePath, ".checkpoints")
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create checkpoints directory: %w", err)
	}

	return &Manager{
		basePath: basePath,
		repo:     repo,
	}, nil
}

// Create creates a new checkpoint
func (m *Manager) Create(name, description string, tags []string) (*Checkpoint, error) {
	checkpoint := &Checkpoint{
		ID:          generateID(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		Tags:        tags,
	}

	// Capture git state
	if m.repo != nil {
		head, err := m.repo.Head()
		if err == nil {
			checkpoint.GitRef = head.Hash().String()
			checkpoint.GitBranch = head.Name().Short()
		}
	}

	// Capture files
	files, err := m.captureFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to capture files: %w", err)
	}
	checkpoint.Files = files

	// Save checkpoint
	if err := m.saveCheckpoint(checkpoint); err != nil {
		return nil, fmt.Errorf("failed to save checkpoint: %w", err)
	}

	return checkpoint, nil
}

// Restore restores a checkpoint
func (m *Manager) Restore(checkpointID string) error {
	checkpoint, err := m.loadCheckpoint(checkpointID)
	if err != nil {
		return fmt.Errorf("failed to load checkpoint: %w", err)
	}

	// Restore files
	for _, file := range checkpoint.Files {
		path := filepath.Join(m.basePath, file.Path)
		
		// Create directory if needed
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		// Write file
		if err := os.WriteFile(path, file.Content, file.Mode); err != nil {
			return fmt.Errorf("failed to write file %s: %w", path, err)
		}

		// Restore modification time
		os.Chtimes(path, file.ModTime, file.ModTime)
	}

	return nil
}

// List returns all checkpoints
func (m *Manager) List() ([]*Checkpoint, error) {
	checkpointsDir := filepath.Join(m.basePath, ".checkpoints")
	
	entries, err := os.ReadDir(checkpointsDir)
	if err != nil {
		return nil, err
	}

	var checkpoints []*Checkpoint
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".tar.gz" {
			continue
		}

		id := entry.Name()[:len(entry.Name())-7] // Remove .tar.gz
		checkpoint, err := m.loadCheckpoint(id)
		if err != nil {
			continue
		}

		checkpoints = append(checkpoints, checkpoint)
	}

	return checkpoints, nil
}

// Delete deletes a checkpoint
func (m *Manager) Delete(checkpointID string) error {
	path := m.checkpointPath(checkpointID)
	return os.Remove(path)
}

// captureFiles captures all files in the workspace
func (m *Manager) captureFiles() ([]FileSnapshot, error) {
	var files []FileSnapshot

	err := filepath.Walk(m.basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip directories and certain files
		if info.IsDir() {
			if shouldSkipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldSkipFile(path) {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Calculate hash
		hash := sha256.Sum256(content)

		relPath, _ := filepath.Rel(m.basePath, path)
		
		files = append(files, FileSnapshot{
			Path:    relPath,
			Hash:    hex.EncodeToString(hash[:]),
			Mode:    info.Mode(),
			ModTime: info.ModTime(),
			Content: content,
		})

		return nil
	})

	return files, err
}

// saveCheckpoint saves a checkpoint to disk
func (m *Manager) saveCheckpoint(checkpoint *Checkpoint) error {
	path := m.checkpointPath(checkpoint.ID)

	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	gzWriter := gzip.NewWriter(file)
	defer gzWriter.Close()

	tarWriter := tar.NewWriter(gzWriter)
	defer tarWriter.Close()

	// Write each file
	for _, file := range checkpoint.Files {
		header := &tar.Header{
			Name:    file.Path,
			Mode:    int64(file.Mode),
			ModTime: file.ModTime,
			Size:    int64(len(file.Content)),
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if _, err := tarWriter.Write(file.Content); err != nil {
			return err
		}
	}

	return nil
}

// loadCheckpoint loads a checkpoint from disk
func (m *Manager) loadCheckpoint(checkpointID string) (*Checkpoint, error) {
	path := m.checkpointPath(checkpointID)

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	checkpoint := &Checkpoint{
		ID: checkpointID,
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		content := make([]byte, header.Size)
		if _, err := io.ReadFull(tarReader, content); err != nil {
			return nil, err
		}

		hash := sha256.Sum256(content)

		checkpoint.Files = append(checkpoint.Files, FileSnapshot{
			Path:    header.Name,
			Hash:    hex.EncodeToString(hash[:]),
			Mode:    os.FileMode(header.Mode),
			ModTime: header.ModTime,
			Content: content,
		})
	}

	return checkpoint, nil
}

// checkpointPath returns the path for a checkpoint
func (m *Manager) checkpointPath(id string) string {
	return filepath.Join(m.basePath, ".checkpoints", id+".tar.gz")
}

// generateID generates a unique checkpoint ID
func generateID() string {
	return fmt.Sprintf("checkpoint-%d", time.Now().UnixNano())
}

// shouldSkipDir returns true if directory should be skipped
func shouldSkipDir(path string) bool {
	skipDirs := []string{
		".git",
		"node_modules",
		"vendor",
		".checkpoints",
		"dist",
		"build",
		"target",
	}

	for _, dir := range skipDirs {
		if filepath.Base(path) == dir {
			return true
		}
	}
	return false
}

// shouldSkipFile returns true if file should be skipped
func shouldSkipFile(path string) bool {
	// Skip checkpoint files
	if filepath.Ext(path) == ".tar.gz" {
		return true
	}
	return false
}
