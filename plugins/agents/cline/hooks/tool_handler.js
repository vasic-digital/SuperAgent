#!/usr/bin/env node
/**
 * HelixAgent Tool Handler Hook for Cline
 *
 * Handles PreToolUse and PostToolUse events:
 * - PreToolUse: Transform helix_* tool calls, encode TOON
 * - PostToolUse: Render AI debate results, progress bars
 */

const fs = require('fs');
const path = require('path');

// Session state storage
const SESSION_FILE = path.join(process.env.HOME || '/tmp', '.helixagent-cline', 'session.json');

// Get hook phase from command line
const phase = process.argv[2] || 'pre';

// Render style from environment
const RENDER_STYLE = process.env.HELIXAGENT_RENDER_STYLE || 'theater';

// ANSI colors
const COLORS = {
  reset: '\x1b[0m',
  bold: '\x1b[1m',
  dim: '\x1b[2m',
  italic: '\x1b[3m',
  cyan: '\x1b[36m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  magenta: '\x1b[35m',
  blue: '\x1b[34m',
};

// Position colors
const POSITION_COLORS = {
  analyst: COLORS.cyan,
  proposer: COLORS.green,
  critic: COLORS.yellow,
  synthesizer: COLORS.magenta,
  mediator: COLORS.blue,
};

// Phase icons
const PHASE_ICONS = {
  initial: '\u{1F50D}',
  validation: '\u2713',
  polish: '\u2728',
  final: '\u{1F4DC}',
};

// HelixAgent tools
const HELIX_TOOLS = [
  'helixagent_debate',
  'helixagent_ensemble',
  'helixagent_task',
  'helixagent_rag',
  'helixagent_memory',
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

  try {
    let response;

    if (phase === 'pre') {
      response = handlePreToolUse(hookData);
    } else {
      response = handlePostToolUse(hookData);
    }

    process.stdout.write(JSON.stringify(response));
  } catch (error) {
    process.stdout.write(JSON.stringify({
      cancel: false,
      contextModification: `[HelixAgent] Tool handler error: ${error.message}`,
    }));
  }
}

/**
 * Handle PreToolUse
 */
function handlePreToolUse(hookData) {
  const toolName = hookData.preToolUse?.toolName || '';
  const parameters = hookData.preToolUse?.parameters || {};

  // Check if this is a HelixAgent tool
  if (HELIX_TOOLS.includes(toolName)) {
    updateSessionStats('tool', toolName);

    // Load session for endpoint
    const sessionState = loadSession();
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
        updateSessionStats('task');
        return {
          cancel: false,
          contextModification: `[HelixAgent] Creating background task: ${parameters.description || parameters.command || 'unnamed'}`,
          preToolUse: {
            toolName,
            parameters: {
              ...parameters,
              _helixagent_endpoint: endpoint,
            },
          },
        };

      default:
        return {
          cancel: false,
          preToolUse: {
            toolName,
            parameters: {
              ...parameters,
              _helixagent_endpoint: endpoint,
            },
          },
        };
    }
  }

  return { cancel: false };
}

/**
 * Handle PostToolUse
 */
function handlePostToolUse(hookData) {
  const toolName = hookData.postToolUse?.toolName || '';
  const result = hookData.postToolUse?.result || {};

  // Check if this is a HelixAgent tool result
  if (toolName.startsWith('helixagent_')) {
    return {
      contextModification: renderHelixResult(toolName, result),
    };
  }

  // Check if result contains debate data
  if (result.debate || result.debateId || result.consensus) {
    return {
      contextModification: renderDebateResult(result),
    };
  }

  // Check if result contains task data
  if (result.taskId || result.progress !== undefined) {
    return {
      contextModification: renderTaskResult(result),
    };
  }

  return {};
}

/**
 * Render HelixAgent tool results
 */
