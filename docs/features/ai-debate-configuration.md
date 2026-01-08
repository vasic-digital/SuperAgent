# AI Debate Configuration System

## Overview

The AI Debate Configuration System provides a comprehensive framework for configuring and managing AI-powered debates with multiple participants, each backed by multiple LLMs with intelligent fallback chains. The system integrates with Cognee AI for enhanced response quality and consensus analysis.

## Key Features

- **Multiple Participants**: Configure 2-10 AI participants with distinct roles and personalities
- **LLM Fallback Chains**: Each participant can have multiple LLMs (first is primary, others are fallbacks)
- **Configurable Retry Rounds**: Global and per-participant maximum retry rounds (1-10)
- **Cognee AI Integration**: Enhanced responses, consensus analysis, and insights generation
- **Comprehensive Validation**: Extensive validation for all configuration parameters
- **Environment Variable Support**: Sensitive data loaded from environment variables
- **Flexible Debate Strategies**: Multiple debate and voting strategies available
- **Memory Management**: Contextual memory retention for coherent multi-round debates

## Configuration Structure

### Global Configuration

```yaml
# Global AI debate settings
enabled: true
maximal_repeat_rounds: 3
debate_timeout: 300000  # 5 minutes in milliseconds
consensus_threshold: 0.75
enable_cognee: true

# Debate strategy configuration
debate_strategy: "structured"  # round_robin, free_form, structured, adversarial, collaborative
voting_strategy: "confidence_weighted"  # majority, weighted, consensus, confidence_weighted, quality_weighted
response_format: "detailed"  # brief, detailed, structured, bullet_points

# Memory and context management
enable_memory: true
memory_retention: 2592000000  # 30 days in milliseconds
max_context_length: 32000

# Quality and performance settings
quality_threshold: 0.7
max_response_time: 30000  # 30 seconds in milliseconds
enable_streaming: false

# Logging and monitoring
enable_debate_logging: true
log_debate_details: true
metrics_enabled: true
```

### Cognee AI Configuration

```yaml
cognee_config:
  enabled: true
  enhance_responses: true
  analyze_consensus: true
  generate_insights: true
  dataset_name: "ai_debate_enhancement"
  max_enhancement_time: 10000  # 10 seconds in milliseconds
  enhancement_strategy: "hybrid"  # semantic_enhancement, contextual_analysis, knowledge_integration, hybrid
  memory_integration: true
  contextual_analysis: true
```

### Participant Configuration

```yaml
participants:
  - name: "Strongest"  # Unique participant name
    role: "Primary Analyst"
    description: "Main analytical participant with comprehensive reasoning capabilities"
    enabled: true
    maximal_repeat_rounds: 3  # Override global setting (optional)
    response_timeout: 30000  # 30 seconds in milliseconds
    weight: 1.5  # Weight for voting/scoring
    priority: 1  # Priority order (1 = highest)
    
    # Debate behavior
    debate_style: "analytical"  # analytical, creative, balanced, aggressive, diplomatic, technical
    argumentation_style: "logical"  # logical, emotional, evidence_based, hypothetical, socratic
    persuasion_level: 0.8  # 0.0 - 1.0
    openness_to_change: 0.3  # 0.0 - 1.0
    
    # Quality settings
    quality_threshold: 0.8  # 0.0 - 1.0
    min_response_length: 100  # Minimum characters
    max_response_length: 2000  # Maximum characters
    
    # Cognee enhancement
    enable_cognee: true
    cognee_settings:
      enhance_responses: true
      analyze_sentiment: true
      extract_entities: true
      generate_summary: true
      dataset_name: "strongest_participant_enhancement"
```

### LLM Configuration Chain

