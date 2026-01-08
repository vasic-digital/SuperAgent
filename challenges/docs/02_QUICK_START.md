# HelixAgent Challenges - Quick Start Guide

Get started with HelixAgent Challenges in 5 minutes.

## Prerequisites

- Go 1.24+
- HelixAgent built (`make build`)
- LLMsVerifier built (in `../LLMsVerifier`)
- At least one LLM provider API key

## Step 1: Configure API Keys

```bash
# Navigate to challenges directory
cd challenges

# Copy the example environment file
cp .env.example .env

# Edit .env with your API keys
nano .env  # or your preferred editor
```

**Minimum required keys for testing:**
```env
# At least one of these:
OPENROUTER_API_KEY=sk-or-v1-your-key-here
# OR
ANTHROPIC_API_KEY=sk-ant-your-key-here
# OR
OPENAI_API_KEY=sk-your-key-here
```

## Step 2: Verify Configuration

```bash
# Check that keys are loaded
./scripts/verify_config.sh
```

Expected output:
```
Checking API key configuration...
✓ OPENROUTER_API_KEY: configured
✓ ANTHROPIC_API_KEY: configured
✗ OPENAI_API_KEY: not configured
...
Found 5 configured providers
Configuration valid!
```

## Step 3: Run Your First Challenge

```bash
# Run provider verification
./scripts/run_challenges.sh provider_verification
```

This will:
1. Load all configured providers
2. Verify each provider's API connectivity
3. Score models based on capabilities
4. Generate results in `results/provider_verification/`

## Step 4: Form AI Debate Group

```bash
# Run debate group formation
./scripts/run_challenges.sh ai_debate_formation
```

This will:
1. Use verification results from Step 3
2. Select top 5 models as primary members
3. Assign 2 fallbacks to each primary
4. Create debate group configuration

## Step 5: Test the Debate Group

```bash
# Run API quality tests
./scripts/run_challenges.sh api_quality_test
```

This will:
1. Send test prompts to the debate group
2. Validate response quality
3. Assert no mock/empty responses
4. Generate test report

## Step 6: View Results

```bash
# View master summary
cat master_results/master_summary_*.md | tail -50

# View detailed logs
ls -la results/ai_debate_formation/*/logs/

# View debate group configuration
cat results/ai_debate_formation/*/results/debate_group.json
```

## Quick Reference

### Available Commands

```bash
# Run specific challenge
./scripts/run_challenges.sh <challenge_name>

# Run all challenges
./scripts/run_all_challenges.sh

# Generate report only
./scripts/generate_report.sh

# Clean old results (older than 30 days)
./scripts/cleanup_results.sh
```

### Challenge Names

| Name | Description |
|------|-------------|
| `provider_verification` | Verify all LLM providers |
| `ai_debate_formation` | Form AI debate group |
| `api_quality_test` | Test API quality |

### Common Issues

**"No API keys configured"**
- Ensure `.env` exists and has at least one key
- Check key format matches examples

**"LLMsVerifier not found"**
- Build LLMsVerifier: `cd ../LLMsVerifier && make build`
- Set `LLMSVERIFIER_PATH` in `.env`

**"Provider timeout"**
- Increase `VERIFICATION_TIMEOUT_SECONDS` in `.env`
- Check network connectivity

## Next Steps

- Read [Challenge Catalog](03_CHALLENGE_CATALOG.md) for all challenges
- Learn about [AI Debate Groups](04_AI_DEBATE_GROUP.md)
- Review [Security Practices](05_SECURITY.md)