function renderHelixResult(toolName, result) {
  switch (toolName) {
    case 'helixagent_debate':
      return renderDebateResult(result);
    case 'helixagent_ensemble':
      return renderEnsembleResult(result);
    case 'helixagent_task':
      return renderTaskResult(result);
    case 'helixagent_rag':
      return renderRAGResult(result);
    case 'helixagent_memory':
      return renderMemoryResult(result);
    default:
      return '';
  }
}

/**
 * Render debate result
 */
function renderDebateResult(result) {
  const lines = [];

  // Title
  lines.push('');
  lines.push('\u2550'.repeat(60));
  lines.push(`${COLORS.bold}AI Debate Ensemble Result${COLORS.reset}`);
  lines.push('\u2550'.repeat(60));

  if (result.topic) {
    lines.push(`Topic: ${result.topic}`);
    lines.push('');
  }

  // Render rounds
  if (result.rounds) {
    for (const round of result.rounds) {
      lines.push(`\u2500\u2500\u2500 Round ${round.number} \u2500\u2500\u2500`);
      lines.push('');

      for (const response of round.responses || []) {
        const color = POSITION_COLORS[response.role] || COLORS.reset;

        if (response.phase) {
          const icon = PHASE_ICONS[response.phase] || '';
          lines.push(`  ${COLORS.dim}${icon} ${response.phase.toUpperCase()}${COLORS.reset}`);
        }

        lines.push(`  ${color}${COLORS.bold}${(response.participant || response.name || 'AI').toUpperCase()}${COLORS.reset} (${response.role})`);

        if (response.provider || response.model) {
          lines.push(`  ${COLORS.dim}[${response.provider || ''}/${response.model || ''}]${COLORS.reset}`);
        }

        lines.push(`    ${COLORS.italic}${response.content}${COLORS.reset}`);

        if (response.confidence !== undefined) {
          lines.push(`  ${COLORS.dim}${renderConfidence(response.confidence)}${COLORS.reset}`);
        }
        lines.push('');
      }
    }
  }

  // Consensus
  if (result.consensus) {
    const status = result.consensus.achieved ? '\u2705 CONSENSUS ACHIEVED' : '\u274C NO CONSENSUS';
    lines.push(`${COLORS.bold}${status}${COLORS.reset}`);
    lines.push(renderConfidence(result.consensus.confidence));
    lines.push('');
    lines.push(`${COLORS.italic}${result.consensus.summary}${COLORS.reset}`);
  }

  // Multi-pass result
  if (result.multi_pass_result) {
    lines.push('');
    lines.push('\u2500'.repeat(40));
    lines.push(`${COLORS.bold}Multi-Pass Validation${COLORS.reset}`);
    lines.push(`Phases: ${result.multi_pass_result.phases_completed || 0}`);
    lines.push(`Confidence: ${renderConfidence(result.multi_pass_result.overall_confidence)}`);
    lines.push(`Quality Improvement: ${((result.multi_pass_result.quality_improvement || 0) * 100).toFixed(1)}%`);
  }

  return lines.join('\n');
}

/**
 * Render ensemble result
 */
function renderEnsembleResult(result) {
  const lines = [];
  lines.push('');
  lines.push('\u2500'.repeat(50));
  lines.push(`${COLORS.bold}AI Debate Ensemble${COLORS.reset}`);
  lines.push('\u2500'.repeat(50));

  if (result.response || result.choices?.[0]?.message?.content) {
    lines.push(result.response || result.choices[0].message.content);
  }

  if (result.confidence !== undefined) {
    lines.push('');
    lines.push(`Confidence: ${renderConfidence(result.confidence)}`);
  }

  return lines.join('\n');
}

/**
 * Render task result
 */
function renderTaskResult(result) {
  const lines = [];
  const taskId = result.taskId || result.id || 'unknown';
  const status = result.status || 'pending';
  const progress = result.progress !== undefined ? result.progress : 0;

  lines.push('');
  lines.push(`Task: ${taskId}`);
  lines.push(`Status: ${status}`);
  lines.push(`Progress: ${renderProgressBar(progress, 30)}`);

  if (result.output) {
    lines.push('');
    lines.push('Output:');
    lines.push(result.output);
  }

  return lines.join('\n');
}

