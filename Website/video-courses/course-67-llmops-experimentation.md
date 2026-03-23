# Video Course 67: LLMOps & A/B Experimentation

## Course Overview

**Duration:** 2.5 hours
**Level:** Intermediate to Advanced
**Prerequisites:** Course 01 (Fundamentals), Course 07 (Advanced Providers)

Master LLM operations with HelixAgent. Learn to run continuous evaluations, design A/B experiments for model and prompt comparison, manage versioned prompt templates, and use the REST API to drive data-driven decisions about your AI stack.

---

## Learning Objectives

By the end of this course, you will be able to:

1. Explain the pillars of LLMOps and why they matter for production AI
2. Create and manage A/B experiments with multiple variants
3. Set up continuous evaluation pipelines with configurable metrics
4. Version, diff, and rollback prompt templates
5. Use the LLMOps REST API to automate experiment workflows
6. Interpret experiment results and make evidence-based model decisions

---

## Module 1: What Is LLMOps? (20 min)

### Video 1.1: The Problem with "Ship and Pray" (10 min)

**Topics:**
- Why deploying LLMs without measurement leads to silent quality degradation
- The three pillars of LLMOps: Evaluate, Experiment, Version
- How HelixAgent's LLMOps module addresses each pillar
- The feedback loop: measure, compare, improve, repeat

**The LLMOps Loop:**
```
Prompt Change -> Evaluation -> Comparison -> Decision
      ^                                        |
      +----------------------------------------+
```

### Video 1.2: LLMOps Architecture in HelixAgent (10 min)

**Topics:**
- `InMemoryContinuousEvaluator` -- runs evaluation pipelines
- `InMemoryExperimentManager` -- manages A/B experiments
- `PromptRegistry` -- stores and versions prompt templates
- `LLMOpsSystem` -- facade that wires everything together
- How these components integrate with the REST API handlers

**Key Types:**
```go
// Evaluation pipeline
type InMemoryContinuousEvaluator struct {
    runs      map[string]*EvaluationRun
    datasets  map[string]*Dataset
    evaluator LLMEvaluator
    registry  PromptRegistry
}

// A/B experiments
type InMemoryExperimentManager struct {
    experiments map[string]*Experiment
    metrics     map[string]map[string][]*metricSample
    assignments map[string]map[string]string
}
```

---

## Module 2: A/B Experiments (40 min)

### Video 2.1: Experiment Design (15 min)

**Topics:**
- What constitutes an A/B experiment for LLMs
- Variants: different prompts, models, temperatures, or configurations
- Traffic split: percentage of requests routed to each variant
- Control variant: the baseline for comparison
- Metrics to track: accuracy, latency, cost, user satisfaction

**Experiment Structure:**
```go
type Experiment struct {
    ID           string
    Name         string
    Variants     []*Variant
    TrafficSplit map[string]float64
    Status       ExperimentStatus // draft, running, paused, completed, cancelled
    Metrics      []string
    TargetMetric string
    Winner       string
}

type Variant struct {
    ID            string
    Name          string
    PromptName    string
    PromptVersion string
    ModelName     string
    Parameters    map[string]interface{}
    IsControl     bool
}
```

### Video 2.2: Creating Experiments via the API (15 min)

**Topics:**
- POST /v1/llmops/experiments to create an experiment
- Required: name, at least 2 variants
- Optional: traffic split, metrics, target metric
- Experiment lifecycle: draft, running, paused, completed, cancelled

**Request Example:**
```json
{
  "name": "prompt-v2-vs-v1",
  "description": "Compare new system prompt against current baseline",
  "variants": [
    {
      "name": "control",
      "prompt_name": "system-prompt",
      "prompt_version": "1.0.0",
      "model_name": "deepseek-chat",
      "parameters": {"temperature": 0.7},
      "is_control": true
    },
    {
      "name": "treatment",
      "prompt_name": "system-prompt",
      "prompt_version": "2.0.0",
      "model_name": "deepseek-chat",
      "parameters": {"temperature": 0.7}
    }
  ],
  "traffic_split": {
    "control": 0.5,
    "treatment": 0.5
  },
  "metrics": ["accuracy", "latency_ms", "cost_per_request"],
  "target_metric": "accuracy"
}
```

