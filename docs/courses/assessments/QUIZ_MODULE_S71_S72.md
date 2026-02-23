# Quiz: Modules S7.1–S7.2 (Advanced AI/ML Modules)

## Instructions

- **Total Questions**: 25
- **Time Limit**: 50 minutes
- **Passing Score**: 80% (20/25)
- **Format**: Multiple choice (single answer unless specified)

---

## Section 1: Agentic Module (Questions 1-5)

### Q1. In the `digital.vasic.agentic` module, what is the purpose of `WorkflowState`?

A) To define the graph structure (nodes and edges)
B) A thread-safe mutable bag of key-value pairs threaded through all node executions
C) The function signature for individual workflow steps
D) The configuration struct for the Workflow executor

---

### Q2. What does `NodeOutput.Done = true` signal in an Agentic workflow?

A) The current node encountered an error
B) The workflow should skip the next node
C) The workflow is complete and should stop executing further nodes
D) The WorkflowState should be reset

---

### Q3. Which field of `NodeOutput` enables dynamic routing at runtime?

A) `Data`
B) `Done`
C) `NextNode`
D) `Error`

---

### Q4. How does `NodeOutput.Data` flow between nodes?

A) It is stored in WorkflowState automatically
B) It is passed as the `input interface{}` parameter to the next node's handler
C) It is broadcast to all nodes in the graph simultaneously
D) It is written to the edge metadata for the next node to read

---

### Q5. In HelixAgent, which component uses the Agentic module for multi-phase development workflows?

A) `internal/services/boot_manager.go`
B) `internal/services/speckit_orchestrator.go`
C) `internal/handlers/chat_handler.go`
D) `internal/debate/orchestrator.go`

---

## Section 2: LLMOps Module (Questions 6-10)

### Q6. What is the difference between `DatasetTypeGolden` and `DatasetTypeProduction` in the LLMOps module?

A) Golden datasets are stored in PostgreSQL; production datasets are in-memory
B) Golden datasets are hand-curated with ground truth; production datasets are sampled from real traffic
C) Golden datasets have expected outputs; production datasets do not allow scoring
D) They are identical; the type field is informational only

---

### Q7. What does `AggregateMetrics.PassRate` measure in an `EvaluationRun`?

A) The fraction of examples where the LLM returned a non-empty response
B) The fraction of examples scoring above a configurable threshold
C) The fraction of examples that completed without timeout
D) The fraction of examples where the response matched exactly

---

### Q8. In the `InMemoryExperimentManager`, what does `TrafficSplit: 0.5` mean?

A) Only 50% of experiments will be analyzed
B) The experiment will run for 50% of the total timeout duration
C) 50% of incoming requests go to the control config; 50% go to the treatment config
D) The control group is limited to 50 examples maximum

---

### Q9. When should you use `mgr.RollbackPrompt()` in the LLMOps prompt versioning workflow?

A) When a new prompt version improves MeanScore by more than 10%
B) When a new prompt version causes quality regression on the evaluation dataset
C) After every deployment to reset to factory defaults
D) When the experiment p-value is below 0.05

---

### Q10. How does HelixAgent's LLMsVerifier startup pipeline relate to `InMemoryContinuousEvaluator`?

A) LLMsVerifier replaces the evaluator entirely and does not use the LLMOps module
B) The LLMsVerifier 8-test verification pipeline is implemented as an evaluation run against a built-in dataset
C) LLMsVerifier only uses the Experiment Manager, not the Evaluator
D) They are unrelated; LLMsVerifier uses direct API calls without evaluation abstraction

---

## Section 3: SelfImprove Module (Questions 11-15)

### Q11. What is the primary difference between `ExplicitFeedback` and `ImplicitFeedback` in the SelfImprove module?

A) ExplicitFeedback is stored permanently; ImplicitFeedback is discarded after one batch
B) ExplicitFeedback is directly provided by users (e.g., thumbs up/down); ImplicitFeedback is inferred from user behavior
C) ExplicitFeedback only supports binary signals; ImplicitFeedback supports continuous scores
D) ExplicitFeedback requires a UserID; ImplicitFeedback is always anonymous

---

### Q12. What is a `PreferencePair` in the SelfImprove module?

A) Two consecutive messages in a conversation context
B) A pair of (prompt, response) examples scored with the same reward model
C) A training example with a preferred (chosen) response and a less preferred (rejected) response for the same prompt
D) A pair of reward models compared during evaluation

---

### Q13. In `SelfRefinementConfig`, when does a `SelfRefinementLoop` stop before reaching `MaxIterations`?

A) When `Done: true` is returned from the critique handler
B) When the reward model score exceeds `ScoreThreshold`
C) When the refined response is identical to the input
D) When the LLM client returns an empty response

---

### Q14. What is the correct order of steps in the `SelfRefinementLoop` for a single iteration?

A) Score → Refine → Critique → Check threshold
B) Critique → Score → Refine → Check threshold
C) Refine → Critique → Score → Check threshold
D) Score → Critique → Refine → Check threshold

---

### Q15. Why is anonymizing feedback data important before calling `RewardModel.Train()`?

A) Because the training algorithm performs better on anonymized data
B) Because feedback data may contain PII (user IDs, conversation content) that must not be used in training
C) Because anonymous data reduces training time by 50%
D) Because the InMemoryRewardModel only accepts anonymous data

---

## Section 4: Planning Module (Questions 16-20)

