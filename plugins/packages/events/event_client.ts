/**
 * HelixAgent Event Client
 *
 * Provides real-time event subscription for CLI agent plugins.
 * Supports SSE, WebSocket, and Webhooks.
 */

import { EventEmitter } from 'events';

// Event types from HelixAgent background task system
export type TaskEventType =
  | 'task.created'
  | 'task.started'
  | 'task.progress'
  | 'task.heartbeat'
  | 'task.paused'
  | 'task.resumed'
  | 'task.completed'
  | 'task.failed'
  | 'task.stuck'
  | 'task.cancelled'
  | 'task.retrying'
  | 'task.deadletter'
  | 'task.log'
  | 'task.resource';

// Event types from HelixAgent AI Debate system
export type DebateEventType =
  | 'debate.started'
  | 'debate.round_started'
  | 'debate.position_submitted'
  | 'debate.validation_phase'
  | 'debate.polish_phase'
  | 'debate.consensus'
  | 'debate.completed'
  | 'debate.failed';

export type EventType = TaskEventType | DebateEventType | string;

// Event payload structures
export interface TaskEvent {
  id: string;
  type: TaskEventType;
  timestamp: string;
  data: {
    taskId: string;
    status?: string;
    progress?: number;
    message?: string;
    output?: string;
    exitCode?: number;
    resources?: {
      cpuPercent: number;
      memoryMb: number;
      ioReadBytes: number;
      ioWriteBytes: number;
    };
  };
}

export interface DebateEvent {
  id: string;
  type: DebateEventType;
  timestamp: string;
  data: {
    debateId: string;
    topic?: string;
    currentRound?: number;
    totalRounds?: number;
    participant?: string;
    position?: string;
    content?: string;
    confidence?: number;
    phase?: 'initial' | 'validation' | 'polish' | 'final';
    consensus?: {
      achieved: boolean;
      confidence: number;
      summary: string;
    };
    error?: string;
  };
}

export type HelixEvent = TaskEvent | DebateEvent | {
  id: string;
  type: EventType;
  timestamp: string;
  data: unknown;
};

// Event subscription options
export interface SubscriptionOptions {
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
  heartbeatInterval?: number;
  filterTypes?: EventType[];
}

// Default subscription options
const defaultSubscriptionOptions: SubscriptionOptions = {
  reconnectInterval: 5000,
  maxReconnectAttempts: 10,
  heartbeatInterval: 30000,
};

/**
 * SSE Event Client for Server-Sent Events
 */
export class SSEEventClient extends EventEmitter {
  private endpoint: string;
  private eventSource: EventSource | null = null;
  private reconnectAttempts: number = 0;
  private opts: SubscriptionOptions;
  private subscribed: boolean = false;

  constructor(endpoint: string, opts?: SubscriptionOptions) {
    super();
    this.endpoint = endpoint;
    this.opts = { ...defaultSubscriptionOptions, ...opts };
  }

  /**
   * Subscribe to events
   */
  subscribe(path: string = '/v1/events'): void {
    if (this.subscribed) {
      return;
    }

    const url = `${this.endpoint}${path}`;

    try {
      this.eventSource = new EventSource(url);

      this.eventSource.onopen = () => {
        this.reconnectAttempts = 0;
        this.subscribed = true;
        this.emit('connected');
      };

      this.eventSource.onmessage = (event) => {
        this.handleEvent(event.data);
      };

      this.eventSource.onerror = (error) => {
        this.emit('error', error);
        this.handleReconnect();
      };

      // Listen for specific event types
      const eventTypes: EventType[] = [
        'task.created', 'task.started', 'task.progress', 'task.heartbeat',
        'task.paused', 'task.resumed', 'task.completed', 'task.failed',
        'task.stuck', 'task.cancelled', 'task.retrying', 'task.deadletter',
        'task.log', 'task.resource',
        'debate.started', 'debate.round_started', 'debate.position_submitted',
        'debate.validation_phase', 'debate.polish_phase', 'debate.consensus',
        'debate.completed', 'debate.failed',
      ];

      for (const type of eventTypes) {
        this.eventSource.addEventListener(type, (event: MessageEvent) => {
          this.handleTypedEvent(type, event.data);
        });
      }
    } catch (error) {
      this.emit('error', error);
    }
  }

