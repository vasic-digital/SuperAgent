# Chapter 1: Getting Started with HelixAgent

Welcome to HelixAgent - an AI-powered ensemble LLM service that combines responses from multiple language models using intelligent aggregation strategies.

## What is HelixAgent?

HelixAgent is a powerful service that:
- Combines multiple LLM providers (Claude, DeepSeek, Gemini, Qwen, and more)
- Uses AI Debate Ensemble for consensus-driven responses
- Provides OpenAI-compatible APIs
- Supports 18+ LLM providers with dynamic selection

## System Requirements

### Minimum Requirements
- CPU: 2 cores
- RAM: 4GB
- Disk: 10GB
- OS: Linux, macOS, or Windows with WSL2

### Recommended Requirements
- CPU: 4+ cores
- RAM: 8GB+
- Disk: 20GB+
- Docker for containerized deployment

### Software Prerequisites
- Go 1.24+ (for building from source)
- Docker and Docker Compose (for containerized deployment)
- PostgreSQL 15+ (for production)
- Redis 7+ (for caching)

## Quick Installation

### Option 1: Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/your-org/helix-agent.git
cd helix-agent

# Start all services
docker-compose up -d

# Verify it's running
curl http://localhost:7061/health
```

### Option 2: Building from Source

```bash
# Clone the repository
git clone https://github.com/your-org/helix-agent.git
cd helix-agent

# Build the binary
make build

# Run HelixAgent
./bin/helixagent
```

## Configuration

### Environment Variables

Create a `.env` file with your API keys:

```bash
# Copy the example configuration
cp .env.example .env

# Edit with your API keys
nano .env
```

Required API keys (at least one provider):
- `CLAUDE_API_KEY` - Anthropic Claude
- `DEEPSEEK_API_KEY` - DeepSeek
- `GEMINI_API_KEY` - Google Gemini
- `QWEN_API_KEY` - Alibaba Qwen
- `OPENROUTER_API_KEY` - OpenRouter

### Configuration Files

Configuration files are located in `configs/`:
- `development.yaml` - Development settings
- `production.yaml` - Production settings
- `multi-provider.yaml` - Multi-provider setup

## Your First API Call

### Health Check

```bash
curl http://localhost:7061/health
```

Expected response:
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

### Simple Completion

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "helixagent-debate",
    "messages": [
      {"role": "user", "content": "Hello, what can you help me with?"}
    ]
  }'
```

### AI Debate Completion

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "helixagent-debate",
    "messages": [
      {"role": "user", "content": "What is the best programming language for AI?"}
    ],
    "stream": true
  }'
```

## Understanding the Response

HelixAgent's AI Debate Ensemble provides theatrical dialogue responses:

```
+------------------------------------------------------------------+
|           HELIXAGENT AI DEBATE ENSEMBLE                          |
+------------------------------------------------------------------+
|  Five AI minds deliberate to synthesize the optimal response.    |
+------------------------------------------------------------------+

[A] THE ANALYST: "Let me analyze this systematically..."
[P] THE PROPOSER: "I propose we consider..."
[C] THE CRITIC: "I must challenge some assumptions..."
[S] THE SYNTHESIZER: "Combining these perspectives..."
[M] THE MEDIATOR: "After weighing all arguments..."

+------------------------------------------------------------------+
                      CONSENSUS REACHED
+------------------------------------------------------------------+

[Final synthesized response]
```

## Next Steps

- [Chapter 2: API Reference](02_API_REFERENCE.md) - Learn the full API
- [Chapter 3: Provider Configuration](03_PROVIDER_CONFIG.md) - Configure LLM providers
- [Chapter 4: MCP Integration](04_MCP_INTEGRATION.md) - Use Model Context Protocol
- [Chapter 5: Advanced Features](05_ADVANCED_FEATURES.md) - Explore advanced capabilities

## Getting Help

- Check the [FAQ](../FAQ.md)
- Review [Troubleshooting](../TROUBLESHOOTING.md)
- Open an issue on GitHub