**Response:**
```json
{
  "id": "exp-abc123",
  "name": "prompt-v2-vs-v1",
  "status": "draft",
  "variants": [...],
  "created_at": "2026-03-23T10:00:00Z"
}
```

### Video 2.3: Monitoring and Analyzing Results (10 min)

**Topics:**
- GET /v1/llmops/experiments to list all experiments
- GET /v1/llmops/experiments/:id to view detailed results
- Traffic assignment: consistent hashing ensures users see the same variant
- Statistical significance: when is enough data collected?
- Determining the winner and completing the experiment

**List Response:**
```json
{
  "experiments": [
    {
      "id": "exp-abc123",
      "name": "prompt-v2-vs-v1",
      "status": "running",
      "winner": ""
    }
  ],
  "total": 1
}
```

---

## Module 3: Continuous Evaluation (35 min)

### Video 3.1: Evaluation Pipelines (15 min)

**Topics:**
- What is continuous evaluation: scheduled quality scoring against golden datasets
- `EvaluationRun`: a single evaluation execution
- `Dataset` and `DatasetSample`: the test data evaluated against
- Configurable metrics: accuracy, coherence, relevance, safety, custom
- The `LLMEvaluator` interface: pluggable evaluation strategy

**Creating an Evaluation Run:**
```json
{
  "name": "weekly-quality-check",
  "description": "Evaluate response quality on golden dataset",
  "dataset": "golden-qa-v3",
  "prompt_name": "system-prompt",
  "prompt_version": "2.0.0",
  "model_name": "helixagent-ensemble",
  "metrics": ["accuracy", "coherence", "relevance"]
}
```

### Video 3.2: Dataset Management (10 min)

**Topics:**
- Dataset types: golden (hand-curated), synthetic (generated), production (sampled)
- `DatasetSample`: input prompt, expected output, metadata
- Versioning datasets alongside prompt versions
- Building datasets from production traffic

**Dataset Structure:**
```go
type Dataset struct {
    ID          string
    Name        string
    Description string
    Type        string // golden, synthetic, production
    Samples     int
    Tags        []string
    CreatedAt   time.Time
}

type DatasetSample struct {
    Input    string
    Expected string
    Metadata map[string]interface{}
}
```

### Video 3.3: Alerting on Quality Regressions (10 min)

**Topics:**
- The `AlertManager` interface for notification on threshold violations
- Configuring minimum score thresholds per metric
- Alert routing: email, Slack, webhook
- Integrating evaluation alerts with Prometheus metrics

---

## Module 4: Prompt Versioning (30 min)

### Video 4.1: The Prompt Registry (15 min)

**Topics:**
- Why prompt versioning matters: reproducibility, rollback, audit
- `PromptVersion`: ID, name, semantic version, content, variables
- The `PromptRegistry` interface: Create, Get, GetLatest, List, Activate, Delete, Render
- Variable substitution with `PromptVariable` types

**PromptVersion Structure:**
```go
type PromptVersion struct {
    ID        string
    Name      string
    Version   string // Semantic version (1.0.0, 2.0.0)
    Content   string
    Variables []PromptVariable
    IsActive  bool
    Author    string
    CreatedAt time.Time
}

type PromptVariable struct {
    Name        string
    Type        string // string, int, float, bool, array
    Required    bool
    Default     interface{}
    Validation  string // Regex pattern
}
```

### Video 4.2: Managing Prompts via the API (10 min)

**Topics:**
- POST /v1/llmops/prompts to create a new version
- GET /v1/llmops/prompts to list all versions
- Activating a version makes it the default for rendering
- Rendering a prompt with variable substitution