  private handleEvent(data: string): void {
    try {
      const event = JSON.parse(data) as HelixEvent;

      // Filter by type if specified
      if (this.opts.filterTypes && this.opts.filterTypes.length > 0) {
        if (!this.opts.filterTypes.includes(event.type)) {
          return;
        }
      }

      this.emit('event', event);
      this.emit(event.type, event);
    } catch (error) {
      this.emit('parse_error', { data, error });
    }
  }

  private handleTypedEvent(type: EventType, data: string): void {
    try {
      const eventData = JSON.parse(data);
      const event: HelixEvent = {
        id: eventData.id || `${type}-${Date.now()}`,
        type,
        timestamp: eventData.timestamp || new Date().toISOString(),
        data: eventData,
      };

      this.emit('event', event);
      this.emit(type, event);
    } catch (error) {
      this.emit('parse_error', { type, data, error });
    }
  }

  private handleReconnect(): void {
    if (this.reconnectAttempts >= (this.opts.maxReconnectAttempts || 10)) {
      this.emit('max_reconnects');
      return;
    }

    this.reconnectAttempts++;
    this.subscribed = false;

    setTimeout(() => {
      this.emit('reconnecting', { attempt: this.reconnectAttempts });
      this.subscribe();
    }, this.opts.reconnectInterval || 5000);
  }

  /**
   * Unsubscribe from events
   */
  unsubscribe(): void {
    if (this.eventSource) {
      this.eventSource.close();
      this.eventSource = null;
    }
    this.subscribed = false;
    this.emit('disconnected');
  }

  /**
   * Check if subscribed
   */
  isSubscribed(): boolean {
    return this.subscribed;
  }
}

/**
 * WebSocket Event Client
 */
export class WebSocketEventClient extends EventEmitter {
  private endpoint: string;
  private ws: WebSocket | null = null;
  private reconnectAttempts: number = 0;
  private opts: SubscriptionOptions;
  private subscribed: boolean = false;
  private pingInterval: NodeJS.Timeout | null = null;

  constructor(endpoint: string, opts?: SubscriptionOptions) {
    super();
    this.endpoint = endpoint.replace(/^http/, 'ws');
    this.opts = { ...defaultSubscriptionOptions, ...opts };
  }

  /**
   * Connect to WebSocket endpoint
   */
  connect(path: string = '/v1/ws/events'): void {
    if (this.subscribed) {
      return;
    }

    const url = `${this.endpoint}${path}`;

    try {
      this.ws = new WebSocket(url);

      this.ws.onopen = () => {
        this.reconnectAttempts = 0;
        this.subscribed = true;
        this.emit('connected');
        this.startPing();
      };

      this.ws.onmessage = (event) => {
        this.handleMessage(event.data);
      };

      this.ws.onerror = (error) => {
        this.emit('error', error);
      };

      this.ws.onclose = (event) => {
        this.stopPing();
        this.subscribed = false;
        this.emit('disconnected', { code: event.code, reason: event.reason });

        if (!event.wasClean) {
          this.handleReconnect();
        }
      };
    } catch (error) {
      this.emit('error', error);
    }
  }

  private handleMessage(data: string): void {
    try {
      const message = JSON.parse(data);

      // Handle ping/pong
      if (message.type === 'pong') {
        this.emit('pong', message);
        return;
      }

      // Filter by type if specified
      if (this.opts.filterTypes && this.opts.filterTypes.length > 0) {
        if (!this.opts.filterTypes.includes(message.type)) {
          return;
        }
      }

      const event: HelixEvent = {
        id: message.id || `ws-${Date.now()}`,
        type: message.type,
        timestamp: message.timestamp || new Date().toISOString(),
        data: message.data || message,
      };

      this.emit('event', event);
      this.emit(event.type, event);
    } catch (error) {
      this.emit('parse_error', { data, error });
    }
  }

