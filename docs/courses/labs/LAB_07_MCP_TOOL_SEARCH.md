# Lab 7: MCP Tool Search Integration

## Lab Overview

**Duration**: 60 minutes
**Difficulty**: Intermediate
**Module**: 13 - MCP Tool Search and Discovery

## Objectives

By completing this lab, you will:
- Use MCP Tool Search to discover available tools
- Implement AI-powered tool suggestions
- Search for MCP adapters
- Build a custom tool discovery workflow

## Prerequisites

- Labs 1-6 completed
- HelixAgent running
- MCP servers configured
- Understanding of MCP protocol basics

---

## Exercise 1: Exploring Tool Search Endpoints (15 minutes)

### Task 1.1: List Available MCP Endpoints

```bash
# Get MCP capabilities
curl http://localhost:7061/v1/mcp/capabilities | jq

# Get MCP tools list
curl http://localhost:7061/v1/mcp/tools | jq

# Get MCP categories
curl http://localhost:7061/v1/mcp/categories | jq

# Get MCP statistics
curl http://localhost:7061/v1/mcp/stats | jq
```

**Record the available tool count**: ____________

### Task 1.2: Understand the Tool Search Response Structure

```json
{
  "results": [
    {
      "name": "Read",
      "description": "Reads a file from the local filesystem",
      "category": "filesystem",
      "parameters": {...}
    }
  ],
  "count": 21,
  "query": "file"
}
```

---

## Exercise 2: Basic Tool Search (15 minutes)

### Task 2.1: Search for File-Related Tools

```bash
# Search for file tools
curl "http://localhost:7061/v1/mcp/tools/search?q=file" | jq

# Verify results have actual tools
curl "http://localhost:7061/v1/mcp/tools/search?q=file" | jq '.count'
```

**Expected tools**: Read, Write, Edit, Glob, FileInfo

### Task 2.2: Search for Different Categories

```bash
# Search for git tools
curl "http://localhost:7061/v1/mcp/tools/search?q=git" | jq

# Search for search tools
curl "http://localhost:7061/v1/mcp/tools/search?q=search" | jq

# Search for web tools
curl "http://localhost:7061/v1/mcp/tools/search?q=web" | jq

# Search for bash/command tools
curl "http://localhost:7061/v1/mcp/tools/search?q=bash" | jq
```

**Document Your Search Results**:
| Query | Count | Example Tools |
|-------|-------|---------------|
| file | | |
| git | | |
| search | | |
| web | | |
| bash | | |

### Task 2.3: POST Method Search

```bash
# Search using POST method
curl -X POST http://localhost:7061/v1/mcp/tools/search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "file operations",
    "limit": 10
  }' | jq
```

---

## Exercise 3: Tool Suggestions (15 minutes)

### Task 3.1: Get Tool Suggestions for Prompts

The tool suggestions endpoint uses AI to recommend tools based on natural language prompts.

```bash
# Get suggestions for listing files
curl "http://localhost:7061/v1/mcp/tools/suggestions?prompt=list%20files%20in%20directory" | jq

# Get suggestions for searching code
curl "http://localhost:7061/v1/mcp/tools/suggestions?prompt=search%20for%20text%20in%20files" | jq

# Get suggestions for editing files
curl "http://localhost:7061/v1/mcp/tools/suggestions?prompt=edit%20a%20configuration%20file" | jq

# Get suggestions for running commands
curl "http://localhost:7061/v1/mcp/tools/suggestions?prompt=run%20a%20shell%20command" | jq
```

### Task 3.2: Understanding Suggestion Response

```json
{
  "prompt": "list files in directory",
  "suggestions": [
    {
      "tool": "Glob",
      "confidence": 0.95,
      "reason": "Glob is ideal for listing files with pattern matching"
    },
    {
      "tool": "Bash",
      "confidence": 0.85,
      "reason": "Bash can execute ls command for file listing"
    }
  ]
}
```

### Task 3.3: Test Various Prompts

Create a table of prompts and their suggested tools:

| Prompt | Top Suggestion | Confidence |
|--------|---------------|------------|
| "read the contents of main.go" | | |
| "find all test files" | | |
| "search for TODO comments" | | |
| "commit my changes" | | |
| "write to config.yaml" | | |

