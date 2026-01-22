#!/usr/bin/env node
/**
 * HelixAgent Post-Tool Hook for Claude Code
 *
 * Renders results after tool execution:
 * - AI debate visualization with multiple render styles
 * - Task progress bars and status
 * - Confidence scores and consensus indicators
 */

const fs = require('fs');
const path = require('path');

// Session state storage
const SESSION_FILE = path.join(process.env.HOME || '/tmp', '.helixagent-claude-session.json');

// Render style (can be configured)
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
  white: '\x1b[37m',
};

// Position colors
const POSITION_COLORS = {
  analyst: COLORS.cyan,
  proposer: COLORS.green,
  critic: COLORS.yellow,
  synthesizer: COLORS.magenta,
  mediator: COLORS.blue,
  default: COLORS.white,
};

// Phase icons
const PHASE_ICONS = {
  initial: '\u{1F50D}',      // Magnifying glass
  validation: '\u2713',       // Check mark
  polish: '\u2728',           // Sparkles
  final: '\u{1F4DC}',         // Scroll
};

/**
 * Main post-tool handler
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
    process.stdout.write(JSON.stringify({}));
    return;
  }

  const toolName = hookData.postToolUse?.toolName || '';
  const result = hookData.postToolUse?.result || {};

  try {
    // Check if this is a HelixAgent tool result
    if (toolName.startsWith('helixagent_')) {
      const rendered = renderHelixResult(toolName, result);
      process.stdout.write(JSON.stringify({
        contextModification: rendered,
      }));
      return;
    }

    // Check if result contains debate data
    if (result.debate || result.debateId || result.consensus) {
      const rendered = renderDebateResult(result);
      process.stdout.write(JSON.stringify({
        contextModification: rendered,
      }));
      return;
    }

    // Check if result contains task data
    if (result.taskId || result.task || result.progress !== undefined) {
      const rendered = renderTaskResult(result);
      process.stdout.write(JSON.stringify({
        contextModification: rendered,
      }));
      return;
    }

    // No special rendering needed
    process.stdout.write(JSON.stringify({}));
  } catch (error) {
    process.stdout.write(JSON.stringify({
      contextModification: `[HelixAgent] Render warning: ${error.message}`,
    }));
  }
}

/**
 * Render HelixAgent-specific tool results
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
 * Render AI debate result
 */
function renderDebateResult(result) {
  const style = RENDER_STYLE;

  switch (style) {
    case 'theater':
      return renderTheaterStyle(result);
    case 'novel':
      return renderNovelStyle(result);
    case 'screenplay':
      return renderScreenplayStyle(result);
    case 'minimal':
      return renderMinimalStyle(result);
    case 'plain':
      return renderPlainStyle(result);
    default:
      return renderTheaterStyle(result);
  }
}

/**
 * Theater style - dramatic presentation
 */
function renderTheaterStyle(result) {
  const lines = [];

  // Title card
  lines.push('');
  lines.push(centerText('\u2554' + '\u2550'.repeat(60) + '\u2557'));
  lines.push(centerText('\u2551' + centerText('AI DEBATE ENSEMBLE RESULT', 60) + '\u2551'));
  lines.push(centerText('\u255A' + '\u2550'.repeat(60) + '\u255D'));
  lines.push('');

  // Topic
  if (result.topic) {
    lines.push(centerText(`${COLORS.bold}Topic:${COLORS.reset} ${result.topic}`));
    lines.push('');
  }

  // Render rounds
  if (result.rounds) {
    for (const round of result.rounds) {
      lines.push(centerText(`${COLORS.bold}\u2501\u2501\u2501 ROUND ${round.number} \u2501\u2501\u2501${COLORS.reset}`));
      lines.push('');

      for (const response of round.responses || []) {
        lines.push(renderTheaterResponse(response));
        lines.push('');
      }
    }
  }

  // Consensus
  if (result.consensus) {
    lines.push(renderConsensus(result.consensus));
  }

  // Multi-pass validation result
  if (result.multi_pass_result) {
    lines.push(renderMultiPassResult(result.multi_pass_result));
  }

  return lines.join('\n');
}