/**
 * Render RAG result
 */
function renderRAGResult(result) {
  const lines = [];
  lines.push('');
  lines.push(`${COLORS.bold}Hybrid RAG Result${COLORS.reset}`);
  lines.push('');

  if (result.answer) {
    lines.push(result.answer);
  }

  if (result.sources) {
    lines.push('');
    lines.push(`${COLORS.dim}Sources:${COLORS.reset}`);
    for (const source of result.sources) {
      lines.push(`  - ${source.title || source.id} (${(source.score || 0).toFixed(2)})`);
    }
  }

  return lines.join('\n');
}

/**
 * Render memory result
 */
function renderMemoryResult(result) {
  const lines = [];
  lines.push('');
  lines.push(`${COLORS.bold}Memory System${COLORS.reset}`);

  if (result.memories) {
    for (const memory of result.memories) {
      lines.push(`  - ${memory.content} (${memory.type || 'general'})`);
    }
  }

  if (result.message) {
    lines.push(result.message);
  }

  return lines.join('\n');
}

/**
 * Format debate start notification
 */
function formatDebateStart(parameters) {
  const topic = parameters.topic || parameters.prompt || 'Unknown topic';
  const lines = [
    '',
    '\u250C' + '\u2500'.repeat(58) + '\u2510',
    '\u2502' + centerText('AI Debate Starting', 58) + '\u2502',
    '\u251C' + '\u2500'.repeat(58) + '\u2524',
    '\u2502  ' + `Topic: ${truncate(topic, 50)}`.padEnd(56) + '\u2502',
    '\u2502  ' + 'Participants: 15 LLMs (5 positions x 3 each)'.padEnd(56) + '\u2502',
    '\u2502  ' + 'Phases: Initial -> Validation -> Polish -> Final'.padEnd(56) + '\u2502',
    '\u2514' + '\u2500'.repeat(58) + '\u2518',
    '',
  ];
  return lines.join('\n');
}

/**
 * Render confidence score
 */
function renderConfidence(confidence) {
  if (confidence === undefined) return '';
  const percent = Math.round((confidence || 0) * 100);
  const bar = renderProgressBar(percent, 10);
  return `[${bar}] ${percent}%`;
}

/**
 * Render progress bar
 */
function renderProgressBar(percent, width) {
  const filled = Math.round((percent / 100) * width);
  const empty = width - filled;
  return '\u2588'.repeat(filled) + '\u2591'.repeat(empty);
}

/**
 * Center text
 */
function centerText(text, width) {
  const padding = Math.max(0, Math.floor((width - text.length) / 2));
  return ' '.repeat(padding) + text + ' '.repeat(width - padding - text.length);
}

/**
 * Truncate text
 */
function truncate(text, length) {
  if (text.length <= length) return text;
  return text.substring(0, length - 3) + '...';
}

/**
 * Load session state
 */
function loadSession() {
  if (!fs.existsSync(SESSION_FILE)) {
    return null;
  }
  try {
    return JSON.parse(fs.readFileSync(SESSION_FILE, 'utf-8'));
  } catch (e) {
    return null;
  }
}

/**
 * Update session statistics
 */
function updateSessionStats(type, value) {
  const sessionState = loadSession();
  if (!sessionState) return;

  if (type === 'tool' && value === 'helixagent_debate') {
    sessionState.debatesStarted = (sessionState.debatesStarted || 0) + 1;
  } else if (type === 'task') {
    sessionState.tasksCreated = (sessionState.tasksCreated || 0) + 1;
  }

  try {
    fs.writeFileSync(SESSION_FILE, JSON.stringify(sessionState, null, 2));
  } catch (e) {
    // Ignore errors
  }
}

main().catch((error) => {
  process.stderr.write(`Tool handler error: ${error.message}\n`);
  process.stdout.write(JSON.stringify({ cancel: false }));
});
