# HelixAgent CLI SDK

Node.js command-line interface for the HelixAgent Protocol Enhancement API, providing access to MCP, LSP, ACP, analytics, plugins, templates, and system management commands.

## Platform Requirements

- Node.js 14+
- No external dependencies (uses built-in `http`/`https` modules)

## Installation

```bash
# Make executable
chmod +x helixagent-cli.js

# Optionally link globally
npm link
# or
ln -s $(pwd)/helixagent-cli.js /usr/local/bin/helixagent-cli
```

## Quick Start

```bash
# Set environment
export HELIXAGENT_URL=http://localhost:8080
export HELIXAGENT_API_KEY=your-api-key

# Check health
helixagent-cli health

# List MCP tools
helixagent-cli mcp:tools

# Get code completions
helixagent-cli lsp:completion main.go 10 5
```

## Commands

### MCP Protocol

```bash
helixagent-cli mcp:tools [server_id]           # List MCP tools
helixagent-cli mcp:call <server_id> <tool>      # Call MCP tool
helixagent-cli mcp:servers                      # List MCP servers
```

### LSP Protocol

```bash
helixagent-cli lsp:completion <file> <line> <char>   # Get code completions
helixagent-cli lsp:hover <file> <line> <char>        # Get hover information
helixagent-cli lsp:definition <file> <line> <char>   # Get definition location
helixagent-cli lsp:diagnostics <file>                # Get file diagnostics
```

### ACP Protocol

```bash
helixagent-cli acp:execute <action> [agent_id]       # Execute agent action
helixagent-cli acp:broadcast <message> <targets...>  # Broadcast to agents
helixagent-cli acp:status [agent_id]                 # Get agent status
```

### Analytics

```bash
helixagent-cli analytics                      # Get all analytics
helixagent-cli analytics:protocol <protocol>  # Protocol-specific analytics
helixagent-cli analytics:health               # System health status
```

### Plugins

```bash
helixagent-cli plugins                                    # List loaded plugins
helixagent-cli plugins:load <path>                        # Load plugin
helixagent-cli plugins:unload <plugin_id>                 # Unload plugin
helixagent-cli plugins:execute <id> <operation> [params]  # Execute operation
helixagent-cli plugins:marketplace [query] [protocol]     # Search marketplace
```

### Templates

```bash
helixagent-cli templates [protocol]              # List templates
helixagent-cli templates:get <template_id>       # Get template details
helixagent-cli templates:generate <id> [config]  # Generate from template
```

### System

```bash
helixagent-cli health    # System health check
helixagent-cli status    # System status
helixagent-cli metrics   # Prometheus metrics
```

## API Reference (Programmatic)

The CLI can also be used as a library:

```javascript
const { HelixAgentCLI } = require('./helixagent-cli');

const cli = new HelixAgentCLI();
cli.baseURL = 'http://localhost:8080';
cli.apiKey = 'your-key';

// MCP
const tools = await cli.mcpListTools();
const result = await cli.mcpCallTool('server-1', 'read_file', { path: '/tmp/test' });

// LSP
const completions = await cli.lspCompletion('main.go', 10, 5);
const hover = await cli.lspHover('main.go', 10, 5);
const definition = await cli.lspDefinition('main.go', 10, 5);
const diagnostics = await cli.lspDiagnostics('main.go');

// ACP
const execResult = await cli.acpExecute('process_data', 'agent-001');
const broadcast = await cli.acpBroadcast('hello', ['agent-001', 'agent-002']);
const status = await cli.acpStatus('agent-001');

// Analytics
const metrics = await cli.analytics();
const health = await cli.analyticsHealth();

// Plugins
const plugins = await cli.plugins();
const loaded = await cli.pluginLoad('/path/to/plugin');
const execPlugin = await cli.pluginExecute('plugin-1', 'transform', { data: '...' });

// System
const systemHealth = await cli.health();
const systemStatus = await cli.status();
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `HELIXAGENT_URL` | `http://localhost:8080` | HelixAgent server URL |
| `HELIXAGENT_API_KEY` | `null` | API key for authentication |

### Client Options

| Property | Default | Description |
|----------|---------|-------------|
| `baseURL` | from env or `http://localhost:8080` | Server URL |
| `apiKey` | from env or `null` | Bearer token |
| `timeout` | `30000` | Request timeout in milliseconds |

## Error Handling

CLI exits with code 1 on errors and prints the error message to stderr:

```bash
$ helixagent-cli mcp:call
Error: Usage: mcp:call <server_id> <tool>

$ helixagent-cli unknown-command
Error: Unknown command: unknown-command
```

Programmatic usage throws errors that can be caught:

```javascript
try {
    const result = await cli.mcpCallTool('server-1', 'tool-1');
} catch (error) {
    console.error(`Failed: ${error.message}`);
}
```

## Output Format

All commands output JSON to stdout, formatted with 2-space indentation:

```bash
$ helixagent-cli health
{
  "status": "healthy",
  "uptime": 3600,
  "version": "1.0.0"
}
```

## Troubleshooting

- **Connection refused**: Verify `HELIXAGENT_URL` points to a running server
- **Request timeout**: Increase timeout or check server responsiveness
- **Authentication errors**: Verify `HELIXAGENT_API_KEY` is set correctly
- **Parse errors**: Server may be returning non-JSON responses; check server logs

## License

Proprietary.