```yaml
llms:  # First LLM is primary, others are fallbacks
  - name: "Claude Primary"
    provider: "claude"  # claude, deepseek, gemini, qwen, zai, ollama, openrouter
    model: "claude-3-5-sonnet-20241022"
    enabled: true
    api_key: "${CLAUDE_API_KEY}"  # Loaded from environment
    base_url: "https://api.anthropic.com/v1"  # Optional, defaults to provider default
    
    # Connection settings
    timeout: 30000  # 30 seconds in milliseconds
    max_retries: 3
    
    # Model parameters
    temperature: 0.1  # 0.0 - 2.0
    max_tokens: 2000
    top_p: 0.9  # 0.0 - 1.0
    frequency_penalty: 0.0  # -2.0 - 2.0
    presence_penalty: 0.0  # -2.0 - 2.0
    stop_sequences: []  # Optional stop sequences
    
    # Performance settings
    weight: 1.0  # Weight within participant's LLM chain
    rate_limit_rps: 10  # Requests per second
    request_timeout: 35000  # 35 seconds in milliseconds
    
    # Capabilities and custom parameters
    capabilities:
      - "reasoning"
      - "analysis"
      - "code_generation"
      - "mathematics"
    custom_params:
      thinking: true
      advanced_reasoning: true
  
  - name: "DeepSeek Fallback"
    provider: "deepseek"
    model: "deepseek-coder"
    enabled: true
    api_key: "${DEEPSEEK_API_KEY}"
    base_url: "https://api.deepseek.com/v1"
    timeout: 25000
    max_retries: 3
    temperature: 0.1
    max_tokens: 2000
    weight: 0.9  # Slightly lower weight than primary
    rate_limit_rps: 15
    request_timeout: 30000
    capabilities:
      - "code_generation"
      - "debugging"
      - "analysis"
    custom_params:
      code_focus: true
      technical_accuracy: high
```

## Configuration Validation

The system performs comprehensive validation on all configuration parameters:

### Global Validation
- `maximal_repeat_rounds`: 1-10
- `consensus_threshold`: 0.0-1.0
- `quality_threshold`: 0.0-1.0
- `debate_timeout`: Must be positive
- `max_response_time`: Must be positive
- `max_context_length`: Must be positive
- At least 2 participants required
- Maximum 10 participants allowed

### Participant Validation
- Unique participant names
- Required name and role
- At least one enabled LLM
- `persuasion_level`: 0.0-1.0
- `openness_to_change`: 0.0-1.0
- `quality_threshold`: 0.0-1.0
- `min_response_length` ≤ `max_response_length`
- Valid debate and argumentation styles

### LLM Validation
- Required name, provider, and model
- Valid provider names (claude, deepseek, gemini, qwen, zai, ollama, openrouter)
- `temperature`: 0.0-2.0
- `max_tokens`: Must be positive
- `top_p`: 0.0-1.0
- `frequency_penalty` and `presence_penalty`: -2.0-2.0
- `weight`: Non-negative
- `rate_limit_rps`: Non-negative

## Environment Variables

Sensitive configuration data should be stored in environment variables:

```yaml
# Example with environment variables
llms:
  - name: "Claude Primary"
    provider: "claude"
    model: "claude-3-5-sonnet-20241022"
    api_key: "${CLAUDE_API_KEY}"  # Loaded from environment
  
  - name: "DeepSeek Fallback"
    provider: "deepseek"
    model: "deepseek-coder"
    api_key: "${DEEPSEEK_API_KEY}"  # Loaded from environment
```

Required environment variables:
- `CLAUDE_API_KEY`: API key for Claude
- `DEEPSEEK_API_KEY`: API key for DeepSeek
- `GEMINI_API_KEY`: API key for Gemini
- `QWEN_API_KEY`: API key for Qwen
- `OPENROUTER_API_KEY`: API key for OpenRouter
- `ZAI_API_KEY`: API key for ZAI

## Usage Examples

### Basic Configuration

