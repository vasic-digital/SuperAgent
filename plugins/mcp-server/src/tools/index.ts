/**
 * HelixAgent MCP Tools
 *
 * Tool implementations for the generic MCP server.
 */

import { HelixAgentTransport } from '../transport';
import type { MCPTool } from '../index';

/**
 * AI Debate Tool
 */
export class DebateTool implements MCPTool {
  description = 'Start an AI debate with 15 LLMs (5 positions x 3 LLMs each) to reach consensus on a topic';

  inputSchema = {
    type: 'object',
    properties: {
      topic: {
        type: 'string',
        description: 'The topic or question to debate',
      },
      rounds: {
        type: 'integer',
        description: 'Number of debate rounds (default: 3)',
        default: 3,
      },
      enable_multi_pass_validation: {
        type: 'boolean',
        description: 'Enable multi-pass validation for higher quality responses',
        default: true,
      },
    },
    required: ['topic'],
  };

  constructor(private transport: HelixAgentTransport) {}

  async execute(args: Record<string, unknown>): Promise<unknown> {
    const body = {
      topic: args.topic,
      rounds: args.rounds ?? 3,
      enable_multi_pass_validation: args.enable_multi_pass_validation ?? true,
    };

    return this.transport.request('POST', '/v1/debates', body);
  }
}

/**
 * Ensemble Tool
 */
export class EnsembleTool implements MCPTool {
  description = 'Get a response from the AI Debate Ensemble (single query, confidence-weighted voting)';

  inputSchema = {
    type: 'object',
    properties: {
      prompt: {
        type: 'string',
        description: 'The prompt to send to the ensemble',
      },
      temperature: {
        type: 'number',
        description: 'Sampling temperature (0-1)',
        default: 0.7,
      },
      max_tokens: {
        type: 'integer',
        description: 'Maximum tokens to generate',
      },
    },
    required: ['prompt'],
  };

  constructor(private transport: HelixAgentTransport) {}

  async execute(args: Record<string, unknown>): Promise<unknown> {
    const body = {
      model: 'helix-debate-ensemble',
      messages: [{ role: 'user', content: args.prompt }],
      temperature: args.temperature ?? 0.7,
      max_tokens: args.max_tokens,
    };

    return this.transport.request('POST', '/v1/chat/completions', body);
  }
}

/**
 * Background Task Tool
 */
export class TaskTool implements MCPTool {
  description = 'Create a background task for long-running operations';

  inputSchema = {
    type: 'object',
    properties: {
      command: {
        type: 'string',
        description: 'The command to execute',
      },
      description: {
        type: 'string',
        description: 'Task description',
      },
      timeout: {
        type: 'integer',
        description: 'Timeout in seconds (default: 300)',
        default: 300,
      },
      working_dir: {
        type: 'string',
        description: 'Working directory for command execution',
      },
    },
    required: ['command'],
  };

  constructor(private transport: HelixAgentTransport) {}

  async execute(args: Record<string, unknown>): Promise<unknown> {
    const body = {
      command: args.command,
      description: args.description,
      timeout: args.timeout ?? 300,
      working_dir: args.working_dir,
    };

    return this.transport.request('POST', '/v1/tasks', body);
  }
}

/**
 * RAG Query Tool
 */
export class RAGTool implements MCPTool {
  description = 'Perform a hybrid RAG query (dense + sparse retrieval with reranking)';

  inputSchema = {
    type: 'object',
    properties: {
      query: {
        type: 'string',
        description: 'The query to search for',
      },
      collection: {
        type: 'string',
        description: 'The collection to search in',
      },
      top_k: {
        type: 'integer',
        description: 'Number of results to return (default: 5)',
        default: 5,
      },
      rerank: {
        type: 'boolean',
        description: 'Enable reranking of results',
        default: true,
      },
    },
    required: ['query'],
  };

  constructor(private transport: HelixAgentTransport) {}

