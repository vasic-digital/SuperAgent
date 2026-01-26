#!/usr/bin/env node
/**
 * HelixAgent Generic MCP Server
 *
 * Provides MCP protocol support for Tier 2-3 CLI agents that don't have
 * rich plugin systems. Supports stdio and SSE transports.
 */

import { createReadStream, createWriteStream } from 'fs';
import { HelixAgentTransport } from './transport';
import { DebateTool, EnsembleTool, TaskTool, RAGTool, MemoryTool, ProvidersTool, ACPTool, LSPTool, EmbeddingsTool, VisionTool, CogneeTool } from './tools';

// Configuration
export interface MCPServerConfig {
  endpoint: string;
  transport: 'stdio' | 'sse';
  port?: number;
  preferHTTP3?: boolean;
  enableTOON?: boolean;
  enableBrotli?: boolean;
}

const defaultConfig: MCPServerConfig = {
  endpoint: 'https://localhost:7061',
  transport: 'stdio',
  preferHTTP3: true,
  enableTOON: true,
  enableBrotli: true,
};

// MCP Protocol types
interface MCPRequest {
  jsonrpc: '2.0';
  id: string | number;
  method: string;
  params?: unknown;
}

interface MCPResponse {
  jsonrpc: '2.0';
  id: string | number;
  result?: unknown;
  error?: {
    code: number;
    message: string;
    data?: unknown;
  };
}

/**
 * Generic MCP Server for HelixAgent
 */
export class HelixAgentMCPServer {
  private config: MCPServerConfig;
  private transport: HelixAgentTransport;
  private tools: Map<string, MCPTool>;

  constructor(config?: Partial<MCPServerConfig>) {
    this.config = { ...defaultConfig, ...config };
    this.transport = new HelixAgentTransport(this.config.endpoint, {
      preferHTTP3: this.config.preferHTTP3,
      enableTOON: this.config.enableTOON,
      enableBrotli: this.config.enableBrotli,
    });

    // Register tools - Core functionality
    this.tools = new Map();
    this.tools.set('helixagent_debate', new DebateTool(this.transport));
    this.tools.set('helixagent_ensemble', new EnsembleTool(this.transport));
    this.tools.set('helixagent_task', new TaskTool(this.transport));
    this.tools.set('helixagent_rag', new RAGTool(this.transport));
    this.tools.set('helixagent_memory', new MemoryTool(this.transport));
    this.tools.set('helixagent_providers', new ProvidersTool(this.transport));

    // Register tools - Protocol tools (ACP, LSP, Embeddings, Vision, Cognee)
    this.tools.set('helixagent_acp', new ACPTool(this.transport));
    this.tools.set('helixagent_lsp', new LSPTool(this.transport));
    this.tools.set('helixagent_embeddings', new EmbeddingsTool(this.transport));
    this.tools.set('helixagent_vision', new VisionTool(this.transport));
    this.tools.set('helixagent_cognee', new CogneeTool(this.transport));
  }

  /**
   * Start the MCP server
   */
  async start(): Promise<void> {
    // Connect to HelixAgent
    await this.transport.connect();

    if (this.config.transport === 'stdio') {
      await this.runStdio();
    } else {
      await this.runSSE();
    }
  }

