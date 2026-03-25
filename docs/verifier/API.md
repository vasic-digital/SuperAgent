# LLMsVerifier API Reference

API documentation for the LLMsVerifier verification pipeline.

## Verification Endpoint

```
GET /v1/startup/verification
```

Returns the current provider verification results including scores, rankings, and debate team selection.

## Scoring Components

Each provider is scored on 5 weighted components:

| Component | Weight | Description |
|-----------|--------|-------------|
| ResponseSpeed | 25% | Time to first token and total latency |
| CostEffectiveness | 25% | Cost per token relative to quality |
| ModelEfficiency | 20% | Output quality per compute unit |
| Capability | 20% | Feature support (tools, vision, streaming) |
| Recency | 10% | Model freshness and update frequency |

## Provider Types

- **API Key providers**: Direct API access (DeepSeek, Gemini, Mistral, OpenRouter, ZAI, Cerebras)
- **OAuth providers**: CLI proxy access (Claude, Qwen) with +0.5 score bonus
- **Free providers**: No authentication required (Zen, OpenRouter :free), scored 6.0-7.0

## Subscription Detection

```
GET /v1/verification/subscription/:provider
```

Returns the detected subscription tier for a provider: `free`, `free_credits`, `free_tier`, `pay_as_you_go`, `monthly`, or `enterprise`.

## Related Documentation

- [LLMsVerifier User Guide](./USER_GUIDE.md)
- [Startup Verification](../../internal/verifier/startup.go)