```yaml
# configs/ai-debate-basic.yaml
enabled: true
maximal_repeat_rounds: 3
debate_timeout: 300000
consensus_threshold: 0.75

participants:
  - name: "Analyst"
    role: "Data Analyst"
    enabled: true
    llms:
      - name: "Primary"
        provider: "claude"
        model: "claude-3-sonnet"
        enabled: true
        api_key: "${CLAUDE_API_KEY}"
        timeout: 30000
        max_tokens: 1000
        temperature: 0.7
    response_timeout: 30000
    weight: 1.0
    priority: 1
    debate_style: "analytical"
    argumentation_style: "logical"
    persuasion_level: 0.5
    openness_to_change: 0.5
    quality_threshold: 0.7
    min_response_length: 50
    max_response_length: 1000
  
  - name: "Critic"
    role: "Quality Critic"
    enabled: true
    llms:
      - name: "Primary"
        provider: "deepseek"
        model: "deepseek-coder"
        enabled: true
        api_key: "${DEEPSEEK_API_KEY}"
        timeout: 30000
        max_tokens: 1000
        temperature: 0.7
    response_timeout: 30000
    weight: 1.0
    priority: 2
    debate_style: "critical"
    argumentation_style: "logical"
    persuasion_level: 0.5
    openness_to_change: 0.5
    quality_threshold: 0.7
    min_response_length: 50
    max_response_length: 1000
```

### Advanced Configuration with Fallbacks

```yaml
# configs/ai-debate-advanced.yaml
enabled: true
maximal_repeat_rounds: 5
debate_timeout: 600000  # 10 minutes
consensus_threshold: 0.8
enable_cognee: true

cognee_config:
  enabled: true
  enhance_responses: true
  analyze_consensus: true
  dataset_name: "advanced_debate_enhancement"
  enhancement_strategy: "hybrid"

participants:
  - name: "PrimaryAnalyst"
    role: "Senior Analyst"
    enabled: true
    maximal_repeat_rounds: 5
    response_timeout: 45000
    weight: 2.0
    priority: 1
    debate_style: "analytical"
    argumentation_style: "evidence_based"
    persuasion_level: 0.9
    openness_to_change: 0.2
    quality_threshold: 0.85
    enable_cognee: true
    
    llms:
      - name: "Claude Premium"
        provider: "claude"
        model: "claude-3-opus-20240229"
        enabled: true
        api_key: "${CLAUDE_API_KEY}"
        timeout: 45000
        max_retries: 3
        temperature: 0.1
        max_tokens: 3000
        weight: 1.0
        rate_limit_rps: 5
        capabilities:
          - "advanced_reasoning"
          - "complex_analysis"
          - "synthesis"
      
      - name: "Claude Fallback"
        provider: "claude"
        model: "claude-3-sonnet-20241022"
        enabled: true
        api_key: "${CLAUDE_API_KEY}"
        timeout: 40000
        max_retries: 3
        temperature: 0.2
        max_tokens: 2500
        weight: 0.9
        rate_limit_rps: 8
      
      - name: "DeepSeek Ultimate Fallback"
        provider: "deepseek"
        model: "deepseek-chat"
        enabled: true
        api_key: "${DEEPSEEK_API_KEY}"
        timeout: 35000
        max_retries: 2
        temperature: 0.3
        max_tokens: 2000
        weight: 0.7
        rate_limit_rps: 10
```

### Programmatic Usage

```go
package main

import (
    "context"
    "log"
    "time"
    
    "dev.helix.agent/internal/config"
    "dev.helix.agent/internal/services"
)

func main() {
    // Load configuration
    loader := config.NewAIDebateConfigLoader("configs/ai-debate-example.yaml")
    cfg, err := loader.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    
    // Create debate service
    debateService, err := services.NewAIDebateService(cfg, nil, nil)
    if err != nil {
        log.Fatalf("Failed to create debate service: %v", err)
    }
    
    // Conduct a debate
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
    defer cancel()
    
    result, err := debateService.ConductDebate(ctx, "Should AI be regulated?", "Initial context about AI regulation")
    if err != nil {
        log.Fatalf("Debate failed: %v", err)
    }
    
    // Process results
    log.Printf("Debate completed in %v", result.Duration)
    log.Printf("Consensus reached: %v", result.Consensus != nil && result.Consensus.Reached)
    log.Printf("Best response: %s", result.BestResponse.Content)
}
```

## Debate Strategies

### Available Strategies

