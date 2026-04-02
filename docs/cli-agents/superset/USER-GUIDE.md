# Apache Superset User Guide

## Overview

Apache Superset is a modern, enterprise-ready business intelligence web application that provides intuitive data visualization and exploration capabilities. It enables users to create interactive dashboards, execute SQL queries, and visualize data from various sources including PostgreSQL, MySQL, BigQuery, Snowflake, and many more.

**Key Features:**
- 50+ visualization types (charts, tables, maps)
- SQL Lab for ad-hoc queries
- Drag-and-drop dashboard builder
- Role-based access control
- Caching layer for performance
- Asynchronous query execution
- REST API for programmatic access
- Plugin architecture for custom visualizations
- Multi-tenant support
- Embedded analytics capabilities

---

## Installation Methods

### Method 1: Docker Compose (Recommended)

The easiest way to run Superset locally:

```bash
# Clone the repository
git clone https://github.com/apache/superset.git
cd superset

# Start with Docker Compose
docker compose -f docker-compose-non-dev.yml up

# Or use the latest image tag version
docker compose -f docker-compose-image-tag.yml up
```

### Method 2: pip Install (Python)

Requirements: Python 3.9-3.11

```bash
# Create virtual environment
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# Install Superset
pip install apache-superset

# Or install specific version
pip install apache-superset==4.1.1
```

### Method 3: Production Docker Setup

```bash
# Custom Docker Compose with PostgreSQL and Redis
mkdir superset-prod && cd superset-prod

# Create docker-compose.yml
cat > docker-compose.yml << 'EOF'
version: "3.8"

x-superset-common: &superset-common
  image: apache/superset:latest
  environment: &superset-env
    SUPERSET_SECRET_KEY: your-secret-key-change-this-in-production
    DATABASE_HOST: postgres
    DATABASE_PORT: 5432
    DATABASE_USER: superset
    DATABASE_PASSWORD: superset
    DATABASE_DB: superset
    REDIS_HOST: redis
    REDIS_PORT: 6379
    SQLALCHEMY_DATABASE_URI: postgresql+psycopg2://superset:superset@postgres:5432/superset
    CELERY_BROKER_URL: redis://redis:6379/0
    CELERY_RESULT_BACKEND: redis://redis:6379/0
  depends_on:
    postgres:
      condition: service_healthy
    redis:
      condition: service_healthy

services:
  postgres:
    image: postgres:16
    environment:
      POSTGRES_USER: superset
      POSTGRES_PASSWORD: superset
      POSTGRES_DB: superset
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U superset"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7
    volumes:
      - redisdata:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  superset-init:
    <<: *superset-common
    command: >
      bash -c "
        superset db upgrade &&
        superset fab create-admin --username admin --firstname Admin --lastname User --email admin@example.com --password admin &&
        superset init &&
        echo 'Superset initialized successfully'
      "
    restart: "no"

  superset:
    <<: *superset-common
    ports:
      - "8088:8088"
    command: >
      bash -c "
        superset db upgrade &&
        gunicorn --bind 0.0.0.0:8088 --workers 4 --timeout 120 'superset.app:create_app()'
      "
    restart: unless-stopped

  superset-worker:
    <<: *superset-common
    command: celery --app=superset.tasks.celery_app:app worker --pool=prefork --concurrency=4
    restart: unless-stopped

  superset-beat:
    <<: *superset-common
    command: celery --app=superset.tasks.celery_app:app beat --schedule=/tmp/celerybeat-schedule
    restart: unless-stopped

volumes:
  pgdata:
  redisdata:
EOF

# Start services
docker compose up -d

# Initialize (only first time)
docker compose run --rm superset-init
```

### Method 4: Kubernetes (Helm)

```bash
# Add Helm repository
helm repo add superset https://apache.github.io/superset
helm repo update

# Install with default values
helm upgrade --install superset superset/superset

# Or with custom values
helm upgrade --install superset superset/superset \
  --values custom-values.yaml
```

### Method 5: Ubuntu/Debian Installation

```bash
# Install system dependencies
sudo apt-get update
sudo apt-get install -y build-essential libssl-dev libffi-dev python3-dev python3-pip libsasl2-dev libldap2-dev python3.10-venv

# Create virtual environment
python3 -m venv venv
source venv/bin/activate

# Upgrade pip
pip install --upgrade pip

# Install Superset
pip install apache-superset

# Install database drivers
pip install psycopg2-binary pymysql pyhive trino sqlalchemy-bigquery
```

