// Package minio provides adapter types for the digital.vasic.storage module.
// This adapter re-exports types and functions to maintain backward compatibility
// with code using the internal/storage/minio package.
package minio

import (
	"context"
	"io"
	"sync"
	"time"

	"digital.vasic.storage/pkg/object"
	"digital.vasic.storage/pkg/s3"
	"github.com/sirupsen/logrus"
)

// Config holds MinIO client configuration.
type Config struct {
	Endpoint          string
	AccessKey         string
	SecretKey         string
	UseSSL            bool
	Region            string
	ConnectTimeout    time.Duration
	RequestTimeout    time.Duration
	MaxRetries        int
	PartSize          int64
	ConcurrentUploads int
}

// BucketConfig holds bucket creation configuration.
type BucketConfig struct {
	Name          string
	Region        string
	Versioning    bool
	ObjectLocking bool
	RetentionDays int
}

// DefaultBucketConfig creates a default bucket configuration.
func DefaultBucketConfig(name string) *BucketConfig {
	return &BucketConfig{
		Name: name,
	}
}

// ObjectInfo holds information about an object.
type ObjectInfo struct {
	Key          string
	Size         int64
	ETag         string
	ContentType  string
	LastModified time.Time
	Metadata     map[string]string
}

// BucketInfo holds information about a bucket.
type BucketInfo struct {
	Name         string
	CreationDate time.Time
}

// LifecycleRule configures object lifecycle management.
type LifecycleRule struct {
	ID             string
	Prefix         string
	Enabled        bool
	ExpirationDays int
	NoncurrentDays int
}

// DefaultLifecycleRule creates a default lifecycle rule.
func DefaultLifecycleRule(id string, expirationDays int) *LifecycleRule {
	return &LifecycleRule{
		ID:             id,
		Enabled:        true,
		ExpirationDays: expirationDays,
	}
}

// PutOption is a functional option for PutObject.
type PutOption func(*putOptions)

type putOptions struct {
	contentType string
	metadata    map[string]string
}

// WithContentType sets the content type for an object.
func WithContentType(ct string) PutOption {
	return func(o *putOptions) {
		o.contentType = ct
	}
}

// WithMetadata sets metadata for an object.
func WithMetadata(m map[string]string) PutOption {
	return func(o *putOptions) {
		o.metadata = m
	}
}

// Client wraps the extracted module's S3 client to provide backward compatibility.
type Client struct {
	extClient *s3.Client
	logger    *logrus.Logger
	mu        sync.RWMutex
	connected bool
	config    *Config
}

// NewClient creates a new MinIO client adapter.
func NewClient(config *Config, logger *logrus.Logger) (*Client, error) {
	if logger == nil {
		logger = logrus.New()
	}

	extConfig := &s3.Config{
		Endpoint:       config.Endpoint,
		AccessKey:      config.AccessKey,
		SecretKey:      config.SecretKey,
		UseSSL:         config.UseSSL,
		Region:         config.Region,
		ConnectTimeout: config.ConnectTimeout,
		RequestTimeout: config.RequestTimeout,
		MaxRetries:     config.MaxRetries,
		PartSize:       config.PartSize,
	}

	extClient, err := s3.NewClient(extConfig, logger)
	if err != nil {
		return nil, err
	}

	return &Client{
		extClient: extClient,
		logger:    logger,
		config:    config,
	}, nil
}

// Connect connects to MinIO.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.extClient.Connect(ctx); err != nil {
		return err
	}
	c.connected = true
	return nil
}

// Close closes the client connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = false
	return c.extClient.Close()
}

// IsConnected returns whether the client is connected.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// HealthCheck checks the health of MinIO.
func (c *Client) HealthCheck(ctx context.Context) error {
	return c.extClient.HealthCheck(ctx)
}

// CreateBucket creates a new bucket.
func (c *Client) CreateBucket(ctx context.Context, config *BucketConfig) error {
	extConfig := object.BucketConfig{
		Name:          config.Name,
		Versioning:    config.Versioning,
		ObjectLocking: config.ObjectLocking,
		RetentionDays: config.RetentionDays,
	}
	return c.extClient.CreateBucket(ctx, extConfig)
}

