#!/usr/bin/env node
/**
 * HelixAgent Prompt Handler Hook for Cline
 *
 * Intercepts user prompts to:
 * - Encode prompts in TOON format for optimization
 * - Detect HelixAgent-specific commands
 * - Track prompt statistics
 */

const fs = require('fs');
const path = require('path');

// Session state storage
const SESSION_FILE = path.join(process.env.HOME || '/tmp', '.helixagent-cline', 'session.json');

// HelixAgent command patterns
const HELIX_COMMANDS = [
  { pattern: /^@helix\s+debate\s+(.+)$/i, tool: 'helixagent_debate', paramKey: 'topic' },
  { pattern: /^@helix\s+ask\s+(.+)$/i, tool: 'helixagent_ensemble', paramKey: 'prompt' },
  { pattern: /^@helix\s+task\s+(.+)$/i, tool: 'helixagent_task', paramKey: 'command' },
  { pattern: /^@helix\s+rag\s+(.+)$/i, tool: 'helixagent_rag', paramKey: 'query' },
  { pattern: /^@helix\s+remember\s+(.+)$/i, tool: 'helixagent_memory', paramKey: 'content', action: 'add' },
  { pattern: /^@helix\s+recall\s+(.+)$/i, tool: 'helixagent_memory', paramKey: 'query', action: 'search' },
];

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

  const userPrompt = hookData.userPromptSubmit?.content || '';

  try {
    // Check for HelixAgent commands
    for (const cmd of HELIX_COMMANDS) {
      const match = userPrompt.match(cmd.pattern);
      if (match) {
        const response = {
          cancel: false,
          toolCall: {
            toolName: cmd.tool,
            parameters: {
              [cmd.paramKey]: match[1],
            },
          },
          contextModification: `[HelixAgent] Routing to ${cmd.tool}...`,
        };

        if (cmd.action) {
          response.toolCall.parameters.action = cmd.action;
        }

        process.stdout.write(JSON.stringify(response));
        return;
      }
    }

    // Update session statistics
    updateSessionStats();

    // No special handling needed
    process.stdout.write(JSON.stringify({ cancel: false }));
  } catch (error) {
    process.stdout.write(JSON.stringify({
      cancel: false,
      contextModification: `[HelixAgent] Prompt handler warning: ${error.message}`,
    }));
  }
}

/**
 * Update session statistics
 */
function updateSessionStats() {
  if (!fs.existsSync(SESSION_FILE)) {
    return;
  }

  try {
    const sessionState = JSON.parse(fs.readFileSync(SESSION_FILE, 'utf-8'));
    sessionState.promptCount = (sessionState.promptCount || 0) + 1;
    fs.writeFileSync(SESSION_FILE, JSON.stringify(sessionState, null, 2));
  } catch (e) {
    // Ignore errors
  }
}

main().catch((error) => {
  process.stderr.write(`Prompt handler error: ${error.message}\n`);
  process.stdout.write(JSON.stringify({ cancel: false }));
});