---

## Quick Start

### 1. Initialize Database

```bash
# Set required environment variable
export FLASK_APP=superset
export SUPERSET_SECRET_KEY=$(openssl rand -base64 42)

# Initialize database
superset db upgrade
```

### 2. Create Admin User

```bash
# Create admin user
superset fab create-admin

# Or non-interactive
superset fab create-admin \
  --username admin \
  --firstname Admin \
  --lastname User \
  --email admin@example.com \
  --password admin
```

### 3. Initialize Superset

```bash
# Create default roles and permissions
superset init

# Load example data (optional)
superset load-examples
```

### 4. Start the Server

```bash
# Development mode
superset run -p 8088 --with-threads --reload --debugger

# Production mode with Gunicorn
gunicorn --bind 0.0.0.0:8088 \
  --workers 4 \
  --timeout 120 \
  --limit-request-line 0 \
  --limit-request-field_size 0 \
  'superset.app:create_app()'
```

### 5. Access the UI

Open browser: http://localhost:8088

Default credentials: admin / admin

---

## CLI Commands Reference

### Database Commands

| Command | Description |
|---------|-------------|
| `superset db upgrade` | Upgrade database schema |
| `superset db downgrade` | Downgrade database schema |
| `superset db current` | Show current revision |
| `superset db history` | Show migration history |
| `superset db migrate` | Create new migration |

### User Management

| Command | Description |
|---------|-------------|
| `superset fab create-admin` | Create admin user |
| `superset fab create-user` | Create regular user |
| `superset fab reset-password` | Reset user password |
| `superset fab list-users` | List all users |

### Initialization

| Command | Description |
|---------|-------------|
| `superset init` | Initialize Superset (roles, permissions) |
| `superset load-examples` | Load example dashboards/charts |
| `superset import-dashboards` | Import dashboards from JSON |
| `superset export-dashboards` | Export dashboards to JSON |

### Security

| Command | Description |
|---------|-------------|
| `superset set-database-uri` | Update database URI |
| `superset sync-permissions` | Sync permissions |
| `superset update-datasources-cache` | Update datasource cache |

### Cache Management

| Command | Description |
|---------|-------------|
| `superset cache-warmup` | Warm up cache |
| `superset cache-datasource-warmup` | Warm up datasource cache |
| `superset celery-beat` | Start Celery beat scheduler |
| `superset celery-worker` | Start Celery worker |

### Development

| Command | Description |
|---------|-------------|
| `superset run` | Run development server |
| `superset run -p 8088` | Run on specific port |
| `superset run --with-threads` | Enable threading |
| `superset run --reload` | Auto-reload on changes |
| `superset run --debugger` | Enable debugger |
| `superset shell` | Open Python shell |
| `superset test` | Run tests |

### Data Commands

| Command | Description |
|---------|-------------|
| `superset import-datasource` | Import datasource YAML |
| `superset export-datasource` | Export datasource YAML |
| `superset refresh-datasource` | Refresh datasource metadata |
| `superset compute-thumbnails` | Compute chart thumbnails |

### Translation

| Command | Description |
|---------|-------------|
| `superset babel-compile` | Compile translations |
| `superset babel-extract` | Extract strings for translation |
| `superset babel-update` | Update translation files |

### Complete Setup Script

```bash
#!/bin/bash
# setup_superset.sh

export FLASK_APP=superset
export SUPERSET_SECRET_KEY=$(openssl rand -base64 42)

echo "Upgrading database..."
superset db upgrade

echo "Creating admin user..."
superset fab create-admin \
  --username admin \
  --firstname Admin \
  --lastname User \
  --email admin@example.com \
  --password admin

echo "Initializing Superset..."
superset init

echo "Loading examples..."
superset load-examples

echo "Starting server..."
superset run -p 8088 --with-threads --reload --debugger
```

---

## Configuration

### Configuration File (superset_config.py)

Create `superset_config.py` in your Python path:

