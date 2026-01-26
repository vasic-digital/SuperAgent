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
export declare class DebateTool implements MCPTool {
    private transport;
    description: string;
    inputSchema: {
        type: string;
        properties: {
            topic: {
                type: string;
                description: string;
            };
            rounds: {
                type: string;
                description: string;
                default: number;
            };
            enable_multi_pass_validation: {
                type: string;
                description: string;
                default: boolean;
            };
        };
        required: string[];
    };
    constructor(transport: HelixAgentTransport);
    execute(args: Record<string, unknown>): Promise<unknown>;
}
/**
 * Ensemble Tool
 */
export declare class EnsembleTool implements MCPTool {
    private transport;
    description: string;
    inputSchema: {
        type: string;
        properties: {
            prompt: {
                type: string;
                description: string;
            };
            temperature: {
                type: string;
                description: string;
                default: number;
            };
            max_tokens: {
                type: string;
                description: string;
            };
        };
        required: string[];
    };
    constructor(transport: HelixAgentTransport);
    execute(args: Record<string, unknown>): Promise<unknown>;
}
/**
 * Background Task Tool
 */
export declare class TaskTool implements MCPTool {
    private transport;
    description: string;
    inputSchema: {
        type: string;
        properties: {
            command: {
                type: string;
                description: string;
            };
            description: {
                type: string;
                description: string;
            };
            timeout: {
                type: string;
                description: string;
                default: number;
            };
            working_dir: {
                type: string;
                description: string;
            };
        };
        required: string[];
    };
    constructor(transport: HelixAgentTransport);
    execute(args: Record<string, unknown>): Promise<unknown>;
}
/**
 * RAG Query Tool
 */
export declare class RAGTool implements MCPTool {
    private transport;
    description: string;
    inputSchema: {
        type: string;
        properties: {
            query: {
                type: string;
                description: string;
            };
            collection: {
                type: string;
                description: string;
            };
            top_k: {
                type: string;
                description: string;
                default: number;
            };
            rerank: {
                type: string;
                description: string;
                default: boolean;
            };
        };
        required: string[];
    };
    constructor(transport: HelixAgentTransport);
    execute(args: Record<string, unknown>): Promise<unknown>;
}
/**
 * Memory Tool
 */
export declare class MemoryTool implements MCPTool {
    private transport;
    description: string;
    inputSchema: {
        type: string;
        properties: {
            action: {
                type: string;
                enum: string[];
                description: string;
            };
            content: {
                type: string;
                description: string;
            };
            query: {
                type: string;
                description: string;
            };
            memory_id: {
                type: string;
                description: string;
            };
            metadata: {
                type: string;
                description: string;
            };
            limit: {
                type: string;
                description: string;
                default: number;
            };
        };
        required: string[];
    };
    constructor(transport: HelixAgentTransport);
    execute(args: Record<string, unknown>): Promise<unknown>;
}
/**
 * Providers Info Tool
 */
export declare class ProvidersTool implements MCPTool {
    private transport;
    description: string;
    inputSchema: {
        type: string;
        properties: {
            provider: {
                type: string;
                description: string;
            };
            include_models: {
                type: string;
                description: string;
                default: boolean;
            };
        };
    };
    constructor(transport: HelixAgentTransport);
    execute(args: Record<string, unknown>): Promise<unknown>;
}
//# sourceMappingURL=index.d.ts.map