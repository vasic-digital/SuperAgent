# All Validated LLMs - Sorted by Score (Strongest to Weakest)

**Generated**: 2026-01-28
**Source**: LLMsVerifier (Single Source of Truth)
**Sorting**: Score Only (NO OAuth Priority)

---

## All Verified LLMs Sorted by Score

| Rank | Model ID | Provider | Score | Auth Type | Status |
|------|----------|----------|-------|-----------|--------|
| 1 | llama-3.3-70b | Cerebras | 7.85 | API Key | Verified |
| 2 | claude-opus-4-5-20251101 | Claude | 7.65 | OAuth | Verified (Trusted) |
| 3 | claude-sonnet-4-5-20250929 | Claude | 7.65 | OAuth | Verified (Trusted) |
| 4 | claude-haiku-4-5-20251001 | Claude | 7.65 | OAuth | Verified (Trusted) |
| 5 | deepseek-reasoner | DeepSeek | 7.33 | API Key | Verified |
| 6 | deepseek-chat | DeepSeek | 7.33 | API Key | Verified |
| 7 | deepseek-coder | DeepSeek | 7.33 | API Key | Verified |
| 8 | mistral-large-latest | Mistral | 7.33 | API Key | Verified |
| 9 | mistral-medium-latest | Mistral | 7.33 | API Key | Verified |
| 10 | codestral-latest | Mistral | 7.33 | API Key | Verified |
| 11 | qwen-max | Qwen | 7.27 | OAuth | Verified (Trusted) |
| 12 | qwen-plus | Qwen | 7.27 | OAuth | Verified (Trusted) |
| 13 | qwen-turbo | Qwen | 7.27 | OAuth | Verified (Trusted) |

*OAuth providers use trusted default score (API verification may fail due to product-restricted tokens)*

---

## AI Debate Team Composition (25 LLMs)

### Position 1: ANALYST

| Role | Model | Provider | Score | Type |
|------|-------|----------|-------|------|
| PRIMARY | llama-3.3-70b | Cerebras | 7.85 | API Key |
| Fallback 1 | claude-sonnet-4-5-20250929 | Claude | 7.65 | OAuth |
| Fallback 2 | claude-haiku-4-5-20251001 | Claude | 7.65 | OAuth |
| Fallback 3 | claude-opus-4-5-20251101 | Claude | 7.65 | OAuth |
| Fallback 4 | deepseek-reasoner | DeepSeek | 7.33 | API Key |

### Position 2: PROPOSER

| Role | Model | Provider | Score | Type |
|------|-------|----------|-------|------|
| PRIMARY | deepseek-chat | DeepSeek | 7.33 | API Key |
| Fallback 1 | deepseek-coder | DeepSeek | 7.33 | API Key |
| Fallback 2 | mistral-large-latest | Mistral | 7.33 | API Key |
| Fallback 3 | mistral-medium-latest | Mistral | 7.33 | API Key |
| Fallback 4 | codestral-latest | Mistral | 7.33 | API Key |

### Position 3: CRITIC

| Role | Model | Provider | Score | Type |
|------|-------|----------|-------|------|
| PRIMARY | qwen-turbo | Qwen | 7.27 | OAuth |
| Fallback 1 | qwen-plus | Qwen | 7.27 | OAuth |
| Fallback 2 | qwen-max | Qwen | 7.27 | OAuth |
| Fallback 3 | llama-3.3-70b | Cerebras | 7.85 | API Key (Reused) |
| Fallback 4 | claude-sonnet-4-5-20250929 | Claude | 7.65 | OAuth (Reused) |

### Position 4: SYNTHESIS

