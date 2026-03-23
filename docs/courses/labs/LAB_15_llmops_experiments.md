# Lab 15: LLMOps Experimentation

## Objective
Set up and run an A/B experiment comparing two LLM providers using the LLMOps API.

## Prerequisites
- HelixAgent running locally
- At least 2 providers configured (e.g., DeepSeek + Gemini)

## Exercise 1: Create a Prompt Version

```bash
curl -X POST http://localhost:7061/v1/llmops/prompts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "summarize-code",
    "version": "1.0.0",
    "content": "Summarize the following code in 2-3 sentences, focusing on its purpose and key patterns: {{code}}"
  }'
```

**Verify:** Response contains prompt ID and version.

## Exercise 2: Create an A/B Experiment

```bash
curl -X POST http://localhost:7061/v1/llmops/experiments \
  -H "Content-Type: application/json" \
  -d '{
    "name": "summarize-provider-test",
    "description": "Compare DeepSeek vs Gemini for code summarization",
    "variants": [
      {"name": "deepseek", "provider": "deepseek", "model": "deepseek-chat", "weight": 50},
      {"name": "gemini", "provider": "gemini", "model": "gemini-pro", "weight": 50}
    ]
  }'
```

## Exercise 3: Run Continuous Evaluation

```bash
curl -X POST http://localhost:7061/v1/llmops/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "name": "weekly-quality",
    "dataset": "code-samples-v1"
  }'
```

## Exercise 4: List and Compare Results

```bash
curl http://localhost:7061/v1/llmops/experiments
curl http://localhost:7061/v1/llmops/prompts
```

## Assessment Questions
1. Why use weighted variants instead of random selection?
2. How many samples are needed for statistical significance?
3. When should you create a new prompt version vs. modify existing?
