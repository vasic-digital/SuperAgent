# MinIO S3-Compatible Storage

This package provides MinIO/S3-compatible object storage integration for HelixAgent, enabling scalable file storage, multipart uploads, and pre-signed URL generation.

## Overview

The MinIO client provides a high-level API for object storage operations with support for:
- Bucket lifecycle management
- Multipart upload for large files
- Pre-signed URLs for secure temporary access
- Object versioning and lifecycle policies
- Connection health monitoring

## Architecture

```
                    +------------------+
                    |   Application    |
                    +--------+---------+
                             |
                             v
                    +--------+---------+
                    |   MinIO Client   |
                    | (this package)   |
                    +--------+---------+
                             |
                             v
                    +--------+---------+
                    |  minio-go SDK    |
                    +--------+---------+
                             |
              +--------------+--------------+
              |                             |
              v                             v
     +--------+--------+          +--------+--------+
     |   MinIO Server  |          |    AWS S3       |
     |   (local/cloud) |          | (compatible)    |
     +-----------------+          +-----------------+
```

## Client Setup

### Basic Configuration

```go
import "dev.helix.agent/internal/storage/minio"

config := &minio.Config{
    Endpoint:   "localhost:9000",
    AccessKey:  "minioadmin",
    SecretKey:  "minioadmin123",
    UseSSL:     false,
    Region:     "us-east-1",
}

client, err := minio.NewClient(config, logger)
if err != nil {
    log.Fatal(err)
}

err = client.Connect(ctx)
if err != nil {
    log.Fatal(err)
}
defer client.Close()
```

### Configuration Options

```go
type Config struct {
    // Connection settings
    Endpoint  string `json:"endpoint" yaml:"endpoint"`
    AccessKey string `json:"access_key" yaml:"access_key"`
    SecretKey string `json:"secret_key" yaml:"secret_key"`
    UseSSL    bool   `json:"use_ssl" yaml:"use_ssl"`
    Region    string `json:"region" yaml:"region"`

    // Connection options
    ConnectTimeout time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
    RequestTimeout time.Duration `json:"request_timeout" yaml:"request_timeout"`
    MaxRetries     int           `json:"max_retries" yaml:"max_retries"`

    // Upload settings
    PartSize          int64 `json:"part_size" yaml:"part_size"`
    ConcurrentUploads int   `json:"concurrent_uploads" yaml:"concurrent_uploads"`

    // Health check
    HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
}
```

### Default Configuration

```go
config := minio.DefaultConfig()
// Returns:
// - Endpoint: "localhost:9000"
// - AccessKey: "minioadmin"
// - SecretKey: "minioadmin123"
// - UseSSL: false
// - Region: "us-east-1"
// - ConnectTimeout: 30s
// - RequestTimeout: 60s
// - MaxRetries: 3
// - PartSize: 16MB
// - ConcurrentUploads: 4
// - HealthCheckInterval: 30s
```

## Bucket Operations

### Create Bucket

```go
bucketConfig := &minio.BucketConfig{
    Name:          "my-bucket",
    RetentionDays: 30,       // -1 for unlimited
    Versioning:    true,
    ObjectLocking: false,
    Public:        false,
}

err := client.CreateBucket(ctx, bucketConfig)
```

### Fluent Bucket Configuration

```go
bucketConfig := minio.DefaultBucketConfig("my-bucket").
    WithRetention(90).
    WithVersioning().
    WithObjectLocking()

err := client.CreateBucket(ctx, bucketConfig)
```

### List Buckets

```go
buckets, err := client.ListBuckets(ctx)
for _, bucket := range buckets {
    fmt.Printf("Bucket: %s, Created: %s\n", bucket.Name, bucket.CreationDate)
}
```

### Check Bucket Exists

```go
exists, err := client.BucketExists(ctx, "my-bucket")
if exists {
    fmt.Println("Bucket exists")
}
```

### Delete Bucket

```go
err := client.DeleteBucket(ctx, "my-bucket")
```

## Object Operations

### Upload Object

```go
file, err := os.Open("document.pdf")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

fileInfo, _ := file.Stat()
err = client.PutObject(ctx, "my-bucket", "documents/doc.pdf", file, fileInfo.Size(),
    minio.WithContentType("application/pdf"),
    minio.WithMetadata(map[string]string{
        "author":  "John Doe",
        "version": "1.0",
    }),
)
```

### Download Object

```go
object, err := client.GetObject(ctx, "my-bucket", "documents/doc.pdf")
if err != nil {
    log.Fatal(err)
}
defer object.Close()

// Read content
content, err := io.ReadAll(object)
```

### List Objects

```go
objects, err := client.ListObjects(ctx, "my-bucket", "documents/")
for _, obj := range objects {
    fmt.Printf("Key: %s, Size: %d, Modified: %s\n",
        obj.Key, obj.Size, obj.LastModified)
}
```

### Get Object Info

```go
info, err := client.StatObject(ctx, "my-bucket", "documents/doc.pdf")
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Size: %d, ContentType: %s, ETag: %s\n",
    info.Size, info.ContentType, info.ETag)
fmt.Printf("Metadata: %v\n", info.Metadata)
```

### Delete Object

```go
err := client.DeleteObject(ctx, "my-bucket", "documents/doc.pdf")
```

### Copy Object

```go
err := client.CopyObject(ctx,
    "source-bucket", "original.pdf",
    "dest-bucket", "copy.pdf",
)
```

## Multipart Upload

For large files (>5GB), the MinIO SDK automatically handles multipart uploads based on the configured `PartSize`.

### Configuration

```go
config := &minio.Config{
    PartSize:          64 * 1024 * 1024,  // 64MB parts
    ConcurrentUploads: 4,                  // 4 parallel uploads
}
```

