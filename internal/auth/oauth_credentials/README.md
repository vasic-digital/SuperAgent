# OAuth Credentials Package

This package provides functionality to read OAuth2 credentials from Claude Code and Qwen Code CLI agents.

## Overview

When users are logged into Claude Code or Qwen Code CLI tools via OAuth2, this package allows HelixAgent to read and use those credentials for provider authentication.

## Components

### Credential Reader (`oauth_credentials.go`)

Main credential reading functionality:

```go
reader := oauth_credentials.NewOAuthCredentialReader()

// Read Claude credentials
claudeCreds, err := reader.GetClaudeCredentials()
if err == nil && claudeCreds.ClaudeAiOauth != nil {
    token := claudeCreds.ClaudeAiOauth.AccessToken
}

// Read Qwen credentials
qwenCreds, err := reader.GetQwenCredentials()
if err == nil {
    token := qwenCreds.AccessToken
}
```

### Token Refresh (`token_refresh.go`)

Automatic token refresh handling:

```go
refresher := oauth_credentials.NewTokenRefresher()
newToken, err := refresher.RefreshClaudeToken(ctx, refreshToken)
```

### CLI Refresh (`cli_refresh.go`)

CLI-based credential refresh utilities.

## Data Types

### ClaudeOAuthCredentials

```go
type ClaudeOAuthCredentials struct {
    ClaudeAiOauth *ClaudeAiOauth `json:"claudeAiOauth"`
}

type ClaudeAiOauth struct {
    AccessToken      string   // OAuth access token
    RefreshToken     string   // OAuth refresh token
    ExpiresAt        int64    // Unix timestamp (milliseconds)
    Scopes           []string // OAuth scopes
    SubscriptionType string   // Claude subscription type
    RateLimitTier    string   // Rate limit tier
}
```

### QwenOAuthCredentials

```go
type QwenOAuthCredentials struct {
    AccessToken  string // OAuth access token
    RefreshToken string // OAuth refresh token
    IDToken      string // ID token (optional)
    ExpiryDate   int64  // Unix timestamp (milliseconds)
    TokenType    string // Token type (Bearer)
    ResourceURL  string // Resource URL (optional)
}
```

## Credential Paths

| Provider | Path |
|----------|------|
| Claude | `~/.claude/.credentials.json` |
| Qwen | `~/.qwen/oauth_creds.json` |

## Usage

### Basic Credential Reading

```go
import "dev.helix.agent/internal/auth/oauth_credentials"

reader := oauth_credentials.NewOAuthCredentialReader()

// Check Claude credentials
claudeCreds, err := reader.GetClaudeCredentials()
if err == nil {
    // Check if token is expired
    if !reader.IsClaudeTokenExpired() {
        token := claudeCreds.ClaudeAiOauth.AccessToken
        // Use token for API calls
    }
}
```

### With Caching

```go
// Credentials are cached for 5 minutes by default
reader := oauth_credentials.NewOAuthCredentialReader()

// First call reads from file
creds1, _ := reader.GetClaudeCredentials()

// Second call within 5 minutes returns cached credentials
creds2, _ := reader.GetClaudeCredentials()
```

### Token Expiration Check

```go
if reader.IsClaudeTokenExpired() {
    // Token needs refresh or re-authentication
}

if reader.IsQwenTokenExpired() {
    // Token needs refresh or re-authentication
}
```

## Limitations

OAuth tokens from CLI tools are product-restricted:
- **Claude**: Tokens are restricted to Claude Code only; cannot be used for general API calls
- **Qwen**: Tokens are for Qwen Portal only; DashScope API requires separate API key

See CLAUDE.md for details on these limitations.

## Testing

```bash
go test -v ./internal/auth/oauth_credentials/...
```

## Files

- `oauth_credentials.go` - Main credential reading functionality
- `token_refresh.go` - Token refresh handling
- `cli_refresh.go` - CLI refresh utilities
