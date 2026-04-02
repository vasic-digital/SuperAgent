# TaskWeaver User Guide

## Overview

TaskWeaver is Microsoft's open-source, code-first agent framework designed for seamlessly planning and executing data analytics tasks. Unlike traditional LLM frameworks that work with text strings, TaskWeaver interprets user requests through executable code snippets and treats user-defined plugins as callable functions. It supports rich data structures like DataFrames, incorporates domain-specific knowledge, and provides secure code execution environments.

**Key Features:**
- Code-first architecture (Python)
- Rich data structure support (DataFrames, arrays)
- Custom plugin system
- Domain-specific knowledge integration
- Stateful conversation memory
- Code verification and validation
- Secure execution with process isolation
- Multi-LLM support (OpenAI, Azure, Ollama)
- Web UI for demos
- Library mode for integration

---

## Installation Methods

### Method 1: pip Install (Recommended)

Requirements: Python 3.10+

```bash
# Create virtual environment
python -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# Install TaskWeaver
pip install taskweaver

# Or install with all optional dependencies
pip install taskweaver[all]
```

### Method 2: Clone from Source

```bash
# Clone repository
git clone https://github.com/microsoft/TaskWeaver.git
cd TaskWeaver

# Install dependencies
pip install -r requirements.txt

# Install in development mode
pip install -e .
```

### Method 3: Using uv (Modern Python)

```bash
# Install uv
curl -sSL https://astral.sh/uv/install.sh | sh

# Clone and install
git clone https://github.com/microsoft/TaskWeaver.git
cd TaskWeaver
uv pip install -e .
```

### Method 4: Docker

```bash
# Pull image
docker pull microsoft/taskweaver:latest

# Run with environment variables
docker run -it \
  -e OPENAI_API_KEY=$OPENAI_API_KEY \
  -v $(pwd)/project:/app/project \
  microsoft/taskweaver:latest
```

### Verify Installation

```bash
# Check version
taskweaver --version

# Show help
taskweaver --help

# Check installation
python -c "import taskweaver; print(taskweaver.__version__)"
```

---

## Quick Start

### 1. Configure API Keys

```bash
# Required for OpenAI
export OPENAI_API_KEY=your_openai_api_key

# Or for Azure OpenAI
export AZURE_OPENAI_API_KEY=your_azure_key
export AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
export AZURE_OPENAI_API_VERSION=2024-02-15-preview

# For Ollama (local models)
export OLLAMA_HOST=http://localhost:11434
```

### 2. Create Project Directory

```bash
# Create project folder
mkdir -p ~/taskweaver-projects/my-project
cd ~/taskweaver-projects/my-project

# TaskWeaver will create necessary subdirectories:
# - plugins/       (custom plugins)
# - code/          (generated code)
# - logs/          (execution logs)
# - config/        (configuration files)
```

### 3. Start TaskWeaver

```bash
# Start CLI mode
taskweaver -p ./

# Or explicitly specify project path
python -m taskweaver -p ./
```

### 4. First Interaction

```
=========================================================
 _____         _     _       __
|_   _|_ _ ___| | _ | |     / /__  ____ __   _____  _____
  | |/ _` / __| |/ /| | /| / / _ \/ __ `/ | / / _ \/ ___/
  | | (_| \__ \   < | |/ |/ /  __/ /_/ /| |/ /  __/ /
  |_|\__,_|___/_|\_\|__/|__/\___/\__,_/ |___/\___/_/
=========================================================
TaskWeaver: I am TaskWeaver, an AI assistant. To get started, could you please enter your request?
Human: Load the sample sales data and show me total revenue by region
```

---

## CLI Commands Reference

### Core Commands

| Command | Description |
|---------|-------------|
| `taskweaver` | Start TaskWeaver CLI |
| `taskweaver -p <path>` | Start with project path |
| `taskweaver --version` | Show version |
| `taskweaver --help` | Show help |

### Project Options

| Flag | Description |
|------|-------------|
| `-p, --project <path>` | Project directory path |
| `-c, --config <file>` | Configuration file path |
| `--skip-verification` | Skip code verification |
| `--execution-timeout <sec>` | Code execution timeout |
| `--max-iterations <n>` | Max planner iterations |
| `--llm-model <model>` | LLM model to use |

### Configuration Commands

```bash
# Initialize new project
taskweaver init ./my-project

# Validate configuration
taskweaver validate-config

# Show current configuration
taskweaver show-config
```

### Advanced Options

```bash
# Use specific LLM model
taskweaver -p ./project --llm-model gpt-4

# Enable debug logging
taskweaver -p ./project --log-level debug

# Set execution timeout
taskweaver -p ./project --execution-timeout 120

# Limit iterations
taskweaver -p ./project --max-iterations 10
```

