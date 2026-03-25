# Authentication Reference

HelixAgent supports three authentication mechanisms: JWT tokens, API keys, and OAuth credentials.

## API Key Authentication

The simplest method. Pass the API key in the request header:

```bash
curl -H "Authorization: Bearer <API_KEY>" http://localhost:7061/v1/chat/completions
# or
curl -H "x-api-key: <API_KEY>" http://localhost:7061/v1/chat/completions
```

API keys are configured via environment variables and validated by the auth middleware in `internal/middleware/`.

## JWT Authentication

JWT tokens are signed with `JWT_SECRET` and include expiration, issuer, and scope claims. Tokens are issued via the auth endpoint and validated on every request.

| Configuration | Description |
|---------------|-------------|
| `JWT_SECRET` | Signing secret (required) |
| `JWT_EXPIRY` | Token lifetime (default: `24h`) |
| `JWT_ISSUER` | Token issuer string |

## OAuth Provider Credentials

For providers that use OAuth (Claude, Qwen, Gemini, Junie), HelixAgent proxies through CLI tools when `<PROVIDER>_USE_OAUTH_CREDENTIALS=true` is set and no API key is available.

- **Claude**: Uses `claude -p --output-format json`
- **Qwen**: Uses ACP via `qwen --acp` with JSON-RPC 2.0
- **Gemini**: Uses `gemini -p --output-format json`

OAuth tokens are product-restricted and cannot be used for general API access. For production deployments, obtain dedicated API keys from each provider's console.

## Auth Module

The extracted `Auth` module (`digital.vasic.auth`) provides JWT, API key, and OAuth authentication with HTTP middleware and token management. See `Auth/` for the full implementation.

## Related Documentation

- [Security Overview](./README.md)
- [Sandboxing](./SANDBOXING.md)
- [SDK Authentication](../sdk/README.md)