  /**
   * Run in stdio mode
   */
  private async runStdio(): Promise<void> {
    const readline = await import('readline');
    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout,
      terminal: false,
    });

    rl.on('line', async (line) => {
      if (!line.trim()) return;

      try {
        const request = JSON.parse(line) as MCPRequest;
        const response = await this.handleRequest(request);
        if (response) {
          process.stdout.write(JSON.stringify(response) + '\n');
        }
      } catch (error) {
        const errorResponse: MCPResponse = {
          jsonrpc: '2.0',
          id: 0,
          error: {
            code: -32700,
            message: 'Parse error',
          },
        };
        process.stdout.write(JSON.stringify(errorResponse) + '\n');
      }
    });

    rl.on('close', () => {
      process.exit(0);
    });
  }

  /**
   * Run in SSE mode
   */
  private async runSSE(): Promise<void> {
    const http = await import('http');
    const port = this.config.port || 7062;

    const server = http.createServer(async (req, res) => {
      if (req.method === 'OPTIONS') {
        res.writeHead(204, {
          'Access-Control-Allow-Origin': '*',
          'Access-Control-Allow-Methods': 'GET, POST, OPTIONS',
          'Access-Control-Allow-Headers': 'Content-Type',
        });
        res.end();
        return;
      }

      if (req.url === '/health') {
        res.writeHead(200, { 'Content-Type': 'application/json' });
        res.end(JSON.stringify({ status: 'healthy' }));
        return;
      }

      if (req.method === 'POST' && req.url === '/mcp') {
        let body = '';
        req.on('data', (chunk) => { body += chunk; });
        req.on('end', async () => {
          try {
            const request = JSON.parse(body) as MCPRequest;
            const response = await this.handleRequest(request);

            res.writeHead(200, {
              'Content-Type': 'application/json',
              'Access-Control-Allow-Origin': '*',
            });
            res.end(JSON.stringify(response));
          } catch (error) {
            res.writeHead(400, { 'Content-Type': 'application/json' });
            res.end(JSON.stringify({
              jsonrpc: '2.0',
              id: 0,
              error: { code: -32700, message: 'Parse error' },
            }));
          }
        });
        return;
      }

      res.writeHead(404);
      res.end('Not found');
    });

    server.listen(port, () => {
      console.error(`HelixAgent MCP Server running on http://localhost:${port}`);
    });
  }

  /**
   * Handle MCP request
   */
  private async handleRequest(request: MCPRequest): Promise<MCPResponse | null> {
    switch (request.method) {
      case 'initialize':
        return this.handleInitialize(request);

      case 'notifications/initialized':
        return null; // No response for notifications

      case 'tools/list':
        return this.handleListTools(request);

      case 'tools/call':
        return this.handleCallTool(request);

      case 'resources/list':
        return this.handleListResources(request);

      case 'resources/read':
        return this.handleReadResource(request);

      default:
        return {
          jsonrpc: '2.0',
          id: request.id,
          error: {
            code: -32601,
            message: `Method not found: ${request.method}`,
          },
        };
    }
  }

  /**
   * Handle initialize
   */
  private handleInitialize(request: MCPRequest): MCPResponse {
    return {
      jsonrpc: '2.0',
      id: request.id,
      result: {
        protocolVersion: '2024-11-05',
        capabilities: {
          tools: { listChanged: true },
          resources: { subscribe: false, listChanged: false },
        },
        serverInfo: {
          name: 'helixagent-mcp-server',
          version: '1.0.0',
        },
      },
    };
  }

  /**
   * Handle tools/list
   */
  private handleListTools(request: MCPRequest): MCPResponse {
    const tools = Array.from(this.tools.entries()).map(([name, tool]) => ({
      name,
      description: tool.description,
      inputSchema: tool.inputSchema,
    }));

    return {
      jsonrpc: '2.0',
      id: request.id,
      result: { tools },
    };
  }

  /**
   * Handle tools/call
   */
  private async handleCallTool(request: MCPRequest): Promise<MCPResponse> {
    const params = request.params as { name: string; arguments?: Record<string, unknown> };
    const toolName = params?.name;
    const toolArgs = params?.arguments || {};

    const tool = this.tools.get(toolName);
    if (!tool) {
      return {
        jsonrpc: '2.0',
        id: request.id,
        error: {
          code: -32602,
          message: `Unknown tool: ${toolName}`,
        },
      };
    }

    try {
      const result = await tool.execute(toolArgs);
      return {
        jsonrpc: '2.0',
        id: request.id,
        result: {
          content: [
            {
              type: 'text',
              text: typeof result === 'string' ? result : JSON.stringify(result, null, 2),
            },
          ],
        },
      };
    } catch (error) {
      return {
        jsonrpc: '2.0',
        id: request.id,
        error: {
          code: -32000,
          message: error instanceof Error ? error.message : 'Tool execution failed',
        },
      };
    }
  }

  /**
   * Handle resources/list
   */
  private handleListResources(request: MCPRequest): MCPResponse {
    return {
      jsonrpc: '2.0',
      id: request.id,
      result: {
        resources: [
          {
            uri: 'helixagent://providers',
            name: 'HelixAgent Providers',
            description: 'List of available LLM providers and their status',
            mimeType: 'application/json',
          },
          {
            uri: 'helixagent://debates',
            name: 'Active Debates',
            description: 'List of currently active AI debates',
            mimeType: 'application/json',
          },
          {
            uri: 'helixagent://tasks',
            name: 'Background Tasks',
            description: 'List of background tasks and their status',
            mimeType: 'application/json',
          },
        ],
      },
    };
  }

  /**
   * Handle resources/read
   */
  private async handleReadResource(request: MCPRequest): Promise<MCPResponse> {
    const params = request.params as { uri: string };
    const uri = params?.uri;

    try {
      let content: unknown;

      switch (uri) {
        case 'helixagent://providers':
          content = await this.transport.request('GET', '/v1/providers');
          break;
        case 'helixagent://debates':
          content = await this.transport.request('GET', '/v1/debates');
          break;
        case 'helixagent://tasks':
          content = await this.transport.request('GET', '/v1/tasks');
          break;
        default:
          return {
            jsonrpc: '2.0',
            id: request.id,
            error: {
              code: -32602,
              message: `Unknown resource: ${uri}`,
            },
          };
      }

      return {
        jsonrpc: '2.0',
        id: request.id,
        result: {
          contents: [
            {
              uri,
              mimeType: 'application/json',
              text: JSON.stringify(content, null, 2),
            },
          ],
        },
      };
    } catch (error) {
      return {
        jsonrpc: '2.0',
        id: request.id,
        error: {
          code: -32000,
          message: error instanceof Error ? error.message : 'Resource read failed',
        },
      };
    }
  }
}

