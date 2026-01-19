# Auth Package

The auth package provides authentication and credential management for HelixAgent, including OAuth2 credential handling for LLM providers.

## Overview

This package implements:

- **OAuth Credential Management**: Read and manage OAuth tokens from CLI tools
- **Token Refresh**: Automatic token refresh handling
- **CLI Integration**: Support for Claude Code and Qwen CLI credentials

## Directory Structure

```
internal/auth/
└── oauth_credentials/
    ├── oauth_credentials.go    # Core credential management
    ├── token_refresh.go        # Token refresh logic
    └── cli_refresh.go          # CLI token refresh utilities
```

## OAuth Credential Sources

### Claude Code

Reads credentials from `~/.claude/.credentials.json`:

```go
creds, err := auth.LoadClaudeCredentials()
if err != nil {
    return err
}

fmt.Printf("Access Token: %s\n", creds.AccessToken)
fmt.Printf("Expires At: %s\n", creds.ExpiresAt)
```

### Qwen CLI

Reads credentials from `~/.qwen/oauth_creds.json`:

```go
creds, err := auth.LoadQwenCredentials()
if err != nil {
    return err
}
```

## Key Components

### Credential Loader

```go
type OAuthCredentials struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    TokenType    string    `json:"token_type"`
    ExpiresAt    time.Time `json:"expires_at"`
    Scope        string    `json:"scope"`
}

creds, err := auth.LoadCredentials(provider)
if err != nil {
    return err
}

if creds.IsExpired() {
    creds, err = auth.RefreshToken(creds)
}
```

### Token Refresh

```go
refresher := auth.NewTokenRefresher(config)

// Auto-refresh on expiry
creds, err := refresher.EnsureValid(ctx, credentials)
if err != nil {
    return err
}
```

### CLI Refresh

Integrates with CLI tools for token refresh:

```go
// Trigger Claude CLI refresh
err := auth.RefreshClaudeCLI()

// Trigger Qwen CLI refresh
err := auth.RefreshQwenCLI()
```

## Configuration

```go
type AuthConfig struct {
    ClaudeCredPath string        // Path to Claude credentials
    QwenCredPath   string        // Path to Qwen credentials
    RefreshBuffer  time.Duration // Refresh before expiry
    EnableRefresh  bool          // Auto-refresh tokens
}
```

### Environment Variables

```bash
# OAuth credential paths
CLAUDE_CRED_PATH=~/.claude/.credentials.json
QWEN_CRED_PATH=~/.qwen/oauth_creds.json

# Enable OAuth usage
CLAUDE_CODE_USE_OAUTH_CREDENTIALS=true
QWEN_CODE_USE_OAUTH_CREDENTIALS=true
```

## Important Limitations

OAuth tokens from CLI tools have product restrictions:

| Provider | Token Source | API Access |
|----------|--------------|------------|
| **Claude** | Claude Code CLI | Restricted to Claude Code only |
| **Qwen** | Qwen CLI | Portal only, not DashScope API |

For general API access, use dedicated API keys instead of OAuth tokens.

## Usage Examples

### Check Credential Validity

```go
creds, err := auth.LoadCredentials("claude")
if err != nil {
    log.Fatal(err)
}

if creds.IsValid() {
    fmt.Println("Credentials are valid")
} else if creds.IsExpired() {
    fmt.Println("Credentials expired, refreshing...")
    creds, err = auth.RefreshToken(creds)
}
```

### Secure Token Storage

```go
// Tokens are stored with restricted permissions
err := auth.SaveCredentials(creds, &auth.SaveOptions{
    Path: "~/.myapp/credentials.json",
    Mode: 0600, // Owner read/write only
})
```

## Testing

```bash
# Run auth tests
go test -v ./internal/auth/...

# Test OAuth credential loading
go test -v -run TestOAuth ./internal/auth/oauth_credentials/

# Test token refresh
go test -v -run TestRefresh ./internal/auth/oauth_credentials/
```

## Security Considerations

1. **Credential Storage**: Store credentials with restrictive permissions (0600)
2. **Token Exposure**: Never log full tokens
3. **Refresh Tokens**: Handle refresh token rotation
4. **Expiry Handling**: Always check token expiry before use

## See Also

- `internal/verifier/adapters/oauth_adapter.go` - OAuth provider verification
- `CLAUDE.md` - OAuth2 authentication limitations