---

## Project Structure

```
my-project/
├── taskweaver_config.json    # Main configuration
├── plugins/                  # Custom plugins
│   ├── __init__.py
│   ├── sql_analyzer.py
│   ├── data_loader.py
│   └── custom_visualization.py
├── code/                     # Generated code (auto-created)
│   └── generated/
├── logs/                     # Execution logs (auto-created)
│   └── sessions/
├── knowledge/                # Domain knowledge files
│   ├── domain_knowledge.json
│   └── examples/
└── data/                     # Project data files
    └── sample_data.csv
```

---

## Configuration

### Configuration File (taskweaver_config.json)

```json
{
  "llm": {
    "api_type": "openai",
    "api_key": "${OPENAI_API_KEY}",
    "api_base": "https://api.openai.com/v1",
    "api_version": "2024-02-01",
    "model": "gpt-4",
    "temperature": 0.0,
    "max_tokens": 4000,
    "top_p": 1.0,
    "frequency_penalty": 0.0,
    "presence_penalty": 0.0
  },
  "embedding": {
    "api_type": "openai",
    "api_key": "${OPENAI_API_KEY}",
    "model": "text-embedding-3-small",
    "batch_size": 100
  },
  "planner": {
    "max_iterations": 10,
    "prompt_path": null
  },
  "code_interpreter": {
    "execution_timeout": 120,
    "max_output_lines": 1000,
    "enable_verification": true,
    "allowed_modules": [
      "pandas",
      "numpy",
      "matplotlib",
      "seaborn",
      "scipy",
      "sklearn"
    ],
    "forbidden_modules": [
      "os",
      "subprocess",
      "sys"
    ]
  },
  "logging": {
    "level": "info",
    "file_path": "./logs/taskweaver.log"
  },
  "session": {
    "max_history_messages": 20,
    "enable_compression": true
  },
  "plugin": {
    "enabled": true,
    "module_paths": ["./plugins"]
  },
  "security": {
    "sandbox_mode": "subprocess",
    "enable_code_verification": true,
    "max_memory_mb": 2048
  }
}
```

### Azure OpenAI Configuration

```json
{
  "llm": {
    "api_type": "azure",
    "api_key": "${AZURE_OPENAI_API_KEY}",
    "api_base": "https://your-resource.openai.azure.com/",
    "api_version": "2024-02-15-preview",
    "deployment_name": "gpt-4-deployment",
    "model": "gpt-4"
  }
}
```

### Ollama (Local) Configuration

```json
{
  "llm": {
    "api_type": "ollama",
    "api_base": "http://localhost:11434",
    "model": "llama3",
    "temperature": 0.7
  }
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |
| `AZURE_OPENAI_API_KEY` | Azure OpenAI key |
| `AZURE_OPENAI_ENDPOINT` | Azure OpenAI endpoint |
| `OLLAMA_HOST` | Ollama server URL |
| `TASKWEAVER_PROJECT` | Default project path |
| `TASKWEAVER_CONFIG` | Config file path |

---

## Usage Examples

### Example 1: Data Analysis Session

```python
# In TaskWeaver CLI
Human: Load the CSV file at data/sales.csv and analyze monthly sales trends

# TaskWeaver will:
# 1. Generate code to load the CSV
# 2. Parse date columns
# 3. Aggregate by month
# 4. Create visualization
# 5. Present results
```

### Example 2: Custom Plugin Creation

```python
# plugins/data_analyzer.py
from taskweaver.plugin import Plugin, register_plugin

@register_plugin
class DataAnalyzer(Plugin):
    def __call__(self, df, analysis_type="summary"):
        """
        Analyze a DataFrame and return insights.
        
        Args:
            df: pandas DataFrame
            analysis_type: type of analysis (summary, correlation, outliers)
        """
        if analysis_type == "summary":
            return {
                "shape": df.shape,
                "columns": list(df.columns),
                "dtypes": df.dtypes.to_dict(),
                "missing": df.isnull().sum().to_dict(),
                "describe": df.describe().to_dict()
            }
        elif analysis_type == "correlation":
            return df.corr().to_dict()
        elif analysis_type == "outliers":
            numeric_cols = df.select_dtypes(include=['number'])
            outliers = {}
            for col in numeric_cols.columns:
                Q1 = df[col].quantile(0.25)
                Q3 = df[col].quantile(0.75)
                IQR = Q3 - Q1
                outliers[col] = df[
                    (df[col] < Q1 - 1.5 * IQR) | 
                    (df[col] > Q3 + 1.5 * IQR)
                ].shape[0]
            return outliers