// Tool interface
export interface MCPTool {
  description: string;
  inputSchema: Record<string, unknown>;
  execute(args: Record<string, unknown>): Promise<unknown>;
}

// CLI entry point
async function main() {
  const args = process.argv.slice(2);
  const config: Partial<MCPServerConfig> = {};

  for (let i = 0; i < args.length; i++) {
    switch (args[i]) {
      case '--endpoint':
        config.endpoint = args[++i];
        break;
      case '--port':
        config.port = parseInt(args[++i], 10);
        break;
      case '--transport':
        config.transport = args[++i] as 'stdio' | 'sse';
        break;
      case '--no-http3':
        config.preferHTTP3 = false;
        break;
      case '--no-toon':
        config.enableTOON = false;
        break;
      case '--no-brotli':
        config.enableBrotli = false;
        break;
      case '--help':
        console.log(`
HelixAgent MCP Server

Usage: helixagent-mcp [options]

Options:
  --endpoint <url>     HelixAgent endpoint (default: https://localhost:7061)
  --port <port>        SSE server port (default: 7062)
  --transport <type>   Transport type: stdio or sse (default: stdio)
  --no-http3           Disable HTTP/3 preference
  --no-toon            Disable TOON protocol
  --no-brotli          Disable Brotli compression
  --help               Show this help message
        `);
        process.exit(0);
    }
  }

  const server = new HelixAgentMCPServer(config);
  await server.start();
}

// Export for programmatic use
export { HelixAgentTransport } from './transport';
export * from './tools';

// Run if invoked directly
if (require.main === module) {
  main().catch((error) => {
    console.error('MCP Server error:', error);
    process.exit(1);
  });
}
