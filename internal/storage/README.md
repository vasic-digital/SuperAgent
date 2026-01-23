# Storage Package

The storage package provides object storage integration for HelixAgent, supporting MinIO and S3-compatible storage systems.

## Overview

This package implements a client for MinIO/S3 object storage, enabling HelixAgent to store and retrieve files, models, embeddings, and other binary data. It supports bucket management, object operations, and lifecycle policies.

## Key Components

### Client

```go
type Client struct {
    config      *Config
    minioClient *minio.Client
    logger      *logrus.Logger
    connected   bool
}
```

The main client for interacting with MinIO/S3 storage.

### Configuration

```go
type Config struct {
    Endpoint        string        `json:"endpoint"`
    AccessKey       string        `json:"access_key"`
    SecretKey       string        `json:"secret_key"`
    Region          string        `json:"region"`
    UseSSL          bool          `json:"use_ssl"`
    BucketName      string        `json:"bucket_name"`
    AutoCreateBucket bool         `json:"auto_create_bucket"`
    Timeout         time.Duration `json:"timeout"`
}
```

## Features

- **Bucket Management**: Create, delete, list buckets
- **Object Operations**: Upload, download, delete, copy objects
- **Multipart Upload**: Large file uploads with resumable support
- **Lifecycle Policies**: Automatic object expiration
- **Pre-signed URLs**: Secure temporary access URLs
- **Metadata**: Custom object metadata

## Subpackages

### internal/storage/minio

MinIO/S3 client implementation:
- `client.go` - Main client implementation
- `config.go` - Configuration management

## Object Operations

```go
// Upload an object
err := client.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
    ContentType: "application/json",
})

// Download an object
object, err := client.GetObject(ctx, bucketName, objectName)

// List objects
objects := client.ListObjects(ctx, bucketName, minio.ListObjectsOptions{
    Prefix:    "models/",
    Recursive: true,
})

// Delete an object
err := client.RemoveObject(ctx, bucketName, objectName)

// Generate pre-signed URL
url, err := client.PresignedGetObject(ctx, bucketName, objectName, 24*time.Hour, nil)
```

## Usage

```go
import "dev.helix.agent/internal/storage/minio"

config := &minio.Config{
    Endpoint:         "localhost:9000",
    AccessKey:        "minioadmin",
    SecretKey:        "minioadmin",
    UseSSL:           false,
    BucketName:       "helixagent",
    AutoCreateBucket: true,
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

## Lifecycle Policies

```go
// Set object expiration
lifecycle := lifecycle.NewConfiguration()
lifecycle.Rules = []lifecycle.Rule{
    {
        ID:     "expire-temp-files",
        Status: "Enabled",
        Expiration: lifecycle.Expiration{
            Days: 7,
        },
        RuleFilter: lifecycle.Filter{
            Prefix: "temp/",
        },
    },
}
err := client.SetBucketLifecycle(ctx, bucketName, lifecycle)
```

## Testing

```bash
go test -v ./internal/storage/...
```

## Related Packages

- `internal/lakehouse` - Data lakehouse integration
- `internal/cache` - Caching layer
