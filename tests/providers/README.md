# Provider Tests

Real integration tests for all LLM providers. All tests use actual API calls - no mocks.

## Running Tests

```bash
# All provider tests
go test -v ./tests/providers/...

# Specific provider
go test -v ./tests/providers/openai_test.go
go test -v ./tests/providers/anthropic_test.go
go test -v ./tests/providers/deepseek_test.go

# With timeout
go test -v -timeout 60s ./tests/providers/...

# Parallel (careful with rate limits)
go test -v -parallel 4 ./tests/providers/...
```

## Test Coverage

| Provider | Tests | Features Tested |
|----------|-------|-----------------|
| OpenAI | 8 | Chat, Streaming, JSON, Tools, Vision, Long Context |
| Anthropic | 8 | Chat, Streaming, Tool Use, Vision, 200K Context, Caching |
| DeepSeek | 8 | Chat, Reasoning, Streaming, JSON, Tools, 64K Context |
| Groq | 6 | Speed, Streaming, Tools, Vision, Multi-Model |
| Mistral | 5 | Chat, Streaming, Tools, Agents, Multi-Model |
| Gemini | 7 | Chat, Streaming, Vision, Tools, 1M Context, JSON |
| Cohere | 5 | Chat, Tools, Streaming, RAG, Multi-Model |
| Perplexity | 4 | Search, Streaming, Citations, Multi-Model |

**Total: 51 test functions**

## Environment Variables

```bash
export OPENAI_API_KEY="sk-..."
export ANTHROPIC_API_KEY="sk-ant-..."
export DEEPSEEK_API_KEY="sk-..."
export GROQ_API_KEY="gsk_..."
export MISTRAL_API_KEY="..."
export GEMINI_API_KEY="..."
export COHERE_API_KEY="..."
export PERPLEXITY_API_KEY="pplx-..."
```

## Test Structure

Each provider test file includes:
- `Test<Provider>_ShortRequest` - Basic completion
- `Test<Provider>_Streaming` - Real-time streaming
- `Test<Provider>_ToolUse/ToolCalling` - Function calling
- `Test<Provider>_Vision` - Image analysis (if supported)
- `Test<Provider>_JSON/JSONMode` - Structured output
- `Test<Provider>_LongContext` - Large context handling
- `Test<Provider>_ErrorHandling` - Error scenarios
- `Test<Provider>_Models/MultipleModels` - Model variants
- `Benchmark<Provider>_Latency/Speed` - Performance benchmarks

## Notes

- Tests require valid API keys
- Tests make real API calls (costs incurred)
- Tests respect timeout contexts
- Tests can be skipped with `-short` flag
- Parallel execution may hit rate limits
