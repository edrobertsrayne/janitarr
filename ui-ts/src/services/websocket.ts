/**
 * WebSocket client for real-time log streaming
 */

import type {
  WSClientMessage,
  WSServerMessage,
  WebSocketFilters,
  LogEntry,
} from '../types';

type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'error';

interface WebSocketClientOptions {
  onLog?: (log: LogEntry) => void;
  onStatusChange?: (status: ConnectionStatus) => void;
  onError?: (error: Error) => void;
}

/**
 * WebSocket client for log streaming with auto-reconnect
 */
export class WebSocketClient {
  private ws: WebSocket | null = null;
  private status: ConnectionStatus = 'disconnected';
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private reconnectTimeout: number | null = null;
  private pingInterval: number | null = null;
  private options: WebSocketClientOptions;
  private url: string;

  constructor(options: WebSocketClientOptions = {}) {
    this.options = options;
    // Use ws:// protocol in development, wss:// in production
    // and construct URL from window.location.origin for robustness
    this.url = `${window.location.origin}/ws/logs`;
  }

  /**
   * Connect to WebSocket server
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      return; // Already connected
    }

    this.setStatus('connecting');

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        this.setStatus('connected');
        this.reconnectAttempts = 0;
        this.startPingInterval();
      };

      this.ws.onmessage = (event) => {
        this.handleMessage(event.data);
      };

      this.ws.onerror = () => {
        this.setStatus('error');
        this.options.onError?.(new Error('WebSocket error'));
      };

      this.ws.onclose = () => {
        this.setStatus('disconnected');
        this.stopPingInterval();
        this.scheduleReconnect();
      };
    } catch (error) {
      this.setStatus('error');
      this.options.onError?.(
        error instanceof Error ? error : new Error('Connection failed')
      );
      this.scheduleReconnect();
    }
  }

  /**
   * Disconnect from WebSocket server
   */
  disconnect(): void {
    if (this.reconnectTimeout !== null) {
      clearTimeout(this.reconnectTimeout);
      this.reconnectTimeout = null;
    }

    this.stopPingInterval();

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.setStatus('disconnected');
  }

  /**
   * Subscribe to logs with optional filters
   */
  subscribe(filters?: WebSocketFilters): void {
    this.send({ type: 'subscribe', filters });
  }

  /**
   * Unsubscribe from logs
   */
  unsubscribe(): void {
    this.send({ type: 'unsubscribe' });
  }

  /**
   * Send a message to the server
   */
  private send(message: WSClientMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    }
  }

  /**
   * Handle incoming messages
   */
  private handleMessage(data: string): void {
    try {
      const message: WSServerMessage = JSON.parse(data);

      switch (message.type) {
        case 'log':
          this.options.onLog?.(message.data);
          break;
        case 'connected':
          // Server confirmed connection
          break;
        case 'pong':
          // Keep-alive response
          break;
      }
    } catch (error) {
      this.options.onError?.(
        new Error('Failed to parse message: ' + data)
      );
    }
  }

  /**
   * Start ping interval to keep connection alive
   */
  private startPingInterval(): void {
    this.stopPingInterval();
    this.pingInterval = window.setInterval(() => {
      this.send({ type: 'ping' });
    }, 30000); // 30 seconds
  }

  /**
   * Stop ping interval
   */
  private stopPingInterval(): void {
    if (this.pingInterval !== null) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  /**
   * Schedule reconnection with exponential backoff
   */
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.setStatus('error');
      this.options.onError?.(
        new Error('Max reconnection attempts reached')
      );
      return;
    }

    // Exponential backoff: 1s, 2s, 4s, 8s, 16s, 30s (max)
    const delay = Math.min(
      1000 * Math.pow(2, this.reconnectAttempts),
      30000
    );

    this.reconnectAttempts++;

    this.reconnectTimeout = window.setTimeout(() => {
      this.connect();
    }, delay);
  }

  /**
   * Update connection status
   */
  private setStatus(status: ConnectionStatus): void {
    this.status = status;
    this.options.onStatusChange?.(status);
  }

  /**
   * Get current connection status
   */
  getStatus(): ConnectionStatus {
    return this.status;
  }
}
