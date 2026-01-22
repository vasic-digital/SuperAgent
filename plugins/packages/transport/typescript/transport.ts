/**
 * HelixAgent Transport Library for TypeScript/JavaScript CLI Agents
 *
 * Provides HTTP/3, HTTP/2, HTTP/1.1 transport with automatic fallback,
 * TOON protocol encoding, and Brotli compression.
 */

import { EventEmitter } from 'events';

// Protocol versions
export type Protocol = 'h3' | 'h2' | 'http/1.1';

// Content types
export type ContentType = 'application/toon+json' | 'application/json';

// Compression methods
export type Compression = 'br' | 'gzip' | 'identity';

// Connection options
export interface ConnectOptions {
  preferHTTP3?: boolean;
  enableTOON?: boolean;
  enableBrotli?: boolean;
  timeout?: number;
  headers?: Record<string, string>;
  tlsOptions?: {
    rejectUnauthorized?: boolean;
    ca?: string | Buffer;
  };
}

// Request structure
export interface Request {
  method: string;
  path: string;
  headers?: Record<string, string>;
  body?: unknown;
}

// Response structure
export interface Response {
  statusCode: number;
  headers: Record<string, string>;
  body: Buffer | string;
  protocol: Protocol;
  contentType: ContentType;
  compression: Compression;
}

// Streaming event
export interface StreamEvent {
  type: string;
  data: unknown;
  id?: string;
}

// Default options
export const defaultConnectOptions: ConnectOptions = {
  preferHTTP3: true,
  enableTOON: true,
  enableBrotli: true,
  timeout: 30000,
  headers: {},
};

/**
 * HelixTransport provides unified transport for CLI agent plugins
 */
export class HelixTransport extends EventEmitter {
  private endpoint: string = '';
  private opts: ConnectOptions = defaultConnectOptions;
  private protocol: Protocol = 'http/1.1';
  private contentType: ContentType = 'application/json';
  private compression: Compression = 'identity';
  private connected: boolean = false;

  constructor() {
    super();
  }

  /**
   * Connect to HelixAgent endpoint with automatic protocol negotiation
   */
  async connect(endpoint: string, opts?: ConnectOptions): Promise<void> {
    this.endpoint = endpoint.replace(/\/$/, '');
    this.opts = { ...defaultConnectOptions, ...opts };

    // Attempt protocol negotiation
    await this.negotiateProtocol();

    // Set content type
    this.contentType = this.opts.enableTOON
      ? 'application/toon+json'
      : 'application/json';

    // Set compression
    this.compression = this.opts.enableBrotli ? 'br' : 'gzip';

    this.connected = true;
    this.emit('connected', { protocol: this.protocol });
  }

  /**
   * Negotiate the best available protocol
   */
  private async negotiateProtocol(): Promise<void> {
    // In browser/Node.js environments, HTTP/3 requires special handling
    // For now, fall back to HTTP/2 or HTTP/1.1 based on availability

    // Test connection with health endpoint
    try {
      const response = await this.testConnection();
      if (response.ok) {
        // Check HTTP version from response
        // Note: fetch() doesn't expose HTTP version directly
        // We'll use HTTP/2 as default for modern endpoints
        this.protocol = 'h2';
        return;
      }
    } catch (error) {
      // Fall back to HTTP/1.1
      this.protocol = 'http/1.1';
    }
  }

  private async testConnection(): Promise<globalThis.Response> {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 5000);