```

Usage:
```
Human: Analyze the sales data for outliers
# TaskWeaver calls: data_analyzer(sales_df, analysis_type="outliers")
```

### Example 3: Domain Knowledge Integration

```json
// knowledge/domain_knowledge.json
{
  "domain": "retail_analytics",
  "concepts": {
    "customer_lifetime_value": {
      "description": "Total revenue expected from a customer",
      "formula": "avg_order_value * purchase_frequency * customer_lifespan",
      "example": "CLV = $50 * 4 purchases/year * 3 years = $600"
    },
    "churn_rate": {
      "description": "Percentage of customers who stop buying",
      "formula": "(customers_lost / total_customers) * 100"
    }
  },
  "common_queries": [
    {
      "description": "Calculate customer lifetime value",
      "code": "df.groupby('customer_id')['revenue'].sum()"
    }
  ]
}
```

### Example 4: Using as Library

```python
# app.py
from taskweaver.app import TaskWeaverApp

# Initialize app
app = TaskWeaverApp(
    project_path="./my-project",
    config_path="./taskweaver_config.json"
)

# Start session
session = app.new_session()

# Send messages
response = session.send_message(
    "Load data.csv and calculate correlation matrix"
)
print(response)

# Continue conversation
response = session.send_message(
    "Now plot a heatmap of the correlations"
)
print(response)

# Close session
session.close()
```

### Example 5: Web UI Mode

```bash
# Start web UI
taskweaver -p ./project --web

# Or explicitly
python -m taskweaver.web -p ./project

# Access at http://localhost:5000
```

### Example 6: Batch Processing

```python
# batch_analysis.py
from taskweaver.app import TaskWeaverApp

app = TaskWeaverApp(project_path="./project")

queries = [
    "Load sales_data.csv",
    "Filter for Q4 2024",
    "Calculate revenue by product category",
    "Create bar chart of top 10 categories",
    "Export results to output.csv"
]

session = app.new_session()
for query in queries:
    response = session.send_message(query)
    print(f"Query: {query}")
    print(f"Response: {response}\n")
session.close()
```

### Example 7: Integration with FastAPI

```python
# api.py
from fastapi import FastAPI
from taskweaver.app import TaskWeaverApp

app = FastAPI()
taskweaver = TaskWeaverApp(project_path="./project")

@app.post("/analyze")
async def analyze(request: str):
    session = taskweaver.new_session()
    try:
        response = session.send_message(request)
        return {"result": response}
    finally:
        session.close()

@app.post("/query")
async def query(sql: str):
    session = taskweaver.new_session()
    try:
        response = session.send_message(f"Execute SQL: {sql}")
        return {"result": response}
    finally:
        session.close()
```

---

## TUI / Interactive Features

### CLI Interaction Modes

**1. Single Query Mode:**
```bash
taskweaver -p ./project --query "Analyze data.csv"
```

**2. Interactive Mode:**
```bash
taskweaver -p ./project
> Load data.csv
> Show summary statistics
> Create histogram of sales
> exit
```

**3. File Input Mode:**
```bash
taskweaver -p ./project --input queries.txt --output results.txt
```

### Web UI Features

The Web UI provides:
- Chat interface with conversation history
- Code execution display
- Visualization rendering
- File upload/download
- Session management
- Plugin browser

### Session Management

```python
# List active sessions
sessions = app.list_sessions()

# Resume session
session = app.resume_session(session_id="session-123")

# Close specific session
app.close_session(session_id="session-123")

# Close all sessions
app.close_all_sessions()
```

---

## Plugin Development

### Basic Plugin Structure

```python
# plugins/my_plugin.py
from taskweaver.plugin import Plugin, register_plugin
from typing import Dict, Any

@register_plugin
class MyPlugin(Plugin):
    """
    Description of what this plugin does.
    """
    
    def __call__(self, param1: str, param2: int = 10) -> Dict[str, Any]:
        """
        Execute the plugin.
        
        Args:
            param1: Description of param1
            param2: Description of param2
            
        Returns:
            Dictionary with results
        """
        # Plugin logic here
        result = {"param1": param1, "param2": param2}
        return result
```

### Plugin with External Dependencies

```python
# plugins/advanced_analyzer.py
from taskweaver.plugin import Plugin, register_plugin
import pandas as pd
import numpy as np
from scipy import stats

