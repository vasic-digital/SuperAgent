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
/**
 * ACP (Agent Communication Protocol) Tool
 */
export declare class ACPTool implements MCPTool {
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
            agent_id: {
                type: string;
                description: string;
            };
            message: {
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
 * LSP (Language Server Protocol) Tool
 */
export declare class LSPTool implements MCPTool {
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
            file_path: {
                type: string;
                description: string;
            };
            line: {
                type: string;
                description: string;
            };
            character: {
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
 * Embeddings Tool
 */
export declare class EmbeddingsTool implements MCPTool {
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
            text: {
                type: string;
                description: string;
            };
            texts: {
                type: string;
                items: {
                    type: string;
                };
                description: string;
            };
            query: {
                type: string;
                description: string;
            };
            model: {
                type: string;
                description: string;
            };
            top_k: {
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
 * Vision Tool
 */
export declare class VisionTool implements MCPTool {
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
            image_url: {
                type: string;
                description: string;
            };
            image_base64: {
                type: string;
                description: string;
            };
            prompt: {
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
 * Cognee (Knowledge Graph) Tool
 */
export declare class CogneeTool implements MCPTool {
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
            search_type: {
                type: string;
                enum: string[];
                description: string;
                default: string;
            };
            dataset: {
                type: string;
                description: string;
            };
        };
        required: string[];
    };
    constructor(transport: HelixAgentTransport);
    execute(args: Record<string, unknown>): Promise<unknown>;
}
//# sourceMappingURL=index.d.ts.map