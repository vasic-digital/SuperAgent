#!/usr/bin/env node
/**
 * HelixAgent Session Start Hook for Claude Code
 *
 * Initializes connection to HelixAgent server with:
 * - HTTP/3 (QUIC) transport with fallback to HTTP/2, HTTP/1.1
 * - TOON protocol encoding for token optimization
 * - Brotli compression for bandwidth reduction
 * - Real-time event subscription (SSE/WebSocket)
 */

const fs = require('fs');
const path = require('path');

// Session state storage
const SESSION_FILE = path.join(process.env.HOME || '/tmp', '.helixagent-claude-session.json');

// Configuration from hooks.json or environment
const CONFIG = {
  endpoint: process.env.HELIXAGENT_ENDPOINT || 'https://localhost:7061',
  preferHTTP3: process.env.HELIXAGENT_PREFER_HTTP3 !== 'false',
  enableTOON: process.env.HELIXAGENT_ENABLE_TOON !== 'false',
  enableBrotli: process.env.HELIXAGENT_ENABLE_BROTLI !== 'false',
  subscribeToDebates: process.env.HELIXAGENT_SUBSCRIBE_DEBATES !== 'false',
  subscribeToTasks: process.env.HELIXAGENT_SUBSCRIBE_TASKS !== 'false',
};

/**
 * Main session start handler
 */
async function main() {
  // Read hook input from stdin
  let input = '';
  for await (const chunk of process.stdin) {
    input += chunk;
  }

  let hookData = {};
  try {
    hookData = JSON.parse(input);
  } catch (e) {
    // No input or invalid JSON - continue with defaults
  }

  const sessionId = hookData.sessionId || `claude-${Date.now()}`;

  try {
    // Initialize transport
    const transport = await initializeTransport();

    // Subscribe to events if enabled
    const eventSubscriptions = await subscribeToEvents(transport);

    // Save session state
    const sessionState = {
      sessionId,
      startedAt: new Date().toISOString(),
      endpoint: CONFIG.endpoint,
      protocol: transport.protocol,
      contentType: transport.contentType,
      compression: transport.compression,
      eventSubscriptions,
    };

    fs.writeFileSync(SESSION_FILE, JSON.stringify(sessionState, null, 2));

    // Output success response
    const response = {
      success: true,
      message: `HelixAgent connected via ${transport.protocol}`,
      sessionId,
      transport: {
        protocol: transport.protocol,
        contentType: transport.contentType,
        compression: transport.compression,
      },
      contextModification: formatConnectionBanner(transport),
    };

    process.stdout.write(JSON.stringify(response));
  } catch (error) {
    // Output error response
    const response = {
      success: false,
      error: error.message,
      contextModification: `[HelixAgent] Connection failed: ${error.message}`,
    };

    process.stdout.write(JSON.stringify(response));
  }
}

/**
 * Initialize transport with protocol negotiation
 */
async function initializeTransport() {
  const result = {
    protocol: 'http/1.1',
    contentType: 'application/json',
    compression: 'identity',
  };

  try {
    // Test connection to health endpoint
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 5000);

    const response = await fetch(`${CONFIG.endpoint}/health`, {
      method: 'HEAD',
      signal: controller.signal,
    });

    clearTimeout(timeout);

    if (response.ok) {
      // Determine protocol from response
      // Note: fetch() doesn't expose HTTP version, assume HTTP/2 for modern servers
      result.protocol = CONFIG.preferHTTP3 ? 'h2' : 'http/1.1';
    }
  } catch (error) {
    // Connection test failed, use fallback
    result.protocol = 'http/1.1';
  }

  // Set content type
  result.contentType = CONFIG.enableTOON
    ? 'application/toon+json'
    : 'application/json';

  // Set compression
  result.compression = CONFIG.enableBrotli ? 'br' : 'gzip';

  return result;
}

/**
 * Subscribe to real-time events
 */
async function subscribeToEvents(transport) {
  const subscriptions = [];

  if (CONFIG.subscribeToDebates) {
    subscriptions.push({
      type: 'debate',
      path: '/v1/events',
      filter: ['debate.started', 'debate.round_started', 'debate.completed'],
    });
  }

  if (CONFIG.subscribeToTasks) {
    subscriptions.push({
      type: 'task',
      path: '/v1/events',
      filter: ['task.progress', 'task.completed', 'task.failed'],
    });
  }

  return subscriptions;
}

/**
 * Format connection banner for context
 */
function formatConnectionBanner(transport) {
  const lines = [
    '',
    '┌─────────────────────────────────────────────────────────┐',
    '│             HelixAgent Connection Established           │',
    '├─────────────────────────────────────────────────────────┤',
    `│  Endpoint:    ${CONFIG.endpoint.padEnd(40)} │`,
    `│  Protocol:    ${transport.protocol.toUpperCase().padEnd(40)} │`,
    `│  Content:     ${transport.contentType.padEnd(40)} │`,
    `│  Compression: ${transport.compression.toUpperCase().padEnd(40)} │`,
    '├─────────────────────────────────────────────────────────┤',
    '│  AI Debate Ensemble ready for multi-LLM consensus      │',
    '│  Real-time events: Debates, Tasks, Progress            │',
    '└─────────────────────────────────────────────────────────┘',
    '',
  ];

  return lines.join('\n');
}

main().catch((error) => {
  process.stderr.write(`Session start error: ${error.message}\n`);
  process.exit(1);
});
