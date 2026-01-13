# Lab 1: Getting Started with HelixAgent

## Lab Overview

**Duration**: 45 minutes
**Difficulty**: Beginner
**Module**: 1 - Introduction to HelixAgent

## Objectives

By completing this lab, you will:
- Clone and set up the HelixAgent repository
- Understand the project structure
- Run your first health check
- Explore the API documentation

## Prerequisites

- Git installed
- Terminal/command line access
- Internet connection
- Optional: Docker installed

## Lab Environment

You will need:
- A terminal window
- A text editor (VS Code recommended)
- A web browser

---

## Exercise 1: Repository Setup (10 minutes)

### Task 1.1: Clone the Repository

```bash
# Clone HelixAgent repository
git clone https://github.com/your-org/helix-agent.git
cd helix-agent
```

**Checkpoint**: Verify you see the following directories:
- [ ] `cmd/`
- [ ] `internal/`
- [ ] `configs/`
- [ ] `docs/`
- [ ] `tests/`

### Task 1.2: Explore Project Structure

```bash
# List top-level directories
ls -la

# View README
cat README.md | head -50
```

**Answer these questions**:
1. What is the main entry point directory? ____________
2. Where are configuration files stored? ____________
3. What Go version is required? ____________

---

## Exercise 2: Configuration Files (10 minutes)

### Task 2.1: Explore Configuration

```bash
# View available configs
ls configs/

# Examine development config
cat configs/development.yaml
```

### Task 2.2: Environment Variables

```bash
# Copy example environment file
cp .env.example .env

# View required variables
grep -E "^[A-Z_]+=" .env.example | head -20
```

**Document the following**:
| Variable | Purpose | Your Value |
|----------|---------|------------|
| `PORT` | | |
| `CLAUDE_API_KEY` | | |
| `DEEPSEEK_API_KEY` | | |

---

## Exercise 3: Build and Run (15 minutes)

### Task 3.1: Build HelixAgent

**Option A: Using Make**
```bash
make build
ls -la bin/
```

**Option B: Using Go directly**
```bash
go build -o bin/helixagent ./cmd/helixagent
```

**Checkpoint**:
- [ ] Binary exists in `bin/` directory
- [ ] No build errors occurred

### Task 3.2: Start the Server

```bash
# Start in development mode
make run-dev

# Or directly
GIN_MODE=debug ./bin/helixagent
```

**Expected output**:
```
[GIN-debug] Listening and serving HTTP on :7061
```

### Task 3.3: Health Check

Open a new terminal:

```bash
# Check if server is running
curl http://localhost:7061/health
```

**Expected response**:
```json
{
  "status": "healthy",
  "version": "1.0.0"
}
```

**Checkpoint**:
- [ ] Server is running on port 7061
- [ ] Health endpoint returns "healthy"

---

## Exercise 4: API Exploration (10 minutes)

### Task 4.1: List Available Models

```bash
curl http://localhost:7061/v1/models | jq
```

**Document the response**:
- Number of models available: ____________
- Model names: ____________

### Task 4.2: Explore Swagger Documentation

Open in your browser:
```
http://localhost:7061/swagger/index.html
```

**Explore these endpoints**:
- [ ] `/v1/chat/completions`
- [ ] `/v1/models`
- [ ] `/health`
- [ ] `/v1/debates`

### Task 4.3: Make Your First API Call

```bash
curl -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "helixagent-debate",
    "messages": [
      {"role": "user", "content": "Hello, HelixAgent!"}
    ]
  }'
```

**Note**: This may fail if no providers are configured. That's expected!

---

## Lab Completion Checklist

- [ ] Repository cloned successfully
- [ ] Project structure understood
- [ ] Configuration files explored
- [ ] Environment file created
- [ ] Server built without errors
- [ ] Server running on port 7061
- [ ] Health check returns "healthy"
- [ ] Swagger UI accessible
- [ ] First API call attempted

---

## Troubleshooting

### Build Fails
```bash
# Ensure Go is installed
go version

# Update dependencies
go mod tidy
go mod download
```

### Server Won't Start
```bash
# Check if port is in use
lsof -i :7061

# Use different port
PORT=8080 ./bin/helixagent
```

### API Calls Fail
- Verify server is running
- Check Content-Type header
- Validate JSON format

---

## Challenge Exercise (Optional)

Create a simple script that:
1. Checks server health
2. Lists available models
3. Reports status

Save as `scripts/health-check.sh`:

```bash
#!/bin/bash
echo "=== HelixAgent Health Check ==="
echo ""
echo "Health Status:"
curl -s http://localhost:7061/health | jq '.status'
echo ""
echo "Available Models:"
curl -s http://localhost:7061/v1/models | jq '.data[].id'
```

---

## Next Lab

Proceed to **Lab 2: Docker Installation** to learn containerized deployment.

---

*Lab Version: 1.0.0*
*Last Updated: January 2026*
