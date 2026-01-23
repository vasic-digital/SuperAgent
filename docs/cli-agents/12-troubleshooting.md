# Troubleshooting

Common issues and solutions for CLI agent plugins.

## Quick Diagnostics

### Check HelixAgent Status

```bash
# Health check
curl http://localhost:7061/health

# Expected response:
# {"status":"healthy","version":"1.0.0"}
```

### Check Configuration

```bash
# Validate agent config
./bin/helixagent --validate-agent-config=opencode:~/.config/opencode/opencode.json

# List supported agents
./bin/helixagent --list-agents
```

### Check MCP Server

```bash
# Test MCP server directly
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | npx @helixagent/mcp-server

# Expected response:
# {"jsonrpc":"2.0","id":1,"result":{"tools":[...]}}
```

---

## Connection Issues

### Issue: "Connection refused"

**Symptoms:**
```
Error: connect ECONNREFUSED 127.0.0.1:7061
```

**Causes:**
1. HelixAgent not running
2. Wrong port configured
3. Firewall blocking connection

**Solutions:**

```bash
# 1. Start HelixAgent
./bin/helixagent

# 2. Check if running
curl http://localhost:7061/health

# 3. Check port
lsof -i :7061

# 4. Check firewall
sudo ufw status
sudo iptables -L -n | grep 7061
```

### Issue: "Connection timeout"

**Symptoms:**
```
Error: ETIMEDOUT
RequestError: timeout of 30000ms exceeded
```

**Causes:**
1. Network latency
2. HelixAgent overloaded
3. Request too large

**Solutions:**

```bash
# 1. Increase timeout in config
{
  "transport": {
    "timeout": 60000
  }
}

# 2. Check HelixAgent load
curl http://localhost:7061/metrics | grep helixagent_requests

# 3. Enable request batching for large payloads
```

### Issue: "SSL certificate error"

**Symptoms:**
```
Error: CERT_HAS_EXPIRED
Error: UNABLE_TO_VERIFY_LEAF_SIGNATURE
```

**Solutions:**

```bash
# 1. For development only - disable verification
export NODE_TLS_REJECT_UNAUTHORIZED=0

# 2. Add custom CA certificate
export NODE_EXTRA_CA_CERTS=/path/to/ca.crt

# 3. Use HTTP for local development
{
  "endpoint": "http://localhost:7061"
}
```

---

## Authentication Issues

### Issue: "401 Unauthorized"

**Symptoms:**
```
HTTP 401: Unauthorized
{"error":"invalid_api_key"}
```

**Solutions:**

```bash
# 1. Check API key is set
echo $HELIXAGENT_API_KEY

# 2. Set API key
export HELIXAGENT_API_KEY="your-api-key"

# 3. Verify in config
{
  "provider": {
    "apiKeyEnv": "HELIXAGENT_API_KEY"
  }
}

# 4. Test directly
curl -H "Authorization: Bearer $HELIXAGENT_API_KEY" \
  http://localhost:7061/v1/models
```

### Issue: "403 Forbidden"

**Symptoms:**
```
HTTP 403: Forbidden
{"error":"insufficient_permissions"}
```

**Causes:**
1. API key lacks required permissions
2. Rate limit exceeded
3. IP not whitelisted

**Solutions:**

```bash
# 1. Check rate limits
curl http://localhost:7061/v1/rate-limit-status \
  -H "Authorization: Bearer $HELIXAGENT_API_KEY"

# 2. Verify API key permissions
# Check your HelixAgent admin console
```

---

## Configuration Issues

### Issue: "Invalid configuration"

**Symptoms:**
```
Error: provider.base_url is required
Error: mcp.helixagent: type is required
```

**Solutions:**

```bash
# 1. Validate config
./bin/helixagent --validate-agent-config=opencode:config.json

# 2. Check required fields
{
  "provider": {
    "type": "openai-compatible",
    "base_url": "http://localhost:7061/v1"  # Required!
  },
  "mcp": {
    "helixagent": {
      "type": "remote",  # Required!
      "url": "http://localhost:7061/v1/mcp"
    }
  }
}

# 3. Regenerate config
./bin/helixagent --generate-agent-config=opencode --agent-config-output=config.json
```

### Issue: "Config file not found"

**Symptoms:**
```
Error: ENOENT: no such file or directory
```

**Solutions:**

```bash
# 1. Create config directory
mkdir -p ~/.config/opencode

# 2. Generate and install config
./bin/helixagent --generate-agent-config=opencode \
  --agent-config-output=~/.config/opencode/opencode.json

# 3. Check permissions
ls -la ~/.config/opencode/
chmod 644 ~/.config/opencode/opencode.json
```

