# Comprehensive LLM Provider Integration Plan

## Executive Summary

This document outlines the complete plan to integrate **50+ new LLM providers** into HelixAgent and LLMsVerifier, achieving global coverage of all available LLM services.

**Current State:** 10 directly implemented providers
**Target State:** 60+ providers with full support

---

## Provider Categories

### Category 1: Major Cloud Providers (Priority: HIGH)

| Provider | Auth Type | Base URL | Models | Status |
|----------|-----------|----------|--------|--------|
| **OpenAI** | API Key | `https://api.openai.com/v1` | GPT-4o, GPT-4.5, o1, o3 | TO ADD |
| **AWS Bedrock** | AWS IAM/Keys | `https://bedrock-runtime.{region}.amazonaws.com` | Claude, Llama, Titan, Mistral | TO ADD |
| **Azure OpenAI** | API Key/AD | `https://{resource}.openai.azure.com` | GPT-4o, GPT-4 Turbo | TO ADD |
| **Google Vertex AI** | Service Account/OAuth | `https://{region}-aiplatform.googleapis.com` | Gemini, PaLM, Claude, Llama | TO ADD |
| **Google AI Studio** | API Key | `https://generativelanguage.googleapis.com/v1beta` | Gemini Pro, Gemini Flash | EXISTING (via Gemini) |

### Category 2: Frontier Model Providers (Priority: HIGH)

| Provider | Auth Type | Base URL | Models | Status |
|----------|-----------|----------|--------|--------|
| **xAI (Grok)** | API Key | `https://api.x.ai/v1` | Grok 3, Grok 4, Grok Vision | TO ADD |
| **Cohere** | API Key | `https://api.cohere.ai/v1` | Command R+, Command A, Aya | TO ADD |
| **AI21 Labs** | API Key | `https://api.ai21.com/studio/v1` | Jamba 1.5, Jurassic-2 | TO ADD |
| **Perplexity** | API Key | `https://api.perplexity.ai` | Sonar Pro, Sonar Small | TO ADD |
| **Reka AI** | API Key | `https://api.reka.ai/v1` | Reka Core, Reka Flash, Reka Edge | TO ADD |

### Category 3: Fast Inference Providers (Priority: HIGH)

| Provider | Auth Type | Base URL | Models | Status |
|----------|-----------|----------|--------|--------|
| **Groq** | API Key | `https://api.groq.com/openai/v1` | Llama 3.3, Mixtral, Whisper | TO ADD |
| **SambaNova** | API Key | `https://api.sambanova.ai/v1` | Llama, Qwen, DeepSeek | TO ADD |
| **Together AI** | API Key | `https://api.together.xyz/v1` | 100+ open models | TO ADD |
| **Fireworks AI** | API Key | `https://api.fireworks.ai/inference/v1` | Llama, Mixtral, DeepSeek | TO ADD |
| **Cerebras** | API Key | `https://api.cerebras.ai/v1` | Llama, Qwen, GPT-OSS | EXISTING |

### Category 4: Chinese Providers (Priority: MEDIUM-HIGH)

| Provider | Auth Type | Base URL | Models | Status |
|----------|-----------|----------|--------|--------|
| **Baidu ERNIE** | API Key | `https://aip.baidubce.com/rpc/2.0/ai_custom` | ERNIE 4.0, ERNIE Speed | TO ADD |
| **ByteDance Doubao** | API Key | `https://ark.cn-beijing.volces.com/api/v3` | Doubao Pro, Doubao Lite | TO ADD |
| **Zhipu AI (GLM)** | API Key | `https://open.bigmodel.cn/api/paas/v4` | GLM-4, GLM-4V, CogView | TO ADD |
| **Moonshot (Kimi)** | API Key | `https://api.moonshot.cn/v1` | Moonshot v1-8k/32k/128k | TO ADD |
| **MiniMax** | API Key | `https://api.minimax.chat/v1` | abab6, abab5.5 | TO ADD |
| **Baichuan** | API Key | `https://api.baichuan-ai.com/v1` | Baichuan-Turbo | TO ADD |
| **01.AI (Yi)** | API Key | `https://api.01.ai/v1` | Yi-Large, Yi-Medium | TO ADD |
| **StepFun** | API Key | `https://api.stepfun.com/v1` | Step-1, Step-2 | TO ADD |
| **SiliconFlow** | API Key | `https://api.siliconflow.cn/v1` | Multiple open models | TO ADD |

### Category 5: Alternative/Aggregator Providers (Priority: MEDIUM)

| Provider | Auth Type | Base URL | Models | Status |
|----------|-----------|----------|--------|--------|
| **DeepInfra** | API Key | `https://api.deepinfra.com/v1/openai` | 100+ models | TO ADD |
| **Replicate** | API Key | `https://api.replicate.com/v1` | Various open models | TO ADD |
| **HuggingFace** | API Key | `https://api-inference.huggingface.co` | 1000+ models | TO ADD |
| **Hyperbolic** | API Key | `https://api.hyperbolic.xyz/v1` | Llama, DeepSeek, Qwen | TO ADD |
| **Novita AI** | API Key | `https://api.novita.ai/v3/openai` | Various models | TO ADD |
| **Lepton AI** | API Key | `https://api.lepton.ai/api/v1` | Llama, Mistral | TO ADD |