/**
 * Render theater-style response
 */
function renderTheaterResponse(response) {
  const color = POSITION_COLORS[response.role] || POSITION_COLORS.default;
  const lines = [];

  // Phase indicator
  if (response.phase) {
    const icon = PHASE_ICONS[response.phase] || '';
    const label = response.phase.toUpperCase();
    lines.push(`  ${COLORS.dim}${icon} ${label}${COLORS.reset}`);
  }

  // Character entrance
  lines.push(`  ${color}${COLORS.bold}${(response.participant || response.name || 'Unknown').toUpperCase()}${COLORS.reset} ${COLORS.dim}(${response.role})${COLORS.reset}`);

  // Provider/model info
  if (response.provider || response.model) {
    lines.push(`  ${COLORS.dim}[${response.provider || ''}/${response.model || ''}]${COLORS.reset}`);
  }

  // Content
  const wrapped = wrapText(response.content || '', 76);
  for (const line of wrapped) {
    lines.push(`    ${COLORS.italic}${line}${COLORS.reset}`);
  }

  // Confidence
  if (response.confidence !== undefined) {
    lines.push(`  ${COLORS.dim}${renderConfidence(response.confidence)}${COLORS.reset}`);
  }

  return lines.join('\n');
}

/**
 * Novel style - narrative prose
 */
function renderNovelStyle(result) {
  const lines = [];

  lines.push(`${COLORS.bold}Chapter: ${result.topic || 'AI Debate'}${COLORS.reset}`);
  lines.push('');

  if (result.rounds) {
    for (const round of result.rounds) {
      lines.push(`${COLORS.dim}--- Round ${round.number} ---${COLORS.reset}`);
      lines.push('');

      for (const response of round.responses || []) {
        const color = POSITION_COLORS[response.role] || POSITION_COLORS.default;
        const name = response.participant || response.name || 'The AI';
        let narrative = `${color}${name}${COLORS.reset}, the ${response.role}, spoke thoughtfully: `;
        narrative += `"${response.content}"`;

        if (response.confidence !== undefined) {
          narrative += ` ${COLORS.dim}(confidence: ${Math.round(response.confidence * 100)}%)${COLORS.reset}`;
        }

        lines.push(...wrapText(narrative, 80));
        lines.push('');
      }
    }
  }

  return lines.join('\n');
}

/**
 * Screenplay style - script format
 */
function renderScreenplayStyle(result) {
  const lines = [];

  lines.push(`${COLORS.bold}INT. DEBATE CHAMBER - ${result.startedAt || 'NOW'}${COLORS.reset}`);
  lines.push('');
  lines.push(`Topic: ${result.topic || 'AI Debate'}`);
  lines.push('');

  if (result.rounds) {
    for (const round of result.rounds) {
      lines.push(`${COLORS.dim}[ROUND ${round.number}]${COLORS.reset}`);
      lines.push('');

      for (const response of round.responses || []) {
        const color = POSITION_COLORS[response.role] || POSITION_COLORS.default;
        const name = (response.participant || response.name || 'AI').toUpperCase();

        lines.push(centerText(`${color}${name}${COLORS.reset}`));
        lines.push(centerText(`(${response.role})`));

        const wrapped = wrapText(response.content || '', 70);
        for (const line of wrapped) {
          lines.push('     ' + line);
        }
        lines.push('');
      }
    }
  }

  return lines.join('\n');
}

/**
 * Minimal style - clean and simple
 */
function renderMinimalStyle(result) {
  const lines = [];

  lines.push(`# ${result.topic || 'AI Debate'}`);
  lines.push('');

  if (result.rounds) {
    for (const round of result.rounds) {
      lines.push(`## Round ${round.number}`);
      lines.push('');

      for (const response of round.responses || []) {
        const conf = response.confidence !== undefined
          ? ` [${Math.round(response.confidence * 100)}%]`
          : '';
        lines.push(`**${response.participant || response.name}** (${response.role})${conf}:`);
        lines.push(response.content || '');
        lines.push('');
      }
    }
  }

  return lines.join('\n');
}