  async execute(args: Record<string, unknown>): Promise<unknown> {
    const body = {
      query: args.query,
      collection: args.collection,
      top_k: args.top_k ?? 5,
      rerank: args.rerank ?? true,
    };

    return this.transport.request('POST', '/v1/rag/query', body);
  }
}

/**
 * Memory Tool
 */
export class MemoryTool implements MCPTool {
  description = 'Access the Mem0-style memory system for storing and retrieving memories';

  inputSchema = {
    type: 'object',
    properties: {
      action: {
        type: 'string',
        enum: ['add', 'search', 'get', 'delete'],
        description: 'The memory action to perform',
      },
      content: {
        type: 'string',
        description: 'Memory content (for add action)',
      },
      query: {
        type: 'string',
        description: 'Search query (for search action)',
      },
      memory_id: {
        type: 'string',
        description: 'Memory ID (for get/delete actions)',
      },
      metadata: {
        type: 'object',
        description: 'Additional metadata for the memory',
      },
      limit: {
        type: 'integer',
        description: 'Maximum number of results (for search action)',
        default: 10,
      },
    },
    required: ['action'],
  };

  constructor(private transport: HelixAgentTransport) {}

  async execute(args: Record<string, unknown>): Promise<unknown> {
    const action = args.action as string;

    switch (action) {
      case 'add':
        return this.transport.request('POST', '/v1/memory', {
          content: args.content,
          metadata: args.metadata,
        });

      case 'search':
        return this.transport.request('POST', '/v1/memory/search', {
          query: args.query,
          limit: args.limit ?? 10,
        });

      case 'get':
        return this.transport.request('GET', `/v1/memory/${args.memory_id}`);

      case 'delete':
        return this.transport.request('DELETE', `/v1/memory/${args.memory_id}`);

      default:
        throw new Error(`Unknown memory action: ${action}`);
    }
  }
}

/**
 * Providers Info Tool
 */
export class ProvidersTool implements MCPTool {
  description = 'Get information about available LLM providers and their verification status';

  inputSchema = {
    type: 'object',
    properties: {
      provider: {
        type: 'string',
        description: 'Specific provider to get info for (optional)',
      },
      include_models: {
        type: 'boolean',
        description: 'Include available models for each provider',
        default: false,
      },
    },
  };

  constructor(private transport: HelixAgentTransport) {}

  async execute(args: Record<string, unknown>): Promise<unknown> {
    if (args.provider) {
      return this.transport.request('GET', `/v1/providers/${args.provider}`);
    }

    const params = args.include_models ? '?include_models=true' : '';
    return this.transport.request('GET', `/v1/providers${params}`);
  }
}

/**
 * ACP (Agent Communication Protocol) Tool
 */
export class ACPTool implements MCPTool {
  description = 'Interact with the Agent Communication Protocol - send messages to agents or list available agents';

  inputSchema = {
    type: 'object',
    properties: {
      action: {
        type: 'string',
        enum: ['list_agents', 'send_message', 'get_capabilities'],
        description: 'The ACP action to perform',
      },
      agent_id: {
        type: 'string',
        description: 'Target agent ID (for send_message)',
      },
      message: {
        type: 'string',
        description: 'Message content (for send_message)',
      },
    },
    required: ['action'],
  };

  constructor(private transport: HelixAgentTransport) {}

  async execute(args: Record<string, unknown>): Promise<unknown> {
    const action = args.action as string;

    switch (action) {
      case 'list_agents':
        return this.transport.request('POST', '/v1/acp', {
          jsonrpc: '2.0',
          id: 1,
          method: 'tools/call',
          params: { name: 'acp_list_agents', arguments: {} },
        });

      case 'send_message':
        return this.transport.request('POST', '/v1/acp', {
          jsonrpc: '2.0',
          id: 1,
          method: 'tools/call',
          params: {
            name: 'acp_send_message',
            arguments: {
              agent_id: args.agent_id,
              message: args.message,
            },
          },
        });

      case 'get_capabilities':
        return this.transport.request('POST', '/v1/acp', {
          jsonrpc: '2.0',
          id: 1,
          method: 'initialize',
          params: {
            protocolVersion: '2024-11-05',
            capabilities: {},
            clientInfo: { name: 'helixagent-mcp', version: '1.0.0' },
          },
        });

      default:
        throw new Error(`Unknown ACP action: ${action}`);
    }
  }
}

