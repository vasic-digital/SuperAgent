// Package adapters provides MCP server adapters.
// This file implements the AWS S3 MCP server adapter.
package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

// AWSS3Config configures the AWS S3 adapter.
type AWSS3Config struct {
	Region          string        `json:"region"`
	AccessKeyID     string        `json:"access_key_id"`
	SecretAccessKey string        `json:"secret_access_key"`
	SessionToken    string        `json:"session_token,omitempty"`
	Endpoint        string        `json:"endpoint,omitempty"` // For S3-compatible services
	Timeout         time.Duration `json:"timeout"`
}

// DefaultAWSS3Config returns default configuration.
func DefaultAWSS3Config() AWSS3Config {
	return AWSS3Config{
		Region:  "us-east-1",
		Timeout: 60 * time.Second,
	}
}

// AWSS3Adapter implements the AWS S3 MCP server.
type AWSS3Adapter struct {
	config AWSS3Config
	client S3Client
}

// S3Client interface for S3 operations (for testability).
type S3Client interface {
	ListBuckets(ctx context.Context) ([]S3Bucket, error)
	ListObjects(ctx context.Context, bucket, prefix string, maxKeys int) ([]S3Object, error)
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	PutObject(ctx context.Context, bucket, key string, body io.Reader, contentType string) error
	DeleteObject(ctx context.Context, bucket, key string) error
	CopyObject(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error
	GetObjectMetadata(ctx context.Context, bucket, key string) (*S3ObjectMetadata, error)
	CreateBucket(ctx context.Context, bucket string) error
	DeleteBucket(ctx context.Context, bucket string) error
}

// S3Bucket represents an S3 bucket.
type S3Bucket struct {
	Name         string    `json:"name"`
	CreationDate time.Time `json:"creation_date"`
}

// S3Object represents an S3 object.
type S3Object struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ETag         string    `json:"etag"`
	StorageClass string    `json:"storage_class"`
}

// S3ObjectMetadata represents S3 object metadata.
type S3ObjectMetadata struct {
	ContentType     string            `json:"content_type"`
	ContentLength   int64             `json:"content_length"`
	LastModified    time.Time         `json:"last_modified"`
	ETag            string            `json:"etag"`
	Metadata        map[string]string `json:"metadata"`
	StorageClass    string            `json:"storage_class"`
	VersionID       string            `json:"version_id,omitempty"`
}

// NewAWSS3Adapter creates a new AWS S3 adapter.
func NewAWSS3Adapter(config AWSS3Config, client S3Client) *AWSS3Adapter {
	return &AWSS3Adapter{
		config: config,
		client: client,
	}
}

// GetServerInfo returns server information.
func (a *AWSS3Adapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "aws-s3",
		Version:     "1.0.0",
		Description: "AWS S3 storage operations including list, get, put, delete, and copy objects",
		Capabilities: []string{
			"bucket_management",
			"object_management",
			"metadata",
			"copy",
		},
	}
}

