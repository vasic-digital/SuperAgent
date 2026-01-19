# MCP Servers Research Documentation

## Status: RESEARCHED
**Date**: 2026-01-19

---

## 1. Official Reference Servers (modelcontextprotocol/servers)

### Active Reference Implementations

| Server | Package | Description | Requirements |
|--------|---------|-------------|--------------|
| Everything | @modelcontextprotocol/server-everything | Test server with prompts, resources, tools | Node.js |
| Fetch | @modelcontextprotocol/server-fetch | Web content fetching and conversion | Node.js |
| Filesystem | @modelcontextprotocol/server-filesystem | Secure file operations | Node.js |
| Git | @modelcontextprotocol/server-git | Repository operations | Node.js, Git |
| Memory | @modelcontextprotocol/server-memory | Knowledge graph memory | Node.js |
| Sequential Thinking | @modelcontextprotocol/server-sequential-thinking | Problem-solving sequences | Node.js |
| Time | @modelcontextprotocol/server-time | Timezone operations | Node.js |

### Archived Servers (Separate Repository)

| Server | Previous Package | Status |
|--------|-----------------|--------|
| Puppeteer | @modelcontextprotocol/server-puppeteer | Archived |
| SQLite | mcp-server-sqlite | Archived |
| PostgreSQL | mcp-server-postgresql | Archived |
| GitHub | mcp-server-github | Archived |
| GitLab | mcp-server-gitlab | Archived |
| Google Drive | mcp-server-google-drive | Archived |
| Brave Search | mcp-server-brave-search | Archived |

---

## 2. Vector Database MCP Servers

### Chroma MCP Server

**Repository**: https://github.com/chroma-core/chroma

**Installation**:
```bash
pip install chromadb
npm install chromadb  # JS client
```

**Server Mode**:
```bash
chroma run --path /chroma_db_path
```

**Features**:
- In-memory or persistent storage
- Default: Sentence Transformers (all-minilm-l6-v2)
- OpenAI, Cohere embedding support
- Metadata filtering
- LangChain/LlamaIndex integration

**Docker**:
```yaml
services:
  chromadb:
    image: chromadb/chroma:latest
    ports:
      - "8000:8000"
    volumes:
      - chroma_data:/chroma/chroma
```

### Qdrant MCP Server

**Repository**: https://github.com/qdrant/qdrant

**Installation**:
```bash
docker run -p 6333:6333 qdrant/qdrant
pip install qdrant-client  # Python client
```

**Features**:
- REST API (port 6333) + gRPC
- JSON payload filtering
- Vector quantization (97% RAM reduction)
- Horizontal scaling with sharding
- Write-ahead logging for durability

**Client Libraries**: Go, Rust, JS/TS, Python, .NET, Java

**Docker**:
```yaml
services:
  qdrant:
    image: qdrant/qdrant:latest
    ports:
      - "6333:6333"
      - "6334:6334"
    volumes:
      - qdrant_storage:/qdrant/storage
```

---

## 3. Design & UI MCP Servers

### Figma Integration Servers

| Server | Source | Features |
|--------|--------|----------|
| Cursor Talk to Figma | Community | Read/modify designs via natural language |
| Framelink Figma MCP | Community | Fetch and simplify file data |
| Figma Chunking MCP | Community | Handle large files with pagination |
| Figma to React | Community | Convert designs to React components |

**Requirements**: Figma API access token

### Adobe Integration

| Server | Features | Requirements |
|--------|----------|--------------|
| Illustrator MCP | JavaScript scripting | macOS + Adobe Illustrator |
| Photoshop MCP | Python API control | Photoshop + Python |

### Collaboration Tools

| Server | Features | Requirements |
|--------|----------|--------------|
| MCP-Miro | Whiteboard operations | Miro OAuth token |

---

## 4. Image Generation MCP Servers

### Cloud-Based Generation

| Server | Model | Requirements |
|--------|-------|--------------|
| Replicate Flux MCP | Flux.1 | Replicate API token |
| FLUX Generator MCP | Black Forest Lab | BFL API key |
| 4o-image | GPT-4 Vision | OpenAI API key |

### Local Generation

| Server | Features | Requirements |
|--------|----------|--------------|
| Stable Diffusion MCP | Local WebUI | GPU + SD WebUI |
| ImageSorcery MCP | Crop, resize, OCR | Python 3.10+ |

### Asset Creation

| Server | Features | Requirements |
|--------|----------|--------------|
| SVGMaker MCP | AI SVG generation | API key |

---

## 5. Development Tool MCP Servers

### From mcpservers.org Registry (5,938+ servers)

**Notable Categories**:
- Search: 302AI Web Search, Brave Search
- Web Scraping: Bright Data, Puppeteer alternatives
- Databases: SQLite, PostgreSQL, MongoDB
- Cloud: AWS, GCP, Azure integrations
- File Systems: S3, Google Drive, Dropbox
- Version Control: GitHub, GitLab, Bitbucket

---

## 6. Integration Strategy

### Phase 1: Core Infrastructure
1. Update npm package definitions in `internal/mcp/preinstaller.go`
2. Add new server configs to `internal/mcp/server_registry.go`
3. Create lazy connection handlers in `internal/mcp/connection_pool.go`

### Phase 2: Vector DB MCPs
1. Create Chroma MCP adapter
2. Create Qdrant MCP adapter
3. Add unified vector operations

### Phase 3: Design MCPs
1. Implement Figma OAuth flow
2. Create design file handlers
3. Add React component generator

### Phase 4: Image MCPs
1. Create cloud generation adapters
2. Implement local SD WebUI integration
3. Add image processing pipeline

---

## 7. Docker Compose Stacks

### mcp-core-stack.yml
```yaml
version: '3.8'
services:
  mcp-filesystem:
    build: ./docker/mcp/filesystem
    volumes:
      - workspace:/workspace:rw
  mcp-git:
    build: ./docker/mcp/git
    volumes:
      - git_repos:/repos:rw
  mcp-memory:
    build: ./docker/mcp/memory
    volumes:
      - memory_data:/data:rw
```

### mcp-vector-stack.yml
```yaml
version: '3.8'
services:
  chromadb:
    image: chromadb/chroma:latest
    ports:
      - "8000:8000"
  qdrant:
    image: qdrant/qdrant:latest
    ports:
      - "6333:6333"
```

### mcp-image-stack.yml
```yaml
version: '3.8'
services:
  stable-diffusion:
    image: automatic1111/stable-diffusion-webui:latest
    deploy:
      resources:
        reservations:
          devices:
            - capabilities: [gpu]
```

---

## 8. Testing Requirements

### Unit Tests
- Server registration
- Connection pooling
- Tool discovery
- Request/response handling

### Integration Tests
- End-to-end tool execution
- Multi-server coordination
- Error handling and fallback
- Rate limiting

### Challenge Scripts
- `mcp_core_challenge.sh`: Test all core servers
- `mcp_vector_challenge.sh`: Test vector DB operations
- `mcp_design_challenge.sh`: Test design integrations
- `mcp_image_challenge.sh`: Test image generation

---

## 9. Security Considerations

- API key management via secure vault
- OAuth token refresh automation
- Rate limiting per provider
- Input validation and sanitization
- Audit logging for all operations
