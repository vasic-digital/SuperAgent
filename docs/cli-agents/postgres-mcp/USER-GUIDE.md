# PostgreSQL MCP User Guide

## Overview

PostgreSQL MCP (Model Context Protocol) Server is an AI agent toolkit that enables natural language interaction with PostgreSQL databases. It provides an MCP server implementation that allows AI assistants to query PostgreSQL databases using natural language, with read-only protection and secure execution.

**Key Features:**
- Natural Language to SQL translation
- Read-only transaction protection
- Multiple LLM support (Anthropic Claude, OpenAI, Ollama)
- Web Interface for browser-based interaction
- CLI Client with prompt caching support
- HTTP/HTTPS mode with authentication
- Docker support for easy deployment
- Hybrid search with BM25+MMR
- Embedding generation for RAG workflows

**Supported Versions:** PostgreSQL 14 and higher

---

## Installation Methods

### Method 1: Docker Quick Start (Recommended)

The fastest way to get started:

```bash
# Clone the repository
git clone https://github.com/pgEdge/pgedge-postgres-mcp.git
cd pgedge-postgres-mcp

# Start with Docker Compose
docker compose up -d
```

### Method 2: pip Install (Python)

```bash
# Install from PyPI (when available)
pip install postgres-mcp

# Or install from source
git clone https://github.com/pgEdge/pgedge-postgres-mcp.git
cd pgedge-postgres-mcp
pip install -e .
```

### Method 3: Using uv (Modern Python)

```bash
# Install uv first
curl -sSL https://astral.sh/uv/install.sh | sh

# Clone and install
git clone https://github.com/pgEdge/pgedge-postgres-mcp.git
cd pgedge-postgres-mcp
uv pip install -e .
```

### Method 4: npm Install (Alternative Implementations)

For the Node.js implementation:

```bash
npx @ahmedmustahid/postgres-mcp-server
```

### Method 5: Docker Hub Image

```bash
docker run -d --name postgres-mcp \
  -e PGHOST=localhost \
  -e PGPORT=5432 \
  -e PGUSER=postgres \
  -e PGPASSWORD=password \
  -e PGDATABASE=mydb \
  -p 8080:8080 \
  pgedge/pgedge-postgres-mcp:latest
```

---

## Quick Start

### 1. Configure Database Connection

Set environment variables:

```bash
export PGHOST=localhost
export PGPORT=5432
export PGUSER=postgres
export PGPASSWORD=your_password
export PGDATABASE=mydb
```

Or create a `.env` file:

```env
PGHOST=localhost
PGPORT=5432
PGUSER=postgres
PGPASSWORD=your_password
PGDATABASE=mydb
MCP_PORT=8080
```

### 2. Start the Server

**Using Docker:**
```bash
docker compose up -d
```

**Using Python:**
```bash
# Start the MCP server
postgres-mcp

# Or with explicit connection
postgres-mcp "postgresql://user:pass@localhost:5432/dbname"
```

**Using Go CLI Client:**
```bash
./bin/pgedge-nl-cli
```

### 3. Test the Connection

```bash
# Health check
curl http://localhost:8080/health

# List available tools
curl -X POST http://localhost:8080/mcp \
  -H "Content-Type: application/json" \
  -d '{"type":"function","name":"tools/list"}'
```

### 4. Connect Your AI Client

**For Claude Desktop:**
Add to `~/.config/claude-desktop/config.json`:

```json
{
  "mcpServers": {
    "postgres": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "PGHOST=host.docker.internal",
        "-e", "PGPORT=5432",
        "-e", "PGUSER=postgres",
        "-e", "PGPASSWORD=password",
        "-e", "PGDATABASE=mydb",
        "pgedge/pgedge-postgres-mcp:latest"
      ]
    }
  }
}
```

**For Claude Code:**
Create `.mcp.json` in your project root:

```json
{
  "mcpServers": {
    "pgedge-postgres": {
      "command": "/path/to/pgedge-mcp-server",
      "env": {
        "PGHOST": "localhost",
        "PGPORT": "5432",
        "PGDATABASE": "myapp_db",
        "PGUSER": "app_user",
        "PGPASSWORD": "app_password"
      }
    }
  }
}
```

**For Cursor:**
Add to Cursor MCP settings:

```json
{
  "mcpServers": {
    "postgres": {
      "command": "npx",
      "args": ["-y", "@pgedge/postgres-mcp@latest"],
      "env": {
        "DATABASE_URL": "postgresql://user:pass@localhost:5432/db"
      }
    }
  }
}
```