**Create Prompt:**
```json
{
  "name": "system-prompt",
  "version": "2.0.0",
  "content": "You are a helpful assistant for {{domain}}. Answer in {{language}}.",
  "variables": [
    {"name": "domain", "type": "string", "required": true},
    {"name": "language", "type": "string", "required": false, "default": "English"}
  ],
  "description": "Updated system prompt with domain specialization",
  "author": "team-lead"
}
```

### Video 4.3: Connecting Prompts to Experiments (5 min)

**Topics:**
- Variants reference prompt name and version
- The evaluation pipeline uses the registry to render prompts
- Workflow: create prompt v2, set up A/B test with v1 vs v2, evaluate, promote winner
- Rollback: re-activate the previous version if quality drops

---

## Module 5: API Endpoints Walkthrough (15 min)

### Video 5.1: Complete Endpoint Reference (15 min)

**Endpoints:**

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/llmops/experiments` | Create an A/B experiment |
| GET | `/v1/llmops/experiments` | List all experiments |
| GET | `/v1/llmops/experiments/:id` | Get experiment details and results |
| POST | `/v1/llmops/evaluate` | Create and run an evaluation |
| GET | `/v1/llmops/prompts` | List all prompt versions |
| POST | `/v1/llmops/prompts` | Create a new prompt version |

**Common Patterns:**
1. Create prompt versions, then reference them in experiment variants
2. Create evaluation pipelines against golden datasets
3. Start experiments, collect metrics, then analyze results
4. Promote the winning configuration to production

---

## Module 6: Hands-On Labs (15 min)

### Lab 1: Run Your First A/B Experiment (5 min)

**Objective:** Compare two prompt versions using the REST API.

**Steps:**
1. Create two prompt versions (v1.0.0 and v2.0.0)
2. Create an experiment with 50/50 traffic split
3. Send 20 test requests and observe variant assignment
4. Check experiment results and identify the winner

### Lab 2: Set Up Continuous Evaluation (5 min)

**Objective:** Create a golden dataset and run an evaluation pipeline.

**Steps:**
1. Create a dataset with 10 question-answer pairs
2. Create an evaluation run with accuracy and coherence metrics
3. Analyze the evaluation results
4. Set up an alert threshold for accuracy below 0.8

### Lab 3: Prompt Version Lifecycle (5 min)

**Objective:** Practice the full prompt versioning workflow.

**Steps:**
1. Create v1.0.0 of a system prompt and activate it
2. Create v2.0.0 with an improvement
3. Render both versions with test variables
4. Use an A/B experiment to determine which performs better
5. Activate the winner and deactivate the loser

---

## Assessment

### Quiz (10 questions)

1. What are the three pillars of LLMOps?
2. What is the minimum number of variants required for an experiment?
3. How does traffic assignment ensure consistency for the same user?
4. What dataset types are supported by the evaluation system?
5. What does the `IsControl` flag on a variant indicate?
6. How do you make a prompt version the active default?
7. What endpoint creates a new evaluation run?
8. What is the difference between `ExperimentStatus` "paused" and "cancelled"?
9. How do `PromptVariable` types support validation?
10. Why should you version datasets alongside prompt versions?

### Practical Assessment

Set up a complete LLMOps pipeline:
1. Create 3 prompt versions (v1, v2, v3) with increasing specificity
2. Build a golden dataset with 20 test cases
3. Run evaluations for each prompt version
4. Set up an A/B experiment for the top 2 performers
5. Analyze results and promote the winner

Deliverables:
1. API request/response logs for each step
2. Evaluation scores comparison table
3. Experiment results with statistical analysis
4. Recommendation document with evidence

---

## Resources

- [LLMOps Module Source](../../LLMOps/llmops/)
- [LLMOps Handler API](../../internal/handlers/llmops_handler.go)
- [Course 07: Advanced Provider Configuration](course-07-advanced-providers.md)
- [HelixAgent Features Overview](../../docs/website/FEATURES.md)
