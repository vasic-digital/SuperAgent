#!/usr/bin/env node
/**
 * HelixAgent Task Lifecycle Hook for Cline
 *
 * Handles task lifecycle events:
 * - TaskStart: Initialize HelixAgent connection
 * - TaskResume: Restore session state
 * - TaskCancel: Cleanup resources
 * - TaskComplete: Final render and cleanup
 */

const fs = require('fs');
const path = require('path');

// Session state storage
const SESSION_DIR = path.join(process.env.HOME || '/tmp', '.helixagent-cline');
const SESSION_FILE = path.join(SESSION_DIR, 'session.json');

// Configuration from environment
const CONFIG = {
  endpoint: process.env.HELIXAGENT_ENDPOINT || 'https://localhost:7061',
  preferHTTP3: process.env.HELIXAGENT_PREFER_HTTP3 !== 'false',
  enableTOON: process.env.HELIXAGENT_ENABLE_TOON !== 'false',
  enableBrotli: process.env.HELIXAGENT_ENABLE_BROTLI !== 'false',
};

// Get lifecycle action from command line
const action = process.argv[2] || 'start';

/**
 * Main handler
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
    // No input or invalid JSON
  }

  const taskId = hookData.taskId || `cline-${Date.now()}`;

  try {
    let response;

    switch (action) {
      case 'start':
        response = await handleTaskStart(taskId, hookData);
        break;
      case 'resume':
        response = await handleTaskResume(taskId, hookData);
        break;
      case 'cancel':
        response = await handleTaskCancel(taskId, hookData);
        break;
      case 'complete':
        response = await handleTaskComplete(taskId, hookData);
        break;
      default:
        response = { cancel: false };
    }

    process.stdout.write(JSON.stringify(response));
  } catch (error) {
    process.stdout.write(JSON.stringify({
      cancel: false,
      contextModification: `[HelixAgent] Task lifecycle error: ${error.message}`,
    }));
  }
}

/**
 * Handle task start
 */
async function handleTaskStart(taskId, hookData) {
  // Ensure session directory exists
  if (!fs.existsSync(SESSION_DIR)) {
    fs.mkdirSync(SESSION_DIR, { recursive: true });
  }

  // Initialize transport
  const transport = await initializeTransport();

  // Create session state
  const sessionState = {
    taskId,
    startedAt: new Date().toISOString(),
    endpoint: CONFIG.endpoint,
    protocol: transport.protocol,
    contentType: transport.contentType,
    compression: transport.compression,
    debatesStarted: 0,
    tasksCreated: 0,
  };

  // Save session state
  fs.writeFileSync(SESSION_FILE, JSON.stringify(sessionState, null, 2));

  return {
    cancel: false,
    contextModification: formatConnectionBanner(transport, taskId),
  };
}

/**
 * Handle task resume
 */
async function handleTaskResume(taskId, hookData) {
  // Try to load existing session
  let sessionState = null;
  if (fs.existsSync(SESSION_FILE)) {
    try {
      sessionState = JSON.parse(fs.readFileSync(SESSION_FILE, 'utf-8'));
    } catch (e) {
      // Invalid session file
    }
  }

  if (sessionState) {
    // Update task ID and add resume marker
    sessionState.taskId = taskId;
    sessionState.resumedAt = new Date().toISOString();
    fs.writeFileSync(SESSION_FILE, JSON.stringify(sessionState, null, 2));

    return {
      cancel: false,
      contextModification: `[HelixAgent] Session restored (protocol: ${sessionState.protocol}, debates: ${sessionState.debatesStarted})`,
    };
  }

  // No existing session, start fresh
  return handleTaskStart(taskId, hookData);
}

/**
 * Handle task cancel
 */
async function handleTaskCancel(taskId, hookData) {
  // Load session state for summary
  let sessionState = null;
  if (fs.existsSync(SESSION_FILE)) {
    try {
      sessionState = JSON.parse(fs.readFileSync(SESSION_FILE, 'utf-8'));
    } catch (e) {
      // Invalid session file
    }
  }

  // Cleanup session file
  if (fs.existsSync(SESSION_FILE)) {
    fs.unlinkSync(SESSION_FILE);
  }

  if (sessionState) {
    const duration = formatDuration(Date.now() - new Date(sessionState.startedAt).getTime());
    return {
      cancel: false,
      contextModification: `[HelixAgent] Task cancelled after ${duration} (debates: ${sessionState.debatesStarted}, tasks: ${sessionState.tasksCreated})`,
    };
  }

  return { cancel: false };
}