// DeleteBucket deletes a bucket.
func (c *Client) DeleteBucket(ctx context.Context, name string) error {
	return c.extClient.DeleteBucket(ctx, name)
}

// BucketExists checks if a bucket exists.
func (c *Client) BucketExists(ctx context.Context, name string) (bool, error) {
	return c.extClient.BucketExists(ctx, name)
}

// ListBuckets lists all buckets.
func (c *Client) ListBuckets(ctx context.Context) ([]BucketInfo, error) {
	buckets, err := c.extClient.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]BucketInfo, len(buckets))
	for i, b := range buckets {
		result[i] = BucketInfo{
			Name:         b.Name,
			CreationDate: b.CreationDate,
		}
	}
	return result, nil
}

// PutObject uploads an object to a bucket.
func (c *Client) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, opts ...PutOption) error {
	o := &putOptions{}
	for _, opt := range opts {
		opt(o)
	}

	var putOpts []object.PutOption
	if o.contentType != "" {
		putOpts = append(putOpts, object.WithContentType(o.contentType))
	}
	if o.metadata != nil {
		putOpts = append(putOpts, object.WithMetadata(o.metadata))
	}

	return c.extClient.PutObject(ctx, bucket, key, reader, size, putOpts...)
}

// GetObject downloads an object from a bucket.
func (c *Client) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return c.extClient.GetObject(ctx, bucket, key)
}

// DeleteObject deletes an object from a bucket.
func (c *Client) DeleteObject(ctx context.Context, bucket, key string) error {
	return c.extClient.DeleteObject(ctx, bucket, key)
}

// ListObjects lists objects in a bucket with optional prefix.
func (c *Client) ListObjects(ctx context.Context, bucket, prefix string) ([]ObjectInfo, error) {
	objects, err := c.extClient.ListObjects(ctx, bucket, prefix)
	if err != nil {
		return nil, err
	}
	result := make([]ObjectInfo, len(objects))
	for i, obj := range objects {
		result[i] = ObjectInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			ETag:         obj.ETag,
			ContentType:  obj.ContentType,
			LastModified: obj.LastModified,
		}
	}
	return result, nil
}

// StatObject returns information about an object.
func (c *Client) StatObject(ctx context.Context, bucket, key string) (*ObjectInfo, error) {
	info, err := c.extClient.StatObject(ctx, bucket, key)
	if err != nil {
		return nil, err
	}
	return &ObjectInfo{
		Key:          info.Key,
		Size:         info.Size,
		ETag:         info.ETag,
		ContentType:  info.ContentType,
		LastModified: info.LastModified,
	}, nil
}

// CopyObject copies an object from source to destination.
func (c *Client) CopyObject(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error {
	src := object.ObjectRef{Bucket: srcBucket, Key: srcKey}
	dst := object.ObjectRef{Bucket: dstBucket, Key: dstKey}
	return c.extClient.CopyObject(ctx, src, dst)
}

// GetPresignedURL generates a presigned URL for downloading an object.
func (c *Client) GetPresignedURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	return c.extClient.GetPresignedURL(ctx, bucket, key, expiry)
}

// GetPresignedPutURL generates a presigned URL for uploading an object.
func (c *Client) GetPresignedPutURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	return c.extClient.GetPresignedPutURL(ctx, bucket, key, expiry)
}

// SetLifecycleRule sets a lifecycle rule on a bucket.
func (c *Client) SetLifecycleRule(ctx context.Context, bucket string, rule *LifecycleRule) error {
	extRule := &s3.LifecycleRule{
		ID:             rule.ID,
		Prefix:         rule.Prefix,
		Enabled:        rule.Enabled,
		ExpirationDays: rule.ExpirationDays,
		NoncurrentDays: rule.NoncurrentDays,
	}
	return c.extClient.SetLifecycleRule(ctx, bucket, extRule)
}

// RemoveLifecycleRule removes a lifecycle rule from a bucket.
func (c *Client) RemoveLifecycleRule(ctx context.Context, bucket, ruleID string) error {
	return c.extClient.RemoveLifecycleRule(ctx, bucket, ruleID)
}
