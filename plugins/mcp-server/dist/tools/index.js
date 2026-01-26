"use strict";
/**
 * HelixAgent MCP Tools
 *
 * Tool implementations for the generic MCP server.
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.ProvidersTool = exports.MemoryTool = exports.RAGTool = exports.TaskTool = exports.EnsembleTool = exports.DebateTool = void 0;
/**
 * AI Debate Tool
 */
class DebateTool {
    transport;
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
    constructor(transport) {
        this.transport = transport;
    }
    async execute(args) {
        const body = {
            topic: args.topic,
            rounds: args.rounds ?? 3,
            enable_multi_pass_validation: args.enable_multi_pass_validation ?? true,
        };
        return this.transport.request('POST', '/v1/debates', body);
    }
}
exports.DebateTool = DebateTool;
/**
 * Ensemble Tool
 */
class EnsembleTool {
    transport;
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
    constructor(transport) {
        this.transport = transport;
    }
    async execute(args) {
        const body = {
            model: 'helix-debate-ensemble',
            messages: [{ role: 'user', content: args.prompt }],
            temperature: args.temperature ?? 0.7,
            max_tokens: args.max_tokens,
        };
        return this.transport.request('POST', '/v1/chat/completions', body);
    }
}
exports.EnsembleTool = EnsembleTool;
/**
 * Background Task Tool
 */
class TaskTool {
    transport;
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
    constructor(transport) {
        this.transport = transport;
    }
    async execute(args) {
        const body = {
            command: args.command,
            description: args.description,
            timeout: args.timeout ?? 300,
            working_dir: args.working_dir,
        };
        return this.transport.request('POST', '/v1/tasks', body);
    }
}
exports.TaskTool = TaskTool;
/**
 * RAG Query Tool
 */
class RAGTool {
    transport;
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
    constructor(transport) {
        this.transport = transport;
    }
    async execute(args) {
        const body = {
            query: args.query,
            collection: args.collection,
            top_k: args.top_k ?? 5,
            rerank: args.rerank ?? true,
        };
        return this.transport.request('POST', '/v1/rag/query', body);
    }
}
exports.RAGTool = RAGTool;
/**
 * Memory Tool
 */
class MemoryTool {
    transport;
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
    constructor(transport) {
        this.transport = transport;
    }
    async execute(args) {
        const action = args.action;
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
exports.MemoryTool = MemoryTool;
/**
 * Providers Info Tool
 */
class ProvidersTool {
    transport;
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
    constructor(transport) {
        this.transport = transport;
    }
    async execute(args) {
        if (args.provider) {
            return this.transport.request('GET', `/v1/providers/${args.provider}`);
        }
        const params = args.include_models ? '?include_models=true' : '';
        return this.transport.request('GET', `/v1/providers${params}`);
    }
}
exports.ProvidersTool = ProvidersTool;
//# sourceMappingURL=index.js.map