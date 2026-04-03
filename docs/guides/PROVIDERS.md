# Provider Configuration Guide

HelixAgent supports 20+ LLM providers. This guide explains how to configure each.

## Quick Reference

| Provider | Environment Variable | Models | Features |
|----------|---------------------|--------|----------|
| OpenAI | `OPENAI_API_KEY` | GPT-4o, GPT-4o-mini | Tools, Vision, Streaming |
| Anthropic | `ANTHROPIC_API_KEY` | Claude 3.5 Sonnet, Haiku | Tools, Vision, 200K context |
| DeepSeek | `DEEPSEEK_API_KEY` | DeepSeek-V3, R1 | Tools, Reasoning, 64K context |
| Groq | `GROQ_API_KEY` | Llama 3.1/3.2, Mixtral | Tools, Vision, 800+ tok/s |
| Mistral | `MISTRAL_API_KEY` | Mistral Small/Large | Tools, Agents |
| Gemini | `GEMINI_API_KEY` | Gemini 2.0 Flash, 1.5 Pro | Tools, Vision, 1M context |
| Cohere | `COHERE_API_KEY` | Command R, R+ | Tools, RAG |
| Perplexity | `PERPLEXITY_API_KEY` | Sonar, Sonar Pro | Search-enhanced |
| Together AI | `TOGETHER_API_KEY` | 100+ open source | Various |
| Fireworks | `FIREWORKS_API_KEY` | Llama, Mixtral | Fast inference |
| Cerebras | `CEREBRAS_API_KEY` | Llama 3.1 | Wafer-scale speed |
| xAI | `XAI_API_KEY` | Grok 2, Grok 3 | Real-time |

## Configuration

### OpenAI

```bash
export OPENAI_API_KEY="sk-..."
```

Recommended models:
- `gpt-4o` - Best quality, vision
- `gpt-4o-mini` - Cost-effective

### Anthropic

```bash
export ANTHROPIC_API_KEY="sk-ant-..."
```

Recommended models:
- `claude-3-5-sonnet-20241022` - Best overall
- `claude-3-5-haiku-20241022` - Fast, cheap

Features:
- 200K context window
- Excellent tool use
- Prompt caching

### DeepSeek

```bash
export DEEPSEEK_API_KEY="sk-..."
```

Recommended models:
- `deepseek-chat` - General purpose
- `deepseek-reasoner` - Chain-of-thought reasoning

### Groq

```bash
export GROQ_API_KEY="gsk_..."
```

Recommended models:
- `llama-3.1-8b-instant` - Fastest
- `llama-3.1-70b-versatile` - Best quality
- `llama-3.2-11b-vision-preview` - Vision

Features:
- 800+ tokens/second
- Very low latency

### Mistral

```bash
export MISTRAL_API_KEY="..."
```

Recommended models:
- `mistral-small-latest` - Fast
- `mistral-large-latest` - Best quality
- `codestral-latest` - Code

### Gemini

```bash
export GEMINI_API_KEY="..."
```

Recommended models:
- `gemini-2.0-flash-exp` - Fast, capable
- `gemini-1.5-pro` - 1M context

## Provider Selection

HelixAgent automatically selects the best available provider. Priority order:

1. Groq (speed)
2. OpenAI (reliability)
3. Anthropic (quality)
4. DeepSeek (reasoning)
5. Others (fallback)

Override with:

```bash
export DEFAULT_PROVIDER="anthropic"
```

## Testing Provider Setup

```bash
# Test all configured providers
make test-providers

# Test specific provider
go test -v ./tests/providers/openai_test.go
```

## Troubleshooting

### Rate Limits

If you hit rate limits:

```bash
# Enable circuit breaker
export CIRCUIT_BREAKER_ENABLED=true

# Use multiple providers
export FALLBACK_PROVIDERS="groq,mistral"
```

### Authentication Errors

Check API key format:

```bash
# OpenAI: sk-...
# Anthropic: sk-ant-...
# Groq: gsk_...
# DeepSeek: sk-...
```

### Timeout Issues

```bash
# Increase timeout
export LLM_TIMEOUT=60s
```