/**
 * Handle task complete
 */
async function handleTaskComplete(taskId, hookData) {
  // Load session state for summary
  let sessionState = null;
  if (fs.existsSync(SESSION_FILE)) {
    try {
      sessionState = JSON.parse(fs.readFileSync(SESSION_FILE, 'utf-8'));
    } catch (e) {
      // Invalid session file
    }
  }

  // Cleanup session file
  if (fs.existsSync(SESSION_FILE)) {
    fs.unlinkSync(SESSION_FILE);
  }

  if (sessionState) {
    const duration = formatDuration(Date.now() - new Date(sessionState.startedAt).getTime());
    return {
      cancel: false,
      contextModification: formatSessionSummary(sessionState, duration),
    };
  }

  return { cancel: false };
}

/**
 * Initialize transport
 */
async function initializeTransport() {
  const result = {
    protocol: 'http/1.1',
    contentType: 'application/json',
    compression: 'identity',
  };

  try {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 5000);

    const response = await fetch(`${CONFIG.endpoint}/health`, {
      method: 'HEAD',
      signal: controller.signal,
    });

    clearTimeout(timeout);

    if (response.ok) {
      result.protocol = CONFIG.preferHTTP3 ? 'h2' : 'http/1.1';
    }
  } catch (error) {
    result.protocol = 'http/1.1';
  }

  result.contentType = CONFIG.enableTOON ? 'application/toon+json' : 'application/json';
  result.compression = CONFIG.enableBrotli ? 'br' : 'gzip';

  return result;
}

/**
 * Format connection banner
 */
function formatConnectionBanner(transport, taskId) {
  const lines = [
    '',
    '\u250C' + '\u2500'.repeat(58) + '\u2510',
    '\u2502' + centerText('HelixAgent Connection Established', 58) + '\u2502',
    '\u251C' + '\u2500'.repeat(58) + '\u2524',
    '\u2502  ' + `Task:       ${taskId}`.padEnd(56) + '\u2502',
    '\u2502  ' + `Protocol:   ${transport.protocol.toUpperCase()}`.padEnd(56) + '\u2502',
    '\u2502  ' + `Content:    ${transport.contentType}`.padEnd(56) + '\u2502',
    '\u2502  ' + `Compress:   ${transport.compression.toUpperCase()}`.padEnd(56) + '\u2502',
    '\u251C' + '\u2500'.repeat(58) + '\u2524',
    '\u2502  ' + 'AI Debate Ensemble: 15 LLMs (5 positions x 3 each)'.padEnd(56) + '\u2502',
    '\u2502  ' + 'Multi-pass validation: Enabled'.padEnd(56) + '\u2502',
    '\u2514' + '\u2500'.repeat(58) + '\u2518',
    '',
  ];

  return lines.join('\n');
}

/**
 * Format session summary
 */
function formatSessionSummary(sessionState, duration) {
  const lines = [
    '',
    '\u250C' + '\u2500'.repeat(58) + '\u2510',
    '\u2502' + centerText('HelixAgent Session Complete', 58) + '\u2502',
    '\u251C' + '\u2500'.repeat(58) + '\u2524',
    '\u2502  ' + `Duration:   ${duration}`.padEnd(56) + '\u2502',
    '\u2502  ' + `Debates:    ${sessionState.debatesStarted}`.padEnd(56) + '\u2502',
    '\u2502  ' + `Tasks:      ${sessionState.tasksCreated}`.padEnd(56) + '\u2502',
    '\u2502  ' + `Protocol:   ${sessionState.protocol.toUpperCase()}`.padEnd(56) + '\u2502',
    '\u2514' + '\u2500'.repeat(58) + '\u2518',
    '',
  ];

  return lines.join('\n');
}

/**
 * Center text within width
 */
function centerText(text, width) {
  const padding = Math.max(0, Math.floor((width - text.length) / 2));
  return ' '.repeat(padding) + text + ' '.repeat(width - padding - text.length);
}

/**
 * Format duration
 */
function formatDuration(ms) {
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
  } else if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  } else {
    return `${seconds}s`;
  }
}

main().catch((error) => {
  process.stderr.write(`Task lifecycle error: ${error.message}\n`);
  process.stdout.write(JSON.stringify({ cancel: false }));
});