### Category 6: Specialized/Edge Providers (Priority: MEDIUM)

| Provider | Auth Type | Base URL | Models | Status |
|----------|-----------|----------|--------|--------|
| **Cloudflare Workers AI** | API Key | `https://api.cloudflare.com/client/v4/accounts/{id}/ai` | 50+ edge models | TO ADD |
| **NVIDIA NIM** | API Key | `https://integrate.api.nvidia.com/v1` | Llama, Nemotron, NV models | TO ADD |
| **Databricks** | API Key | `https://{workspace}.cloud.databricks.com/serving-endpoints` | DBRX, custom models | TO ADD |
| **Upstage** | API Key | `https://api.upstage.ai/v1/solar` | Solar Pro, Solar Mini | TO ADD |
| **Anyscale** | API Key | `https://api.endpoints.anyscale.com/v1` | Open models | TO ADD |

### Category 7: Free/No-Key Providers (Priority: HIGH)

| Provider | Auth Type | Base URL | Models | Status |
|----------|-----------|----------|--------|--------|
| **Puter.js** | None | `https://api.puter.com/ai/chat` | 500+ models | TO ADD |
| **mlvoca** | None | `https://mlvoca.com/api/generate` | DeepSeek R1, TinyLlama | TO ADD |
| **g4f Providers** | None | Various | Multiple | TO ADD |
| **OllamaFreeAPI** | None | Gateway | 50+ models | TO ADD |
| **GitHub Models** | GitHub OAuth | `https://models.github.ai/v1` | GPT-4o, Llama, DeepSeek | TO ADD |

### Category 8: Enterprise/Managed Providers (Priority: LOW)

| Provider | Auth Type | Base URL | Models | Status |
|----------|-----------|----------|--------|--------|
| **IBM watsonx** | API Key/IAM | `https://us-south.ml.cloud.ibm.com` | Granite, Llama | TO ADD |
| **Oracle OCI AI** | API Key/IAM | `https://inference.generativeai.{region}.oci.oraclecloud.com` | Cohere, Llama | TO ADD |
| **Salesforce Einstein** | OAuth | `https://api.salesforce.com/einstein/llm` | Einstein GPT | TO ADD |

---

## Authentication Types Summary

### 1. API Key Authentication
- **Providers:** 45+ providers
- **Implementation:** Bearer token in Authorization header
- **Storage:** `.env` file with `{PROVIDER}_API_KEY`

### 2. OAuth 2.0 / OAuth 1.0
- **Providers:** Claude CLI, Qwen CLI, GitHub Models, Salesforce
- **Implementation:** Token refresh flow
- **Storage:** Token files with refresh capability

### 3. AWS IAM / Service Account
- **Providers:** AWS Bedrock, Google Vertex AI
- **Implementation:** SDK-based authentication
- **Storage:** IAM credentials or service account JSON

### 4. No Authentication (Free)
- **Providers:** Puter.js, mlvoca, OllamaFreeAPI, g4f
- **Implementation:** Direct HTTP calls
- **Notes:** May have rate limits or require device ID

### 5. Device/Session Based
- **Providers:** Zen (OpenCode)
- **Implementation:** X-Device-ID header
- **Storage:** Generated device ID

---

## Dynamic Credential Acquisition

### Methods to Implement

#### 1. **CLI Tool Authentication Extraction**
```
Provider: Claude, Qwen, GitHub CLI, AWS CLI, gcloud
Method: Read credentials from CLI config files
Files:
  - ~/.claude/.credentials.json
  - ~/.qwen/oauth_creds.json
  - ~/.config/gh/hosts.yml
  - ~/.aws/credentials
  - ~/.config/gcloud/application_default_credentials.json
```

#### 2. **Browser Session Extraction**
```
Provider: Google AI Studio, Perplexity, ChatGPT
Method: Extract session tokens from browser storage
Files:
  - ~/.config/google-chrome/Default/Cookies
  - Browser localStorage dumps
Note: Requires user consent and browser extension
```

#### 3. **Free Tier Auto-Registration**
```
Provider: OpenRouter, Groq, Cohere, AI21
Method: Automated account creation with temp email
Services:
  - Temp email APIs (guerrillamail, tempail)
  - CAPTCHA solving services (optional)
Note: Must comply with ToS
```

#### 4. **OAuth Device Flow**
```
Provider: GitHub Models, Azure AD
Method: Device authorization grant flow
Process:
  1. Request device code
  2. Display user code
  3. User authenticates in browser
  4. Poll for access token
```