---

## CLI Commands Reference

### Server Commands

| Command | Description |
|---------|-------------|
| `postgres-mcp` | Start the MCP server (stdio mode) |
| `postgres-mcp --http` | Start HTTP server mode |
| `postgres-mcp --port 8080` | Start on custom port |
| `postgres-mcp --config config.yaml` | Use configuration file |

### Connection Options

| Flag | Environment Variable | Description |
|------|---------------------|-------------|
| `--host` | `PGHOST` | PostgreSQL host |
| `--port` | `PGPORT` | PostgreSQL port |
| `--user` | `PGUSER` | PostgreSQL username |
| `--password` | `PGPASSWORD` | PostgreSQL password |
| `--database` | `PGDATABASE` | PostgreSQL database name |
| `--ssl-mode` | `PGSSLMODE` | SSL mode (disable, require, verify-ca, verify-full) |

### Go CLI Client Commands

```bash
# Start interactive CLI
./bin/pgedge-nl-cli

# Query with natural language
./bin/pgedge-nl-cli "Show me all users who signed up last month"

# Execute SQL directly
./bin/pgedge-nl-cli --sql "SELECT * FROM users LIMIT 10"

# Show schema information
./bin/pgedge-nl-cli --schema

# Enable verbose output
./bin/pgedge-nl-cli --verbose
```

### MCP Inspector (Debugging)

```bash
# Install MCP Inspector
npx @modelcontextprotocol/inspector

# Inspect available tools
npx @modelcontextprotocol/inspector \
  --cli http://localhost:8080/mcp \
  --transport http \
  --method tools/list
```

---

## TUI / Interactive Commands

### Web Interface

Access the web UI at `http://localhost:8080` after starting the server.

**Features:**
- Natural language query input
- SQL query visualization
- Result table display
- Schema browser
- Query history

### CLI Interactive Mode

```bash
./bin/pgedge-nl-cli
```

**Interactive Commands:**

```
> show tables                    # List all tables
> describe users                 # Show table schema
> query "top 10 customers"       # Natural language query
> sql SELECT * FROM orders       # Direct SQL execution
> explain last                   # Explain last query plan
> history                        # Show query history
> help                           # Show help
> quit                           # Exit CLI
```

---

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PGHOST` | localhost | PostgreSQL host |
| `PGPORT` | 5432 | PostgreSQL port |
| `PGUSER` | postgres | PostgreSQL username |
| `PGPASSWORD` | - | PostgreSQL password |
| `PGDATABASE` | postgres | PostgreSQL database |
| `MCP_PORT` | 8080 | MCP server port |
| `MCP_HOST` | 0.0.0.0 | MCP server bind address |
| `MCP_LOG_LEVEL` | info | Logging level |
| `OPENAI_API_KEY` | - | OpenAI API key |
| `ANTHROPIC_API_KEY` | - | Anthropic API key |
| `OLLAMA_HOST` | http://localhost:11434 | Ollama host |

### YAML Configuration File

Create `config.yaml`:

```yaml
# Database connection
database:
  host: localhost
  port: 5432
  user: postgres
  password: ${PGPASSWORD}
  database: myapp
  ssl_mode: require

# MCP Server settings
server:
  port: 8080
  host: 0.0.0.0
  cors_origins:
    - http://localhost:3000
    - http://localhost:8080

# LLM configuration
llm:
  provider: anthropic  # or openai, ollama
  model: claude-3-sonnet-20240229
  api_key: ${ANTHROPIC_API_KEY}
  temperature: 0.1
  max_tokens: 4096

# Security settings
security:
  read_only: true
  max_rows: 10000
  query_timeout: 30s
  allowed_schemas:
    - public
    - analytics

# Authentication (optional)
auth:
  enabled: true
  jwt_secret: ${JWT_SECRET}
  token_expiry: 24h

# Features
features:
  embeddings: true
  vector_search: true
  query_caching: true
  explain_queries: true
```

### Multiple Database Configuration

```yaml
# config.multi.yaml
databases:
  - name: production
    host: prod.db.company.com
    port: 5432
    user: readonly
    password: ${PROD_PASSWORD}
    database: app_production
    
  - name: analytics
    host: analytics.db.company.com
    port: 5432
    user: analyst
    password: ${ANALYTICS_PASSWORD}
    database: warehouse

default_database: production
```

### Docker Compose Configuration

```yaml
version: '3.8'

