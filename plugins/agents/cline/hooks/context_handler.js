#!/usr/bin/env node
/**
 * HelixAgent Context Handler Hook for Cline
 *
 * Handles PreCompact events to:
 * - Save HelixAgent session context before compaction
 * - Preserve debate history and task state
 * - Store important memories
 */

const fs = require('fs');
const path = require('path');

// Session state storage
const SESSION_DIR = path.join(process.env.HOME || '/tmp', '.helixagent-cline');
const SESSION_FILE = path.join(SESSION_DIR, 'session.json');
const CONTEXT_FILE = path.join(SESSION_DIR, 'context.json');

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
    process.stdout.write(JSON.stringify({ cancel: false }));
    return;
  }

  try {
    // Load current session state
    let sessionState = null;
    if (fs.existsSync(SESSION_FILE)) {
      sessionState = JSON.parse(fs.readFileSync(SESSION_FILE, 'utf-8'));
    }

    if (sessionState) {
      // Save context for restoration after compaction
      const context = {
        savedAt: new Date().toISOString(),
        sessionState,
        hookData: {
          taskId: hookData.taskId,
          contextBefore: hookData.preCompact?.contextBefore,
        },
        summary: generateContextSummary(sessionState),
      };

      fs.writeFileSync(CONTEXT_FILE, JSON.stringify(context, null, 2));

      // Return context modification with summary
      const response = {
        cancel: false,
        contextModification: formatContextSummary(sessionState),
      };

      process.stdout.write(JSON.stringify(response));
      return;
    }

    // No active session
    process.stdout.write(JSON.stringify({ cancel: false }));
  } catch (error) {
    process.stdout.write(JSON.stringify({
      cancel: false,
      contextModification: `[HelixAgent] Context save warning: ${error.message}`,
    }));
  }
}

/**
 * Generate context summary
 */
function generateContextSummary(sessionState) {
  return {
    endpoint: sessionState.endpoint,
    protocol: sessionState.protocol,
    debatesStarted: sessionState.debatesStarted || 0,
    tasksCreated: sessionState.tasksCreated || 0,
    promptCount: sessionState.promptCount || 0,
    duration: calculateDuration(sessionState.startedAt),
  };
}

/**
 * Format context summary for display
 */
function formatContextSummary(sessionState) {
  const duration = calculateDuration(sessionState.startedAt);
  const lines = [
    '',
    '[HelixAgent Context Summary - Preserved for Compaction]',
    `  Endpoint:    ${sessionState.endpoint}`,
    `  Protocol:    ${sessionState.protocol}`,
    `  Duration:    ${duration}`,
    `  Debates:     ${sessionState.debatesStarted || 0}`,
    `  Tasks:       ${sessionState.tasksCreated || 0}`,
    `  Prompts:     ${sessionState.promptCount || 0}`,
    '',
  ];

  return lines.join('\n');
}

/**
 * Calculate duration from start time
 */
function calculateDuration(startedAt) {
  if (!startedAt) return 'unknown';

  const ms = Date.now() - new Date(startedAt).getTime();
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const hours = Math.floor(minutes / 60);

  if (hours > 0) {
    return `${hours}h ${minutes % 60}m`;
  } else if (minutes > 0) {
    return `${minutes}m ${seconds % 60}s`;
  } else {
    return `${seconds}s`;
  }
}

main().catch((error) => {
  process.stderr.write(`Context handler error: ${error.message}\n`);
  process.stdout.write(JSON.stringify({ cancel: false }));
});