#### 5. **API Key Generation APIs**
```
Provider: Some providers offer API key generation endpoints
Method: Programmatic key creation after login
```

### Credential Storage

**File:** `.dynamic.env` (git-ignored)
```env
# Auto-generated credentials - DO NOT COMMIT
# Generated at: 2026-01-22T19:30:00Z

# Free tier accounts
OPENROUTER_API_KEY=sk-or-v1-xxx
GROQ_API_KEY=gsk_xxx
COHERE_API_KEY=xxx

# OAuth tokens (with refresh)
GITHUB_MODELS_TOKEN=gho_xxx
GITHUB_MODELS_REFRESH_TOKEN=ghr_xxx

# Session tokens
PERPLEXITY_SESSION=xxx

# Device IDs
ZEN_DEVICE_ID=xxx
```

---

## Feature Support Matrix

| Provider | Chat | Stream | Tools | Vision | Embed | Audio | MCP | LSP |
|----------|------|--------|-------|--------|-------|-------|-----|-----|
| OpenAI | YES | YES | YES | YES | YES | YES | - | - |
| xAI Grok | YES | YES | YES | YES | - | - | YES | - |
| Cohere | YES | YES | YES | - | YES | - | - | - |
| Groq | YES | YES | YES | - | - | YES | - | - |
| Perplexity | YES | YES | - | - | - | - | - | - |
| AWS Bedrock | YES | YES | YES | YES | YES | - | - | - |
| Azure OpenAI | YES | YES | YES | YES | YES | YES | - | - |
| Vertex AI | YES | YES | YES | YES | YES | - | - | - |
| HuggingFace | YES | YES | - | YES | YES | YES | - | - |

---

## Implementation Order

### Phase 1: High-Value Providers (Week 1-2)
1. OpenAI (most requested)
2. xAI Grok (frontier model)
3. Groq (fast inference)
4. Cohere (enterprise)
5. Together AI (aggregator)

### Phase 2: Cloud Providers (Week 3-4)
1. AWS Bedrock
2. Azure OpenAI
3. Google Vertex AI
4. NVIDIA NIM

### Phase 3: Chinese Providers (Week 5-6)
1. Zhipu AI (GLM)
2. Baidu ERNIE
3. ByteDance Doubao
4. Moonshot Kimi
5. MiniMax

### Phase 4: Free Providers (Week 7)
1. Puter.js
2. GitHub Models
3. mlvoca
4. Enhanced OpenRouter free models

### Phase 5: Remaining Providers (Week 8-10)
1. All remaining providers
2. Dynamic credential system
3. Tests and challenges

---

## Files to Create/Modify

### New Provider Implementations
```
internal/llm/providers/
├── openai/openai.go
├── xai/xai.go
├── groq/groq.go
├── cohere/cohere.go
├── together/together.go
├── fireworks/fireworks.go
├── aws_bedrock/bedrock.go
├── azure_openai/azure.go
├── vertex/vertex.go
├── nvidia_nim/nvidia.go
├── perplexity/perplexity.go
├── ai21/ai21.go
├── reka/reka.go
├── zhipu/zhipu.go
├── baidu_ernie/ernie.go
├── bytedance_doubao/doubao.go
├── moonshot/moonshot.go
├── minimax/minimax.go
├── baichuan/baichuan.go
├── stepfun/stepfun.go
├── siliconflow/siliconflow.go
├── deepinfra/deepinfra.go
├── replicate/replicate.go
├── huggingface/huggingface.go
├── hyperbolic/hyperbolic.go
├── novita/novita.go
├── cloudflare/cloudflare.go
├── databricks/databricks.go
├── upstage/upstage.go
├── puter/puter.go
├── mlvoca/mlvoca.go
├── github_models/github.go
└── sambanova/sambanova.go
```

### LLMsVerifier Updates
```
LLMsVerifier/
├── internal/verifier/providers/
│   ├── openai_adapter.go
│   ├── xai_adapter.go
│   ├── groq_adapter.go
│   └── ... (one per provider)
├── internal/verifier/scoring/
│   └── provider_scoring.go (updated)
└── configs/
    └── providers.yaml (new provider configs)
```

### Dynamic Credentials
```
internal/auth/
├── dynamic_credentials.go
├── cli_extractors/
│   ├── claude_extractor.go
│   ├── qwen_extractor.go
│   ├── github_extractor.go
│   ├── aws_extractor.go
│   └── gcloud_extractor.go
├── free_registration/
│   ├── temp_email.go
│   ├── auto_register.go
│   └── rate_limiter.go
└── oauth_flows/
    ├── device_flow.go
    └── token_manager.go
```

---

## Success Criteria

1. **Provider Count:** 60+ providers fully integrated
2. **Test Coverage:** 100% for all new code
3. **Verification:** All providers pass LLMsVerifier 8-test pipeline
4. **Documentation:** Complete API docs for each provider
5. **Challenges:** 50+ challenge scripts validating real usage
6. **Website:** Fully functional with Docker/Podman compose
