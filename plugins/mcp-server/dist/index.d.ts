#!/usr/bin/env node
/**
 * HelixAgent Generic MCP Server
 *
 * Provides MCP protocol support for Tier 2-3 CLI agents that don't have
 * rich plugin systems. Supports stdio and SSE transports.
 */
export interface MCPServerConfig {
    endpoint: string;
    transport: 'stdio' | 'sse';
    port?: number;
    preferHTTP3?: boolean;
    enableTOON?: boolean;
    enableBrotli?: boolean;
}
/**
 * Generic MCP Server for HelixAgent
 */
export declare class HelixAgentMCPServer {
    private config;
    private transport;
    private tools;
    constructor(config?: Partial<MCPServerConfig>);
    /**
     * Start the MCP server
     */
    start(): Promise<void>;
    /**
     * Run in stdio mode
     */
    private runStdio;
    /**
     * Run in SSE mode
     */
    private runSSE;
    /**
     * Handle MCP request
     */
    private handleRequest;
    /**
     * Handle initialize
     */
    private handleInitialize;
    /**
     * Handle tools/list
     */
    private handleListTools;
    /**
     * Handle tools/call
     */
    private handleCallTool;
    /**
     * Handle resources/list
     */
    private handleListResources;
    /**
     * Handle resources/read
     */
    private handleReadResource;
}
export interface MCPTool {
    description: string;
    inputSchema: Record<string, unknown>;
    execute(args: Record<string, unknown>): Promise<unknown>;
}
export { HelixAgentTransport } from './transport';
export * from './tools';
//# sourceMappingURL=index.d.ts.map