// ListTools returns available tools.
func (a *AWSS3Adapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "s3_list_buckets",
			Description: "List all S3 buckets",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "s3_list_objects",
			Description: "List objects in an S3 bucket",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"bucket": map[string]interface{}{
						"type":        "string",
						"description": "Bucket name",
					},
					"prefix": map[string]interface{}{
						"type":        "string",
						"description": "Object key prefix filter",
						"default":     "",
					},
					"max_keys": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of objects to return",
						"default":     1000,
					},
				},
				"required": []string{"bucket"},
			},
		},
		{
			Name:        "s3_get_object",
			Description: "Get an object from S3",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"bucket": map[string]interface{}{
						"type":        "string",
						"description": "Bucket name",
					},
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Object key",
					},
				},
				"required": []string{"bucket", "key"},
			},
		},
		{
			Name:        "s3_put_object",
			Description: "Upload an object to S3",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"bucket": map[string]interface{}{
						"type":        "string",
						"description": "Bucket name",
					},
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Object key",
					},
					"content": map[string]interface{}{
						"type":        "string",
						"description": "Object content",
					},
					"content_type": map[string]interface{}{
						"type":        "string",
						"description": "Content MIME type",
						"default":     "application/octet-stream",
					},
				},
				"required": []string{"bucket", "key", "content"},
			},
		},
		{
			Name:        "s3_delete_object",
			Description: "Delete an object from S3",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"bucket": map[string]interface{}{
						"type":        "string",
						"description": "Bucket name",
					},
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Object key",
					},
				},
				"required": []string{"bucket", "key"},
			},
		},
		{
			Name:        "s3_copy_object",
			Description: "Copy an object within or between S3 buckets",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"source_bucket": map[string]interface{}{
						"type":        "string",
						"description": "Source bucket name",
					},
					"source_key": map[string]interface{}{
						"type":        "string",
						"description": "Source object key",
					},
					"destination_bucket": map[string]interface{}{
						"type":        "string",
						"description": "Destination bucket name",
					},
					"destination_key": map[string]interface{}{
						"type":        "string",
						"description": "Destination object key",
					},
				},
				"required": []string{"source_bucket", "source_key", "destination_bucket", "destination_key"},
			},
		},
		{
			Name:        "s3_get_object_metadata",
			Description: "Get metadata for an S3 object",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"bucket": map[string]interface{}{
						"type":        "string",
						"description": "Bucket name",
					},
					"key": map[string]interface{}{
						"type":        "string",
						"description": "Object key",
					},
				},
				"required": []string{"bucket", "key"},
			},
		},
		{
			Name:        "s3_create_bucket",
			Description: "Create a new S3 bucket",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"bucket": map[string]interface{}{
						"type":        "string",
						"description": "Bucket name",
					},
				},
				"required": []string{"bucket"},
			},
		},
		{
			Name:        "s3_delete_bucket",
			Description: "Delete an empty S3 bucket",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"bucket": map[string]interface{}{
						"type":        "string",
						"description": "Bucket name",
					},
				},
				"required": []string{"bucket"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *AWSS3Adapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "s3_list_buckets":
		return a.listBuckets(ctx)
	case "s3_list_objects":
		return a.listObjects(ctx, args)
	case "s3_get_object":
		return a.getObject(ctx, args)
	case "s3_put_object":
		return a.putObject(ctx, args)
	case "s3_delete_object":
		return a.deleteObject(ctx, args)
	case "s3_copy_object":
		return a.copyObject(ctx, args)
	case "s3_get_object_metadata":
		return a.getObjectMetadata(ctx, args)
	case "s3_create_bucket":
		return a.createBucket(ctx, args)
	case "s3_delete_bucket":
		return a.deleteBucket(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *AWSS3Adapter) listBuckets(ctx context.Context) (*ToolResult, error) {
	buckets, err := a.client.ListBuckets(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d buckets:\n\n", len(buckets)))

	for _, bucket := range buckets {
		sb.WriteString(fmt.Sprintf("- %s (created: %s)\n", bucket.Name, bucket.CreationDate.Format(time.RFC3339)))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *AWSS3Adapter) listObjects(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	bucket, _ := args["bucket"].(string)
	prefix, _ := args["prefix"].(string)
	maxKeys := getIntArg(args, "max_keys", 1000)

	objects, err := a.client.ListObjects(ctx, bucket, prefix, maxKeys)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Objects in bucket '%s'", bucket))
	if prefix != "" {
		sb.WriteString(fmt.Sprintf(" with prefix '%s'", prefix))
	}
	sb.WriteString(fmt.Sprintf(" (%d objects):\n\n", len(objects)))

	for _, obj := range objects {
		sb.WriteString(fmt.Sprintf("- %s\n", obj.Key))
		sb.WriteString(fmt.Sprintf("  Size: %d bytes, Modified: %s\n", obj.Size, obj.LastModified.Format(time.RFC3339)))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *AWSS3Adapter) getObject(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	bucket, _ := args["bucket"].(string)
	key, _ := args["key"].(string)

	reader, err := a.client.GetObject(ctx, bucket, key)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}
	defer reader.Close()

	content, err := io.ReadAll(reader)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(content)}},
	}, nil
}

func (a *AWSS3Adapter) putObject(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	bucket, _ := args["bucket"].(string)
	key, _ := args["key"].(string)
	content, _ := args["content"].(string)
	contentType, _ := args["content_type"].(string)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	err := a.client.PutObject(ctx, bucket, key, strings.NewReader(content), contentType)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Successfully uploaded object '%s' to bucket '%s'", key, bucket)}},
	}, nil
}

func (a *AWSS3Adapter) deleteObject(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	bucket, _ := args["bucket"].(string)
	key, _ := args["key"].(string)

	err := a.client.DeleteObject(ctx, bucket, key)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Successfully deleted object '%s' from bucket '%s'", key, bucket)}},
	}, nil
}

func (a *AWSS3Adapter) copyObject(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	srcBucket, _ := args["source_bucket"].(string)
	srcKey, _ := args["source_key"].(string)
	dstBucket, _ := args["destination_bucket"].(string)
	dstKey, _ := args["destination_key"].(string)

	err := a.client.CopyObject(ctx, srcBucket, srcKey, dstBucket, dstKey)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Successfully copied '%s/%s' to '%s/%s'", srcBucket, srcKey, dstBucket, dstKey)}},
	}, nil
}

func (a *AWSS3Adapter) getObjectMetadata(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	bucket, _ := args["bucket"].(string)
	key, _ := args["key"].(string)

	metadata, err := a.client.GetObjectMetadata(ctx, bucket, key)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	data, _ := json.MarshalIndent(metadata, "", "  ")
	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(data)}},
	}, nil
}

func (a *AWSS3Adapter) createBucket(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	bucket, _ := args["bucket"].(string)

	err := a.client.CreateBucket(ctx, bucket)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Successfully created bucket '%s'", bucket)}},
	}, nil
}

func (a *AWSS3Adapter) deleteBucket(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	bucket, _ := args["bucket"].(string)

	err := a.client.DeleteBucket(ctx, bucket)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Successfully deleted bucket '%s'", bucket)}},
	}, nil
}