services:
  postgres-mcp:
    image: pgedge/pgedge-postgres-mcp:latest
    ports:
      - "8080:8080"
    environment:
      - PGHOST=postgres
      - PGPORT=5432
      - PGUSER=postgres
      - PGPASSWORD=postgres
      - PGDATABASE=mydb
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - MCP_PORT=8080
    volumes:
      - ./config.yaml:/app/config.yaml
    depends_on:
      - postgres
    restart: unless-stopped

  postgres:
    image: postgres:16
    environment:
      - POSTGRES_USER=postgres
      - POSTGRES_PASSWORD=postgres
      - POSTGRES_DB=mydb
    volumes:
      - pgdata:/var/lib/postgresql/data
    ports:
      - "5432:5432"

volumes:
  pgdata:
```

---

## Usage Examples

### Example 1: Basic Natural Language Query

```bash
# Using CLI
./bin/pgedge-nl-cli "Find all orders placed in the last 30 days"

# Response:
# Generated SQL: SELECT * FROM orders WHERE created_at >= NOW() - INTERVAL '30 days';
# [Results displayed in table format]
```

### Example 2: Complex Analytics Query

```bash
./bin/pgedge-nl-cli ""
```

### Example 3: Schema Exploration

```bash
# In interactive mode
> describe users

# Output:
# Table: users
# Columns:
#   - id (integer, PK)
#   - email (varchar, unique)
#   - created_at (timestamp)
#   - last_login (timestamp, nullable)
# Indexes:
#   - users_pkey (id)
#   - users_email_idx (email)
```

### Example 4: Using with Claude Code

```bash
# Start Claude Code in project directory
claude-code

# In Claude:
> Find all customers who made purchases over $1000 in the last month
# Claude uses MCP to:
# 1. Query schema
# 2. Generate SQL
# 3. Execute query
# 4. Present results
```

### Example 5: Vector Search (RAG)

```sql
-- Enable vector extension first
CREATE EXTENSION IF NOT EXISTS vector;

-- Create table with embeddings
CREATE TABLE documents (
    id SERIAL PRIMARY KEY,
    content TEXT,
    embedding VECTOR(1536)
);
```

```bash
# Query using natural language
./bin/pgedge-nl-cli "Find documents similar to 'machine learning'"
```

### Example 6: Integration with Application

```python
# Python example using MCP client
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

server_params = StdioServerParameters(
    command="postgres-mcp",
    args=[],
    env={"PGHOST": "localhost", "PGDATABASE": "mydb"}
)

async with stdio_client(server_params) as (read, write):
    async with ClientSession(read, write) as session:
        await session.initialize()
        
        # List available tools
        tools = await session.list_tools()
        
        # Execute natural language query
        result = await session.call_tool(
            "natural_language_query",
            {"query": "Show me top 10 customers by revenue"}
        )
        print(result)
```

### Example 7: Multi-Database Setup

```bash
# Start with multiple databases
export MCP_CONFIG=config.multi.yaml
postgres-mcp --http

# Query specific database
./bin/pgedge-nl-cli --database analytics "Show monthly revenue trends"
```

---

## MCP Tools Reference

### Available Tools

| Tool | Description |
|------|-------------|
| `query` | Execute SQL query (read-only) |
| `natural_language_query` | Convert natural language to SQL and execute |
| `get_schema` | Get database schema information |
| `list_tables` | List all tables in database |
| `describe_table` | Get detailed table information |
| `search_similar` | Vector similarity search |
| `explain_query` | Get query execution plan |
| `get_statistics` | Get table/index statistics |

### MCP Resources

| Resource | Description |
|----------|-------------|
| `schema://{table}` | Table schema details |
| `statistics://{table}` | Table statistics |
| `indexes://{table}` | Index information |

### MCP Prompts

| Prompt | Description |
|--------|-------------|
| `setup_vector_search` | Guide for setting up vector search |
| `database_exploration` | Interactive database exploration |
| `query_optimization` | Query optimization suggestions |

---

## Troubleshooting

### Connection Issues

**Problem:** Cannot connect to PostgreSQL

**Solutions:**
```bash
# Test PostgreSQL connection
psql -h localhost -U postgres -d mydb -c "SELECT 1"

# Check if PostgreSQL is running
pg_isready -h localhost -p 5432

# Verify environment variables
echo $PGHOST $PGPORT $PGUSER

# Check firewall settings
# Ensure port 5432 is accessible

# For Docker: check network
Docker network ls
Docker inspect <network_name>
```