1. **round_robin**: Each participant takes turns in a fixed order
2. **free_form**: Participants respond in any order based on availability
3. **structured**: Organized rounds with specific roles and timing
4. **adversarial**: Designed for opposing viewpoints and debate
5. **collaborative**: Focus on building consensus and agreement

### Voting Strategies

1. **majority**: Simple majority vote
2. **weighted**: Weighted by participant weights
3. **consensus**: Requires high agreement level
4. **confidence_weighted**: Weighted by response confidence scores
5. **quality_weighted**: Weighted by response quality scores

## Fallback Mechanism

The system implements intelligent fallback mechanisms:

1. **Primary LLM Failure**: Automatically tries next LLM in chain
2. **Retry Logic**: Configurable retry attempts per LLM
3. **Timeout Handling**: Respects timeout settings at multiple levels
4. **Quality Fallback**: Falls back if response quality is too low
5. **Consensus Fallback**: Continues debate if consensus isn't reached

## Error Handling

The configuration system provides comprehensive error handling:

- **Validation Errors**: Detailed validation messages for all configuration issues
- **Loading Errors**: Clear error messages for file loading and parsing issues
- **Environment Variable Errors**: Helpful messages for missing environment variables
- **Runtime Errors**: Graceful handling of LLM failures and timeouts

## Performance Considerations

- **Concurrent Processing**: Multiple participants respond concurrently
- **Timeout Management**: Configurable timeouts at multiple levels
- **Rate Limiting**: Built-in rate limiting per LLM provider
- **Memory Management**: Configurable memory retention and cleanup
- **Caching**: Response caching where appropriate

## Monitoring and Metrics

The system provides comprehensive metrics:

- **Debate Statistics**: Total debates, success rate, consensus rate
- **Performance Metrics**: Response times, timeout rates, retry rates
- **Quality Metrics**: Average confidence, quality scores, agreement levels
- **System Health**: Provider availability, error rates, resource usage

## Security Considerations

- **API Key Management**: Environment variable based API key management
- **Input Validation**: Comprehensive validation of all configuration inputs
- **Rate Limiting**: Built-in rate limiting to prevent abuse
- **Timeout Protection**: Multiple timeout layers to prevent hanging
- **Error Sanitization**: Careful error message handling to avoid information leakage

## Testing

The system includes comprehensive test coverage:

- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end configuration loading and validation
- **Chaos Tests**: Testing with malformed and edge-case configurations
- **Performance Tests**: Load testing and performance validation
- **Security Tests**: Security vulnerability testing

## Troubleshooting

### Common Issues

1. **Configuration Loading Fails**
   - Check file path and permissions
   - Validate YAML syntax
   - Ensure all required fields are present

2. **Environment Variables Not Loading**
   - Verify environment variable names
   - Check variable values are set correctly
   - Ensure no typos in variable references

3. **Participant Validation Fails**
   - Verify all participants have unique names
   - Ensure each participant has at least one enabled LLM
   - Check all required fields are present

4. **LLM Provider Errors**
   - Verify API keys are correct and valid
   - Check network connectivity to providers
   - Review rate limiting settings

### Debug Mode

Enable debug logging to troubleshoot issues:

```yaml
logging:
  level: "debug"
  enable_request_logging: true
  enable_response_logging: true
```

## Best Practices

1. **Use Environment Variables**: Store API keys and sensitive data in environment variables
2. **Configure Timeouts Appropriately**: Set realistic timeouts based on your use case
3. **Implement Fallback Chains**: Always configure multiple LLMs for reliability
4. **Monitor Performance**: Use built-in metrics to monitor system performance
5. **Test Configurations**: Thoroughly test configurations before production deployment
6. **Regular Updates**: Keep LLM models and configurations updated
7. **Security First**: Follow security best practices for API key management

## Conclusion

The AI Debate Configuration System provides a robust, flexible, and comprehensive framework for configuring AI-powered debates. With its extensive validation, fallback mechanisms, Cognee AI integration, and detailed monitoring, it enables reliable and intelligent multi-participant AI debates for various use cases.