/**
 * LSP (Language Server Protocol) Tool
 */
export class LSPTool implements MCPTool {
  description = 'Interact with Language Server Protocol - get diagnostics, find definitions, and more';

  inputSchema = {
    type: 'object',
    properties: {
      action: {
        type: 'string',
        enum: ['list_servers', 'get_diagnostics', 'go_to_definition', 'find_references'],
        description: 'The LSP action to perform',
      },
      file_path: {
        type: 'string',
        description: 'Path to the file (required for diagnostics/definition/references)',
      },
      line: {
        type: 'integer',
        description: 'Line number (for go_to_definition/find_references)',
      },
      character: {
        type: 'integer',
        description: 'Character position (for go_to_definition/find_references)',
      },
    },
    required: ['action'],
  };

  constructor(private transport: HelixAgentTransport) {}

  async execute(args: Record<string, unknown>): Promise<unknown> {
    const action = args.action as string;

    switch (action) {
      case 'list_servers':
        return this.transport.request('POST', '/v1/lsp', {
          jsonrpc: '2.0',
          id: 1,
          method: 'tools/call',
          params: { name: 'lsp_list_servers', arguments: {} },
        });

      case 'get_diagnostics':
        return this.transport.request('POST', '/v1/lsp', {
          jsonrpc: '2.0',
          id: 1,
          method: 'tools/call',
          params: {
            name: 'lsp_get_diagnostics',
            arguments: { file_path: args.file_path },
          },
        });

      case 'go_to_definition':
        return this.transport.request('POST', '/v1/lsp', {
          jsonrpc: '2.0',
          id: 1,
          method: 'tools/call',
          params: {
            name: 'lsp_go_to_definition',
            arguments: {
              file_path: args.file_path,
              line: args.line,
              character: args.character,
            },
          },
        });

      case 'find_references':
        return this.transport.request('POST', '/v1/lsp', {
          jsonrpc: '2.0',
          id: 1,
          method: 'tools/call',
          params: {
            name: 'lsp_find_references',
            arguments: {
              file_path: args.file_path,
              line: args.line,
              character: args.character,
            },
          },
        });

      default:
        throw new Error(`Unknown LSP action: ${action}`);
    }
  }
}

/**
 * Embeddings Tool
 */
export class EmbeddingsTool implements MCPTool {
  description = 'Generate and search vector embeddings for semantic search and similarity matching';

  inputSchema = {
    type: 'object',
    properties: {
      action: {
        type: 'string',
        enum: ['generate', 'search', 'list_providers', 'get_stats'],
        description: 'The embeddings action to perform',
      },
      text: {
        type: 'string',
        description: 'Text to embed (for generate action)',
      },
      texts: {
        type: 'array',
        items: { type: 'string' },
        description: 'Array of texts to embed (for batch generate)',
      },
      query: {
        type: 'string',
        description: 'Search query (for search action)',
      },
      model: {
        type: 'string',
        description: 'Embedding model to use (e.g., text-embedding-3-small)',
      },
      top_k: {
        type: 'integer',
        description: 'Number of results to return (for search)',
        default: 5,
      },
    },
    required: ['action'],
  };

  constructor(private transport: HelixAgentTransport) {}

  async execute(args: Record<string, unknown>): Promise<unknown> {
    const action = args.action as string;

    switch (action) {
      case 'generate':
        return this.transport.request('POST', '/v1/embeddings/generate', {
          input: args.texts ?? [args.text],
          model: args.model ?? 'text-embedding-3-small',
        });

      case 'search':
        return this.transport.request('POST', '/v1/embeddings/search', {
          query: args.query,
          top_k: args.top_k ?? 5,
        });

      case 'list_providers':
        return this.transport.request('GET', '/v1/embeddings/providers');

      case 'get_stats':
        return this.transport.request('GET', '/v1/embeddings/stats');

      default:
        throw new Error(`Unknown embeddings action: ${action}`);
    }
  }
}

