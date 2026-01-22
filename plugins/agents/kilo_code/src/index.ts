/**
 * HelixAgent Integration for Kilo-Code
 *
 * Provides HTTP/3 transport, TOON protocol encoding, event subscription,
 * and rich UI rendering for AI debate visualization.
 */

// Transport exports
export {
  HelixTransport,
  createTransport,
  type ConnectOptions,
  type Request,
  type Response,
  type Protocol,
  type ContentType,
  type Compression,
} from './transport';

// Events exports
export {
  EventClient,
  SSEEventClient,
  WebSocketEventClient,
  createEventClient,
  type HelixEvent,
  type TaskEvent,
  type DebateEvent,
  type SubscriptionOptions,
} from './events';

// UI exports
export {
  DebateRenderer,
  ProgressRenderer,
  createDebateRenderer,
  createProgressRenderer,
  type RenderStyle,
  type RendererConfig,
  type DebateState,
  type DebateRound,
  type DebateResponse,
} from './ui';

// Main client class
export class HelixAgentClient {
  private transport: import('./transport').HelixTransport;
  private eventClient: import('./events').EventClient | null = null;
  private debateRenderer: import('./ui').DebateRenderer;
  private progressRenderer: import('./ui').ProgressRenderer;
  private config: HelixAgentClientConfig;

  constructor(config?: Partial<HelixAgentClientConfig>) {
    this.config = {
      endpoint: 'https://localhost:7061',
      preferHTTP3: true,
      enableTOON: true,
      enableBrotli: true,
      debateRenderStyle: 'theater',
      progressStyle: 'unicode',
      subscribeToDebates: true,
      subscribeToTasks: true,
      ...config,
    };

    // Initialize components (lazy import to avoid circular deps)
    const { createTransport } = require('./transport');
    const { createDebateRenderer, createProgressRenderer } = require('./ui');

    this.transport = createTransport();
    this.debateRenderer = createDebateRenderer({
      style: this.config.debateRenderStyle as import('./ui').RenderStyle,
    });
    this.progressRenderer = createProgressRenderer(
      this.config.progressStyle as 'ascii' | 'unicode' | 'block' | 'dots'
    );
  }

  /**
   * Connect to HelixAgent server
   */
  async connect(): Promise<void> {
    await this.transport.connect(this.config.endpoint, {
      preferHTTP3: this.config.preferHTTP3,
      enableTOON: this.config.enableTOON,
      enableBrotli: this.config.enableBrotli,
    });

    // Subscribe to events if configured
    if (this.config.subscribeToDebates || this.config.subscribeToTasks) {
      const { createEventClient } = require('./events');
      this.eventClient = createEventClient(this.config.endpoint);
      this.eventClient.subscribe();
    }
  }

  /**
   * Start an AI debate
   */
  async debate(topic: string, options?: DebateOptions): Promise<DebateResult> {
    const response = await this.transport.do({
      method: 'POST',
      path: '/v1/debates',
      body: {
        topic,
        rounds: options?.rounds ?? 3,
        enable_multi_pass_validation: options?.enableMultiPassValidation ?? true,
        validation_config: options?.validationConfig,
      },
    });

    const result = JSON.parse(response.body.toString());

    // Render the result
    const rendered = this.debateRenderer.renderDebate(result);
    console.log(rendered);

    return result;
  }

  /**
   * Get ensemble response
   */
  async ensemble(prompt: string, options?: EnsembleOptions): Promise<EnsembleResult> {
    const response = await this.transport.do({
      method: 'POST',
      path: '/v1/chat/completions',
      body: {
        model: 'helix-debate-ensemble',
        messages: [{ role: 'user', content: prompt }],
        temperature: options?.temperature ?? 0.7,
        max_tokens: options?.maxTokens,
        stream: options?.stream ?? false,
      },
    });

    return JSON.parse(response.body.toString());
  }

  /**
   * Create a background task
   */
  async createTask(command: string, options?: TaskOptions): Promise<TaskResult> {
    const response = await this.transport.do({
      method: 'POST',
      path: '/v1/tasks',
      body: {
        command,
        description: options?.description,
        timeout: options?.timeout ?? 300,
        working_dir: options?.workingDir,
      },
    });

    return JSON.parse(response.body.toString());
  }

  /**
   * Query RAG system
   */
  async rag(query: string, options?: RAGOptions): Promise<RAGResult> {
    const response = await this.transport.do({
      method: 'POST',
      path: '/v1/rag/query',
      body: {
        query,
        collection: options?.collection,
        top_k: options?.topK ?? 5,
        rerank: options?.rerank ?? true,
      },
    });

    return JSON.parse(response.body.toString());
  }

