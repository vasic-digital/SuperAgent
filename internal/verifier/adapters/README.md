# Verifier Adapters Package

This package provides adapters for verifying different types of LLM providers during the startup verification pipeline.

## Overview

The verifier adapters implement specialized verification logic for:
- OAuth-authenticated providers (Claude, Qwen)
- Free/anonymous providers (Zen/OpenCode)
- Extended provider configurations

## Components

### OAuth Adapter (`oauth_adapter.go`)

Handles verification of OAuth2-authenticated providers:

```go
adapter := adapters.NewOAuthAdapter(config)
result, err := adapter.Verify(ctx, provider)
```

**Supported Providers:**
- Claude (OAuth from `~/.claude/.credentials.json`)
- Qwen (OAuth from `~/.qwen/oauth_creds.json`)

**Note:** OAuth tokens from CLI tools are product-restricted and cannot be used for general API calls. See CLAUDE.md for details.

### Free Adapter (`free_adapter.go`)

Handles verification of free/anonymous providers:

```go
adapter := adapters.NewFreeAdapter(config)
result, err := adapter.Verify(ctx, provider)
```

**Supported Providers:**
- Zen (OpenCode) - Anonymous access
- OpenRouter free models (`:free` suffix)

**Score Range:** 6.0 - 7.0 (reduced verification)

### Extended Providers Adapter (`extended_providers_adapter.go`)

Handles additional provider configurations and extensions:

```go
adapter := adapters.NewExtendedProvidersAdapter(config)
providers, err := adapter.DiscoverProviders(ctx)
```

## Architecture

```
┌─────────────────────────────────────────────┐
│         Startup Verification Pipeline       │
│                     │                        │
│                     ▼                        │
│  ┌─────────────────────────────────────┐   │
│  │          Adapter Selection          │   │
│  │                                     │   │
│  │  ┌──────────┐ ┌────────┐ ┌──────┐ │   │
│  │  │  OAuth   │ │  Free  │ │ Ext. │ │   │
│  │  │ Adapter  │ │Adapter │ │Adapt.│ │   │
│  │  └────┬─────┘ └───┬────┘ └──┬───┘ │   │
│  │       │           │          │     │   │
│  └───────┼───────────┼──────────┼─────┘   │
│          ▼           ▼          ▼         │
│  ┌─────────────────────────────────────┐   │
│  │         Provider Verification        │   │
│  │   8-test pipeline per provider      │   │
│  └─────────────────────────────────────┘   │
└─────────────────────────────────────────────┘
```

## Verification Process

### OAuth Providers

1. Read credentials from CLI credential files
2. Check token validity and expiration
3. Attempt API call with OAuth token
4. If restricted, mark with "trust on API failure" option
5. Calculate verification score

### Free Providers

1. Check provider availability
2. Perform reduced verification (basic API call)
3. Apply reduced score range (6.0-7.0)
4. Handle rate limiting gracefully

### API Key Providers

1. Validate API key format
2. Run full 8-test verification pipeline
3. Calculate weighted score (5 components)
4. Apply OAuth bonus (+0.5) if applicable

## Configuration

```yaml
verifier:
  adapters:
    oauth:
      enabled: true
      trust_on_api_failure: true
      credential_paths:
        claude: "~/.claude/.credentials.json"
        qwen: "~/.qwen/oauth_creds.json"

    free:
      enabled: true
      score_range:
        min: 6.0
        max: 7.0
      providers:
        - zen
        - openrouter_free

    extended:
      enabled: true
      discover_additional: true
```

## Usage

```go
import "dev.helix.agent/internal/verifier/adapters"

// Create adapters
oauthAdapter := adapters.NewOAuthAdapter(adapters.OAuthConfig{
    TrustOnAPIFailure: true,
})

freeAdapter := adapters.NewFreeAdapter(adapters.FreeConfig{
    MinScore: 6.0,
    MaxScore: 7.0,
})

// Use in verification pipeline
for _, provider := range providers {
    var result *adapters.VerificationResult

    switch provider.AuthType {
    case "oauth":
        result, _ = oauthAdapter.Verify(ctx, provider)
    case "free":
        result, _ = freeAdapter.Verify(ctx, provider)
    default:
        result, _ = defaultAdapter.Verify(ctx, provider)
    }
}
```

## Testing

```bash
go test -v ./internal/verifier/adapters/...
```

## Files

- `oauth_adapter.go` - OAuth provider verification
- `free_adapter.go` - Free provider verification
- `free_adapter_test.go` - Free adapter tests
- `free_adapter_comprehensive_test.go` - Comprehensive tests
- `extended_providers_adapter.go` - Extended configurations
- `extended_providers_adapter_test.go` - Extended adapter tests