### Issue: "JSON syntax error"

**Symptoms:**
```
SyntaxError: Unexpected token
Error: invalid JSON at position X
```

**Solutions:**

```bash
# 1. Validate JSON syntax
cat config.json | python -m json.tool

# 2. Common issues:
#    - Trailing commas
#    - Missing quotes
#    - Unescaped characters

# 3. Use a JSON linter
npm install -g jsonlint
jsonlint config.json
```

---

## MCP Server Issues

### Issue: "MCP server failed to start"

**Symptoms:**
```
Error: Failed to spawn MCP server
Error: Command not found: helixagent-mcp
```

**Solutions:**

```bash
# 1. Install MCP server globally
npm install -g @helixagent/mcp-server

# 2. Check installation
which helixagent-mcp

# 3. Use npx instead
{
  "mcp": {
    "helixagent": {
      "command": ["npx", "@helixagent/mcp-server"]
    }
  }
}

# 4. Verify PATH
echo $PATH
```

### Issue: "Tool not found"

**Symptoms:**
```
Error: Unknown tool: helix_chat
MCP error: tool not registered
```

**Solutions:**

```bash
# 1. List available tools
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | npx @helixagent/mcp-server

# 2. Check tool name spelling
# Available tools: helix_chat, helix_embeddings, helix_debate, helix_vision, helix_task

# 3. Update MCP server
npm update -g @helixagent/mcp-server
```

### Issue: "MCP communication error"

**Symptoms:**
```
Error: Invalid JSON-RPC response
Error: Parse error in MCP message
```

**Solutions:**

```bash
# 1. Enable MCP debug logging
DEBUG=mcp:* npx @helixagent/mcp-server

# 2. Check message format
# Request must be valid JSON-RPC 2.0

# 3. Test with simple request
echo '{"jsonrpc":"2.0","id":1,"method":"ping"}' | npx @helixagent/mcp-server
```

---

## TOON Protocol Issues

### Issue: "TOON encoding failed"

**Symptoms:**
```
Error: Failed to encode TOON
Error: Invalid TOON format
```

**Solutions:**

```bash
# 1. Disable TOON temporarily
{
  "transport": {
    "enableTOON": false
  }
}

# 2. Check for unsupported data types
# TOON supports: strings, numbers, booleans, arrays, objects
# Not supported: dates, undefined, functions

# 3. Debug encoding
const codec = new TOONCodec({ level: 'standard' });
console.log(codec.encode({ test: 'value' }));
```

### Issue: "Content-Type mismatch"

**Symptoms:**
```
Error: Expected application/toon+json but got application/json
```

**Solutions:**

```bash
# 1. Check Accept header
curl -H "Accept: application/toon+json" http://localhost:7061/v1/chat/completions

# 2. Server may not support TOON - use JSON fallback
{
  "transport": {
    "contentFallback": ["toon", "json"]
  }
}
```

---

## Event Subscription Issues

### Issue: "SSE connection dropped"

**Symptoms:**
```
EventSource connection lost
Error: Connection closed unexpectedly
```

**Solutions:**

```typescript
// 1. Implement reconnection logic
const events = new SSEClient({
  reconnectInterval: 5000,
  maxRetries: 10,
});

// 2. Handle reconnection events
events.on('reconnecting', ({ attempt }) => {
  console.log(`Reconnecting... attempt ${attempt}`);
});

// 3. Use Last-Event-ID for recovery
events.on('connected', () => {
  // Subscription restored from lastEventId
});
```

### Issue: "Missing events"

**Symptoms:**
- Events received out of order
- Some events never arrive

**Solutions:**

```bash
# 1. Check event subscription
{
  "events": {
    "subscriptions": ["task.*", "debate.*"]
  }
}

# 2. Use WebSocket for critical events
{
  "events": {
    "transport": "websocket"
  }
}

# 3. Enable event buffering
{
  "events": {
    "bufferSize": 100
  }
}
```

### Issue: "WebSocket handshake failed"

**Symptoms:**
```
Error: WebSocket handshake failed
Error: Unexpected response code: 400
```

**Solutions:**

```bash
# 1. Check WebSocket endpoint
ws://localhost:7061/v1/ws/tasks/123  # Correct
http://localhost:7061/v1/ws/tasks/123  # Wrong!

# 2. Check for proxy issues
# Proxies may not support WebSocket - use SSE instead

# 3. Verify Origin header is allowed
```

---

## UI Rendering Issues

### Issue: "Garbled output"

**Symptoms:**
- Box drawing characters appear as question marks
- Colors not displaying correctly

**Solutions:**

