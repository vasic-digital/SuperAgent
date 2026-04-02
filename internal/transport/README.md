# HTTP/3 Transport Package

This package provides HTTP/3 client and server implementations with fallback to HTTP/2 and HTTP/1.1, plus Brotli compression support.

## Overview

The transport package implements:

- **HTTP/3 Client** (`http3_client.go`): Full-featured HTTP client with HTTP/3 (QUIC) support, automatic fallback, retry logic, and Brotli decompression
- **HTTP/3 Server** (`http3.go`): Server implementation supporting HTTP/3, HTTP/2, and HTTP/1.1
- **Compression Support**: Brotli and Gzip compression/decompression

## HTTP/3 Client

### Basic Usage

```go
import "dev.helix.agent/internal/transport"

// Create client with default configuration
client := transport.NewHTTP3Client(nil).HTTPClient()

// Simple GET request
ctx := context.Background()
resp, err := client.Get(ctx, "https://api.example.com/data")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

// POST with JSON
jsonData := []byte(`{"key":"value"}`)
resp, err = client.Post(url, "application/json", bytes.NewReader(jsonData))
```

### Custom Configuration

```go
config := &transport.HTTP3ClientConfig{
    EnableHTTP3:     true,                    // Enable HTTP/3 (QUIC)
    EnableHTTP2:     true,                    // Enable HTTP/2 fallback
    EnableBrotli:    true,                    // Enable Brotli decompression
    Timeout:         120 * time.Second,       // Request timeout
    DialTimeout:     30 * time.Second,        // Connection timeout
    IdleConnTimeout: 90 * time.Second,        // Idle connection timeout
    MaxIdleConns:    100,                     // Max idle connections
    MaxRetries:      3,                       // Max retry attempts
    RetryDelay:      1 * time.Second,         // Initial retry delay
    MaxRetryDelay:   30 * time.Second,        // Max retry delay
    RetryMultiplier: 2.0,                     // Exponential backoff multiplier
}

client := transport.NewHTTP3Client(config).HTTPClient()
```

### Using with LLM Providers

```go
// Replace standard http.Client in providers
client := transport.NewHTTP3Client(nil).HTTPClient()

// Use in provider initialization
provider := &Provider{
    apiKey:     apiKey,
    baseURL:    baseURL,
    httpClient: client,  // HTTP/3 enabled client
}
```

### Advanced Usage with HTTP3Client

```go
// Get the HTTP3Client wrapper for more control
h3Client := transport.NewHTTP3Client(nil)

// Access the underlying http.Client
httpClient := h3Client.HTTPClient()

// Use helper methods
ctx := context.Background()
resp, err := h3Client.Get(ctx, "https://api.example.com/data")
resp, err = h3Client.PostJSON(ctx, url, jsonBytes)

// Get configuration
config := h3Client.GetConfig()

// Close connections when done
defer h3Client.Close()
```

### Global Client Instance

```go
// Get global singleton instance
client := transport.GetGlobalHTTP3Client().HTTPClient()

// Set custom global instance
transport.SetGlobalHTTP3Client(myCustomClient)

// Reset to defaults
transport.ResetGlobalHTTP3Client()
```

## HTTP/3 Round Tripper

For use with standard `http.Client`:

```go
roundTripper := transport.NewHTTP3RoundTripper(true) // enable HTTP/3
client := &http.Client{
    Transport: roundTripper,
    Timeout:   30 * time.Second,
}
```

## Features

### Protocol Negotiation
- Attempts HTTP/3 (QUIC) first
- Falls back to HTTP/2 or HTTP/1.1 automatically
- Transparent to calling code

### Retry Logic
- Exponential backoff with jitter
- Configurable max retries
- Automatic retry on network errors
- Respects context cancellation

### Compression
- Automatic Brotli decompression
- Configurable compression levels
- Transparent to calling code

### Connection Pooling
- Idle connection reuse
- Configurable pool sizes
- Automatic cleanup

## Configuration Constants

| Option | Default | Description |
|--------|---------|-------------|
| EnableHTTP3 | true | Enable HTTP/3 support |
| EnableHTTP2 | true | Enable HTTP/2 fallback |
| EnableBrotli | true | Enable Brotli compression |
| Timeout | 120s | Total request timeout |
| DialTimeout | 30s | Connection establishment timeout |
| MaxRetries | 3 | Maximum retry attempts |

## Testing

Run tests with:

```bash
go test ./internal/transport/... -v
```

## Requirements

- Go 1.25+
- `github.com/quic-go/quic-go` for HTTP/3
- `github.com/andybalholm/brotli` for Brotli compression

## Compliance

This implementation satisfies **CONST-023**: HTTP/3 Support Required
