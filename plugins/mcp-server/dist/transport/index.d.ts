/**
 * HelixAgent Transport for Generic MCP Server
 *
 * Provides HTTP/3, HTTP/2, HTTP/1.1 transport with TOON protocol support.
 */
export interface TransportOptions {
    preferHTTP3?: boolean;
    enableTOON?: boolean;
    enableBrotli?: boolean;
    timeout?: number;
}
/**
 * HelixAgent Transport Class
 */
export declare class HelixAgentTransport {
    private endpoint;
    private options;
    private protocol;
    private contentType;
    private compression;
    private connected;
    constructor(endpoint: string, options?: TransportOptions);
    /**
     * Connect to HelixAgent endpoint
     */
    connect(): Promise<void>;
    /**
     * Make a request to HelixAgent
     */
    request(method: string, path: string, body?: unknown): Promise<unknown>;
    /**
     * Get connection info
     */
    getConnectionInfo(): {
        protocol: string;
        contentType: string;
        compression: string;
    };
    /**
     * Encode to TOON format
     */
    private encodeTOON;
    /**
     * Decode from TOON format
     */
    private decodeTOON;
}
//# sourceMappingURL=index.d.ts.map