```python
# superset_config.py
import os

# Database
SQLALCHEMY_DATABASE_URI = os.getenv(
    'SQLALCHEMY_DATABASE_URI',
    'postgresql+psycopg2://superset:superset@localhost:5432/superset'
)

# Secret key (REQUIRED for production)
SECRET_KEY = os.getenv('SUPERSET_SECRET_KEY', 'your-secret-key')

# Caching
CACHE_CONFIG = {
    'CACHE_TYPE': 'RedisCache',
    'CACHE_DEFAULT_TIMEOUT': 300,
    'CACHE_KEY_PREFIX': 'superset_',
    'CACHE_REDIS_HOST': 'redis',
    'CACHE_REDIS_PORT': 6379,
    'CACHE_REDIS_DB': 0,
}

# Celery
class CeleryConfig:
    BROKER_URL = 'redis://redis:6379/0'
    CELERY_RESULT_BACKEND = 'redis://redis:6379/0'
    CELERY_IMPORTS = ('superset.sql_lab',)
    CELERY_ANNOTATIONS = {'tasks.add': {'rate_limit': '10/s'}}

CELERY_CONFIG = CeleryConfig

# Features
FEATURE_FLAGS = {
    'EMBEDDED_SUPERSET': True,
    'ENABLE_TEMPLATE_PROCESSING': True,
    'DASHBOARD_RBAC': True,
    'ALERT_REPORTS': True,
}

# Authentication
AUTH_TYPE = 1  # 1 = DB, 2 = LDAP, 3 = OAuth
AUTH_USER_REGISTRATION = True
AUTH_USER_REGISTRATION_ROLE = 'Public'

# Timeouts
SQLLAB_ASYNC_TIME_LIMIT_SEC = 60 * 60 * 6  # 6 hours
SUPERSET_WEBSERVER_TIMEOUT = 300

# Mapbox API (for geospatial charts)
MAPBOX_API_KEY = os.getenv('MAPBOX_API_KEY', '')

# SMTP (for alerts/reports)
SMTP_HOST = 'smtp.gmail.com'
SMTP_PORT = 587
SMTP_STARTTLS = True
SMTP_SSL = False
SMTP_USER = 'your-email@gmail.com'
SMTP_PASSWORD = 'your-password'
SMTP_MAIL_FROM = 'your-email@gmail.com'
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `SUPERSET_SECRET_KEY` | Secret key for sessions | Required |
| `SQLALCHEMY_DATABASE_URI` | Metadata database URI | SQLite |
| `REDIS_HOST` | Redis hostname | localhost |
| `REDIS_PORT` | Redis port | 6379 |
| `SUPERSET_ENV` | Environment (production/development) | development |
| `MAPBOX_API_KEY` | Mapbox API key | - |
| `SMTP_HOST` | SMTP server host | - |

### Database Connection Examples

```python
# PostgreSQL
SQLALCHEMY_DATABASE_URI = 'postgresql+psycopg2://user:pass@host:5432/dbname'

# MySQL
SQLALCHEMY_DATABASE_URI = 'mysql+pymysql://user:pass@host:3306/dbname'

# SQLite (development only)
SQLALCHEMY_DATABASE_URI = 'sqlite:////path/to/superset.db'

# BigQuery
SQLALCHEMY_DATABASE_URI = 'bigquery://project-id'

# Snowflake
SQLALCHEMY_DATABASE_URI = 'snowflake://user:pass@account/dbname/schema'

# Redshift
SQLALCHEMY_DATABASE_URI = 'redshift+psycopg2://user:pass@host:5439/dbname'
```

---

## Usage Examples

### Example 1: Complete Local Setup

```bash
# 1. Create project directory
mkdir ~/superset-local && cd ~/superset-local

# 2. Create virtual environment
python3 -m venv venv
source venv/bin/activate

# 3. Install Superset
pip install apache-superset

# 4. Install PostgreSQL driver
pip install psycopg2-binary

# 5. Set environment
export FLASK_APP=superset
export SUPERSET_SECRET_KEY=$(openssl rand -base64 42)

# 6. Initialize
superset db upgrade
superset fab create-admin --username admin --firstname Admin --lastname User --email admin@example.com --password admin
superset init
superset load-examples

# 7. Start server
superset run -p 8088
```

### Example 2: Connect to PostgreSQL Database

```bash
# 1. Install driver
pip install psycopg2-binary

# 2. Start Superset
superset run

# 3. In UI: Sources > Databases > + Database
# 4. Connection String:
#    postgresql+psycopg2://user:password@localhost:5432/mydb
# 5. Test Connection
# 6. Save
```

### Example 3: Create Dashboard via API

```python
import requests

# Authenticate
auth = requests.post('http://localhost:8088/api/v1/security/login', json={
    'username': 'admin',
    'password': 'admin',
    'provider': 'db'
})
access_token = auth.json()['access_token']