---

## Exercise 4: Adapter Search (10 minutes)

### Task 4.1: Search for MCP Adapters

MCP adapters provide pre-built integrations with external services.

```bash
# Search for GitHub adapter
curl "http://localhost:7061/v1/mcp/adapters/search?q=github" | jq

# Search for filesystem adapter
curl "http://localhost:7061/v1/mcp/adapters/search?q=filesystem" | jq

# Search for database adapters
curl "http://localhost:7061/v1/mcp/adapters/search?q=postgres" | jq

# Search for communication adapters
curl "http://localhost:7061/v1/mcp/adapters/search?q=slack" | jq
```

### Task 4.2: Available Adapter Categories

| Category | Adapters | Description |
|----------|----------|-------------|
| Core | filesystem, memory, fetch | Basic operations |
| VCS | git, github, gitlab | Version control |
| Database | postgres, sqlite, redis, mongodb | Data storage |
| Cloud | docker, kubernetes, aws-s3 | Cloud services |
| Communication | slack, notion | Team tools |
| Search | brave-search | Web search |
| Vector | chroma, qdrant, weaviate | Vector databases |

---

## Exercise 5: Building a Tool Discovery Workflow (5 minutes)

### Task 5.1: Create a Discovery Script

Create a script that discovers tools based on user intent:

```bash
#!/bin/bash
# tool_discovery.sh

HELIXAGENT_URL="${HELIXAGENT_URL:-http://localhost:7061}"

discover_tools() {
    local prompt="$1"
    local encoded_prompt=$(echo "$prompt" | sed 's/ /%20/g')

    echo "=== Discovering tools for: $prompt ==="
    echo ""

    # Get suggestions
    echo "Tool Suggestions:"
    curl -s "${HELIXAGENT_URL}/v1/mcp/tools/suggestions?prompt=${encoded_prompt}" | jq '.suggestions[] | "\(.tool): \(.confidence * 100)% - \(.reason)"'

    echo ""
    echo "Search Results:"
    # Extract keywords and search
    curl -s "${HELIXAGENT_URL}/v1/mcp/tools/search?q=$(echo $prompt | awk '{print $1}')" | jq '.results[].name'
}

# Example usage
discover_tools "read file contents"
discover_tools "search for patterns"
discover_tools "execute command"
```

### Task 5.2: Test the Discovery Workflow

```bash
chmod +x tool_discovery.sh
./tool_discovery.sh
```

---

## Lab Completion Checklist

- [ ] Explored MCP capabilities endpoint
- [ ] Performed tool searches with different queries
- [ ] Used POST method for search
- [ ] Tested tool suggestions with prompts
- [ ] Explored adapter search
- [ ] Built a discovery workflow script

---

## Tool Search Reference

### Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/v1/mcp/tools` | GET | List all tools |
| `/v1/mcp/tools/search` | GET/POST | Search tools |
| `/v1/mcp/tools/suggestions` | GET | AI suggestions |
| `/v1/mcp/adapters/search` | GET | Search adapters |
| `/v1/mcp/categories` | GET | List categories |
| `/v1/mcp/stats` | GET | Usage statistics |

### Query Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `q` | Search query | `?q=file` |
| `prompt` | Natural language prompt | `?prompt=list%20files` |
| `limit` | Result limit | `?limit=10` |
| `category` | Filter by category | `?category=filesystem` |

---

## Troubleshooting

### Empty Search Results
- Check if MCP tools are registered
- Verify HelixAgent is running
- Try different search terms

### Suggestions Not Working
- Check AI provider configuration
- Verify prompt encoding
- Try simpler prompts

### Adapter Not Found
- Adapter may not be configured
- Check adapter availability
- Review MCP server configuration

---

## Challenge Exercise (Optional)

Build an interactive tool discovery CLI that:
1. Takes user input as natural language
2. Suggests appropriate tools
3. Displays tool parameters
4. Offers to execute the selected tool

---

## Next Lab

Proceed to **Lab 8: Multi-Pass Validation** to learn advanced AI debate configuration.

---

*Lab Version: 1.0.0*
*Last Updated: January 2026*
