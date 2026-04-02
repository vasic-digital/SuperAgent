# Fauxpilot User Guide

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

### Prerequisites

- Docker and docker compose >= 1.28
- NVIDIA GPU with Compute Capability >= 6.0
- nvidia-container-toolkit
- curl and zstd
- At least 50GB local storage (for models)

### Method 1: Install Script

```bash
# Clone the repository
git clone https://github.com/moyix/fauxpilot.git
cd fauxpilot

# Run setup script
./setup.sh
```

### Method 2: Manual Setup

```bash
# Clone repository
git clone https://github.com/moyix/fauxpilot.git
cd fauxpilot

# Setup will download and convert models
# Choose model based on your GPU VRAM
```

### Method 3: Docker Setup

```bash
# Build Docker images
docker compose build

# Download models
./setup.sh

# Launch services
./launch.sh
```

---

## Quick Start

```bash
# Navigate to fauxpilot directory
cd fauxpilot

# Run setup (first time only)
./setup.sh

# Launch Fauxpilot
./launch.sh

# Test the service
curl -s -H "Accept: application/json" \
  -H "Content-type: application/json" \
  -X POST \
  -d '{"prompt":"def hello","max_tokens":100,"temperature":0.1,"stop":["\n\n"]}' \
  http://localhost:5000/v1/engines/codegen/completions
```

---

## CLI Commands

### Command: setup.sh

**Description:** Download and configure models for Fauxpilot.

**Usage:**
```bash
./setup.sh
```

**Interactive Options:**
- Select model based on GPU VRAM:
  - `codegen-350M-multi` - Smallest model
  - `codegen-2B-multi` - 2B parameters (recommended for 8GB VRAM)
  - `codegen-6B-multi` - 6B parameters (requires 13GB+ VRAM)
  - `codegen-16B-multi` - Largest model (requires more VRAM)

**Examples:**
```bash
# Run setup interactively
./setup.sh

# Follow prompts to select model
```

### Command: launch.sh

**Description:** Start the Fauxpilot server.

**Usage:**
```bash
./launch.sh
```

**Examples:**
```bash
# Launch Fauxpilot services
./launch.sh

# Services will be available at http://localhost:5000
```

### Command: shutdown.sh

**Description:** Stop the Fauxpilot server.

**Usage:**
```bash
./shutdown.sh
```

**Examples:**
```bash
# Stop Fauxpilot
./shutdown.sh
```

---

## TUI/Interactive Commands

Fauxpilot runs as a server and is typically used via IDE integration or API calls.

---

## Configuration

### Docker Compose Configuration

**File:** `docker-compose.yaml`

```yaml
version: '3.8'
services:
  triton:
    image: fauxpilot-triton
    runtime: nvidia
    environment:
      - NVIDIA_VISIBLE_DEVICES=all
    volumes:
      - ./models:/models
    ports:
      - "8000:8000"
  
  copilot_proxy:
    image: fauxpilot-proxy
    ports:
      - "5000:5000"
    environment:
      - TRITON_HOST=triton
      - TRITON_PORT=8000
```

### Environment Configuration

**File:** `.env` (created by setup.sh)

```bash
MODEL_DIR=./models/codegen-2B-multi
HF_CACHE_DIR=./cache
GPU_LAYERS=all
```

### GPU Support Matrix

| Model | VRAM Required | GPUs |
|-------|--------------|------|
| codegen-350M-multi | ~2GB | 1 |
| codegen-2B-multi | ~6GB | 1 |
| codegen-6B-multi | ~13GB | 1-2 |
| codegen-16B-multi | ~35GB | 2-4 |

### Multi-GPU Setup

If you have multiple GPUs, the model can be split across them:
```bash
# Two RTX 3080 (10GB each) can run 6B model
# Set in .env or during setup
```

---

## API/Protocol Endpoints

### OpenAI-Compatible API

Fauxpilot provides an OpenAI-compatible API at `http://localhost:5000`.

#### Endpoint: POST /v1/engines/codegen/completions

**Description:** Generate code completions.

**Request:**
```json
{
  "prompt": "def hello",
  "max_tokens": 100,
  "temperature": 0.1,
  "stop": ["\n\n"]
}
```

**Response:**
```json
{
  "id": "cmpl-OCButmOAbNedOMOxjPc0v9skuLdk7",
  "model": "codegen",
  "object": "text_completion",
  "created": 1692885668,
  "choices": [
    {
      "text": "(self):\n        return \"Hello World!\"",
      "index": 0,
      "finish_reason": "stop",
      "logprobs": null
    }
  ],
  "usage": {
    "completion_tokens": 11,
    "prompt_tokens": 2,
    "total_tokens": 13
  }
}
```

#### Endpoint: POST /v1/completions

**Description:** Alternative completions endpoint.

**Request:**
```bash
curl -X POST http://localhost:5000/v1/completions \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "// Calculate sum of array in JavaScript\nfunction sum",
    "max_tokens": 50,
    "temperature": 0.2
  }'
```

### VS Code Configuration

Add to `settings.json`:

```json
{
  "github.copilot.advanced": {
    "debug.overrideEngine": "codegen",
    "debug.testOverrideProxyUrl": "http://localhost:5000",
    "debug.overrideProxyUrl": "http://localhost:5000"
  }
}
```