headers = {'Authorization': f'Bearer {access_token}'}

# Create chart
chart = requests.post('http://localhost:8088/api/v1/chart/', 
    headers=headers,
    json={
        'slice_name': 'Sales by Region',
        'viz_type': 'pie',
        'datasource_id': 1,
        'datasource_type': 'table',
        'params': {...}
    }
)

# Create dashboard
dashboard = requests.post('http://localhost:8088/api/v1/dashboard/',
    headers=headers,
    json={
        'dashboard_title': 'Sales Dashboard',
        'slug': 'sales-dashboard',
        'position_json': {...}
    }
)
```

### Example 4: Docker with Custom Drivers

```dockerfile
# Dockerfile
FROM apache/superset:latest

USER root

# Install additional database drivers
RUN pip install psycopg2-binary pymysql pyhive trino \
    sqlalchemy-bigquery snowflake-sqlalchemy

# Copy custom configuration
COPY superset_config.py /app/pythonpath/

USER superset
```

```yaml
# docker-compose.override.yml
version: '3.8'
services:
  superset:
    build: .
    environment:
      - SUPERSET_SECRET_KEY=your-secret-key
    volumes:
      - ./superset_config.py:/app/pythonpath/superset_config.py
```

### Example 5: Enable Alerts and Reports

```python
# superset_config.py
FEATURE_FLAGS = {
    'ALERT_REPORTS': True,
}

# Celery configuration for scheduled reports
class CeleryConfig:
    BROKER_URL = 'redis://redis:6379/0'
    CELERY_RESULT_BACKEND = 'redis://redis:6379/0'
    CELERYBEAT_SCHEDULE = {
        'reports-scheduler': {
            'task': 'reports.scheduler',
            'schedule': crontab(minute='*', hour='*'),
        },
        'reports-prune-log': {
            'task': 'reports.prune_log',
            'schedule': crontab(minute=0, hour=0),
        },
    }

CELERY_CONFIG = CeleryConfig
```

### Example 6: Embedded Dashboard

```python
# superset_config.py
FEATURE_FLAGS = {
    'EMBEDDED_SUPERSET': True,
}

# Guest token configuration
GUEST_TOKEN_JWT_SECRET = "your-secret"
GUEST_TOKEN_JWT_ALGO = "HS256"
```

```javascript
// Embed in your application
import { embedDashboard } from "@superset-ui/embedded-sdk";

embedDashboard({
  id: "dashboard-id",
  supersetDomain: "http://localhost:8088",
  mountPoint: document.getElementById("dashboard"),
  fetchGuestToken: () => fetchGuestTokenFromBackend(),
  dashboardUiConfig: {
    hideTitle: true,
    hideTab: true,
  },
});
```

---

## TUI / Web Interface

### SQL Lab

Interactive SQL editor with:
- Multi-tab query editor
- Schema browser
- Query history
- Results export
- Query scheduling

Access: http://localhost:8088/sqllab

### Chart Builder

Drag-and-drop interface for:
- Selecting datasource
- Choosing visualization type
- Configuring metrics/dimensions
- Customizing appearance

Access: Charts > + Chart

### Dashboard Builder

Visual dashboard editor:
- Drag charts onto canvas
- Resize and position
- Add tabs and filters
- Configure refresh intervals

Access: Dashboards > + Dashboard

### Dataset Editor

Manage datasets:
- Define metrics
- Create calculated columns
- Set column types
- Configure caching

Access: Datasets > Edit

---

## Troubleshooting

### Installation Issues

**Problem:** `pip install` fails with compilation errors

**Solutions:**
```bash
# Install system dependencies
# Ubuntu/Debian:
sudo apt-get install build-essential libssl-dev libffi-dev python3-dev

# macOS:
brew install openssl libffi

# Use pre-built wheels
pip install --only-binary :all: apache-superset
```

### Database Connection Issues

**Problem:** Cannot connect to database

**Solutions:**
```bash
# Test connection string
python -c "from sqlalchemy import create_engine; e = create_engine('your-uri'); e.connect()"

# Check network connectivity
telnet db-host 5432

# Verify credentials
# Ensure database user has proper permissions

# For Docker: use host.docker.internal for localhost
docker run -e SQLALCHEMY_DATABASE_URI=postgresql://user:pass@host.docker.internal:5432/db ...
```

### Secret Key Issues

**Problem:** "A SECRET_KEY must be configured"

**Solutions:**
```bash
# Generate secret key
openssl rand -base64 42

