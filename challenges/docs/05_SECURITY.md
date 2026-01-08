# HelixAgent Challenges - Security Practices

This document outlines security practices for handling API keys and sensitive data in the Challenges System.

## Golden Rules

1. **NEVER** commit API keys to version control
2. **ALWAYS** use `.env` for actual credentials
3. **ALWAYS** redact keys in logs and reports
4. **ALWAYS** use `.example` files for templates

## File Classification

### Git-Versioned (Safe to Commit)

| File | Purpose |
|------|---------|
| `.env.example` | Template with placeholder values |
| `config/*.yaml.example` | Configuration templates |
| `*_redacted.json` | Reports with redacted data |
| `docs/*.md` | Documentation |
| `*.go` | Source code (no hardcoded keys) |

### Git-Ignored (NEVER Commit)

| File | Purpose |
|------|---------|
| `.env` | Actual API keys |
| `config/*.yaml` | Actual configurations |
| `results/*/logs/*` | May contain sensitive data |
| `*_actual.json` | Unredacted reports |

## API Key Handling

### Storage

```
challenges/
├── .env.example     # Template (versioned)
├── .env             # Actual keys (ignored)
└── .gitignore       # Ensures .env is ignored
```

### Loading

```go
// Keys are loaded from environment
func LoadAPIKeys() map[string]string {
    keys := make(map[string]string)

    // Load from .env file
    godotenv.Load(".env")

    // Get keys from environment
    keys["anthropic"] = os.Getenv("ANTHROPIC_API_KEY")
    keys["openai"] = os.Getenv("OPENAI_API_KEY")
    // ...

    return keys
}
```

### Redaction

```go
// Redact API keys for logging/storage
func RedactAPIKey(key string) string {
    if len(key) <= 8 {
        return "*****"
    }
    // Show first 4 characters only
    return key[:4] + strings.Repeat("*", len(key)-4)
}

// Example:
// Input:  "sk-ant-api-123456789abcdef"
// Output: "sk-a***********************"
```

### Validation

```go
// Validate key format without exposing value
func ValidateAPIKey(provider, key string) error {
    switch provider {
    case "anthropic":
        if !strings.HasPrefix(key, "sk-ant-") {
            return errors.New("invalid Anthropic key format")
        }
    case "openai":
        if !strings.HasPrefix(key, "sk-") {
            return errors.New("invalid OpenAI key format")
        }
    // ...
    }
    return nil
}
```

## Configuration Files

### Template (Versioned)

`config/config.yaml.example`:
```yaml
providers:
  anthropic:
    api_key: "${ANTHROPIC_API_KEY}"  # Placeholder
    model: "claude-3-opus"

  openai:
    api_key: "${OPENAI_API_KEY}"     # Placeholder
    model: "gpt-4-turbo"
```

### Actual (Ignored)

`config/config.yaml`:
```yaml
providers:
  anthropic:
    api_key: "sk-ant-api-actual-key-here"  # Real value
    model: "claude-3-opus"
```

### Runtime Resolution

```go
// Resolve ${VAR} placeholders
func ResolveConfig(config string) string {
    re := regexp.MustCompile(`\$\{([^}]+)\}`)
    return re.ReplaceAllStringFunc(config, func(match string) string {
        varName := match[2:len(match)-1]
        return os.Getenv(varName)
    })
}
```

## Logging Security

### Request Logging

```go
// Log API request with redacted headers
func LogAPIRequest(req *http.Request) LogEntry {
    return LogEntry{
        Timestamp: time.Now(),
        Method:    req.Method,
        URL:       req.URL.String(),
        Headers:   redactHeaders(req.Header),  // Redact Authorization
    }
}

func redactHeaders(headers http.Header) http.Header {
    safe := headers.Clone()
    if auth := safe.Get("Authorization"); auth != "" {
        safe.Set("Authorization", "Bearer ****")
    }
    if key := safe.Get("X-API-Key"); key != "" {
        safe.Set("X-API-Key", "****")
    }
    return safe
}
```

### Response Logging

```go
// Log response without sensitive data
func LogAPIResponse(resp *http.Response, body []byte) LogEntry {
    return LogEntry{
        Timestamp:  time.Now(),
        StatusCode: resp.StatusCode,
        Headers:    redactHeaders(resp.Header),
        BodyLength: len(body),
        BodyPreview: truncate(string(body), 500),  // Preview only
    }
}
```

## Report Generation

### Redacted Reports (Versioned)

```json
{
  "challenge": "provider_verification",
  "providers": [
    {
      "name": "anthropic",
      "api_key": "sk-a*****",
      "status": "verified"
    }
  ]
}
```

### Full Reports (Ignored)

```json
{
  "challenge": "provider_verification",
  "providers": [
    {
      "name": "anthropic",
      "api_key": "sk-ant-api-123456789",
      "status": "verified"
    }
  ]
}
```

## .gitignore Configuration

```gitignore
# API Keys and Credentials
.env
.env.local
.env.*.local
!.env.example

# Configuration with secrets
config/*.yaml
!config/*.yaml.example

# Logs that may contain sensitive data
*.log
logs/
results/

# Generated files with actual data
*_actual.json
*_actual.yaml

# IDE and OS files
.idea/
.vscode/
.DS_Store
```

## Pre-commit Checks

### Secret Detection Script

```bash
#!/bin/bash
# scripts/check_secrets.sh

# Patterns that indicate secrets
PATTERNS=(
    "sk-ant-"
    "sk-[a-zA-Z0-9]{20,}"
    "AIzaSy"
    "nvapi-"
    "hf_[a-zA-Z0-9]+"
)

# Check staged files
for pattern in "${PATTERNS[@]}"; do
    if git diff --cached | grep -E "$pattern"; then
        echo "ERROR: Potential secret detected: $pattern"
        exit 1
    fi
done

echo "No secrets detected"
exit 0
```

### Git Hook Setup

```bash
# .git/hooks/pre-commit
#!/bin/bash
./scripts/check_secrets.sh
```

## Emergency Procedures

### If Secrets Are Committed

1. **Immediately** rotate the exposed keys
2. Remove from git history:
   ```bash
   git filter-branch --force --index-filter \
     "git rm --cached --ignore-unmatch <file>" \
     --prune-empty --tag-name-filter cat -- --all
   ```
3. Force push (with caution):
   ```bash
   git push --force --all
   ```
4. Notify team members

### Key Rotation

1. Generate new keys from provider dashboard
2. Update `.env` with new keys
3. Verify functionality
4. Revoke old keys
5. Document rotation in security log

## Security Checklist

Before each commit:
- [ ] `.env` is in `.gitignore`
- [ ] No API keys in source code
- [ ] Config files use `${VAR}` placeholders
- [ ] Logs have redacted headers
- [ ] Reports have redacted keys
- [ ] Pre-commit hook is active

Before each release:
- [ ] All secrets are in `.env` only
- [ ] Example files have placeholder values
- [ ] Documentation doesn't contain real keys
- [ ] Test with fresh environment
