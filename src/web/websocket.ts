/**
 * WebSocket log streaming server
 *
 * Provides real-time log streaming to connected clients with subscription filtering.
 */

import type { ServerWebSocket } from "bun";
import type { WSClientMessage, WSServerMessage, WebSocketFilters } from "./types";
import type { LogEntry } from "../types";

/** WebSocket connection data */
interface WebSocketData {
  filters?: WebSocketFilters;
}

/** Set of all connected WebSocket clients */
const clients = new Set<ServerWebSocket<WebSocketData>>();

/**
 * WebSocket handlers for Bun.serve
 */
export const websocketHandlers = {
  /**
   * Handle new WebSocket connection
   */
  open(ws: ServerWebSocket<WebSocketData>) {
    clients.add(ws);
    console.log(`WebSocket client connected. Total clients: ${clients.size}`);

    // Send welcome message
    const welcomeMsg: WSServerMessage = {
      type: "connected",
      message: "WebSocket connection established",
    };
    ws.send(JSON.stringify(welcomeMsg));
  },

  /**
   * Handle incoming WebSocket messages
   */
  message(ws: ServerWebSocket<WebSocketData>, message: string | Buffer) {
    try {
      const data = typeof message === "string" ? message : message.toString();
      const msg = JSON.parse(data) as WSClientMessage;

      switch (msg.type) {
        case "subscribe":
          // Update subscription filters
          ws.data.filters = msg.filters;
          console.log("Client subscribed with filters:", msg.filters);
          break;

        case "unsubscribe":
          // Clear subscription filters
          ws.data.filters = undefined;
          console.log("Client unsubscribed");
          break;

        case "ping": {
          // Respond with pong
          const pongMsg: WSServerMessage = { type: "pong" };
          ws.send(JSON.stringify(pongMsg));
          break;
        }

        default:
          console.warn("Unknown WebSocket message type:", msg);
      }
    } catch (error) {
      console.error("Failed to process WebSocket message:", error);
    }
  },

  /**
   * Handle WebSocket connection close
   */
  close(ws: ServerWebSocket<WebSocketData>) {
    clients.delete(ws);
    console.log(`WebSocket client disconnected. Total clients: ${clients.size}`);
  },

  /**
   * Handle WebSocket errors
   */
  error(ws: ServerWebSocket<WebSocketData>, error: Error) {
    console.error("WebSocket error:", error);
  },
};

/**
 * Broadcast a new log entry to all connected clients
 *
 * Clients will receive the log only if it matches their subscription filters.
 */
export function broadcastLog(log: LogEntry): void {
  if (clients.size === 0) {
    return; // No clients connected, skip broadcasting
  }

  const logMsg: WSServerMessage = {
    type: "log",
    data: log,
  };

  const message = JSON.stringify(logMsg);

  for (const client of clients) {
    // Check if log matches client's filters
    if (shouldSendToClient(log, client.data.filters)) {
      try {
        client.send(message);
      } catch (error) {
        console.error("Failed to send log to WebSocket client:", error);
      }
    }
  }
}

/**
 * Determine if a log entry should be sent to a client based on their filters
 */
function shouldSendToClient(log: LogEntry, filters?: WebSocketFilters): boolean {
  if (!filters) {
    return true; // No filters, send all logs
  }

  // Filter by log type
  if (filters.types && filters.types.length > 0) {
    if (!filters.types.includes(log.type)) {
      return false;
    }
  }

  // Filter by server ID/name
  if (filters.servers && filters.servers.length > 0) {
    if (!log.serverName || !filters.servers.includes(log.serverName)) {
      return false;
    }
  }

  return true;
}

/**
 * Get the current count of connected WebSocket clients
 */
export function getClientCount(): number {
  return clients.size;
}

/**
 * Close all WebSocket connections
 */
export function closeAllClients(): void {
  for (const client of clients) {
    try {
      client.close();
    } catch (error) {
      console.error("Failed to close WebSocket client:", error);
    }
  }
  clients.clear();
}
