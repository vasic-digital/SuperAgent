# OpenHands - Usage Guide

## Table of Contents

1. [Getting Started](#getting-started)
2. [Common Workflows](#common-workflows)
3. [Advanced Features](#advanced-features)
4. [Best Practices](#best-practices)
5. [Troubleshooting](#troubleshooting)

---

## Getting Started

### First-Time Setup

1. **Install Dependencies**
   ```bash
   # Check Docker installation
   docker --version
   
   # Clone repository
   git clone https://github.com/OpenHands/OpenHands.git
   cd OpenHands
   ```

2. **Configure LLM**
   ```bash
   # Interactive setup
   make setup-config
   
   # Or manually create config.toml
   cp config.template.toml config.toml
   # Edit config.toml with your API keys
   ```

3. **Start OpenHands**
   ```bash
   # Full application (recommended)
   make run
   
   # Access at http://localhost:3001
   ```

### Basic Configuration

**Minimal config.toml:**
```toml
[llm]
model = "gpt-4o"
api_key = "sk-..."

[core]
workspace_base = "./workspace"
```

**Using Environment Variables:**
```bash
export LLM_API_KEY="sk-..."
export LLM_MODEL="claude-3-5-sonnet-20241022"
export WORKSPACE_BASE="$(pwd)/workspace"
make run
```

---

## Common Workflows

### 1. Code Development

**Create a New Project:**
```
# In the OpenHands chat:
> Create a Python Flask API with endpoints for user CRUD operations

OpenHands will:
1. Create project structure
2. Initialize virtual environment
3. Install Flask and dependencies
4. Create app.py with CRUD endpoints
5. Create database models
6. Add error handling
```

**Modify Existing Code:**
```
> Add authentication middleware to the Flask app

OpenHands will:
1. Read existing app structure
2. Create auth middleware
3. Update routes to use authentication
4. Add JWT token handling
```

**Debug Issues:**
```
> The tests are failing with a database connection error

OpenHands will:
1. Read error logs
2. Check database configuration
3. Verify connection settings
4. Suggest and implement fixes
```

### 2. Web Development

**Build a Frontend:**
```
> Create a React component for a data table with sorting and filtering

OpenHands will:
1. Check existing React setup
2. Create the component
3. Add TypeScript types
4. Implement sorting logic
5. Add filter functionality
```

**API Integration:**
```
> Connect the frontend to the /api/users endpoint

OpenHands will:
1. Create API client functions
2. Add error handling
3. Update components to use data
4. Add loading states
```

### 3. Repository Management

**Explore a New Repository:**
```
> Clone and analyze the facebook/react repository

OpenHands will:
1. Clone the repository
2. Analyze project structure
3. Identify key components
4. Summarize architecture
5. Find entry points
```

**Git Operations:**
```
> Commit these changes with a descriptive message

OpenHands will:
1. Show git status
2. Stage modified files
3. Create commit with message
4. Optional: Push to remote
```

### 4. Testing

**Run Tests:**
```
> Run the test suite and fix any failures

OpenHands will:
1. Discover test files
2. Run tests
3. Analyze failures
4. Fix issues
5. Re-run tests
```

**Add Test Coverage:**
```
> Add unit tests for the auth module

OpenHands will:
1. Read existing auth code
2. Create test file
3. Write test cases
4. Ensure coverage
```

### 5. DevOps Tasks

**Docker Configuration:**
```
> Create a Dockerfile for this Python application

OpenHands will:
1. Analyze application requirements
2. Create optimized Dockerfile
3. Add .dockerignore
4. Test build
```

**CI/CD Setup:**
```
> Set up GitHub Actions for testing and deployment

OpenHands will:
1. Create .github/workflows/ci.yml
2. Configure test job
3. Add linting
4. Set up deployment
```

### 6. Data Analysis

**Process CSV Data:**
```
> Analyze the sales_data.csv file and create visualizations

OpenHands will:
1. Load CSV file
2. Clean and process data
3. Generate summary statistics
4. Create matplotlib plots
5. Save results
```

```python
# Example task file (task.txt)
Load the dataset at /workspace/data.csv
1. Clean missing values
2. Calculate aggregate statistics
3. Generate a report
4. Save to output/report.md
```

Then run:
```bash
python -m openhands.core.main -f task.txt
```

### 7. Web Scraping

**Extract Web Data:**
```
> Scrape product information from example.com/products

OpenHands will:
1. Navigate to the website
2. Extract product data
3. Handle pagination
4. Save to structured format
```

---

## Advanced Features

### 1. Multiple Runtimes

**Docker Runtime (Default):**
```toml
[core]
runtime = "docker"
```

**Local Runtime (Development):**
```toml
[core]
runtime = "local"
```
**Warning:** Local runtime has no isolation!

**Kubernetes Runtime:**
```toml
[core]
runtime = "kubernetes"

[kubernetes]
namespace = "openhands"
pvc_storage_size = "10Gi"
```

### 2. Custom Agents

**Configure Agent Tools:**
```toml
[agent]
enable_browsing = true
enable_jupyter = true
enable_editor = true
enable_think = true
```

**Use Different Agents:**
```toml
[core]
default_agent = "BrowsingAgent"  # For web tasks
# or
default_agent = "RepoExplorerAgent"  # For code exploration
```

### 3. Context Management

**Enable Condenser:**
```toml
[condenser]
type = "llm"  # Automatically summarize old context
llm_config = "condenser"
max_size = 100
```

**Disable History Truncation:**
```toml
[agent]
enable_history_truncation = false
```

### 4. MCP Integration

**Add MCP Server:**
```toml
[mcp]
stdio_servers = [
    {name = "github", command = "npx", args = ["@github/mcp-server"], env = {GITHUB_TOKEN = "${GITHUB_TOKEN}"}}
]
```

Then use:
```
> List my GitHub repositories

> Create an issue for this bug
```

### 5. Microagents

**Repository-Specific Microagents:**

Create `.openhands/microagents/` in your project:

```markdown
---
triggers:
- api
- endpoint
---

When working with API endpoints in this project:
1. Always validate input with Pydantic
2. Use dependency injection for services
3. Add OpenAPI documentation
4. Include rate limiting
```

### 6. Trajectory Replay

**Save Trajectory:**
```toml
[core]
save_trajectory_path = "./trajectories"
```

**Replay Trajectory:**
```toml
[core]
replay_trajectory_path = "./trajectories/session_001.json"
```

### 7. Multi-Model Setup

**Use Different Models for Different Tasks:**
```toml
[llm]
model = "gpt-4o"  # Main model

[llm.secondary_model]
model = "gpt-4o-mini"
for_routing = true
max_input_tokens = 128000
```

---

## Best Practices

### 1. Workspace Organization

```
workspace/
├── project1/
│   ├── src/
│   ├── tests/
│   └── requirements.txt
├── project2/
└── shared/
```

**Configure workspace mount:**
```toml
[core]
workspace_base = "./workspace"
```

### 2. Cost Management

**Set Budget Limits:**
```toml
[core]
max_budget_per_task = 5.0  # $5 USD limit
max_iterations = 100
```

**Monitor Usage:**
- Check logs for token counts
- Use cheaper models for simple tasks
- Enable prompt caching

### 3. Security

**Enable Confirmation Mode:**
```toml
[security]
confirmation_mode = true  # Ask before dangerous operations
```

**Use Security Analyzer:**
```toml
[security]
enable_security_analyzer = true
security_analyzer = "llm"
```

### 4. Performance

**Optimize Context:**
```toml
[llm]
max_message_chars = 10000  # Limit observation size

[condenser]
type = "amortized"  # Smart context compression
```

**Enable Caching:**
```toml
[llm]
caching_prompt = true
```

### 5. Development Workflow

**Pre-commit Hooks:**
```bash
make install-pre-commit-hooks
pre-commit run --all-files
```

**Linting:**
```bash
# Backend
pre-commit run --config ./dev_config/python/.pre-commit-config.yaml

# Frontend
cd frontend && npm run lint:fix
```

---

## Troubleshooting

### Common Issues

**1. Docker Connection Failed**

```bash
# Error: Cannot connect to Docker daemon
# Solution:
sudo usermod -aG docker $USER
# Log out and back in
```

**2. LLM API Key Not Set**

```bash
# Error: API key not found
# Solution:
export LLM_API_KEY="your-key"
# Or configure in config.toml
```

**3. Port Already in Use**

```bash
# Error: Port 3000/3001 already in use
# Solution:
make run BACKEND_PORT=3002 FRONTEND_PORT=3003
```

**4. Container Build Fails**

```bash
# Clear Docker cache
docker system prune -a
# Rebuild
make build
```

**5. Out of Context Space**

```toml
# Enable condenser
[condenser]
type = "llm"

# Or increase limits
[llm]
max_input_tokens = 200000
```

**6. Runtime Timeout**

```toml
# Increase timeout
[sandbox]
timeout = 300
```

### Debug Mode

```bash
# Enable debug logging
export DEBUG=1
export LOG_LEVEL=debug
make run

# View logs
ls logs/
cat logs/llm/2024-*/prompt_*.log
```

### Getting Help

**In OpenHands:**
```
> How do I use the browser tool?

> Explain the error: [paste error]
```

**External Resources:**
- Documentation: [docs.openhands.dev](https://docs.openhands.dev)
- Slack: [dub.sh/openhands](https://dub.sh/openhands)
- GitHub Issues: [OpenHands Issues](https://github.com/OpenHands/OpenHands/issues)

---

## Quick Reference Card

### Essential Commands

| Command | Purpose |
|---------|---------|
| `make build` | Build the project |
| `make run` | Start OpenHands |
| `make setup-config` | Configure LLM |
| `make test` | Run tests |
| `make docker-run` | Run with Docker |

### CLI Options

| Option | Purpose |
|--------|---------|
| `-t "task"` | Specify task |
| `-f file.txt` | Task from file |
| `-c config.toml` | Custom config |
| `--no-auto-continue` | Interactive mode |

### Key Configuration

| Section | Purpose |
|---------|---------|
| `[llm]` | Model settings |
| `[agent]` | Agent tools |
| `[sandbox]` | Runtime config |
| `[security]` | Security settings |
| `[condenser]` | Context compression |

---

*For more details, see the [API Reference](./API.md) and [Architecture](./ARCHITECTURE.md) documentation.*