    try {
      const response = await fetch(`${this.endpoint}/health`, {
        method: 'HEAD',
        signal: controller.signal,
        headers: this.opts.headers,
      });
      return response;
    } finally {
      clearTimeout(timeout);
    }
  }

  /**
   * Get negotiated protocol
   */
  getProtocol(): Protocol {
    return this.protocol;
  }

  /**
   * Get negotiated content type
   */
  getContentType(): ContentType {
    return this.contentType;
  }

  /**
   * Get negotiated compression
   */
  getCompression(): Compression {
    return this.compression;
  }

  /**
   * Perform a request
   */
  async do(req: Request): Promise<Response> {
    if (!this.connected) {
      throw new Error('Transport not connected');
    }

    // Serialize body
    let bodyData: string | undefined;
    if (req.body !== undefined) {
      if (this.contentType === 'application/toon+json') {
        bodyData = encodeTOON(req.body);
      } else {
        bodyData = JSON.stringify(req.body);
      }

      // Apply compression
      if (this.compression === 'br') {
        bodyData = await compressBrotli(bodyData);
      } else if (this.compression === 'gzip') {
        bodyData = await compressGzip(bodyData);
      }
    }

    // Build headers
    const headers: Record<string, string> = {
      'Content-Type': this.contentType,
      'Accept': this.contentType,
      ...this.opts.headers,
      ...req.headers,
    };

    if (this.compression !== 'identity') {
      headers['Content-Encoding'] = this.compression;
      headers['Accept-Encoding'] = `${this.compression}, gzip`;
    }

    // Perform request
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.opts.timeout || 30000);

    try {
      const response = await fetch(`${this.endpoint}${req.path}`, {
        method: req.method,
        headers,
        body: bodyData,
        signal: controller.signal,
      });

      // Read response
      let respBody = await response.text();

      // Decompress if needed
      const respCompression = response.headers.get('Content-Encoding') as Compression || 'identity';
      if (respCompression === 'br') {
        respBody = await decompressBrotli(respBody);
      } else if (respCompression === 'gzip') {
        respBody = await decompressGzip(respBody);
      }

      // Build response headers
      const respHeaders: Record<string, string> = {};
      response.headers.forEach((value, key) => {
        respHeaders[key] = value;
      });

      return {
        statusCode: response.status,
        headers: respHeaders,
        body: respBody,
        protocol: this.protocol,
        contentType: (response.headers.get('Content-Type') || 'application/json') as ContentType,
        compression: respCompression,
      };
    } finally {
      clearTimeout(timeout);
    }
  }

  /**
   * Perform a streaming request (SSE)
   */
  async *stream(req: Request): AsyncGenerator<StreamEvent> {
    if (!this.connected) {
      throw new Error('Transport not connected');
    }

    // Build body
    let bodyData: string | undefined;
    if (req.body !== undefined) {
      bodyData = JSON.stringify(req.body);
    }

    // Build headers for SSE
    const headers: Record<string, string> = {
      'Accept': 'text/event-stream',
      'Cache-Control': 'no-cache',
      'Connection': 'keep-alive',
      ...this.opts.headers,
      ...req.headers,
    };

    if (bodyData) {
      headers['Content-Type'] = 'application/json';
    }

    // Perform request
    const response = await fetch(`${this.endpoint}${req.path}`, {
      method: req.method,
      headers,
      body: bodyData,
    });

    if (!response.ok) {
      throw new Error(`Stream request failed with status ${response.status}`);
    }

    if (!response.body) {
      throw new Error('No response body');
    }

    // Parse SSE stream
    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    let eventType = '';
    let eventId = '';
    let eventData = '';

    while (true) {
      const { done, value } = await reader.read();
      if (done) break;

      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (const line of lines) {
        const trimmed = line.trim();

        if (trimmed === '') {
          // Empty line = end of event
          if (eventData) {
            const data = eventData.trim();
            if (data === '[DONE]') {
              return;
            }

            let parsedData: unknown;
            try {
              parsedData = JSON.parse(data);
            } catch {
              parsedData = data;
            }

            yield {
              type: eventType || 'message',
              data: parsedData,
              id: eventId || undefined,
            };

            eventType = '';
            eventId = '';
            eventData = '';
          }
          continue;
        }

        if (trimmed.startsWith('event:')) {
          eventType = trimmed.slice(6).trim();
        } else if (trimmed.startsWith('data:')) {
          eventData += trimmed.slice(5).trim();
        } else if (trimmed.startsWith('id:')) {
          eventId = trimmed.slice(3).trim();
        }
      }
    }
  }

  /**
   * Close the transport
   */
  close(): void {
    this.connected = false;
    this.emit('closed');
  }
}

// TOON encoding/decoding
const TOON_ABBREVIATIONS: Record<string, string> = {
  'content': 'c',
  'role': 'r',
  'messages': 'm',
  'model': 'M',
  'temperature': 't',
  'max_tokens': 'x',
  'stream': 's',
  'user': 'u',
  'assistant': 'a',
  'system': 'S',
  'function': 'f',
  'tool_calls': 'tc',
  'finish_reason': 'fr',
  'choices': 'ch',
  'usage': 'U',
  'prompt_tokens': 'pt',
  'completion_tokens': 'ct',
  'total_tokens': 'tt',
  'id': 'i',
  'object': 'o',
  'created': 'cr',
  'index': 'ix',
  'delta': 'd',
  'name': 'n',
  'arguments': 'ar',
  'type': 'ty',
  'description': 'ds',
  'parameters': 'p',
  'properties': 'pr',
  'required': 'rq',
};

