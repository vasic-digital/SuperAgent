# SP6: Website & Course Updates — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Website reflects current project state (41 modules, 43 providers), video courses cover every module, user manuals are complete step-by-step guides, changelog is current.

**Architecture:** Update 5 HTML pages with new module/feature content, update SEO metadata, extend existing video courses with new module references, create comprehensive CHANGELOG.md, rebuild minified assets.

**Tech Stack:** HTML/CSS/JS, PostCSS, Autoprefixer, CSSNano, UglifyJS, npm

**Spec:** `docs/superpowers/specs/2026-03-25-comprehensive-completion-design.md` (SP6 section)

**Depends on:** SP5 complete

---

### Task 1: Update Website Index Page

**Files:**
- Modify: `Website/public/index.html`

- [ ] **Step 1: Read current index.html**

Read `Website/public/index.html` to understand current structure: hero section, feature highlights, statistics, module count references.

- [ ] **Step 2: Update statistics**

Find and update:
- Module count: 35 -> 41
- Provider count: verify 43
- Test count: update to reflect SP2 additions
- Challenge count: 492
- Update copyright year and last-updated date

- [ ] **Step 3: Add feature highlights for new modules**

Add cards/sections for:
- DocProcessor: "Automated documentation analysis and feature extraction"
- HelixQA: "Cross-platform autonomous QA with crash detection"
- LLMOrchestrator: "Headless CLI agent orchestration with circuit breakers"
- VisionEngine: "Computer vision UI analysis with NavigationGraph"

- [ ] **Step 4: Verify rendering**

```bash
cd Website && npm run dev &
# Open http://localhost:8080 and verify
```

- [ ] **Step 5: Commit**

```bash
git add Website/public/index.html
git commit -m "docs(website): update index.html with 41 modules, 4 new feature highlights"
```

---

### Task 2: Update Website Features Page

**Files:**
- Modify: `Website/public/features.html`

- [ ] **Step 1: Read current features.html**

- [ ] **Step 2: Add sections for 6 new modules**

Add feature descriptions for DocProcessor, HelixQA, LLMOrchestrator, VisionEngine, LLMsVerifier, MCP-Servers with:
- Feature title and icon
- 3-4 bullet points of capabilities
- Link to user manual

- [ ] **Step 3: Update security scanning section**

Update to reflect all 7 containerized scanners: Snyk, SonarQube, Trivy, Gosec, Semgrep, KICS, Grype.

- [ ] **Step 4: Update test statistics**

Update test file counts, challenge counts, coverage metrics from SP2 work.

- [ ] **Step 5: Commit**

```bash
git add Website/public/features.html
git commit -m "docs(website): update features.html with 6 new modules, security scanner details"
```

---

### Task 3: Update Website Pricing and Contact Pages

**Files:**
- Modify: `Website/public/pricing.html`
- Modify: `Website/public/contact.html`

- [ ] **Step 1: Verify pricing.html feature matrix**

Read `pricing.html`. Check that the feature matrix includes all current capabilities. Add new modules if they affect tiers.

- [ ] **Step 2: Verify contact.html links**

Check all links in `contact.html` are valid (GitHub repo, docs links, email).

- [ ] **Step 3: Commit if changes needed**

```bash
git add Website/public/pricing.html Website/public/contact.html
git commit -m "docs(website): verify and update pricing and contact pages"
```

---

### Task 4: Update Website Changelog

**Files:**
- Modify: `Website/public/changelog.html`

- [ ] **Step 1: Read current changelog.html**

- [ ] **Step 2: Add comprehensive entry for SP1-SP5 work**

Add a new version entry covering:

```html
<div class="changelog-entry">
    <h3>v2.x.x — Comprehensive Completion (2026-03-25)</h3>

    <h4>Critical Fixes (SP1)</h4>
    <ul>
        <li>Fixed duplicate GetAgentPool() method causing compilation errors</li>
        <li>Moved skills route registration out of per-request handler closure</li>
        <li>Removed 5 dead adapter packages and stale backup directory</li>
    </ul>

    <h4>Test Coverage (SP2)</h4>
    <ul>
        <li>Orchestrated all 492 challenge scripts (was 64)</li>
        <li>Added tests for 12 under-covered packages</li>
        <li>Expanded fuzz testing from 1 to 6 targets</li>
    </ul>

    <h4>Safety & Security (SP3)</h4>
    <ul>
        <li>Fixed channel leaks in Gemini and Qwen ACP providers</li>
        <li>Added context cancellation to all cleanup goroutines</li>
        <li>Ran all 7 security scanners, resolved all HIGH/CRITICAL findings</li>
        <li>Added 6 new stress tests</li>
    </ul>

    <h4>Performance (SP4)</h4>
    <ul>
        <li>Added lazy loading to MCP adapters, formatters, VectorDB, embeddings, BigData</li>
        <li>Created 10 monitoring validation tests</li>
        <li>Established 10 benchmark baselines</li>
        <li>Added backpressure mechanisms to 4 hot paths</li>
    </ul>

    <h4>Documentation (SP5)</h4>
    <ul>
        <li>Fixed 32 broken documentation links</li>
        <li>Added 6 modules to MODULES.md (35 -> 41)</li>
        <li>Created 7 new user manuals and 6 new video courses</li>
        <li>Synchronized CLAUDE.md, AGENTS.md, CONSTITUTION.md</li>
    </ul>
</div>
```

