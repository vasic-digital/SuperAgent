# HelixAgent Python SDK

> **Status: Available**
>
> The Python SDK is implemented and available at `/sdk/python/`.
> Install from source or wait for PyPI publication.

A comprehensive Python SDK for interacting with the HelixAgent AI orchestration platform, providing easy access to multi-provider LLM capabilities and OpenAI-compatible API.

## Installation

```bash
# Install from source
cd sdk/python
pip install -e .

# Or when published to PyPI:
# pip install helixagent-sdk
```

## Quick Start

```python
from helixagent import HelixAgent

# Initialize client
client = HelixAgent(
    api_key="your-api-key",
    base_url="https://api.helixagent.ai"
)

# Simple chat completion
response = client.chat.completions.create(
    model="helixagent-ensemble",
    messages=[
        {"role": "user", "content": "Explain quantum computing"}
    ]
)

print(response.choices[0].message.content)
```

## Authentication

The SDK supports multiple authentication methods:

```python
# API Key authentication
client = HelixAgent(api_key="your-api-key")

# JWT Token authentication
client = HelixAgent(token="your-jwt-token")

# Custom base URL
client = HelixAgent(
    api_key="your-api-key",
    base_url="http://localhost:7061"
)
```

## Chat Completions

### Basic Chat Completion

```python
from helixagent import HelixAgent

client = HelixAgent(api_key="your-api-key")

response = client.chat.completions.create(
    model="helixagent-ensemble",
    messages=[
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "What is machine learning?"}
    ],
    max_tokens=500,
    temperature=0.7
)

print(response.choices[0].message.content)
print(f"Usage: {response.usage.total_tokens} tokens")
```

### Streaming Chat Completion

```python
response = client.chat.completions.create(
    model="deepseek-chat",
    messages=[{"role": "user", "content": "Tell me a story"}],
    stream=True
)

for chunk in response:
    if chunk.choices[0].delta.content:
        print(chunk.choices[0].delta.content, end="")
```

### Ensemble Completion

```python
# Advanced ensemble with custom configuration
response = client.ensemble.completions.create(
    messages=[{"role": "user", "content": "What is the future of AI?"}],
    ensemble_config={
        "strategy": "confidence_weighted",
        "min_providers": 3,
        "confidence_threshold": 0.8,
        "fallback_to_best": True
    }
)

print(f"Ensemble result: {response.choices[0].message.content}")
print(f"Providers used: {response.ensemble.providers_used}")
print(f"Confidence: {response.ensemble.confidence_score}")
```

## Text Completions

### Basic Text Completion

```python
response = client.completions.create(
    model="qwen-max",
    prompt="The future of technology is",
    max_tokens=100,
    temperature=0.8,
    stop=["\n", "."]
)

print(response.choices[0].text)
```

### Streaming Text Completion

```python
response = client.completions.create(
    model="openrouter/grok-4",
    prompt="Write a haiku about programming:",
    stream=True
)

for chunk in response:
    print(chunk.choices[0].text, end="")
```

## AI Debate System

### Creating a Basic Debate

```python
debate_config = {
    "debateId": "ai-ethics-debate-001",
    "topic": "Should AI systems have ethical constraints built into their core architecture?",
    "maximal_repeat_rounds": 3,
    "consensus_threshold": 0.75,
    "participants": [
        {
            "name": "EthicsExpert",
            "role": "AI Ethics Specialist",
            "llms": [{
                "provider": "claude",
                "model": "claude-3-5-sonnet-20241022"
            }]
        },
        {
            "name": "AIScientist",
            "role": "AI Research Scientist",
            "llms": [{
                "provider": "deepseek",
                "model": "deepseek-coder"
            }]
        }
    ]
}

debate = client.debates.create(debate_config)
print(f"Debate created: {debate.debateId}")
```

### Monitoring Debate Progress

```python
# Get debate status
status = client.debates.get_status("ai-ethics-debate-001")
print(f"Status: {status.status}")
print(f"Progress: Round {status.current_round}/{status.total_rounds}")

# Wait for completion
while status.status not in ["completed", "failed"]:
    time.sleep(5)
    status = client.debates.get_status("ai-ethics-debate-001")
```

### Getting Debate Results