### Q16. Which planning algorithm is most appropriate for decomposing a software feature into milestones and steps when the overall structure is known in advance?

A) MCTS (Monte Carlo Tree Search)
B) Tree of Thoughts
C) HiPlan (Hierarchical Planning)
D) A/B Experiment Manager

---

### Q17. In MCTS, what does the `ExplorationC` parameter (also called the UCB constant) control?

A) The maximum number of child nodes per parent
B) The balance between exploring new nodes and exploiting known high-scoring nodes
C) The maximum depth of the search tree
D) The timeout per node evaluation

---

### Q18. What is the `BeamWidth` parameter in `TreeOfThoughtsConfig`?

A) The number of new thoughts generated per node (also called BranchingFactor)
B) The maximum depth of the thought tree
C) The number of top-scoring branches kept at each depth level (beam search)
D) The minimum score a thought must achieve to be expanded

---

### Q19. For a task like "find a mathematical proof for a theorem" where the correct approach is completely unknown, which planning algorithm is most suitable?

A) HiPlan — because it decomposes the problem hierarchically
B) MCTS — because mathematics is game-like
C) Tree of Thoughts — because it explores multiple thought branches and backtracks
D) SelfRefinementLoop — because proofs improve iteratively

---

### Q20. In HelixAgent's SpecKit orchestrator, when is the `GranularityRefactoring` level triggered?

A) For any change to a single file
B) For large refactoring tasks requiring hierarchical decomposition across multiple components
C) When the user explicitly requests HiPlan via the API
D) Only when the Agentic module is not available

---

## Section 5: Benchmark Module (Questions 21-25)

### Q21. What does `pass@1` measure in the HumanEval benchmark?

A) The accuracy of the model on 1-shot prompts
B) The fraction of coding problems where the model's first-attempt code passes all unit tests
C) The model's accuracy when given one example before the question
D) The pass rate on the easiest 1% of problems

---

### Q22. In `RunConfig`, what is the effect of setting `MaxExamples: 50` on a benchmark with 164 total examples?

A) The benchmark runs all 164 examples but only reports results for the first 50
B) The benchmark runs only the first 50 examples, reducing cost and time
C) The benchmark repeats 50 examples multiple times for statistical stability
D) The benchmark randomly samples 50 examples from the full dataset each time

---

### Q23. What does a `ComparisonReport.PValue < 0.05` indicate?

A) The winning provider's advantage may be due to random chance
B) The difference in performance between the two providers is statistically significant at the 95% confidence level
C) The experiment needs at least 5 more examples to be conclusive
D) The cost difference between providers exceeds 5%

---

### Q24. Why must benchmark runs always use resource limits like `GOMAXPROCS=2` and `Parallelism: 2`?

A) The Benchmark module requires these settings to compile correctly
B) The HelixAgent API enforces a maximum of 2 concurrent benchmark requests
C) The host runs mission-critical processes and exceeding 30-40% resource usage can cause system crashes
D) These limits improve benchmark accuracy by reducing interference

---

### Q25. Which HelixAgent component uses `BenchmarkResult` metrics to compute provider scoring components like `CostEffectiveness` and `ModelEfficiency`?

A) `internal/handlers/chat_handler.go`
B) `internal/verifier/startup.go` and the LLMsVerifier pipeline
C) `internal/services/debate_service.go`
D) `internal/adapters/containers/adapter.go`

---

## Answer Key

| Q | Answer | Q | Answer | Q | Answer | Q | Answer | Q | Answer |
|---|--------|---|--------|---|--------|---|--------|---|--------|
| 1 | B | 6 | B | 11 | B | 16 | C | 21 | B |
| 2 | C | 7 | B | 12 | C | 17 | B | 22 | B |
| 3 | C | 8 | C | 13 | B | 18 | C | 23 | B |
| 4 | B | 9 | B | 14 | D | 19 | C | 24 | C |
| 5 | B | 10 | B | 15 | B | 20 | B | 25 | B |

---

## Scoring

- **23-25 (92-100%)**: Excellent — Ready for Level 6 certification
- **20-22 (80-88%)**: Good — Review missed topics and retry
- **17-19 (68-76%)**: Fair — Additional study recommended before certification
- **Below 17**: Review Modules S7.1-S7.2 before attempting certification

## Level 6 Certification Requirements

After passing this quiz with 80%+:

1. Complete Labs 9-13 (all 5 AI/ML module labs)
2. Pass all 5 module challenge scripts with 100% test pass rate:
   - `./challenges/scripts/agentic_challenge.sh`
   - `./challenges/scripts/llmops_challenge.sh`
   - `./challenges/scripts/selfimprove_challenge.sh`
   - `./challenges/scripts/planning_challenge.sh`
   - `./challenges/scripts/benchmark_challenge.sh`
3. Build a complete end-to-end AI agent pipeline demonstrating:
   - Agentic workflow with at least 4 nodes and dynamic routing
   - LLMOps A/B experiment with documented results
   - SelfRefinementLoop with >20% quality improvement
   - HiPlan or MCTS on a real planning task
   - Benchmark comparison report with p-value analysis

---

*Quiz Version: 1.0.0*
*Last Updated: February 2026*
*Modules Covered: S7.1.1 (Agentic), S7.1.2 (LLMOps), S7.1.3 (SelfImprove), S7.2.1 (Planning), S7.2.2 (Benchmark)*