# Set environment variable
export SUPERSET_SECRET_KEY='your-generated-key'

# Or in superset_config.py
SECRET_KEY = 'your-generated-key'
```

### Memory Issues

**Problem:** Out of memory errors

**Solutions:**
```bash
# Reduce Celery workers
CELERYD_CONCURRENCY = 2

# Enable query caching
CACHE_CONFIG = {...}

# Limit query results
SQL_MAX_ROW = 100000

# Increase swap (Linux)
sudo fallocate -l 4G /swapfile
sudo chmod 600 /swapfile
sudo mkswap /swapfile
sudo swapon /swapfile
```

### Permission Issues

**Problem:** Permission denied errors

**Solutions:**
```bash
# Fix ownership
sudo chown -R $(whoami):$(whoami) ~/superset

# Check file permissions
ls -la

# Run with proper user (Docker)
docker run -u $(id -u):$(id -g) ...
```

### CORS Issues

**Problem:** CORS errors when embedding

**Solutions:**
```python
# superset_config.py
ENABLE_CORS = True
CORS_OPTIONS = {
    'supports_credentials': True,
    'allow_headers': ['*'],
    'resources': ['*'],
    'origins': ['http://localhost:3000', 'https://yourdomain.com'],
}
```

### Slow Query Performance

**Problem:** Queries timing out

**Solutions:**
```python
# superset_config.py
# Increase timeouts
SQLLAB_ASYNC_TIME_LIMIT_SEC = 60 * 60 * 6
SUPERSET_WEBSERVER_TIMEOUT = 600

# Enable async query execution
SQLLAB_ASYNC_TIME_LIMIT_SEC = 60 * 5

# Configure caching
CACHE_CONFIG = {
    'CACHE_TYPE': 'RedisCache',
    'CACHE_DEFAULT_TIMEOUT': 86400,
}
```

### Common Error Messages

| Error | Solution |
|-------|----------|
| "No module named superset" | Install: `pip install apache-superset` |
| "Address already in use" | Change port: `superset run -p 8089` |
| "Permission denied: superset.db" | Fix permissions or change DB URI |
| "CSRF token missing" | Clear browser cookies/cache |
| "Database locked" | Use PostgreSQL/MySQL instead of SQLite |
| "Unable to load dialect" | Install database driver |

### Getting Help

```bash
# Check version
superset version

# Verbose logging
superset --debug run

# Check configuration
superset show-unix-socket

# Documentation
# https://superset.apache.org/docs/intro

# GitHub Issues
# https://github.com/apache/superset/issues

# Slack Community
# https://join.slack.com/t/apache-superset/shared_invite/...
```

---

## Best Practices

### Security

1. **Change Default Credentials:**
   ```bash
   superset fab create-admin  # Create new admin
   # Delete default admin user
   ```

2. **Use Strong Secret Key:**
   ```bash
   openssl rand -base64 42
   ```

3. **Enable HTTPS in Production:**
   ```python
   SESSION_COOKIE_SECURE = True
   SESSION_COOKIE_HTTPONLY = True
   ```

4. **Use Database for Metadata:**
   - PostgreSQL or MySQL recommended
   - Avoid SQLite in production

### Performance

1. **Enable Caching:**
   ```python
   CACHE_CONFIG = {'CACHE_TYPE': 'RedisCache', ...}
   ```

2. **Use Celery for Async Queries:**
   ```python
   CELERY_CONFIG = CeleryConfig
   ```

3. **Configure Database Connection Pool:**
   ```python
   SQLALCHEMY_ENGINE_OPTIONS = {
       'pool_size': 10,
       'max_overflow': 20,
   }
   ```

### Maintenance

1. **Regular Backups:**
   ```bash
   # Export dashboards
   superset export-dashboards -f dashboards.json
   
   # Backup database
   pg_dump superset > superset_backup.sql
   ```

2. **Monitor Logs:**
   ```bash
   docker logs superset_app -f
   ```

3. **Update Regularly:**
   ```bash
   pip install --upgrade apache-superset
   superset db upgrade
   superset init
   ```

---

## Resources

- **Website:** https://superset.apache.org
- **GitHub:** https://github.com/apache/superset
- **Documentation:** https://superset.apache.org/docs/intro
- **Docker Hub:** https://hub.docker.com/r/apache/superset
- **Helm Charts:** https://github.com/apache/superset/tree/master/helm/superset

---

*Last Updated: April 2026*
