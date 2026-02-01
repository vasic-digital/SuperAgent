package bigdata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
)

// DataLakeClient manages data lake operations (S3/MinIO)
type DataLakeClient struct {
	client     *minio.Client
	bucketName string
	logger     *logrus.Logger
}

// ConversationArchive represents archived conversation data
type ConversationArchive struct {
	ConversationID string                 `json:"conversation_id"`
	UserID         string                 `json:"user_id"`
	SessionID      string                 `json:"session_id"`
	StartedAt      time.Time              `json:"started_at"`
	CompletedAt    time.Time              `json:"completed_at"`
	MessageCount   int                    `json:"message_count"`
	EntityCount    int                    `json:"entity_count"`
	TotalTokens    int64                  `json:"total_tokens"`
	Messages       []ArchivedMessage      `json:"messages"`
	Entities       []ArchivedEntity       `json:"entities"`
	DebateRounds   []ArchivedDebateRound  `json:"debate_rounds,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ArchivedMessage represents a single message in archive
type ArchivedMessage struct {
	MessageID string                 `json:"message_id"`
	Role      string                 `json:"role"`
	Content   string                 `json:"content"`
	Model     string                 `json:"model,omitempty"`
	Tokens    int                    `json:"tokens"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ArchivedEntity represents an entity in archive
type ArchivedEntity struct {
	EntityID   string                 `json:"entity_id"`
	Type       string                 `json:"entity_type"`
	Name       string                 `json:"name"`
	Value      string                 `json:"value"`
	Confidence float64                `json:"confidence"`
	FirstSeen  time.Time              `json:"first_seen"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// ArchivedDebateRound represents a debate round in archive
type ArchivedDebateRound struct {
	Round        int                    `json:"round"`
	Position     string                 `json:"position"`
	Provider     string                 `json:"provider"`
	Model        string                 `json:"model"`
	Response     string                 `json:"response"`
	Confidence   float64                `json:"confidence"`
	ResponseTime int64                  `json:"response_time_ms"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// DataLakeConfig defines data lake configuration
type DataLakeConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Region          string
	UseSSL          bool
}

// NewDataLakeClient creates a new data lake client
func NewDataLakeClient(config DataLakeConfig, logger *logrus.Logger) (*DataLakeClient, error) {
	if logger == nil {
		logger = logrus.New()
	}

	// Initialize MinIO client
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
		Region: config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	dlc := &DataLakeClient{
		client:     client,
		bucketName: config.BucketName,
		logger:     logger,
	}

	// Ensure bucket exists
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, config.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		logger.WithField("bucket", config.BucketName).Info("Creating data lake bucket")
		err = client.MakeBucket(ctx, config.BucketName, minio.MakeBucketOptions{
			Region: config.Region,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	logger.WithFields(logrus.Fields{
		"endpoint": config.Endpoint,
		"bucket":   config.BucketName,
	}).Info("Data lake client initialized")

	return dlc, nil
}

// ArchiveConversation archives a conversation to the data lake
func (dlc *DataLakeClient) ArchiveConversation(
	ctx context.Context,
	archive ConversationArchive,
) error {
	// Convert to JSON
	data, err := json.Marshal(archive)
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	// Generate Hive-style partition path
	partition := dlc.formatPartition(archive.StartedAt)
	objectKey := fmt.Sprintf(
		"conversations/%s/conversation_%s.json",
		partition,
		archive.ConversationID,
	)

	// Upload to data lake
	_, err = dlc.client.PutObject(
		ctx,
		dlc.bucketName,
		objectKey,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{
			ContentType: "application/json",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upload conversation: %w", err)
	}

	dlc.logger.WithFields(logrus.Fields{
		"conversation_id": archive.ConversationID,
		"object_key":      objectKey,
		"size_bytes":      len(data),
	}).Info("Conversation archived to data lake")

	return nil
}

// GetConversation retrieves an archived conversation
func (dlc *DataLakeClient) GetConversation(
	ctx context.Context,
	conversationID string,
	timestamp time.Time,
) (*ConversationArchive, error) {
	// Generate object key
	partition := dlc.formatPartition(timestamp)
	objectKey := fmt.Sprintf(
		"conversations/%s/conversation_%s.json",
		partition,
		conversationID,
	)

	// Download from data lake
	object, err := dlc.client.GetObject(ctx, dlc.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	defer func() { _ = object.Close() }()

	// Read and unmarshal
	data, err := io.ReadAll(object)
	if err != nil {
		return nil, fmt.Errorf("failed to read conversation data: %w", err)
	}

	var archive ConversationArchive
	if err := json.Unmarshal(data, &archive); err != nil {
		return nil, fmt.Errorf("failed to unmarshal conversation: %w", err)
	}

	return &archive, nil
}

// ListConversations lists conversations in a date range
func (dlc *DataLakeClient) ListConversations(
	ctx context.Context,
	startDate, endDate time.Time,
) ([]string, error) {
	var conversationIDs []string

	// Iterate through date range
	currentDate := startDate
	for !currentDate.After(endDate) {
		partition := dlc.formatPartition(currentDate)
		prefix := fmt.Sprintf("conversations/%s/", partition)

		// List objects with prefix
		objectCh := dlc.client.ListObjects(ctx, dlc.bucketName, minio.ListObjectsOptions{
			Prefix:    prefix,
			Recursive: false,
		})

		for object := range objectCh {
			if object.Err != nil {
				dlc.logger.WithError(object.Err).Warn("Error listing object")
				continue
			}

			// Extract conversation ID from filename
			filename := filepath.Base(object.Key)
			if len(filename) > 14 { // "conversation_" prefix
				conversationID := filename[14 : len(filename)-5] // Remove ".json"
				conversationIDs = append(conversationIDs, conversationID)
			}
		}

		currentDate = currentDate.AddDate(0, 0, 1) // Next day
	}

	return conversationIDs, nil
}

// PathExists checks if a path exists in the data lake
func (dlc *DataLakeClient) PathExists(ctx context.Context, path string) (bool, error) {
	// Try to stat the object
	_, err := dlc.client.StatObject(ctx, dlc.bucketName, path, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check path existence: %w", err)
	}
	return true, nil
}

// DeleteConversation deletes an archived conversation
func (dlc *DataLakeClient) DeleteConversation(
	ctx context.Context,
	conversationID string,
	timestamp time.Time,
) error {
	partition := dlc.formatPartition(timestamp)
	objectKey := fmt.Sprintf(
		"conversations/%s/conversation_%s.json",
		partition,
		conversationID,
	)

	err := dlc.client.RemoveObject(ctx, dlc.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete conversation: %w", err)
	}

	dlc.logger.WithFields(logrus.Fields{
		"conversation_id": conversationID,
		"object_key":      objectKey,
	}).Info("Conversation deleted from data lake")

	return nil
}

// ArchiveDebateResults archives debate results to data lake
func (dlc *DataLakeClient) ArchiveDebateResults(
	ctx context.Context,
	debateID string,
	timestamp time.Time,
	results map[string]interface{},
) error {
	data, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal debate results: %w", err)
	}

	partition := dlc.formatPartition(timestamp)
	objectKey := fmt.Sprintf(
		"debates/%s/debate_%s.json",
		partition,
		debateID,
	)

	_, err = dlc.client.PutObject(
		ctx,
		dlc.bucketName,
		objectKey,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{
			ContentType: "application/json",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upload debate results: %w", err)
	}

	return nil
}

// ArchiveEntities archives entity snapshots to data lake
func (dlc *DataLakeClient) ArchiveEntities(
	ctx context.Context,
	timestamp time.Time,
	entities []ArchivedEntity,
) error {
	data, err := json.Marshal(entities)
	if err != nil {
		return fmt.Errorf("failed to marshal entities: %w", err)
	}

	partition := dlc.formatPartition(timestamp)
	objectKey := fmt.Sprintf(
		"entities/%s/entities_snapshot_%d.json",
		partition,
		timestamp.UnixNano(),
	)

	_, err = dlc.client.PutObject(
		ctx,
		dlc.bucketName,
		objectKey,
		bytes.NewReader(data),
		int64(len(data)),
		minio.PutObjectOptions{
			ContentType: "application/json",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to upload entities: %w", err)
	}

	return nil
}

// GetStorageStats retrieves storage statistics
func (dlc *DataLakeClient) GetStorageStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count objects and calculate total size
	var totalObjects int64
	var totalSize int64

	objectCh := dlc.client.ListObjects(ctx, dlc.bucketName, minio.ListObjectsOptions{
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			continue
		}
		totalObjects++
		totalSize += object.Size
	}

	stats["total_objects"] = totalObjects
	stats["total_size_bytes"] = totalSize
	stats["total_size_gb"] = float64(totalSize) / (1024 * 1024 * 1024)
	stats["bucket_name"] = dlc.bucketName

	return stats, nil
}

// formatPartition formats a date into Hive-style partition path
func (dlc *DataLakeClient) formatPartition(t time.Time) string {
	return fmt.Sprintf(
		"year=%d/month=%02d/day=%02d",
		t.Year(),
		t.Month(),
		t.Day(),
	)
}

// ListDirectories lists all directory-like prefixes under a given path
func (dlc *DataLakeClient) ListDirectories(ctx context.Context, path string) ([]string, error) {
	// Ensure path ends with /
	if !strings.HasSuffix(path, "/") && path != "" {
		path = path + "/"
	}

	var directories []string
	seen := make(map[string]bool)

	// List objects with prefix
	objectCh := dlc.client.ListObjects(ctx, dlc.bucketName, minio.ListObjectsOptions{
		Prefix:    path,
		Recursive: false,
	})

	for object := range objectCh {
		if object.Err != nil {
			continue
		}

		// Extract directory from object key
		key := object.Key
		if strings.HasPrefix(key, path) {
			relPath := strings.TrimPrefix(key, path)
			// Get first component (directory name)
			if idx := strings.Index(relPath, "/"); idx > 0 {
				dirName := relPath[:idx]
				fullDir := path + dirName + "/"
				if !seen[fullDir] {
					seen[fullDir] = true
					directories = append(directories, fullDir)
				}
			}
		}
	}

	return directories, nil
}

// GetMetadata retrieves metadata for a path (object or pseudo-directory)
func (dlc *DataLakeClient) GetMetadata(ctx context.Context, path string) (*struct {
	ModTime time.Time
	Size    int64
	IsDir   bool
}, error) {
	// Remove trailing slash for object metadata
	objectPath := strings.TrimSuffix(path, "/")

	// Try to get object stats
	stat, err := dlc.client.StatObject(ctx, dlc.bucketName, objectPath, minio.StatObjectOptions{})
	if err != nil {
		// If object doesn't exist, treat as directory
		// For directories, we need to check if any objects exist with this prefix
		// For simplicity, return current time as mod time
		return &struct {
			ModTime time.Time
			Size    int64
			IsDir   bool
		}{
			ModTime: time.Now(),
			Size:    0,
			IsDir:   true,
		}, nil
	}

	return &struct {
		ModTime time.Time
		Size    int64
		IsDir   bool
	}{
		ModTime: stat.LastModified,
		Size:    stat.Size,
		IsDir:   strings.HasSuffix(path, "/"),
	}, nil
}

// DeletePath deletes an object or directory (recursive if needed)
func (dlc *DataLakeClient) DeletePath(ctx context.Context, path string, recursive bool) error {
	// Remove trailing slash for object operations
	path = strings.TrimSuffix(path, "/")

	if recursive {
		// List all objects with prefix and delete them
		objectCh := dlc.client.ListObjects(ctx, dlc.bucketName, minio.ListObjectsOptions{
			Prefix:    path + "/",
			Recursive: true,
		})

		var objects []minio.ObjectInfo
		for object := range objectCh {
			if object.Err != nil {
				continue
			}
			objects = append(objects, object)
		}

		// Delete all objects
		for _, obj := range objects {
			err := dlc.client.RemoveObject(ctx, dlc.bucketName, obj.Key, minio.RemoveObjectOptions{})
			if err != nil {
				return fmt.Errorf("failed to delete object %s: %w", obj.Key, err)
			}
		}

		dlc.logger.WithField("path", path).Debugf("Deleted %d objects recursively", len(objects))
	} else {
		// Try to delete as single object
		err := dlc.client.RemoveObject(ctx, dlc.bucketName, path, minio.RemoveObjectOptions{})
		if err != nil {
			return fmt.Errorf("failed to delete object %s: %w", path, err)
		}
	}

	return nil
}

// Close closes the data lake client
func (dlc *DataLakeClient) Close() error {
	// MinIO client doesn't require explicit closing
	return nil
}