/**
 * Plain style - no formatting
 */
function renderPlainStyle(result) {
  const lines = [];

  lines.push(`Topic: ${result.topic || 'AI Debate'}`);
  lines.push('');

  if (result.rounds) {
    for (const round of result.rounds) {
      lines.push(`Round ${round.number}:`);
      lines.push('');

      for (const response of round.responses || []) {
        lines.push(`${response.participant || response.name} (${response.role}): ${response.content}`);
        lines.push('');
      }
    }
  }

  return lines.join('\n');
}

/**
 * Render consensus result
 */
function renderConsensus(consensus) {
  const lines = [];
  const status = consensus.achieved ? '\u2705 CONSENSUS ACHIEVED' : '\u274C NO CONSENSUS';

  lines.push(centerText(`${COLORS.bold}${status}${COLORS.reset}`));
  lines.push(centerText(renderConfidence(consensus.confidence)));
  lines.push('');
  lines.push(`${COLORS.italic}${consensus.summary}${COLORS.reset}`);

  return lines.join('\n');
}

/**
 * Render multi-pass validation result
 */
function renderMultiPassResult(multiPass) {
  const lines = [];

  lines.push('');
  lines.push(centerText('\u2500'.repeat(40)));
  lines.push(centerText(`${COLORS.bold}Multi-Pass Validation${COLORS.reset}`));
  lines.push('');
  lines.push(`  Phases Completed: ${multiPass.phases_completed || 0}`);
  lines.push(`  Overall Confidence: ${renderConfidence(multiPass.overall_confidence)}`);
  lines.push(`  Quality Improvement: ${((multiPass.quality_improvement || 0) * 100).toFixed(1)}%`);
  lines.push('');

  if (multiPass.final_response) {
    lines.push(`${COLORS.bold}Final Response:${COLORS.reset}`);
    lines.push(multiPass.final_response);
  }

  return lines.join('\n');
}

/**
 * Render ensemble result
 */
function renderEnsembleResult(result) {
  const lines = [];

  lines.push('');
  lines.push('\u2500'.repeat(60));
  lines.push(`${COLORS.bold}AI Debate Ensemble Result${COLORS.reset}`);
  lines.push('\u2500'.repeat(60));

  if (result.response) {
    lines.push('');
    lines.push(result.response);
  }

  if (result.confidence !== undefined) {
    lines.push('');
    lines.push(`Confidence: ${renderConfidence(result.confidence)}`);
  }

  if (result.providers_used) {
    lines.push(`Providers: ${result.providers_used.join(', ')}`);
  }

  return lines.join('\n');
}

/**
 * Render task result with progress bar
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
      lines.push(`  - ${source.title || source.id} (score: ${(source.score || 0).toFixed(2)})`);
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
  lines.push('');

  if (result.memories) {
    for (const memory of result.memories) {
      lines.push(`  - ${memory.content} (${memory.type})`);
    }
  }

  if (result.message) {
    lines.push(result.message);
  }

  return lines.join('\n');
}

/**
 * Render confidence score
 */
function renderConfidence(confidence) {
  if (confidence === undefined) return '';
  const percent = Math.round(confidence * 100);
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
function centerText(text, width = 80) {
  // Strip ANSI codes for length calculation
  const plainText = text.replace(/\x1b\[[0-9;]*m/g, '');
  const padding = Math.max(0, Math.floor((width - plainText.length) / 2));
  return ' '.repeat(padding) + text;
}

/**
 * Wrap text to width
 */
function wrapText(text, width) {
  const words = text.split(/\s+/);
  const lines = [];
  let currentLine = '';

  for (const word of words) {
    if (currentLine.length + word.length + 1 <= width) {
      currentLine += (currentLine ? ' ' : '') + word;
    } else {
      if (currentLine) {
        lines.push(currentLine);
      }
      currentLine = word;
    }
  }

  if (currentLine) {
    lines.push(currentLine);
  }

  return lines;
}

main().catch((error) => {
  process.stderr.write(`Post-tool error: ${error.message}\n`);
  process.stdout.write(JSON.stringify({}));
});