  /**
   * Access memory system
   */
  async memory(action: MemoryAction, params: MemoryParams): Promise<MemoryResult> {
    let method: string;
    let path: string;
    let body: Record<string, unknown> | undefined;

    switch (action) {
      case 'add':
        method = 'POST';
        path = '/v1/memory';
        body = { content: params.content, metadata: params.metadata };
        break;
      case 'search':
        method = 'POST';
        path = '/v1/memory/search';
        body = { query: params.query, limit: params.limit ?? 10 };
        break;
      case 'get':
        method = 'GET';
        path = `/v1/memory/${params.memoryId}`;
        break;
      case 'delete':
        method = 'DELETE';
        path = `/v1/memory/${params.memoryId}`;
        break;
      default:
        throw new Error(`Unknown memory action: ${action}`);
    }

    const response = await this.transport.do({ method, path, body });
    return JSON.parse(response.body.toString());
  }

  /**
   * Subscribe to debate events
   */
  onDebateEvent(handler: (event: import('./events').DebateEvent) => void): void {
    if (this.eventClient) {
      this.eventClient.on('debate.started', handler);
      this.eventClient.on('debate.round_started', handler);
      this.eventClient.on('debate.position_submitted', handler);
      this.eventClient.on('debate.completed', handler);
    }
  }

  /**
   * Subscribe to task events
   */
  onTaskEvent(handler: (event: import('./events').TaskEvent) => void): void {
    if (this.eventClient) {
      this.eventClient.on('task.progress', handler);
      this.eventClient.on('task.completed', handler);
      this.eventClient.on('task.failed', handler);
    }
  }

  /**
   * Render progress bar
   */
  renderProgress(percent: number, label?: string): string {
    return this.progressRenderer.render(percent, label);
  }

  /**
   * Close connection
   */
  close(): void {
    this.transport.close();
    this.eventClient?.unsubscribe();
  }

  /**
   * Get connection info
   */
  getConnectionInfo(): ConnectionInfo {
    return {
      endpoint: this.config.endpoint,
      protocol: this.transport.getProtocol(),
      contentType: this.transport.getContentType(),
      compression: this.transport.getCompression(),
    };
  }
}

// Configuration types
export interface HelixAgentClientConfig {
  endpoint: string;
  preferHTTP3: boolean;
  enableTOON: boolean;
  enableBrotli: boolean;
  debateRenderStyle: string;
  progressStyle: string;
  subscribeToDebates: boolean;
  subscribeToTasks: boolean;
}

export interface DebateOptions {
  rounds?: number;
  enableMultiPassValidation?: boolean;
  validationConfig?: {
    enableValidation?: boolean;
    enablePolish?: boolean;
    validationTimeout?: number;
    polishTimeout?: number;
    minConfidenceToSkip?: number;
    maxValidationRounds?: number;
    showPhaseIndicators?: boolean;
  };
}

export interface DebateResult {
  debateId: string;
  topic: string;
  rounds: import('./ui').DebateRound[];
  consensus?: {
    achieved: boolean;
    confidence: number;
    summary: string;
  };
  multi_pass_result?: {
    phases_completed: number;
    overall_confidence: number;
    quality_improvement: number;
    final_response: string;
  };
}

export interface EnsembleOptions {
  temperature?: number;
  maxTokens?: number;
  stream?: boolean;
}

export interface EnsembleResult {
  id: string;
  choices: Array<{
    message: {
      role: string;
      content: string;
    };
    finish_reason: string;
  }>;
  usage?: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}

export interface TaskOptions {
  description?: string;
  timeout?: number;
  workingDir?: string;
}

export interface TaskResult {
  taskId: string;
  status: string;
  progress?: number;
  output?: string;
}

export interface RAGOptions {
  collection?: string;
  topK?: number;
  rerank?: boolean;
}

export interface RAGResult {
  answer: string;
  sources: Array<{
    id: string;
    title?: string;
    content: string;
    score: number;
  }>;
}

export type MemoryAction = 'add' | 'search' | 'get' | 'delete';

export interface MemoryParams {
  content?: string;
  query?: string;
  memoryId?: string;
  metadata?: Record<string, unknown>;
  limit?: number;
}

export interface MemoryResult {
  success: boolean;
  memories?: Array<{
    id: string;
    content: string;
    type: string;
    createdAt: string;
  }>;
  message?: string;
}

export interface ConnectionInfo {
  endpoint: string;
  protocol: import('./transport').Protocol;
  contentType: import('./transport').ContentType;
  compression: import('./transport').Compression;
}

// Factory function
export function createHelixAgentClient(config?: Partial<HelixAgentClientConfig>): HelixAgentClient {
  return new HelixAgentClient(config);
}
