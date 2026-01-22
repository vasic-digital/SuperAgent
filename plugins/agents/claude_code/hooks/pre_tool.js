#!/usr/bin/env node
/**
 * HelixAgent Pre-Tool Hook for Claude Code
 *
 * Transforms tool calls before execution:
 * - Intercepts helix_* tool calls for HelixAgent routing
 * - Encodes request bodies in TOON format for token optimization
 * - Adds compression headers
 */

const fs = require('fs');
const path = require('path');

// Session state storage
const SESSION_FILE = path.join(process.env.HOME || '/tmp', '.helixagent-claude-session.json');

// TOON abbreviations for common fields
const TOON_ABBREVIATIONS = {
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

// HelixAgent-specific tools
const HELIX_TOOLS = [
  'helixagent_debate',
  'helixagent_ensemble',
  'helixagent_task',
  'helixagent_rag',
  'helixagent_memory',
];

/**
 * Main pre-tool handler
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
    process.stdout.write(JSON.stringify({ cancel: false }));
    return;
  }

  const toolName = hookData.preToolUse?.toolName || '';
  const parameters = hookData.preToolUse?.parameters || {};

  try {
    // Load session state
    let sessionState = null;
    if (fs.existsSync(SESSION_FILE)) {
      sessionState = JSON.parse(fs.readFileSync(SESSION_FILE, 'utf-8'));
    }

    // Check if this is a HelixAgent tool
    if (HELIX_TOOLS.includes(toolName)) {
      const response = await handleHelixTool(toolName, parameters, sessionState);
      process.stdout.write(JSON.stringify(response));
      return;
    }

    // Check if TOON encoding should be applied to request bodies
    if (sessionState?.contentType === 'application/toon+json') {
      const transformed = transformToTOON(parameters);
      const response = {
        cancel: false,
        preToolUse: {
          toolName,
          parameters: transformed,
        },
      };
      process.stdout.write(JSON.stringify(response));
      return;
    }

    // No transformation needed
    process.stdout.write(JSON.stringify({ cancel: false }));
  } catch (error) {
    // On error, don't cancel - let the tool proceed
    process.stdout.write(JSON.stringify({
      cancel: false,
      contextModification: `[HelixAgent] Pre-tool warning: ${error.message}`,
    }));
  }
}

/**
 * Handle HelixAgent-specific tools
 */
async function handleHelixTool(toolName, parameters, sessionState) {
  const endpoint = sessionState?.endpoint || 'https://localhost:7061';

  switch (toolName) {
    case 'helixagent_debate':
      return {
        cancel: false,
        contextModification: formatDebateStart(parameters),
        preToolUse: {
          toolName,
          parameters: {
            ...parameters,
            _helixagent_endpoint: endpoint,
            _helixagent_protocol: sessionState?.protocol || 'http/1.1',
          },
        },
      };

    case 'helixagent_ensemble':
      return {
        cancel: false,
        contextModification: '[HelixAgent] Routing to AI Debate Ensemble...',
        preToolUse: {
          toolName,
          parameters: {
            ...parameters,
            _helixagent_endpoint: endpoint,
          },
        },
      };

    case 'helixagent_task':
      return {
        cancel: false,
        contextModification: `[HelixAgent] Creating background task: ${parameters.description || 'unnamed'}`,
        preToolUse: {
          toolName,
          parameters: {
            ...parameters,
            _helixagent_endpoint: endpoint,
          },
        },
      };

    case 'helixagent_rag':
      return {
        cancel: false,
        contextModification: '[HelixAgent] Executing hybrid RAG query...',
        preToolUse: {
          toolName,
          parameters: {
            ...parameters,
            _helixagent_endpoint: endpoint,
          },
        },
      };

    case 'helixagent_memory':
      return {
        cancel: false,
        contextModification: '[HelixAgent] Accessing memory system...',
        preToolUse: {
          toolName,
          parameters: {
            ...parameters,
            _helixagent_endpoint: endpoint,
          },
        },
      };

    default:
      return { cancel: false };
  }
}

/**
 * Transform parameters to TOON format
 */
function transformToTOON(obj) {
  if (typeof obj !== 'object' || obj === null) {
    return obj;
  }

  if (Array.isArray(obj)) {
    return obj.map(transformToTOON);
  }

  const result = {};
  for (const [key, value] of Object.entries(obj)) {
    const newKey = TOON_ABBREVIATIONS[key] || key;
    result[newKey] = transformToTOON(value);
  }
  return result;
}

/**
 * Format debate start notification
 */
function formatDebateStart(parameters) {
  const topic = parameters.topic || parameters.prompt || 'Unknown topic';
  const participants = parameters.participants || 15;

  const lines = [
    '',
    '┌─────────────────────────────────────────────────────────┐',
    '│                  AI Debate Starting                     │',
    '├─────────────────────────────────────────────────────────┤',
    `│  Topic:        ${truncate(topic, 38).padEnd(40)} │`,
    `│  Participants: ${String(participants).padEnd(40)} │`,
    `│  Positions:    Analyst, Proposer, Critic, Synthesizer   │`,
    '│                Mediator (5 positions x 3 LLMs)          │',
    '├─────────────────────────────────────────────────────────┤',
    '│  Phases: Initial -> Validation -> Polish -> Final       │',
    '└─────────────────────────────────────────────────────────┘',
    '',
  ];

  return lines.join('\n');
}

/**
 * Truncate string with ellipsis
 */
function truncate(str, length) {
  if (str.length <= length) return str;
  return str.substring(0, length - 3) + '...';
}

main().catch((error) => {
  process.stderr.write(`Pre-tool error: ${error.message}\n`);
  process.stdout.write(JSON.stringify({ cancel: false }));
});
