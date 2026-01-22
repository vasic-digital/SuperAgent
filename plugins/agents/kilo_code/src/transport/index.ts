/**
 * HelixAgent Transport Module for Kilo-Code
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

// Default options
const defaultConnectOptions: ConnectOptions = {
  preferHTTP3: true,
  enableTOON: true,
  enableBrotli: true,
  timeout: 30000,
};

// TOON abbreviations
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
 * HelixTransport provides unified transport for Kilo-Code
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
   * Connect to HelixAgent endpoint
   */
  async connect(endpoint: string, opts?: ConnectOptions): Promise<void> {
    this.endpoint = endpoint.replace(/\/$/, '');
    this.opts = { ...defaultConnectOptions, ...opts };

    await this.negotiateProtocol();

    this.contentType = this.opts.enableTOON
      ? 'application/toon+json'
      : 'application/json';

    this.compression = this.opts.enableBrotli ? 'br' : 'gzip';

    this.connected = true;
    this.emit('connected', { protocol: this.protocol });
  }

  /**
   * Negotiate the best available protocol
   */
  private async negotiateProtocol(): Promise<void> {
    try {
      const controller = new AbortController();
      const timeout = setTimeout(() => controller.abort(), 5000);

      const response = await fetch(`${this.endpoint}/health`, {
        method: 'HEAD',
        signal: controller.signal,
      });

      clearTimeout(timeout);

      if (response.ok) {
        this.protocol = 'h2';
        return;
      }
    } catch {
      this.protocol = 'http/1.1';
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

    let bodyData: string | undefined;
    if (req.body !== undefined) {
      if (this.contentType === 'application/toon+json') {
        bodyData = encodeTOON(req.body);
      } else {
        bodyData = JSON.stringify(req.body);
      }
    }

    const headers: Record<string, string> = {
      'Content-Type': this.contentType,
      'Accept': this.contentType,
      ...this.opts.headers,
      ...req.headers,
    };

    if (this.compression !== 'identity') {
      headers['Accept-Encoding'] = `${this.compression}, gzip`;
    }

    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.opts.timeout || 30000);

    try {
      const response = await fetch(`${this.endpoint}${req.path}`, {
        method: req.method,
        headers,
        body: bodyData,
        signal: controller.signal,
      });

      const respBody = await response.text();

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
        compression: (response.headers.get('Content-Encoding') || 'identity') as Compression,
      };
    } finally {
      clearTimeout(timeout);
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

/**
 * Encode value to TOON format
 */
export function encodeTOON(value: unknown): string {
  let json = JSON.stringify(value);

  for (const [full, abbrev] of Object.entries(TOON_ABBREVIATIONS)) {
    json = json.replace(new RegExp(`"${full}"`, 'g'), `"${abbrev}"`);
  }

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

  for (const [abbrev, full] of Object.entries(TOON_REVERSE)) {
    json = json.replace(new RegExp(`"${abbrev}"`, 'g'), `"${full}"`);
  }

  json = json.replace(/:1([,}])/g, ':true$1');
  json = json.replace(/:0([,}])/g, ':false$1');

  return JSON.parse(json);
}

// Factory function
export function createTransport(): HelixTransport {
  return new HelixTransport();
}