- [ ] **Step 3: Commit**

```bash
git add Website/public/changelog.html
git commit -m "docs(website): add comprehensive changelog entry for SP1-SP5 completion work"
```

---

### Task 5: Update Existing Video Courses with New Module References

**Files:**
- Modify: `Website/video-courses/course-01-*.md` through `course-05-*.md`
- Modify: `Website/video-courses/course-07-*.md` (debate)
- Modify: `Website/video-courses/course-12-*.md` (MCP)
- Modify: `Website/video-courses/course-15-*.md` (memory)
- Modify: `Website/video-courses/course-17-*.md` (security)

- [ ] **Step 1: Update courses 01-05 (core curriculum)**

Add "See Also" references to the 4 new module courses (70-73) where relevant. For example, in the architecture course, mention DocProcessor for documentation analysis.

- [ ] **Step 2: Update course 07 (debate)**

Add references to debate performance optimizer (SP4), reflexion framework improvements.

- [ ] **Step 3: Update course 12 (MCP)**

Update adapter count to 45+. Reference LLMOrchestrator (course 72) for CLI agent management.

- [ ] **Step 4: Update course 15 (memory)**

Add HelixMemory fusion pipeline details, reference VisionEngine for visual memory.

- [ ] **Step 5: Update course 17 (security)**

Expand scanner coverage to all 7 tools. Reference security scanning course (74) for deep dive.

- [ ] **Step 6: Commit**

```bash
git add Website/video-courses/
git commit -m "docs(courses): update existing courses 01-17 with new module cross-references"
```

---

### Task 6: Create Comprehensive CHANGELOG.md

**Files:**
- Create or modify: `CHANGELOG.md` (project root)

- [ ] **Step 1: Check if CHANGELOG.md exists at root**

```bash
test -f CHANGELOG.md && echo "EXISTS" || echo "CREATE"
```

- [ ] **Step 2: Write or append comprehensive entry**

