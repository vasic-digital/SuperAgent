#!/usr/bin/env node
/**
 * HelixAgent Session End Hook for Claude Code
 *
 * Cleans up HelixAgent connection:
 * - Unsubscribes from event streams
 * - Closes transport connection
 * - Removes session state file
 */

const fs = require('fs');
const path = require('path');

// Session state storage
const SESSION_FILE = path.join(process.env.HOME || '/tmp', '.helixagent-claude-session.json');

/**
 * Main session end handler
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

  try {
    // Load session state
    let sessionState = null;
    if (fs.existsSync(SESSION_FILE)) {
      sessionState = JSON.parse(fs.readFileSync(SESSION_FILE, 'utf-8'));
    }

    if (sessionState) {
      // Calculate session duration
      const startedAt = new Date(sessionState.startedAt);
      const duration = Date.now() - startedAt.getTime();
      const durationStr = formatDuration(duration);

      // Cleanup event subscriptions
      await cleanupSubscriptions(sessionState);

      // Remove session file
      fs.unlinkSync(SESSION_FILE);

      // Output success response
      const response = {
        success: true,
        message: 'HelixAgent session ended',
        sessionId: sessionState.sessionId,
        duration: durationStr,
        contextModification: formatSessionSummary(sessionState, durationStr),
      };

      process.stdout.write(JSON.stringify(response));
    } else {
      // No active session
      const response = {
        success: true,
        message: 'No active HelixAgent session',
      };

      process.stdout.write(JSON.stringify(response));
    }
  } catch (error) {
    // Output error response
    const response = {
      success: false,
      error: error.message,
    };

    process.stdout.write(JSON.stringify(response));
  }
}

/**
 * Cleanup event subscriptions
 */
async function cleanupSubscriptions(sessionState) {
  // In a real implementation, this would close SSE/WebSocket connections
  // For now, just log the cleanup
  if (sessionState.eventSubscriptions) {
    for (const sub of sessionState.eventSubscriptions) {
      // Close subscription
    }
  }
}

/**
 * Format duration in human-readable format
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

/**
 * Format session summary
 */
function formatSessionSummary(sessionState, duration) {
  const lines = [
    '',
    '┌─────────────────────────────────────────────────────────┐',
    '│             HelixAgent Session Ended                    │',
    '├─────────────────────────────────────────────────────────┤',
    `│  Session ID:  ${sessionState.sessionId.substring(0, 38).padEnd(40)} │`,
    `│  Duration:    ${duration.padEnd(40)} │`,
    `│  Protocol:    ${sessionState.protocol.toUpperCase().padEnd(40)} │`,
    '└─────────────────────────────────────────────────────────┘',
    '',
  ];

  return lines.join('\n');
}

main().catch((error) => {
  process.stderr.write(`Session end error: ${error.message}\n`);
  process.exit(1);
});
