# Snow CLI User Guide

## Table of Contents
1. [Installation](#installation)
2. [Quick Start](#quick-start)
3. [CLI Commands](#cli-commands)
4. [TUI/Interactive Commands](#tuiinteractive-commands)
5. [Configuration](#configuration)
6. [API/Protocol Endpoints](#api-protocol-endpoints)
7. [Usage Examples](#usage-examples)
8. [Troubleshooting](#troubleshooting)

---

## Installation

### Method 1: Package Manager (PIP)

```bash
pip install snowflake-cli
```

Or upgrade:
```bash
pip install --upgrade snowflake-cli
```

### Method 2: Package Manager (PIPX)

```bash
pipx install snowflake-cli
```

### Method 3: Homebrew

```bash
brew tap snowflakedb/snowflake-cli
brew update
brew install snowflake-cli
```

### Method 4: Linux Package Managers

**Debian/Ubuntu:**
```bash
# Download .deb from releases page
sudo dpkg -i snowflake-cli-<version>.deb
```

**RHEL/CentOS/Fedora:**
```bash
# Download .rpm from releases page
sudo rpm -i snowflake-cli-<version>.rpm
```

### Method 5: WinGet (Windows)

```powershell
winget install Snowflake.CLI
```

### Method 6: Install Script

**macOS/Linux:**
```bash
curl -fsSL https://sfc-repo.snowflakecomputing.com/snowflake-cli/install.sh | bash
```

---

## Quick Start

```bash
# Verify installation
snow --version
snow --help

# Add a connection
snow connection add --connection-name myaccount \
  --account myaccount \
  --user myuser \
  --password

# Test connection
snow connection test

# Set default connection
snow connection set-default myaccount

# Execute SQL
snow sql -q "SELECT CURRENT_VERSION()"
```

---

## CLI Commands

### Global Options

| Option | Description | Example |
|--------|-------------|---------|
| `--version` | Show version | `snow --version` |
| `--info` | Show CLI information | `snow --info` |
| `--config-file` | Specify config file | `snow --config-file /path/to/config.toml` |
| `--install-completion` | Install shell completion | `snow --install-completion` |
| `--show-completion` | Show completion script | `snow --show-completion` |
| `--help` | Show help | `snow --help` |

### Command: snow connection

**Description:** Manage connections to Snowflake.

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `add` | Add a new connection |
| `list` | List all connections |
| `test` | Test a connection |
| `set-default` | Set default connection |

**Usage:**
```bash
snow connection add [options]
```

**Options:**
| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `--connection-name` | string | Yes | Connection name |
| `--account` | string | Yes | Account identifier |
| `--user` | string | Yes | Username |
| `--password` | string | No | Password (prompted if not provided) |
| `--database` | string | No | Default database |
| `--schema` | string | No | Default schema |
| `--warehouse` | string | No | Default warehouse |
| `--role` | string | No | Default role |

**Examples:**
```bash
# Add connection interactively
snow connection add

# Add connection with options
snow connection add \
  --connection-name prod \
  --account xy12345 \
  --user admin \
  --password \
  --database PROD_DB \
  --warehouse COMPUTE_WH

# List connections
snow connection list

# Test connection
snow connection test -c prod

# Set default
snow connection set-default prod
```

### Command: snow sql

**Description:** Execute Snowflake SQL queries.

**Usage:**
```bash
snow sql [options]
```

**Options:**
| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `-q, --query` | string | Yes | SQL query to execute |
| `-f, --file` | string | Yes | SQL file to execute |
| `-c, --connection` | string | No | Connection name |

**Examples:**
```bash
# Execute query
snow sql -q "SELECT CURRENT_VERSION()"

# Execute from file
snow sql -f queries/setup.sql

# Multi-line query
snow sql -q "CREATE DATABASE IF NOT EXISTS mydb; USE mydb;"

# With specific connection
snow sql -c prod -q "SELECT * FROM large_table LIMIT 10"
```

### Command: snow object

**Description:** Manage Snowflake objects.

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `list` | List objects |
| `describe` | Describe object |

**Examples:**
```bash
# List warehouses
snow object list warehouse

# List databases
snow object list database

# Describe table
snow object describe table mydb.public.users
```

### Command: snow stage

**Description:** Manage stages.

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `list` | List stages |
| `create` | Create stage |
| `drop` | Drop stage |
| `copy` | Copy files to/from stage |
| `list-files` | List files in stage |

**Examples:**
```bash
# List stages
snow stage list

# Create stage
snow stage create my_stage

# Copy file to stage
snow stage copy my_stage file:///path/to/file.csv

# List files in stage
snow stage list-files my_stage
```

### Command: snow snowpark

**Description:** Manage procedures and functions.

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `build` | Build Snowpark project |
| `deploy` | Deploy Snowpark project |
| `run` | Run Snowpark procedure |

**Examples:**
```bash
# Build project
snow snowpark build

# Deploy
snow snowpark deploy

# Run procedure
snow snowpark run my_procedure
```

### Command: snow streamlit

**Description:** Manage Streamlit apps.

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `init` | Initialize Streamlit project |
| `deploy` | Deploy Streamlit app |
| `list` | List Streamlit apps |
| `describe` | Describe Streamlit app |
| `drop` | Drop Streamlit app |

**Examples:**
```bash
# Initialize project
snow streamlit init my_app

# Deploy
snow streamlit deploy my_app

# List apps
snow streamlit list

# Open app
snow streamlit open my_app
```

### Command: snow app

**Description:** Manage Snowflake Native Apps.

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `init` | Initialize app project |
| `run` | Run app locally |
| `deploy` | Deploy app |
| `teardown` | Remove app |
| `version` | Manage app versions |

**Examples:**
```bash
# Initialize app
snow app init my_native_app

# Run locally
snow app run

# Deploy to account
snow app deploy

# Create version
snow app version create v1.0.0
```

### Command: snow spcs

**Description:** Manage Snowpark Container Services.

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `compute-pool` | Manage compute pools |
| `service` | Manage services |
| `image-registry` | Manage image registry |
| `image-repository` | Manage image repositories |
| `job` | Manage jobs |

**Examples:**
```bash
# List compute pools
snow spcs compute-pool list

# Create service
snow spcs service create my_service --compute-pool my_pool

# List services
snow spcs service list

# View logs
snow spcs service logs my_service
```

### Command: snow cortex

**Description:** Access Snowflake Cortex AI services.

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `complete` | Text completion |
| `extract_answer` | Extract answer from text |
| `sentiment` | Analyze sentiment |
| `summarize` | Summarize text |
| `translate` | Translate text |

**Examples:**
```bash
# Complete text
snow cortex complete "What is Snowflake?"

# Analyze sentiment
snow cortex sentiment "I love this product!"

# Summarize
snow cortex summarize --file article.txt

# Translate
snow cortex translate "Hello" --from en --to fr
```

### Command: snow notebook

**Description:** Manage notebooks.

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `list` | List notebooks |
| `create` | Create notebook |
| `drop` | Drop notebook |
| `execute` | Execute notebook |

**Examples:**
```bash
# List notebooks
snow notebook list

# Create notebook
snow notebook create my_notebook

# Execute notebook
snow notebook execute my_notebook
```

### Command: snow git

**Description:** Manage git repositories in Snowflake.

**Subcommands:**

| Subcommand | Description |
|------------|-------------|
| `list` | List git repositories |
| `create` | Create repository |
| `drop` | Drop repository |
| `execute` | Execute files from repository |

**Examples:**
```bash
# List repositories
snow git list

# Create repository
snow git create my_repo --url https://github.com/org/repo.git

# Execute SQL from repo
snow git execute my_repo /path/to/script.sql
```

---

## TUI/Interactive Commands

Snow CLI primarily operates in command-line mode. Some commands support interactive prompts when required parameters are not provided.

---

## Configuration

### Configuration File Format

Snow CLI uses TOML configuration:

**File Location:** `~/.snowflake/config.toml`

```toml
[connections]
[connections.default]
account = "xy12345"
user = "admin"
password = "..."
database = "PROD"
schema = "PUBLIC"
warehouse = "COMPUTE_WH"
role = "ACCOUNTADMIN"

[connections.dev]
account = "xy12345"
user = "developer"
password = "..."
database = "DEV"
schema = "TEST"
warehouse = "DEV_WH"

[cli]
output_format = "table"
verbose = false
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `SNOWFLAKE_HOME` | Config directory path |
| `SNOWFLAKE_DEFAULT_CONNECTION_NAME` | Default connection |
| `SNOWFLAKE_ACCOUNT` | Account identifier |
| `SNOWFLAKE_USER` | Username |
| `SNOWFLAKE_PASSWORD` | Password |
| `SNOWFLAKE_DATABASE` | Default database |
| `SNOWFLAKE_SCHEMA` | Default schema |
| `SNOWFLAKE_WAREHOUSE` | Default warehouse |
| `SNOWFLAKE_ROLE` | Default role |

### Shell Completion

```bash
# Install completion
snow --install-completion

# Show completion for manual installation
snow --show-completion

# Add to shell profile (bash)
eval "$(snow --show-completion)"
```

---

## API/Protocol Endpoints

### Snowflake REST API

Snow CLI uses Snowflake's REST API for communication.

### Supported Cortex Models

| Category | Models |
|----------|--------|
| Large | reka-core, llama3-70b, mistral-large |
| Medium | snowflake-arctic, reka-flash, mixtral-8x7b, llama2-70b-chat |
| Small | llama3-8b, mistral-7b, gemma-7b |

---

## Usage Examples

### Example 1: Basic Setup

```bash
# Install Snow CLI
pip install snowflake-cli

# Add connection
snow connection add --connection-name prod \
  --account xy12345 \
  --user admin \
  --password

# Test connection
snow connection test

# Execute query
snow sql -q "SELECT CURRENT_VERSION()"
```

### Example 2: Database Management

```bash
# Create database
snow sql -q "CREATE DATABASE IF NOT EXISTS analytics;"

# Create schema
snow sql -q "CREATE SCHEMA IF NOT EXISTS analytics.raw;"

# Create table
snow sql -q "CREATE TABLE analytics.raw.events (
  id VARCHAR,
  event_time TIMESTAMP,
  user_id VARCHAR,
  event_type VARCHAR
);"

# Load data
snow stage copy @analytics.raw.my_stage file:///data/events.csv
snow sql -q "COPY INTO analytics.raw.events FROM @analytics.raw.my_stage;"
```

### Example 3: Streamlit App Deployment

```bash
# Initialize app
snow streamlit init sales_dashboard

# Edit app.py with your code
# ...

# Deploy
snow streamlit deploy sales_dashboard

# Open in browser
snow streamlit open sales_dashboard
```

### Example 4: Native App Development

```bash
# Initialize native app
snow app init my_data_app

# Develop app files
# ...

# Run locally
snow app run

# Test in account
snow app deploy

# Create version
snow app version create v1.0.0
snow app version create-patch v1.0.1
```

### Example 5: SPCS (Container Services)

```bash
# Create compute pool
snow sql -q "CREATE COMPUTE POOL my_pool ..."

# Build and push image
snow spcs image-registry login
docker build -t my_image .
docker tag my_image registry.snowflake.com/.../my_image
docker push registry.snowflake.com/.../my_image

# Create service
snow spcs service create my_service \
  --compute-pool my_pool \
  --image my_image

# Monitor service
snow spcs service list
snow spcs service status my_service
snow spcs service logs my_service
```

### Example 6: Cortex AI

```bash
# Text completion
snow cortex complete "Explain data warehousing"

# Code generation
snow cortex complete "Write a SQL query to find top customers"

# Sentiment analysis
snow cortex sentiment "This product is amazing!"

# Translation
snow cortex translate "Hello world" --from en --to es

# Summarization
snow cortex summarize --file long_document.txt
```

### Example 7: Git Integration

```bash
# Setup git repository
snow git create my_repo \
  --url https://github.com/org/snowflake-scripts.git \
  --secret github_token

# List files
snow git list-files my_repo

# Execute SQL script
snow git execute my_repo /migrations/v1_setup.sql

# Schedule with tasks
snow sql -f <(echo "
CREATE TASK git_migration
  SCHEDULE = 'USING CRON 0 0 * * * UTC'
  AS CALL SYSTEM$GIT_EXECUTE('my_repo', '/migrations/*.sql');
")
```

---

## Troubleshooting

### Issue: Connection refused

**Solution:**
```bash
# Verify account identifier format
# Should be: xy12345.region.cloud (e.g., xy12345.us-east-1.aws)

# Test connection
snow connection test -c myconnection

# Check network connectivity
ping <account>.snowflakecomputing.com
```

### Issue: Authentication failed

**Solution:**
```bash
# Update credentials
snow connection add --connection-name myconnection --password

# Or edit config file directly
# ~/.snowflake/config.toml
```

### Issue: Command not found

**Solution:**
```bash
# Check installation
which snow
pip show snowflake-cli

# Ensure bin directory in PATH
export PATH="$HOME/.local/bin:$PATH"
```

### Issue: SQL execution errors

**Solution:**
```bash
# Enable verbose output
snow --info

# Check query syntax
# Verify permissions
# Check object exists
```

### Issue: SPCS image push fails

**Solution:**
```bash
# Login to registry
snow spcs image-registry login

# Verify image tag format
# Should match: registry.snowflake.com/db/schema/repo/image:tag
```

---

## Additional Resources

- **Snowflake Documentation:** https://docs.snowflake.com
- **Snowflake CLI Docs:** https://docs.snowflake.com/en/developer-guide/snowflake-cli/index
- **Snowpark Docs:** https://docs.snowflake.com/en/developer-guide/snowpark/index
- **Cortex Docs:** https://docs.snowflake.com/en/user-guide/snowflake-cortex