  private handleReconnect(): void {
    if (this.reconnectAttempts >= (this.opts.maxReconnectAttempts || 10)) {
      this.emit('max_reconnects');
      return;
    }

    this.reconnectAttempts++;

    setTimeout(() => {
      this.emit('reconnecting', { attempt: this.reconnectAttempts });
      this.connect();
    }, this.opts.reconnectInterval || 5000);
  }

  private startPing(): void {
    this.pingInterval = setInterval(() => {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'ping', timestamp: Date.now() }));
      }
    }, this.opts.heartbeatInterval || 30000);
  }

  private stopPing(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  /**
   * Send a message
   */
  send(message: unknown): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      throw new Error('WebSocket not connected');
    }
  }

  /**
   * Disconnect
   */
  disconnect(): void {
    this.stopPing();
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
    this.subscribed = false;
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.subscribed && this.ws?.readyState === WebSocket.OPEN;
  }
}

/**
 * Unified Event Client that auto-selects transport
 */
export class EventClient extends EventEmitter {
  private sseClient: SSEEventClient | null = null;
  private wsClient: WebSocketEventClient | null = null;
  private preferredTransport: 'sse' | 'websocket';

  constructor(
    private endpoint: string,
    private opts?: SubscriptionOptions & { preferWebSocket?: boolean }
  ) {
    super();
    this.preferredTransport = opts?.preferWebSocket ? 'websocket' : 'sse';
  }

  /**
   * Subscribe to events using the best available transport
   */
  subscribe(path?: string): void {
    if (this.preferredTransport === 'websocket') {
      this.wsClient = new WebSocketEventClient(this.endpoint, this.opts);
      this.forwardEvents(this.wsClient);
      this.wsClient.connect(path || '/v1/ws/events');
    } else {
      this.sseClient = new SSEEventClient(this.endpoint, this.opts);
      this.forwardEvents(this.sseClient);
      this.sseClient.subscribe(path || '/v1/events');
    }
  }

  /**
   * Subscribe to task events
   */
  subscribeToTask(taskId: string): void {
    if (this.preferredTransport === 'websocket') {
      this.wsClient = new WebSocketEventClient(this.endpoint, this.opts);
      this.forwardEvents(this.wsClient);
      this.wsClient.connect(`/v1/ws/tasks/${taskId}`);
    } else {
      this.sseClient = new SSEEventClient(this.endpoint, this.opts);
      this.forwardEvents(this.sseClient);
      this.sseClient.subscribe(`/v1/tasks/${taskId}/events`);
    }
  }

  /**
   * Subscribe to debate events
   */
  subscribeToDebate(debateId: string): void {
    if (this.preferredTransport === 'websocket') {
      this.wsClient = new WebSocketEventClient(this.endpoint, this.opts);
      this.forwardEvents(this.wsClient);
      this.wsClient.connect(`/v1/ws/debates/${debateId}`);
    } else {
      this.sseClient = new SSEEventClient(this.endpoint, this.opts);
      this.forwardEvents(this.sseClient);
      this.sseClient.subscribe(`/v1/debates/${debateId}/events`);
    }
  }

  private forwardEvents(client: EventEmitter): void {
    const events = [
      'connected', 'disconnected', 'error', 'event',
      'reconnecting', 'max_reconnects', 'parse_error',
    ];

    for (const event of events) {
      client.on(event, (...args) => this.emit(event, ...args));
    }
  }

  /**
   * Unsubscribe from all events
   */
  unsubscribe(): void {
    this.sseClient?.unsubscribe();
    this.wsClient?.disconnect();
  }

  /**
   * Check if subscribed
   */
  isSubscribed(): boolean {
    return (this.sseClient?.isSubscribed() || this.wsClient?.isConnected()) ?? false;
  }
}

// Factory functions
export function createSSEClient(endpoint: string, opts?: SubscriptionOptions): SSEEventClient {
  return new SSEEventClient(endpoint, opts);
}

export function createWebSocketClient(endpoint: string, opts?: SubscriptionOptions): WebSocketEventClient {
  return new WebSocketEventClient(endpoint, opts);
}

export function createEventClient(
  endpoint: string,
  opts?: SubscriptionOptions & { preferWebSocket?: boolean }
): EventClient {
  return new EventClient(endpoint, opts);
}