/**
 * Vision Tool
 */
export class VisionTool implements MCPTool {
  description = 'Analyze images using AI vision capabilities - extract text (OCR), describe content, and more';

  inputSchema = {
    type: 'object',
    properties: {
      action: {
        type: 'string',
        enum: ['analyze', 'ocr', 'describe'],
        description: 'The vision action to perform',
      },
      image_url: {
        type: 'string',
        description: 'URL of the image to analyze',
      },
      image_base64: {
        type: 'string',
        description: 'Base64-encoded image data (alternative to URL)',
      },
      prompt: {
        type: 'string',
        description: 'Custom prompt for image analysis',
      },
    },
    required: ['action'],
  };

  constructor(private transport: HelixAgentTransport) {}

  async execute(args: Record<string, unknown>): Promise<unknown> {
    const action = args.action as string;

    switch (action) {
      case 'analyze':
        return this.transport.request('POST', '/v1/vision', {
          jsonrpc: '2.0',
          id: 1,
          method: 'tools/call',
          params: {
            name: 'vision_analyze_image',
            arguments: {
              image_url: args.image_url ?? args.image_base64,
              prompt: args.prompt ?? 'Analyze this image and describe what you see',
            },
          },
        });

      case 'ocr':
        return this.transport.request('POST', '/v1/vision', {
          jsonrpc: '2.0',
          id: 1,
          method: 'tools/call',
          params: {
            name: 'vision_ocr',
            arguments: { image_url: args.image_url ?? args.image_base64 },
          },
        });

      case 'describe':
        return this.transport.request('POST', '/v1/vision', {
          jsonrpc: '2.0',
          id: 1,
          method: 'tools/call',
          params: {
            name: 'vision_analyze_image',
            arguments: {
              image_url: args.image_url ?? args.image_base64,
              prompt: 'Provide a detailed description of this image',
            },
          },
        });

      default:
        throw new Error(`Unknown vision action: ${action}`);
    }
  }
}

/**
 * Cognee (Knowledge Graph) Tool
 */
export class CogneeTool implements MCPTool {
  description = 'Interact with Cognee knowledge graph - add content, search, and visualize knowledge relationships';

  inputSchema = {
    type: 'object',
    properties: {
      action: {
        type: 'string',
        enum: ['add', 'search', 'cognify', 'visualize', 'get_status', 'list_datasets'],
        description: 'The Cognee action to perform',
      },
      content: {
        type: 'string',
        description: 'Content to add to the knowledge graph',
      },
      query: {
        type: 'string',
        description: 'Search query',
      },
      search_type: {
        type: 'string',
        enum: ['insights', 'summaries', 'chunks'],
        description: 'Type of search results to return',
        default: 'insights',
      },
      dataset: {
        type: 'string',
        description: 'Dataset name',
      },
    },
    required: ['action'],
  };

  constructor(private transport: HelixAgentTransport) {}

  async execute(args: Record<string, unknown>): Promise<unknown> {
    const action = args.action as string;

    switch (action) {
      case 'add':
        return this.transport.request('POST', '/v1/cognee/add', {
          content: args.content,
          dataset: args.dataset ?? 'default',
        });

      case 'search':
        return this.transport.request('POST', '/v1/cognee/search', {
          query: args.query,
          search_type: args.search_type ?? 'insights',
        });

      case 'cognify':
        return this.transport.request('POST', '/v1/cognee/cognify', {
          dataset: args.dataset ?? 'default',
        });

      case 'visualize':
        return this.transport.request('GET', '/v1/cognee/visualize');

      case 'get_status':
        return this.transport.request('GET', '/v1/cognee/status');

      case 'list_datasets':
        return this.transport.request('GET', '/v1/cognee/datasets');

      default:
        throw new Error(`Unknown Cognee action: ${action}`);
    }
  }
}