```python
results = client.debates.get_results("ai-ethics-debate-001")

print(f"Topic: {results.topic}")
print(f"Consensus achieved: {results.consensus.achieved}")
print(f"Final position: {results.consensus.final_position}")

for participant in results.participants:
    print(f"{participant.name}: {participant.total_responses} responses, "
          f"avg quality: {participant.avg_quality_score}")
```

### Advanced Debate with Cognee Enhancement

```python
debate_config = {
    "debateId": "enhanced-debate-001",
    "topic": "How should society regulate artificial general intelligence?",
    "enable_cognee": True,
    "cognee_config": {
        "dataset_name": "agi_regulation_debate",
        "enhancement_strategy": "hybrid",
        "max_enhancement_time": 30000
    },
    "participants": [
        {
            "name": "PolicyMaker",
            "role": "Government Policy Advisor",
            "enable_cognee": True,
            "cognee_settings": {
                "enhance_responses": True,
                "analyze_sentiment": True,
                "dataset_name": "policy_debate_data"
            }
        },
        {
            "name": "AIRiskExpert",
            "role": "AI Safety Researcher",
            "enable_cognee": True
        }
    ]
}

debate = client.debates.create(debate_config)
```

## Model Context Protocol (MCP)

### Getting MCP Capabilities

```python
capabilities = client.mcp.capabilities()
print(f"MCP Version: {capabilities.version}")
print(f"Available providers: {capabilities.providers}")
```

### Listing MCP Tools

```python
tools = client.mcp.tools()
for tool in tools.tools:
    print(f"Tool: {tool.name} - {tool.description}")
```

### Executing MCP Tools

```python
result = client.mcp.tools.call(
    name="read_file",
    arguments={"path": "/etc/hostname"}
)
print(f"Result: {result.result}")
```

### MCP Prompts

```python
prompts = client.mcp.prompts()
for prompt in prompts.prompts:
    print(f"Prompt: {prompt.name} - {prompt.description}")
```

### MCP Resources

```python
resources = client.mcp.resources()
for resource in resources.resources:
    print(f"Resource: {resource.name} - {resource.description}")
```

## Provider Management

### Listing Available Providers

```python
providers = client.providers.list()
for provider in providers.providers:
    print(f"{provider.name}: {provider.status} - {len(provider.models)} models")
```

### Provider Health Check

```python
health = client.providers.health()
print(f"Overall status: {health.status}")
for name, status in health.providers.items():
    print(f"{name}: {status.status} (response time: {status.response_time}s)")
```

## Error Handling

The SDK provides comprehensive error handling:

```python
from helixagent.exceptions import (
    AuthenticationError,
    RateLimitError,
    ProviderError,
    ValidationError
)

try:
    response = client.chat.completions.create(
        model="invalid-model",
        messages=[{"role": "user", "content": "Hello"}]
    )
except ValidationError as e:
    print(f"Validation error: {e}")
except ProviderError as e:
    print(f"Provider error: {e}")
except RateLimitError as e:
    print(f"Rate limit exceeded: {e}")
except AuthenticationError as e:
    print(f"Authentication failed: {e}")
```

## Advanced Configuration

### Custom HTTP Client

```python
import requests
from helixagent.client import HelixAgentClient

# Custom session with proxy
session = requests.Session()
session.proxies = {"https": "https://proxy.company.com:7061"}

client = HelixAgentClient(
    api_key="your-api-key",
    session=session
)
```

### Timeout Configuration

```python
client = HelixAgent(
    api_key="your-api-key",
    timeout=30.0,  # 30 seconds
    max_retries=3
)
```

### Logging

```python
import logging

logging.basicConfig(level=logging.DEBUG)

# SDK will use the standard Python logging
client = HelixAgent(api_key="your-api-key")
```

## Best Practices

### 1. Error Handling

Always implement proper error handling:

```python
def safe_completion(model, messages):
    try:
        response = client.chat.completions.create(
            model=model,
            messages=messages,
            max_tokens=1000
        )
        return response.choices[0].message.content
    except RateLimitError:
        time.sleep(60)  # Wait before retry
        return safe_completion(model, messages)
    except ProviderError as e:
        # Fallback to different provider
        if model.startswith("claude"):
            return safe_completion("deepseek-chat", messages)
        raise e
```

### 2. Resource Management

Use context managers for long-running operations:

```python
with client as session:
    # All operations within this block use the same session
    response1 = session.chat.completions.create(...)
    response2 = session.debates.create(...)
```

### 3. Batch Operations

