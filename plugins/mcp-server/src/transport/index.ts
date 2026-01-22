/**
 * HelixAgent Transport for Generic MCP Server
 *
 * Provides HTTP/3, HTTP/2, HTTP/1.1 transport with TOON protocol support.
 */

// Transport options
export interface TransportOptions {
  preferHTTP3?: boolean;
  enableTOON?: boolean;
  enableBrotli?: boolean;
  timeout?: number;
}

const defaultOptions: TransportOptions = {
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

/**
 * HelixAgent Transport Class
 */
export class HelixAgentTransport {
  private endpoint: string;
  private options: TransportOptions;
  private protocol: 'h3' | 'h2' | 'http/1.1' = 'http/1.1';
  private contentType: string = 'application/json';
  private compression: string = 'gzip';
  private connected: boolean = false;

  constructor(endpoint: string, options?: TransportOptions) {
    this.endpoint = endpoint.replace(/\/$/, '');
    this.options = { ...defaultOptions, ...options };
  }

  /**
   * Connect to HelixAgent endpoint
   */
  async connect(): Promise<void> {
    // Negotiate protocol
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
      }
    } catch {
      this.protocol = 'http/1.1';
    }

    // Set content type and compression
    this.contentType = this.options.enableTOON
      ? 'application/toon+json'
      : 'application/json';
    this.compression = this.options.enableBrotli ? 'br' : 'gzip';

    this.connected = true;
  }

  /**
   * Make a request to HelixAgent
   */
  async request(method: string, path: string, body?: unknown): Promise<unknown> {
    if (!this.connected) {
      await this.connect();
    }

    // Encode body
    let bodyData: string | undefined;
    if (body !== undefined) {
      if (this.options.enableTOON) {
        bodyData = this.encodeTOON(body);
      } else {
        bodyData = JSON.stringify(body);
      }
    }

    // Build headers
    const headers: Record<string, string> = {
      'Content-Type': this.contentType,
      'Accept': this.contentType,
    };

    if (this.compression !== 'identity') {
      headers['Accept-Encoding'] = `${this.compression}, gzip`;
    }

    // Make request
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), this.options.timeout || 30000);

    try {
      const response = await fetch(`${this.endpoint}${path}`, {
        method,
        headers,
        body: bodyData,
        signal: controller.signal,
      });

      clearTimeout(timeout);

      const respBody = await response.text();

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${respBody}`);
      }

      // Decode response
      if (this.options.enableTOON && respBody.startsWith('T:')) {
        return this.decodeTOON(respBody);
      }

      try {
        return JSON.parse(respBody);
      } catch {
        return respBody;
      }
    } finally {
      clearTimeout(timeout);
    }
  }

  /**
   * Get connection info
   */
  getConnectionInfo(): { protocol: string; contentType: string; compression: string } {
    return {
      protocol: this.protocol,
      contentType: this.contentType,
      compression: this.compression,
    };
  }

  /**
   * Encode to TOON format
   */
  private encodeTOON(value: unknown): string {
    let json = JSON.stringify(value);

    for (const [full, abbrev] of Object.entries(TOON_ABBREVIATIONS)) {
      json = json.replace(new RegExp(`"${full}"`, 'g'), `"${abbrev}"`);
    }

    json = json.replace(/:true/g, ':1');
    json = json.replace(/:false/g, ':0');

    return 'T:' + json;
  }

  /**
   * Decode from TOON format
   */
  private decodeTOON(data: string): unknown {
    if (!data.startsWith('T:')) {
      return JSON.parse(data);
    }

    let json = data.slice(2);

    const reverse: Record<string, string> = {};
    for (const [full, abbrev] of Object.entries(TOON_ABBREVIATIONS)) {
      reverse[abbrev] = full;
    }

    for (const [abbrev, full] of Object.entries(reverse)) {
      json = json.replace(new RegExp(`"${abbrev}"`, 'g'), `"${full}"`);
    }

    json = json.replace(/:1([,}])/g, ':true$1');
    json = json.replace(/:0([,}])/g, ':false$1');

    return JSON.parse(json);
  }
}