### Manual Control

For very large files, you can control the part size:

```go
err := client.PutObject(ctx, "my-bucket", "large-file.zip", reader, size,
    minio.WithContentType("application/zip"),
)
// SDK automatically splits into parts based on PartSize config
```

### Best Practices

- Use part size of 64MB-128MB for files >1GB
- Enable concurrent uploads (4-8) for better throughput
- Monitor upload progress for long-running uploads

## Pre-signed URLs

### Generate Download URL

```go
url, err := client.GetPresignedURL(ctx, "my-bucket", "documents/doc.pdf", 24*time.Hour)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Download URL (valid 24h): %s\n", url)
```

### Generate Upload URL

```go
url, err := client.GetPresignedPutURL(ctx, "my-bucket", "uploads/file.zip", time.Hour)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Upload URL (valid 1h): %s\n", url)
```

### Security Considerations

- Set appropriate expiry times (shorter is more secure)
- URLs expose bucket/object paths
- Combine with bucket policies for additional control
- Monitor pre-signed URL usage

## Lifecycle Policies

### Set Lifecycle Rule

```go
rule := &minio.LifecycleRule{
    ID:                 "expire-temp-files",
    Prefix:             "temp/",
    Enabled:            true,
    ExpirationDays:     7,
    NoncurrentDays:     3,
    DeleteMarkerExpiry: false,
}

err := client.SetLifecycleRule(ctx, "my-bucket", rule)
```

### Fluent Rule Configuration

```go
rule := minio.DefaultLifecycleRule("cleanup-logs", 30).
    WithPrefix("logs/").
    WithNoncurrentExpiry(7)

err := client.SetLifecycleRule(ctx, "my-bucket", rule)
```

### Remove Lifecycle Rule

```go
err := client.RemoveLifecycleRule(ctx, "my-bucket", "expire-temp-files")
```

### Common Lifecycle Patterns

```go
// Expire temporary files after 7 days
tempRule := &minio.LifecycleRule{
    ID:             "expire-temp",
    Prefix:         "temp/",
    Enabled:        true,
    ExpirationDays: 7,
}

// Archive old logs after 90 days, delete after 365
logRule := &minio.LifecycleRule{
    ID:             "archive-logs",
    Prefix:         "logs/",
    Enabled:        true,
    ExpirationDays: 365,
    NoncurrentDays: 90,
}

// Clean up delete markers
cleanupRule := &minio.LifecycleRule{
    ID:                 "cleanup-markers",
    Prefix:             "",
    Enabled:            true,
    DeleteMarkerExpiry: true,
}
```

## Health Checks

### Connection Health

```go
err := client.HealthCheck(ctx)
if err != nil {
    log.Printf("MinIO unhealthy: %v", err)
}
```

### Connection Status

```go
if client.IsConnected() {
    fmt.Println("Connected to MinIO")
}
```

## Data Types

### ObjectInfo

```go
type ObjectInfo struct {
    Key          string
    Size         int64
    LastModified time.Time
    ContentType  string
    ETag         string
    Metadata     map[string]string
}
```

### BucketInfo

```go
type BucketInfo struct {
    Name         string
    CreationDate time.Time
}
```

### BucketConfig

```go
type BucketConfig struct {
    Name          string `json:"name" yaml:"name"`
    RetentionDays int    `json:"retention_days" yaml:"retention_days"`
    Versioning    bool   `json:"versioning" yaml:"versioning"`
    ObjectLocking bool   `json:"object_locking" yaml:"object_locking"`
    Public        bool   `json:"public" yaml:"public"`
}
```

### LifecycleRule

```go
type LifecycleRule struct {
    ID                 string `json:"id" yaml:"id"`
    Prefix             string `json:"prefix" yaml:"prefix"`
    Enabled            bool   `json:"enabled" yaml:"enabled"`
    ExpirationDays     int    `json:"expiration_days" yaml:"expiration_days"`
    NoncurrentDays     int    `json:"noncurrent_days" yaml:"noncurrent_days"`
    DeleteMarkerExpiry bool   `json:"delete_marker_expiry" yaml:"delete_marker_expiry"`
}
```

## Environment Configuration

```bash
# MinIO connection
export MINIO_ENDPOINT="localhost:9000"
export MINIO_ACCESS_KEY="minioadmin"
export MINIO_SECRET_KEY="minioadmin123"
export MINIO_USE_SSL="false"
export MINIO_REGION="us-east-1"

# Upload settings
export MINIO_PART_SIZE="16777216"  # 16MB
export MINIO_CONCURRENT_UPLOADS="4"
```

## Docker Compose

```yaml
services:
  minio:
    image: minio/minio:latest
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin123
    command: server /data --console-address ":9001"
    volumes:
      - minio-data:/data
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:9000/minio/health/live"]
      interval: 30s
      timeout: 20s
      retries: 3

volumes:
  minio-data:
```

## Testing

```bash
# Run unit tests
go test -v ./internal/storage/minio/...

# Run with coverage
go test -v -coverprofile=coverage.out ./internal/storage/minio/...

# Run benchmark tests
go test -bench=. ./internal/storage/minio/...

# Integration tests (requires MinIO)
make test-infra-start
go test -v -tags=integration ./internal/storage/minio/...
make test-infra-stop
```

## Files

| File | Description |
|------|-------------|
| `client.go` | Main MinIO client implementation |
| `config.go` | Configuration types and validation |
| `client_test.go` | Unit tests |
| `config_test.go` | Configuration tests |
| `minio_bench_test.go` | Benchmark tests |

## Related Packages

- `internal/storage/` - Storage abstraction layer
- `internal/lakehouse/` - Data lakehouse integration
- `internal/cache/` - Caching layer