For multiple requests, consider batching:

```python
# Instead of multiple individual requests
responses = []
for prompt in prompts:
    response = client.completions.create(
        model="helixagent-ensemble",
        prompt=prompt
    )
    responses.append(response)

# Consider using async version if available
import asyncio

async def batch_completions(prompts):
    tasks = [
        client.completions.acreate(
            model="helixagent-ensemble",
            prompt=prompt
        )
        for prompt in prompts
    ]
    return await asyncio.gather(*tasks)
```

### 4. Monitoring and Metrics

```python
# Enable detailed logging
client.enable_logging()

# Get usage statistics
usage = client.get_usage_stats()
print(f"Total requests: {usage.total_requests}")
print(f"Total tokens: {usage.total_tokens}")
print(f"Success rate: {usage.success_rate}%")
```

## API Reference

### Classes

- `HelixAgent`: Main client class
- `ChatCompletions`: Chat completion operations
- `Completions`: Text completion operations
- `Ensemble`: Ensemble operations
- `Debates`: AI debate operations
- `MCP`: Model Context Protocol operations
- `Providers`: Provider management

### Exceptions

- `HelixAgentError`: Base exception
- `AuthenticationError`: Authentication failures
- `RateLimitError`: Rate limit exceeded
- `ProviderError`: Provider-specific errors
- `ValidationError`: Input validation errors
- `NetworkError`: Network connectivity issues

## Embeddings API

### Generate Embeddings

```python
embeddings = client.embeddings.create(
    model="text-embedding-3-small",
    input=[
        "The quick brown fox jumps over the lazy dog",
        "Machine learning is a subset of artificial intelligence"
    ]
)

for i, embedding in enumerate(embeddings.data):
    print(f"Embedding {i}: dimensions={len(embedding.embedding)}")
```

### Semantic Search with Embeddings

```python
import numpy as np

def cosine_similarity(a, b):
    return np.dot(a, b) / (np.linalg.norm(a) * np.linalg.norm(b))

# Create embeddings for search
query_embed = client.embeddings.create(
    model="text-embedding-3-small",
    input=["How does authentication work?"]
)

# Compare with document embeddings
similarity = cosine_similarity(
    query_embed.data[0].embedding,
    document_embed.data[0].embedding
)
print(f"Similarity: {similarity:.4f}")
```

## Vision API

### Image Analysis

```python
response = client.vision.analyze(
    model="gpt-4-vision-preview",
    messages=[
        {
            "role": "user",
            "content": [
                {"type": "text", "text": "What's in this image?"},
                {
                    "type": "image_url",
                    "image_url": {"url": "https://example.com/image.jpg"}
                }
            ]
        }
    ]
)

print(response.choices[0].message.content)
```

### Base64 Image Upload

```python
import base64

with open("image.png", "rb") as f:
    base64_image = base64.b64encode(f.read()).decode()

response = client.vision.analyze(
    model="claude-3-5-sonnet-20241022",
    messages=[
        {
            "role": "user",
            "content": [
                {"type": "text", "text": "Describe this diagram"},
                {
                    "type": "image_url",
                    "image_url": {"url": f"data:image/png;base64,{base64_image}"}
                }
            ]
        }
    ]
)
```

### OCR (Optical Character Recognition)

```python
response = client.vision.ocr(
    image_url="https://example.com/document.png",
    languages=["en", "de"],
    output_format="markdown"
)

print(f"Extracted text:\n{response.text}")
print(f"Confidence: {response.confidence:.2f}")
```

## Background Tasks API

### Create Background Task

```python
task = client.tasks.create(
    type="llm_completion",
    payload={
        "model": "claude-3-5-sonnet-20241022",
        "messages": [
            {"role": "user", "content": "Analyze this large codebase..."}
        ],
        "max_tokens": 10000
    },
    priority="high",
    timeout=600  # 10 minutes
)

print(f"Task created: {task.id}")
```

### Poll Task Status

```python
import time

while True:
    status = client.tasks.get_status(task.id)
    print(f"Status: {status.status}, Progress: {status.progress}%")

    if status.status == "completed":
        result = client.tasks.get_result(task.id)
        print(f"Result: {result.output}")
        break
    elif status.status == "failed":
        print(f"Task failed: {status.error}")
        break

    time.sleep(2)
```

### Server-Sent Events (SSE) Streaming