### MCP Server Not Responding

**Problem:** Server starts but doesn't respond to queries

**Solutions:**
```bash
# Check server logs
docker logs postgres-mcp

# Verify port is not in use
lsof -i :8080

# Restart server
docker compose restart postgres-mcp

# Check configuration syntax
postgres-mcp --config config.yaml --validate
```

### Authentication Errors

**Problem:** API key errors for LLM providers

**Solutions:**
```bash
# Verify API keys
echo $ANTHROPIC_API_KEY
echo $OPENAI_API_KEY

# Test API key
curl -H "x-api-key: $ANTHROPIC_API_KEY" \
  https://api.anthropic.com/v1/models

# Check key permissions
# Ensure key has access to required models
```

### Natural Language Query Failures

**Problem:** NL queries not generating correct SQL

**Solutions:**
- Check schema is loaded: `./bin/pgedge-nl-cli --schema`
- Verify table permissions: `\dp` in psql
- Try more specific queries
- Check LLM provider status

### Performance Issues

**Problem:** Slow query responses

**Solutions:**
```bash
# Enable query caching
export MCP_CACHE_ENABLED=true

# Increase timeout
export MCP_QUERY_TIMEOUT=60s

# Check PostgreSQL performance
EXPLAIN ANALYZE SELECT ...

# Add indexes for frequently queried columns
```

### Docker-Specific Issues

**Problem:** Container cannot reach PostgreSQL

**Solutions:**
```bash
# For local PostgreSQL, use host.docker.internal
docker run -e PGHOST=host.docker.internal ...

# Or use docker-compose with depends_on
# Ensure postgres service starts first

# Check Docker network
docker network inspect bridge

# Use custom network
docker network create mcp-network
docker run --network mcp-network ...
```

### Read-Only Protection

**Problem:** Cannot execute write operations

**Explanation:** By design, PostgreSQL MCP runs all queries in read-only transactions for safety.

**For write operations:**
- Use direct psql connection
- Or use pgEdge RAG Server for write-enabled operations

### Common Error Messages

| Error | Solution |
|-------|----------|
| "Connection refused" | Check PostgreSQL is running and accessible |
| "Authentication failed" | Verify credentials |
| "Database does not exist" | Check PGDATABASE value |
| "SSL required" | Set PGSSLMODE=require |
| "Query timeout" | Increase timeout or optimize query |
| "LLM API error" | Check API key and rate limits |

### Getting Help

```bash
# Check version
postgres-mcp --version

# Verbose logging
postgres-mcp --log-level debug

# Documentation
# https://docs.pgedge.com/pgedge-postgres-mcp-server

# GitHub Issues
# https://github.com/pgEdge/pgedge-postgres-mcp/issues
```

---

## Security Best Practices

1. **Use Read-Only Users:**
   ```sql
   CREATE USER mcp_readonly WITH PASSWORD 'secure_password';
   GRANT CONNECT ON DATABASE mydb TO mcp_readonly;
   GRANT USAGE ON SCHEMA public TO mcp_readonly;
   GRANT SELECT ON ALL TABLES IN SCHEMA public TO mcp_readonly;
   ```

2. **Enable SSL:**
   ```bash
   export PGSSLMODE=require
   ```

3. **Use Environment Variables:** Never hardcode credentials

4. **Limit Row Counts:**
   ```yaml
   security:
     max_rows: 1000
   ```

5. **Restrict Schemas:**
   ```yaml
   security:
     allowed_schemas:
       - public
   ```

---

## Advanced Configuration

### Custom Embeddings Provider

```yaml
embeddings:
  provider: openai  # or ollama, cohere
  model: text-embedding-3-small
  dimensions: 1536
  batch_size: 100
```

### Query Optimization

```yaml
optimization:
  enable_query_caching: true
  cache_ttl: 3600
  auto_explain: true
  suggest_indexes: true
```

### Logging Configuration

```yaml
logging:
  level: info
  format: json
  output: /var/log/postgres-mcp.log
  query_logging: true
  slow_query_threshold: 1s
```

---

## Resources

- **GitHub:** https://github.com/pgEdge/pgedge-postgres-mcp
- **Documentation:** https://docs.pgedge.com/pgedge-postgres-mcp-server
- **Docker Hub:** https://hub.docker.com/r/pgedge/pgedge-postgres-mcp
- **Website:** https://www.pgedge.com

---

*Last Updated: April 2026*