@register_plugin
class AdvancedAnalyzer(Plugin):
    """
    Advanced statistical analysis plugin.
    """
    
    def __call__(self, df: pd.DataFrame, test_type: str = "normality"):
        if test_type == "normality":
            results = {}
            for col in df.select_dtypes(include=[np.number]).columns:
                stat, p_value = stats.shapiro(df[col].dropna())
                results[col] = {
                    "statistic": stat,
                    "p_value": p_value,
                    "is_normal": p_value > 0.05
                }
            return results
        elif test_type == "anova":
            # ANOVA implementation
            pass
```

---

## Troubleshooting

### Installation Issues

**Problem:** `pip install taskweaver` fails

**Solutions:**
```bash
# Upgrade pip
pip install --upgrade pip

# Install with specific Python version
python3.10 -m pip install taskweaver

# Check Python version
python --version  # Must be 3.10+

# Install from source
git clone https://github.com/microsoft/TaskWeaver.git
cd TaskWeaver
pip install -e .
```

### API Key Issues

**Problem:** "API key not found" errors

**Solutions:**
```bash
# Set environment variable
export OPENAI_API_KEY="your-key-here"

# Or add to shell profile
echo 'export OPENAI_API_KEY="your-key"' >> ~/.bashrc
source ~/.bashrc

# Or use .env file
echo "OPENAI_API_KEY=your-key" > .env
source .env
```

### Code Execution Failures

**Problem:** Code fails to execute or times out

**Solutions:**
```json
{
  "code_interpreter": {
    "execution_timeout": 300,
    "max_memory_mb": 4096,
    "enable_verification": false
  }
}
```

### Plugin Not Found

**Problem:** Custom plugins not loading

**Solutions:**
```bash
# Check plugin path in config
"plugin": {
  "module_paths": ["./plugins", "./custom_plugins"]
}

# Verify plugin structure
# Must have @register_plugin decorator
# Must be in configured module path

# Check for syntax errors
python -m py_compile plugins/my_plugin.py
```

### Memory Issues

**Problem:** Out of memory during execution

**Solutions:**
```json
{
  "security": {
    "max_memory_mb": 4096,
    "sandbox_mode": "subprocess"
  },
  "code_interpreter": {
    "max_output_lines": 100
  }
}
```

### LLM Connection Issues

**Problem:** Cannot connect to LLM API

**Solutions:**
```bash
# Test API connectivity
curl https://api.openai.com/v1/models \
  -H "Authorization: Bearer $OPENAI_API_KEY"

# Check proxy settings
export HTTP_PROXY=http://proxy.company.com:8080
export HTTPS_PROXY=http://proxy.company.com:8080

# For Azure: verify endpoint
export AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com/
```

### Common Error Messages

| Error | Solution |
|-------|----------|
| "No module named taskweaver" | Install: `pip install taskweaver` |
| "Project not found" | Check `-p` path exists |
| "Invalid configuration" | Validate JSON syntax |
| "Code execution timeout" | Increase `execution_timeout` |
| "Memory limit exceeded" | Increase `max_memory_mb` |
| "LLM API error" | Check API key and quotas |
| "Plugin import failed" | Check Python syntax in plugin |

### Getting Help

```bash
# Enable debug logging
taskweaver -p ./project --log-level debug

# Check version
taskweaver --version

# Validate config
taskweaver validate-config

# Documentation
# https://github.com/microsoft/TaskWeaver/blob/main/README.md

# Issues
# https://github.com/microsoft/TaskWeaver/issues
```

---

## Best Practices

### 1. Project Organization
```
project/
├── taskweaver_config.json
├── plugins/          # Custom plugins
├── data/            # Data files
├── output/          # Generated outputs
└── knowledge/       # Domain knowledge
```

### 2. Plugin Naming
- Use descriptive names
- Include docstrings
- Type hints for parameters
- Return structured data

### 3. Session Management
- Close sessions when done
- Reuse sessions for related queries
- Monitor memory usage

### 4. Security
```json
{
  "security": {
    "sandbox_mode": "subprocess",
    "enable_code_verification": true,
    "max_memory_mb": 2048
  }
}
```

### 5. Error Handling
```python
# In plugins
from taskweaver.plugin import Plugin, register_plugin

@register_plugin
class SafePlugin(Plugin):
    def __call__(self, data):
        try:
            result = self.process(data)
            return {"success": True, "data": result}
        except Exception as e:
            return {"success": False, "error": str(e)}
```

---

## Resources

- **GitHub:** https://github.com/microsoft/TaskWeaver
- **Paper:** https://arxiv.org/abs/2311.17541
- **Blog:** https://www.microsoft.com/en-us/research/blog/taskweaver/
- **Issues:** https://github.com/microsoft/TaskWeaver/issues

---

*Last Updated: April 2026*