Follow [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
# Changelog

All notable changes to HelixAgent are documented in this file.

## [Unreleased]

### Fixed
- Duplicate GetAgentPool() method in debate/comprehensive/integration.go
- Skills routes registered inside per-request handler closure
- Channel leaks in Gemini and Qwen ACP providers
- Context cancellation missing in query_optimizer cleanup loop
- OAuth manager nil-safety guard
- 32 broken documentation links across docs/

### Added
- Dead code verification challenge
- Test coverage completeness challenge
- 60+ test files for 12 under-covered packages
- 6 fuzz test targets (JSON, schema, protocol, template, config)
- 3 precondition tests (database, redis, API health)
- goleak goroutine leak detection in 5 critical packages
- 6 stress tests (rate limiter, ensemble, debate, streaming, cache, db pool)
- Lazy service provider for on-demand handler initialization
- sync.Once lazy loading for MCP adapters, formatters, VectorDB, embeddings, BigData
- 10 monitoring validation tests for Prometheus metric accuracy
- 10 benchmark baselines documented in docs/performance/BENCHMARKS.md
- Backpressure: exponential backoff in debate optimizer, SSE connection caps, queue depth metrics
- LRU eviction wrapper for debate performance optimizer cache
- Circuit breaker near-cap warning metric
- 6 modules added to MODULES.md (DocProcessor, HelixQA, LLMOrchestrator, VisionEngine, LLMsVerifier, MCP-Servers)
- SQL schema index and guide with ER diagram
- 7 new user manuals (38-44)
- 6 new video courses (70-75)
- 11 new challenge scripts
- docs/ directories for DocProcessor, HelixQA, LLMOrchestrator, VisionEngine
- docs/security/AUTHENTICATION.md, docs/operations/RATE_LIMITING.md, docs/observability/OPENTELEMETRY.md

### Removed
- internal/background/backup/ (stale package duplication, 364KB)
- 5 dead adapter packages (background, observability, events, http, helixqa)
- MODULES.md backup files

### Changed
- Challenge orchestrator expanded from 64 to 492 scripts with tiered execution
- MODULES.md header updated from 33 to 41 modules
- docs/README.md updated to current date and content
- API reference consolidated to single canonical file
- Governance docs (CLAUDE.md, AGENTS.md, CONSTITUTION.md) synchronized with 41 modules
- Website updated with 6 new module features, security scanner details, changelog
- Existing video courses updated with new module cross-references
```

- [ ] **Step 3: Commit**

```bash
git add CHANGELOG.md
git commit -m "docs: create comprehensive CHANGELOG.md covering SP1-SP6 completion work"
```

---

### Task 7: Rebuild Website Assets

**Files:**
- Modify: `Website/styles/main.css` (PostCSS output)
- Modify: `Website/scripts/main.js` (UglifyJS output)

- [ ] **Step 1: Install dependencies if needed**

```bash
cd Website && npm install
```

- [ ] **Step 2: Run build**

```bash
npm run build
```

Expected: CSS minified via PostCSS/CSSNano/Autoprefixer, JS minified via UglifyJS.

- [ ] **Step 3: Verify build output**

```bash
ls -la Website/styles/main.css Website/scripts/main.js
```

- [ ] **Step 4: Preview site**

```bash
npm run preview
```

Verify all pages render correctly at all breakpoints.

- [ ] **Step 5: Commit built assets**

```bash
git add Website/styles/ Website/scripts/
git commit -m "build(website): rebuild minified CSS/JS assets"
```

---

### Task 8: Update SEO Metadata

**Files:**
- Modify: `Website/public/index.html`
- Modify: `Website/public/features.html`

- [ ] **Step 1: Update meta descriptions**

```html
<meta name="description" content="HelixAgent — AI-powered ensemble LLM service with 43 providers, 41 modules, intelligent aggregation, multi-LLM debate, and comprehensive security scanning.">
```

- [ ] **Step 2: Update Open Graph tags**

Update `og:title`, `og:description` for social sharing.

- [ ] **Step 3: Validate all internal links in HTML pages**

Check that all `href` attributes in the 7 HTML pages point to valid targets (other pages, anchors, or external URLs):

```bash
for f in Website/public/*.html; do
    grep -oP 'href="\K[^"]+' "$f" | while read link; do
        # Skip external links and anchors
        case "$link" in http*|mailto*|#*) continue ;; esac
        target="Website/public/$link"
        if [ ! -f "$target" ]; then
            echo "BROKEN: $f -> $link"
        fi
    done
done
```

Expected: Zero broken internal links.

- [ ] **Step 4: Commit**

```bash
git add Website/public/
git commit -m "docs(website): update SEO metadata with current module and provider counts"
```

---

### Task 9: Create Website Completeness Challenge

**Files:**
- Create: `challenges/scripts/website_content_completeness_challenge.sh`

- [ ] **Step 1: Write challenge**

Validates:
- All 7 HTML pages exist and are non-empty
- Module count in index.html matches 41
- Features page mentions all 6 new modules
- Changelog has entry for current work
- `npm run build` succeeds
- Video course count matches expected (43)
- User manual count matches expected (44)
- VIDEO_METADATA.md has entries for courses 70-75

- [ ] **Step 2: Make executable and commit**

```bash
chmod +x challenges/scripts/website_content_completeness_challenge.sh
git add challenges/scripts/website_content_completeness_challenge.sh
git commit -m "test(challenges): add website content completeness challenge"
```

---

### Task 10: Final SP6 Validation

- [ ] **Step 1: Verify website renders**

```bash
cd Website && npm run preview
```

Check all 7 pages at desktop and mobile breakpoints.

- [ ] **Step 2: Verify video course count**

```bash
ls Website/video-courses/course-*.md | wc -l  # Should be 43+
```

- [ ] **Step 3: Verify user manual count**

```bash
ls Website/user-manuals/*.md | wc -l  # Should be 44+
```

- [ ] **Step 4: Run website challenge**

```bash
./challenges/scripts/website_content_completeness_challenge.sh
```

- [ ] **Step 5: Verify CHANGELOG.md exists and is comprehensive**

```bash
wc -l CHANGELOG.md  # Should be substantial
```

- [ ] **Step 6: Tag completion**

```bash
git tag sp6-complete
```

---

### Final Project Validation

After all 6 sub-projects are complete:

- [ ] **Step 1: Full build**

```bash
go build ./...
```

- [ ] **Step 2: Full vet**

```bash
go vet ./internal/...
```

- [ ] **Step 3: Full unit tests**

```bash
GOMAXPROCS=2 nice -n 19 go test ./internal/... -short -count=1 -p 1
```

- [ ] **Step 4: Race detection**

```bash
GOMAXPROCS=2 nice -n 19 go test -race ./internal/... -short -count=1 -p 1
```

- [ ] **Step 5: Run all challenges**

```bash
./challenges/scripts/run_all_challenges.sh
```

- [ ] **Step 6: Tag final completion**

```bash
git tag comprehensive-completion-v1.0
```