```bash
# 1. Check terminal encoding
echo $LANG
# Should be: en_US.UTF-8 or similar

# 2. Set UTF-8 encoding
export LANG=en_US.UTF-8
export LC_ALL=en_US.UTF-8

# 3. Use plain render style
{
  "ui": {
    "renderStyle": "plain"
  }
}
```

### Issue: "Progress bar not updating"

**Symptoms:**
- Progress bar stuck
- No animation

**Solutions:**

```typescript
// 1. Ensure terminal supports cursor movement
if (process.stdout.isTTY) {
  // Use animated progress bar
} else {
  // Use simple text output
}

// 2. Flush output
process.stdout.write('\r' + progressBar);

// 3. Check event subscription
events.on('task.progress', (data) => {
  updateProgressBar(data.percent);
});
```

---

## Agent-Specific Issues

### Claude Code

**Issue: "Hooks not executing"**

```bash
# 1. Check hook registration
cat ~/.claude/plugins/helixagent-integration/.claude-plugin/plugin.json

# 2. Verify hook scripts exist
ls -la ~/.claude/plugins/helixagent-integration/hooks/

# 3. Check hook permissions
chmod +x ~/.claude/plugins/helixagent-integration/hooks/*.js

# 4. Enable plugin debugging
CLAUDE_DEBUG=plugins claude
```

### OpenCode

**Issue: "MCP server not loading"**

```bash
# 1. Validate opencode.json
cat ~/.config/opencode/opencode.json | python -m json.tool

# 2. Check MCP server configuration
{
  "mcpServers": {
    "helixagent": {
      "type": "sse",  # or "stdio"
      "url": "http://localhost:7061/v1/mcp"
    }
  }
}

# 3. Test MCP server independently
curl http://localhost:7061/v1/mcp/tools/list
```

### Cline

**Issue: "Hooks timing out"**

```bash
# 1. Increase hook timeout in .clinerules
{
  "hookTimeout": 30000
}

# 2. Make hooks async
module.exports = async function(input) {
  // Hook code
};

# 3. Check for blocking operations
# Avoid synchronous I/O in hooks
```

---

## Debug Commands

### Full Debug Session

```bash
#!/bin/bash

echo "=== HelixAgent Plugin Debug ==="

# 1. Check HelixAgent
echo "1. HelixAgent status:"
curl -s http://localhost:7061/health || echo "NOT RUNNING"

# 2. Check environment
echo -e "\n2. Environment:"
echo "HELIXAGENT_ENDPOINT=$HELIXAGENT_ENDPOINT"
echo "HELIXAGENT_API_KEY=${HELIXAGENT_API_KEY:0:10}..."

# 3. Check config
echo -e "\n3. Configuration:"
if [ -f ~/.config/opencode/opencode.json ]; then
  cat ~/.config/opencode/opencode.json | python -m json.tool 2>/dev/null || echo "INVALID JSON"
else
  echo "Config not found"
fi

# 4. Test MCP
echo -e "\n4. MCP server:"
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | \
  timeout 5 npx @helixagent/mcp-server 2>/dev/null || echo "MCP FAILED"

# 5. Test API
echo -e "\n5. API test:"
curl -s -X POST http://localhost:7061/v1/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $HELIXAGENT_API_KEY" \
  -d '{"model":"helixagent-debate","messages":[{"role":"user","content":"test"}]}' \
  | head -c 200

echo -e "\n\n=== Debug complete ==="
```

### Log Collection

```bash
# Collect all relevant logs
mkdir -p /tmp/helixagent-debug

# HelixAgent logs
journalctl -u helixagent > /tmp/helixagent-debug/helixagent.log

# Plugin logs
cat ~/.claude/logs/* > /tmp/helixagent-debug/claude-code.log 2>/dev/null

# System info
uname -a > /tmp/helixagent-debug/system.txt
node --version >> /tmp/helixagent-debug/system.txt
go version >> /tmp/helixagent-debug/system.txt

# Create archive
tar -czvf helixagent-debug.tar.gz /tmp/helixagent-debug/
```

---

## Getting Help

### Resources

- **Documentation**: [docs.helixagent.dev](https://docs.helixagent.dev)
- **GitHub Issues**: [github.com/helixagent/helixagent/issues](https://github.com/helixagent/helixagent/issues)
- **Discord**: [discord.gg/helixagent](https://discord.gg/helixagent)

### Filing a Bug Report

Include the following:
1. HelixAgent version (`./bin/helixagent --version`)
2. Plugin version
3. CLI agent and version
4. Operating system
5. Error messages (full stack trace)
6. Steps to reproduce
7. Configuration (redact API keys)

```bash
# Generate bug report template
./bin/helixagent --debug-info > bug-report.txt
```