### Vim Configuration

Locate copilot.vim installation and modify:

```bash
# Find installation directory
# Then run:
sed -i 's|https://copilot-proxy.githubusercontent.com|http://localhost:5000|g' \
  /path/to/copilot.vim/autoload/copilot.vim
```

---

## Usage Examples

### Example 1: Basic Setup

```bash
# Clone repository
git clone https://github.com/moyix/fauxpilot.git
cd fauxpilot

# Run setup
./setup.sh
# Select model based on your GPU

# Launch server
./launch.sh

# Test with curl
curl -s http://localhost:5000/v1/engines/codegen/completions \
  -H "Content-Type: application/json" \
  -d '{"prompt":"def fibonacci","max_tokens":50}'
```

### Example 2: VS Code Integration

```bash
# 1. Install GitHub Copilot extension in VS Code
# 2. Open settings.json
# 3. Add configuration:

{
  "github.copilot.advanced": {
    "debug.overrideEngine": "codegen",
    "debug.testOverrideProxyUrl": "http://localhost:5000",
    "debug.overrideProxyUrl": "http://localhost:5000"
  }
}

# 4. Restart VS Code
# 5. Copilot will now use your local Fauxpilot server
```

### Example 3: Direct API Usage

```bash
# Python script using OpenAI library
python << 'EOF'
import openai

openai.api_base = "http://localhost:5000/v1"
openai.api_key = "dummy"

response = openai.Completion.create(
    model="codegen",
    prompt="def quicksort(arr):",
    max_tokens=100,
    temperature=0.1
)

print(response.choices[0].text)
EOF
```

### Example 4: Multi-GPU Setup

```bash
# For 6B model on two GPUs
# Edit docker-compose.yaml to specify GPUs

services:
  triton:
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: 2
              capabilities: [gpu]

# Or use environment variable
export NVIDIA_VISIBLE_DEVICES=0,1
./launch.sh
```

### Example 5: Offline Deployment

```bash
# On machine with internet:
./setup.sh
# Download models

# Export Docker images
docker save > fauxpilot-triton.tar fauxpilot-main-triton
docker save > fauxpilot-proxy.tar fauxpilot-main-copilot_proxy

# Copy to offline machine:
# - fauxpilot-triton.tar
# - fauxpilot-proxy.tar
# - docker-compose.yaml
# - .env
# - models/ directory

# On offline machine:
docker load < fauxpilot-triton.tar
docker load < fauxpilot-proxy.tar
./launch.sh
```

---

## Troubleshooting

### Issue: Docker permission denied

**Solution:**
```bash
sudo chown $USER /var/run/docker.sock
sudo systemctl restart docker

# Or add user to docker group
sudo usermod -aG docker $USER
# Log out and back in
```

### Issue: NVIDIA container toolkit not installed

**Solution:**
```bash
# Install nvidia-container-toolkit
distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
curl -s -L https://nvidia.github.io/libnvidia-container/gpgkey | \
  sudo apt-key add -
curl -s -L https://nvidia.github.io/libnvidia-container/$distribution/libnvidia-container.list | \
  sudo tee /etc/apt/sources.list.d/nvidia-container-toolkit.list
sudo apt-get update
sudo apt-get install -y nvidia-container-toolkit
sudo nvidia-ctk runtime configure --runtime=docker
sudo systemctl restart docker
```

### Issue: GPU not detected

**Solution:**
```bash
# Verify NVIDIA driver
nvidia-smi

# Check Docker can see GPU
docker run --rm --gpus all nvidia/cuda:11.0-base nvidia-smi

# If using WSL2:
# Ensure WSL2 has GPU support
```

### Issue: Out of memory

**Solution:**
```bash
# Use smaller model
./setup.sh
# Select codegen-350M-multi or codegen-2B-multi

# Or limit GPU layers in .env
GPU_LAYERS=20
```

### Issue: VRAM overflow

**Solution:**
```bash
# If using 6B model with 16GB VRAM, switch to 2B
./shutdown.sh
./setup.sh
# Select codegen-2B-multi
./launch.sh
```

### Issue: Model download fails

**Solution:**
```bash
# Check disk space
df -h

# Check internet connection
ping huggingface.co

# Manual download
# Download model files from HuggingFace/Moyix
# Place in models/ directory
# Re-run setup
```

### Issue: Slow completions

**Solution:**
- Ensure GPU is being used (check `nvidia-smi`)
- Use smaller model
- Reduce max_tokens
- Check GPU temperature and throttling

### Issue: VS Code not connecting

**Solution:**
```bash
# Verify server is running
curl http://localhost:5000/v1/engines/codegen/completions \
  -H "Content-Type: application/json" \
  -d '{"prompt":"test","max_tokens":1}'

# Check VS Code settings.json format
# Ensure no proxy blocking localhost:5000
```

---

## Additional Resources

- **GitHub Repository:** https://github.com/moyix/fauxpilot
- **Wiki:** https://github.com/moyix/fauxpilot/wiki
- **Discussion Forum:** GitHub Discussions
- **SalesForce CodeGen:** https://github.com/salesforce/CodeGen
- **NVIDIA Triton:** https://github.com/triton-inference-server/server