| Role | Model | Provider | Score | Type |
|------|-------|----------|-------|------|
| PRIMARY | claude-haiku-4-5-20251001 | Claude | 7.65 | OAuth (Reused) |
| Fallback 1 | claude-opus-4-5-20251101 | Claude | 7.65 | OAuth (Reused) |
| Fallback 2 | deepseek-reasoner | DeepSeek | 7.33 | API Key (Reused) |
| Fallback 3 | deepseek-chat | DeepSeek | 7.33 | API Key (Reused) |
| Fallback 4 | deepseek-coder | DeepSeek | 7.33 | API Key (Reused) |

### Position 5: MEDIATOR

| Role | Model | Provider | Score | Type |
|------|-------|----------|-------|------|
| PRIMARY | mistral-large-latest | Mistral | 7.33 | API Key (Reused) |
| Fallback 1 | mistral-medium-latest | Mistral | 7.33 | API Key (Reused) |
| Fallback 2 | codestral-latest | Mistral | 7.33 | API Key (Reused) |
| Fallback 3 | qwen-turbo | Qwen | 7.27 | OAuth (Reused) |
| Fallback 4 | qwen-plus | Qwen | 7.27 | OAuth (Reused) |

---

## Team Statistics

| Metric | Value |
|--------|-------|
| Total LLMs in Team | 25 |
| Unique LLMs | 13 |
| LLM Reuse Count | 12 |
| Positions | 5 |
| LLMs per Position | 5 (1 primary + 4 fallbacks) |
| Sorted by Score | Yes (NO OAuth priority) |
| Minimum Score | 5.0 |

---

## Verified Providers Summary

| Provider | Auth Type | Score | Models Verified | In Debate Team |
|----------|-----------|-------|-----------------|----------------|
| Cerebras | API Key | 7.85 | 1 | Yes (Position 1 Primary) |
| Claude | OAuth | 7.65 | 3 | Yes (Position 4 Primary + Fallbacks) |
| DeepSeek | API Key | 7.33 | 3 | Yes (Position 2 Primary + Fallbacks) |
| Mistral | API Key | 7.33 | 3 | Yes (Position 5 Primary + Fallbacks) |
| Qwen | OAuth | 7.27 | 3 | Yes (Position 3 Primary + Fallbacks) |

---

## Unverified Providers (Not in Debate Team)

| Provider | Score | Reason |
|----------|-------|--------|
| Hyperbolic | 8.08 | Verification failed |
| Fireworks | 8.06 | Verification failed |
| SambaNova | 8.06 | Verification failed |
| NVIDIA | 8.06 | Verification failed |
| HuggingFace | 7.84 | Verification failed |
| Cloudflare | 7.80 | Verification failed |
| Novita | 7.79 | Verification failed |
| Gemini | 7.75 | Verification failed |
| Replicate | 7.70 | Verification failed |
| Codestral | 7.55 | Verification failed |
| Upstage | 7.53 | Verification failed |
| Inference | 7.35 | Verification failed |
| Zen | 7.29 | All models failed quality validation |
| ZAI | 7.28 | Verification failed |
| SiliconFlow | 7.23 | Verification failed |
| Kimi | 7.11 | Verification failed |
| OpenRouter | 1.25 | No models verified |
| Vercel | 1.25 | No models verified |
| NLPCloud | 1.25 | No models verified |
| VulaVula | 1.25 | No models verified |
| Baseten | 1.25 | No models verified |
| Modal | 1.25 | No models verified |
| Chutes | 1.25 | No models verified |
| Sarvam | 1.25 | No models verified |

---

## Key Points

1. **NO OAuth Priority**: All LLMs sorted purely by score
2. **Cerebras** scored highest (7.85) - gets primary position 1
3. **Claude** OAuth verified (7.65) - distributed across positions
4. **DeepSeek** verified (7.33) - position 2 primary + fallbacks
5. **Mistral** verified (7.33) - position 5 primary + fallbacks
6. **Qwen** OAuth verified (7.27) - position 3 primary + fallbacks
7. **LLM Reuse**: 12 instances reused to fill all 25 positions

---

*Report generated by HelixAgent*
*LLMsVerifier is the Single Source of Truth for all LLM verification and scoring*