```python
for event in client.tasks.stream_events(task.id):
    if event.type == "progress":
        print(f"Progress: {event.progress}%")
    elif event.type == "output":
        print(f"Output: {event.data}")
    elif event.type == "completed":
        print("Task completed!")
        break
    elif event.type == "error":
        print(f"Error: {event.error}")
        break
```

### WebSocket Real-Time Updates

```python
import asyncio
import websockets

async def monitor_task(task_id):
    async with client.tasks.websocket(task_id) as ws:
        async for message in ws:
            event = json.loads(message)
            print(f"[{event['type']}] {event.get('message', '')}")

            if event["type"] in ("completed", "failed"):
                break

asyncio.run(monitor_task(task.id))
```

## Webhooks

### Configure Webhook

```python
webhook = client.webhooks.create(
    url="https://your-server.com/webhooks/helixagent",
    events=["task.completed", "task.failed", "debate.finished"],
    secret="your-webhook-secret",
    headers={"X-Custom-Header": "value"}
)

print(f"Webhook created: {webhook.id}")
```

### Verify Webhook Signature (Flask Example)

```python
from flask import Flask, request
from helixagent import verify_webhook_signature

app = Flask(__name__)

@app.route("/webhooks/helixagent", methods=["POST"])
def webhook_handler():
    body = request.get_data()
    signature = request.headers.get("X-HelixAgent-Signature")

    if not verify_webhook_signature(body, signature, "your-webhook-secret"):
        return "Invalid signature", 401

    event = request.get_json()

    if event["type"] == "task.completed":
        handle_task_completed(event)
    elif event["type"] == "debate.finished":
        handle_debate_finished(event)

    return "OK", 200
```

## Async Operations

### Full Async Client

```python
import asyncio
from helixagent import AsyncHelixAgent

async def main():
    client = AsyncHelixAgent(api_key="your-api-key")

    # Async chat completion
    response = await client.chat.completions.create(
        model="helixagent-ensemble",
        messages=[{"role": "user", "content": "Hello"}]
    )
    print(response.choices[0].message.content)

    await client.close()

asyncio.run(main())
```

### Concurrent Requests with asyncio.gather

```python
async def batch_process(prompts: list[str]) -> list[str]:
    async with AsyncHelixAgent(api_key="your-api-key") as client:
        # Limit concurrency with semaphore
        semaphore = asyncio.Semaphore(5)

        async def process_one(prompt: str) -> str:
            async with semaphore:
                response = await client.chat.completions.create(
                    model="helixagent-ensemble",
                    messages=[{"role": "user", "content": prompt}]
                )
                return response.choices[0].message.content

        results = await asyncio.gather(*[process_one(p) for p in prompts])
        return results

results = asyncio.run(batch_process(["Prompt 1", "Prompt 2", "Prompt 3"]))
```

### Streaming with Async Generators

```python
async def stream_response(prompt: str):
    async with AsyncHelixAgent(api_key="your-api-key") as client:
        response = await client.chat.completions.create(
            model="deepseek-chat",
            messages=[{"role": "user", "content": prompt}],
            stream=True
        )

        async for chunk in response:
            if chunk.choices[0].delta.content:
                yield chunk.choices[0].delta.content

async def main():
    async for text in stream_response("Tell me a story"):
        print(text, end="", flush=True)

asyncio.run(main())
```

## Retry Strategies

### Exponential Backoff with tenacity

```python
from tenacity import (
    retry,
    stop_after_attempt,
    wait_exponential,
    retry_if_exception_type
)
from helixagent.exceptions import RateLimitError, ProviderError

@retry(
    stop=stop_after_attempt(5),
    wait=wait_exponential(multiplier=1, min=4, max=60),
    retry=retry_if_exception_type((RateLimitError, ProviderError))
)
def reliable_completion(model: str, messages: list) -> str:
    response = client.chat.completions.create(
        model=model,
        messages=messages
    )
    return response.choices[0].message.content
```

### Circuit Breaker Pattern

```python
from circuitbreaker import circuit

@circuit(failure_threshold=3, recovery_timeout=30)
def safe_request(model: str, messages: list):
    return client.chat.completions.create(
        model=model,
        messages=messages
    )

try:
    response = safe_request("helixagent-ensemble", messages)
except CircuitBreakerError:
    # Circuit is open, use fallback
    response = fallback_response()
```

## Testing Utilities

### Mock Client