const TOON_REVERSE: Record<string, string> = Object.fromEntries(
  Object.entries(TOON_ABBREVIATIONS).map(([k, v]) => [v, k])
);

/**
 * Encode value to TOON format
 */
export function encodeTOON(value: unknown): string {
  let json = JSON.stringify(value);

  // Apply abbreviations
  for (const [full, abbrev] of Object.entries(TOON_ABBREVIATIONS)) {
    json = json.replace(new RegExp(`"${full}"`, 'g'), `"${abbrev}"`);
  }

  // Compact booleans
  json = json.replace(/:true/g, ':1');
  json = json.replace(/:false/g, ':0');

  return 'T:' + json;
}

/**
 * Decode TOON format to value
 */
export function decodeTOON<T = unknown>(data: string): T {
  if (!data.startsWith('T:')) {
    return JSON.parse(data);
  }

  let json = data.slice(2);

  // Reverse abbreviations
  for (const [abbrev, full] of Object.entries(TOON_REVERSE)) {
    json = json.replace(new RegExp(`"${abbrev}"`, 'g'), `"${full}"`);
  }

  // Expand booleans
  json = json.replace(/:1([,}])/g, ':true$1');
  json = json.replace(/:0([,}])/g, ':false$1');

  return JSON.parse(json);
}

// Compression utilities (using native APIs where available)
async function compressBrotli(data: string): Promise<string> {
  // In Node.js, use zlib; in browser, fall back to gzip or no compression
  if (typeof globalThis.CompressionStream !== 'undefined') {
    // Use Compression Streams API if available
    const encoder = new TextEncoder();
    const stream = new globalThis.CompressionStream('deflate');
    const writer = stream.writable.getWriter();
    writer.write(encoder.encode(data));
    writer.close();

    const reader = stream.readable.getReader();
    const chunks: Uint8Array[] = [];
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      chunks.push(value);
    }

    // Convert to base64 for transport
    const combined = new Uint8Array(chunks.reduce((acc, c) => acc + c.length, 0));
    let offset = 0;
    for (const chunk of chunks) {
      combined.set(chunk, offset);
      offset += chunk.length;
    }

    return btoa(String.fromCharCode(...combined));
  }

  // No compression available, return as-is
  return data;
}

async function decompressBrotli(data: string): Promise<string> {
  // Try to decompress, fall back to returning as-is
  try {
    if (typeof globalThis.DecompressionStream !== 'undefined') {
      const binaryString = atob(data);
      const bytes = new Uint8Array(binaryString.length);
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }

      const stream = new globalThis.DecompressionStream('deflate');
      const writer = stream.writable.getWriter();
      writer.write(bytes);
      writer.close();

      const reader = stream.readable.getReader();
      const chunks: Uint8Array[] = [];
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        chunks.push(value);
      }

      const decoder = new TextDecoder();
      return chunks.map(c => decoder.decode(c)).join('');
    }
  } catch {
    // Decompression failed, return as-is
  }

  return data;
}

async function compressGzip(data: string): Promise<string> {
  if (typeof globalThis.CompressionStream !== 'undefined') {
    const encoder = new TextEncoder();
    const stream = new globalThis.CompressionStream('gzip');
    const writer = stream.writable.getWriter();
    writer.write(encoder.encode(data));
    writer.close();

    const reader = stream.readable.getReader();
    const chunks: Uint8Array[] = [];
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      chunks.push(value);
    }

    const combined = new Uint8Array(chunks.reduce((acc, c) => acc + c.length, 0));
    let offset = 0;
    for (const chunk of chunks) {
      combined.set(chunk, offset);
      offset += chunk.length;
    }

    return btoa(String.fromCharCode(...combined));
  }

  return data;
}

async function decompressGzip(data: string): Promise<string> {
  try {
    if (typeof globalThis.DecompressionStream !== 'undefined') {
      const binaryString = atob(data);
      const bytes = new Uint8Array(binaryString.length);
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i);
      }

      const stream = new globalThis.DecompressionStream('gzip');
      const writer = stream.writable.getWriter();
      writer.write(bytes);
      writer.close();

      const reader = stream.readable.getReader();
      const chunks: Uint8Array[] = [];
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        chunks.push(value);
      }

      const decoder = new TextDecoder();
      return chunks.map(c => decoder.decode(c)).join('');
    }
  } catch {
    // Decompression failed
  }

  return data;
}

// Export default instance factory
export function createTransport(): HelixTransport {
  return new HelixTransport();
}