```python
from unittest.mock import MagicMock, patch
from helixagent import HelixAgent
from helixagent.types import ChatCompletionResponse, Choice, Message

def test_my_feature():
    # Create mock response
    mock_response = ChatCompletionResponse(
        id="test",
        choices=[Choice(message=Message(content="Mock response"))],
        usage={"total_tokens": 10}
    )

    with patch.object(HelixAgent, 'chat') as mock_chat:
        mock_chat.completions.create.return_value = mock_response

        client = HelixAgent(api_key="test-key")
        result = my_feature(client)

        assert result == "Mock response"
        mock_chat.completions.create.assert_called_once()
```

### Recording and Playback with VCR

```python
import vcr

@vcr.use_cassette('fixtures/chat_completion.yaml')
def test_chat_completion():
    client = HelixAgent(api_key="test-key")

    # First run: records actual API response
    # Subsequent runs: replays recorded response
    response = client.chat.completions.create(
        model="helixagent-ensemble",
        messages=[{"role": "user", "content": "Hello"}]
    )

    assert response.choices[0].message.content is not None
```

### Pytest Fixtures

```python
import pytest
from helixagent import HelixAgent

@pytest.fixture
def client():
    return HelixAgent(
        api_key="test-key",
        base_url="http://localhost:7061"
    )

@pytest.fixture
def mock_client(mocker):
    client = HelixAgent(api_key="test-key")
    mocker.patch.object(client.chat.completions, 'create')
    return client

def test_feature(mock_client):
    mock_client.chat.completions.create.return_value = mock_response
    # Test your feature
```

## Observability

### OpenTelemetry Integration

```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from helixagent import HelixAgent

# Configure tracer
trace.set_tracer_provider(TracerProvider())
tracer = trace.get_tracer("helixagent-client")

client = HelixAgent(
    api_key="your-api-key",
    tracer=tracer
)

# Traces are automatically created for each API call
with tracer.start_as_current_span("my-operation") as span:
    response = client.chat.completions.create(
        model="helixagent-ensemble",
        messages=[{"role": "user", "content": "Hello"}]
    )
    # Span includes: request details, response time, token usage, errors
```

### Structured Logging

```python
import structlog

structlog.configure(
    processors=[
        structlog.stdlib.filter_by_level,
        structlog.stdlib.add_logger_name,
        structlog.stdlib.add_log_level,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.JSONRenderer()
    ],
    logger_factory=structlog.stdlib.LoggerFactory(),
)

client = HelixAgent(
    api_key="your-api-key",
    logger=structlog.get_logger()
)

# Logs include structured data:
# {"event": "api_request", "method": "chat.completions.create", "model": "helixagent-ensemble"}
# {"event": "api_response", "status": 200, "tokens": 150, "latency_ms": 234}
```

### Prometheus Metrics

```python
from prometheus_client import Counter, Histogram

# SDK automatically exports metrics if prometheus_client is installed
# helixagent_requests_total{method="chat.completions.create", status="success"}
# helixagent_request_duration_seconds{method="chat.completions.create"}
# helixagent_tokens_used_total{type="input"}
# helixagent_tokens_used_total{type="output"}

client = HelixAgent(
    api_key="your-api-key",
    enable_metrics=True,
    metrics_prefix="myapp_helixagent"
)
```

## CLI Integration

### Command-Line Usage

```python
# helixagent-cli.py
import click
from helixagent import HelixAgent

@click.group()
@click.option('--api-key', envvar='HELIXAGENT_API_KEY')
@click.pass_context
def cli(ctx, api_key):
    ctx.obj = HelixAgent(api_key=api_key)

@cli.command()
@click.argument('prompt')
@click.option('--model', default='helixagent-ensemble')
@click.pass_obj
def chat(client, prompt, model):
    """Send a chat completion request."""
    response = client.chat.completions.create(
        model=model,
        messages=[{"role": "user", "content": prompt}]
    )
    click.echo(response.choices[0].message.content)

@cli.command()
@click.argument('topic')
@click.pass_obj
def debate(client, topic):
    """Start an AI debate on a topic."""
    debate = client.debates.create({"topic": topic})
    click.echo(f"Debate started: {debate.debateId}")

if __name__ == '__main__':
    cli()
```

## Requirements

- Python 3.8+
- requests
- pydantic
- aiohttp (for async operations)

## License

MIT License - see LICENSE file for details.