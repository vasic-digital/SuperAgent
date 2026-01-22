/**
 * HelixAgent UI Module for Kilo-Code
 *
 * Provides rich terminal UI for AI debate visualization and progress tracking.
 */

// Render styles
export type RenderStyle = 'theater' | 'novel' | 'screenplay' | 'minimal' | 'plain';

// Phase icons
const PHASE_ICONS = {
  initial: '\u{1F50D}',      // Magnifying glass
  validation: '\u2713',       // Check mark
  polish: '\u2728',           // Sparkles
  final: '\u{1F4DC}',         // Scroll
};

// Position colors (ANSI)
const POSITION_COLORS: Record<string, string> = {
  analyst: '\x1b[36m',       // Cyan
  proposer: '\x1b[32m',      // Green
  critic: '\x1b[33m',        // Yellow
  synthesizer: '\x1b[35m',   // Magenta
  mediator: '\x1b[34m',      // Blue
  default: '\x1b[37m',       // White
};

const RESET = '\x1b[0m';
const BOLD = '\x1b[1m';
const DIM = '\x1b[2m';
const ITALIC = '\x1b[3m';

// Debate round data
export interface DebateRound {
  number: number;
  responses: DebateResponse[];
  consensus?: {
    achieved: boolean;
    confidence: number;
    summary: string;
  };
}

export interface DebateResponse {
  participant: string;
  role: string;
  content: string;
  confidence?: number;
  provider?: string;
  model?: string;
  phase?: 'initial' | 'validation' | 'polish' | 'final';
}

export interface DebateState {
  debateId: string;
  topic: string;
  currentRound: number;
  totalRounds: number;
  currentPhase: 'initial' | 'validation' | 'polish' | 'final';
  rounds: DebateRound[];
  participants: string[];
  startedAt: string;
  completedAt?: string;
}

// Renderer configuration
export interface RendererConfig {
  style: RenderStyle;
  colorScheme: '256' | 'truecolor' | 'none';
  showConfidence: boolean;
  showPhaseIndicators: boolean;
  showTimestamps: boolean;
  maxContentWidth: number;
  animate: boolean;
}

const defaultConfig: RendererConfig = {
  style: 'theater',
  colorScheme: '256',
  showConfidence: true,
  showPhaseIndicators: true,
  showTimestamps: false,
  maxContentWidth: 80,
  animate: true,
};

/**
 * Debate Renderer
 */
export class DebateRenderer {
  private config: RendererConfig;

  constructor(config?: Partial<RendererConfig>) {
    this.config = { ...defaultConfig, ...config };
  }

  renderDebate(state: DebateState): string {
    switch (this.config.style) {
      case 'theater':
        return this.renderTheaterStyle(state);
      case 'novel':
        return this.renderNovelStyle(state);
      case 'screenplay':
        return this.renderScreenplayStyle(state);
      case 'minimal':
        return this.renderMinimalStyle(state);
      case 'plain':
        return this.renderPlainStyle(state);
      default:
        return this.renderTheaterStyle(state);
    }
  }

  renderResponse(response: DebateResponse): string {
    switch (this.config.style) {
      case 'theater':
        return this.renderTheaterResponse(response);
      case 'novel':
        return this.renderNovelResponse(response);
      case 'screenplay':
        return this.renderScreenplayResponse(response);
      case 'minimal':
        return this.renderMinimalResponse(response);
      case 'plain':
        return this.renderPlainResponse(response);
      default:
        return this.renderTheaterResponse(response);
    }
  }

  renderPhaseIndicator(phase: string): string {
    if (!this.config.showPhaseIndicators) return '';

    const icon = PHASE_ICONS[phase as keyof typeof PHASE_ICONS] || '';
    const labels: Record<string, string> = {
      initial: 'INITIAL RESPONSE',
      validation: 'VALIDATION',
      polish: 'POLISH & IMPROVE',
      final: 'FINAL CONCLUSION',
    };

    return `${icon} ${labels[phase] || phase.toUpperCase()}`;
  }

  renderConfidence(confidence?: number): string {
    if (!this.config.showConfidence || confidence === undefined) return '';

    const percent = Math.round(confidence * 100);
    const bar = this.renderProgressBar(percent, 10);
    return `[${bar}] ${percent}%`;
  }

  renderProgressBar(percent: number, width: number = 20): string {
    const filled = Math.round((percent / 100) * width);
    const empty = width - filled;
    return '\u2588'.repeat(filled) + '\u2591'.repeat(empty);
  }

  // Theater style
  private renderTheaterStyle(state: DebateState): string {
    const lines: string[] = [];

    lines.push('');
    lines.push(this.centerText('\u2554' + '\u2550'.repeat(60) + '\u2557'));
    lines.push(this.centerText('\u2551' + this.centerText('AI DEBATE ENSEMBLE', 60) + '\u2551'));
    lines.push(this.centerText('\u255A' + '\u2550'.repeat(60) + '\u255D'));
    lines.push('');
    lines.push(this.centerText(`${BOLD}Topic:${RESET} ${state.topic}`));
    lines.push(this.centerText(`Round ${state.currentRound} of ${state.totalRounds}`));
    lines.push('');

    for (const round of state.rounds) {
      lines.push(this.renderRoundHeader(round.number));
      lines.push('');

      for (const response of round.responses) {
        lines.push(this.renderTheaterResponse(response));
        lines.push('');
      }

      if (round.consensus) {
        lines.push(this.renderConsensus(round.consensus));
        lines.push('');
      }
    }

    return lines.join('\n');
  }

  private renderTheaterResponse(response: DebateResponse): string {
    const color = POSITION_COLORS[response.role] || POSITION_COLORS.default;
    const lines: string[] = [];

    if (response.phase) {
      lines.push(`  ${DIM}${this.renderPhaseIndicator(response.phase)}${RESET}`);
    }

    lines.push(`  ${color}${BOLD}${response.participant.toUpperCase()}${RESET} ${DIM}(${response.role})${RESET}`);

    if (response.provider || response.model) {
      lines.push(`  ${DIM}[${response.provider || ''}/${response.model || ''}]${RESET}`);
    }

    const wrapped = this.wrapText(response.content, this.config.maxContentWidth - 4);
    for (const line of wrapped) {
      lines.push(`    ${ITALIC}${line}${RESET}`);
    }

    if (response.confidence !== undefined) {
      lines.push(`  ${DIM}${this.renderConfidence(response.confidence)}${RESET}`);
    }

    return lines.join('\n');
  }

  // Novel style
  private renderNovelStyle(state: DebateState): string {
    const lines: string[] = [];

    lines.push(`${BOLD}Chapter: ${state.topic}${RESET}`);
    lines.push('');

    for (const round of state.rounds) {
      lines.push(`${DIM}--- Round ${round.number} ---${RESET}`);
      lines.push('');

      for (const response of round.responses) {
        lines.push(this.renderNovelResponse(response));
        lines.push('');
      }
    }

    return lines.join('\n');
  }

  private renderNovelResponse(response: DebateResponse): string {
    const color = POSITION_COLORS[response.role] || POSITION_COLORS.default;

    let narrative = `${color}${response.participant}${RESET}, the ${response.role}, spoke thoughtfully: `;
    narrative += `"${response.content}"`;

    if (response.confidence !== undefined) {
      narrative += ` ${DIM}(confidence: ${Math.round(response.confidence * 100)}%)${RESET}`;
    }

    return this.wrapText(narrative, this.config.maxContentWidth).join('\n');
  }

  // Screenplay style
  private renderScreenplayStyle(state: DebateState): string {
    const lines: string[] = [];

    lines.push(`${BOLD}INT. DEBATE CHAMBER - ${state.startedAt}${RESET}`);
    lines.push('');
    lines.push(`Topic: ${state.topic}`);
    lines.push('');

    for (const round of state.rounds) {
      lines.push(`${DIM}[ROUND ${round.number}]${RESET}`);
      lines.push('');

      for (const response of round.responses) {
        lines.push(this.renderScreenplayResponse(response));
        lines.push('');
      }
    }

    return lines.join('\n');
  }

  private renderScreenplayResponse(response: DebateResponse): string {
    const color = POSITION_COLORS[response.role] || POSITION_COLORS.default;
    const lines: string[] = [];

    lines.push(this.centerText(`${color}${response.participant.toUpperCase()}${RESET}`));
    lines.push(this.centerText(`(${response.role})`));

    const wrapped = this.wrapText(response.content, this.config.maxContentWidth - 10);
    for (const line of wrapped) {
      lines.push('     ' + line);
    }

    return lines.join('\n');
  }

  // Minimal style
  private renderMinimalStyle(state: DebateState): string {
    const lines: string[] = [];

    lines.push(`# ${state.topic}`);
    lines.push('');

    for (const round of state.rounds) {
      lines.push(`## Round ${round.number}`);
      lines.push('');

      for (const response of round.responses) {
        lines.push(this.renderMinimalResponse(response));
        lines.push('');
      }
    }

    return lines.join('\n');
  }

  private renderMinimalResponse(response: DebateResponse): string {
    const conf = response.confidence !== undefined
      ? ` [${Math.round(response.confidence * 100)}%]`
      : '';

    return `**${response.participant}** (${response.role})${conf}:\n${response.content}`;
  }

  // Plain style
  private renderPlainStyle(state: DebateState): string {
    const lines: string[] = [];

    lines.push(`Topic: ${state.topic}`);
    lines.push('');

    for (const round of state.rounds) {
      lines.push(`Round ${round.number}:`);
      lines.push('');

      for (const response of round.responses) {
        lines.push(this.renderPlainResponse(response));
        lines.push('');
      }
    }

    return lines.join('\n');
  }

  private renderPlainResponse(response: DebateResponse): string {
    return `${response.participant} (${response.role}): ${response.content}`;
  }

  // Helpers
  private renderRoundHeader(roundNumber: number): string {
    return this.centerText(`${BOLD}\u2501\u2501\u2501 ROUND ${roundNumber} \u2501\u2501\u2501${RESET}`);
  }

  private renderConsensus(consensus: { achieved: boolean; confidence: number; summary: string }): string {
    const lines: string[] = [];
    const status = consensus.achieved ? '\u2705 CONSENSUS ACHIEVED' : '\u274C NO CONSENSUS';

    lines.push(this.centerText(`${BOLD}${status}${RESET}`));
    lines.push(this.centerText(this.renderConfidence(consensus.confidence)));
    lines.push('');
    lines.push(`${ITALIC}${consensus.summary}${RESET}`);

    return lines.join('\n');
  }

  private centerText(text: string, width?: number): string {
    const w = width || this.config.maxContentWidth;
    const plainText = text.replace(/\x1b\[[0-9;]*m/g, '');
    const padding = Math.max(0, Math.floor((w - plainText.length) / 2));
    return ' '.repeat(padding) + text;
  }

  private wrapText(text: string, width: number): string[] {
    const words = text.split(/\s+/);
    const lines: string[] = [];
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
}

/**
 * Progress Bar Renderer
 */
export class ProgressRenderer {
  private style: 'ascii' | 'unicode' | 'block' | 'dots';
  private width: number;

  constructor(style: 'ascii' | 'unicode' | 'block' | 'dots' = 'unicode', width: number = 30) {
    this.style = style;
    this.width = width;
  }

  render(progress: number, label?: string): string {
    const percent = Math.min(100, Math.max(0, progress));
    const filled = Math.round((percent / 100) * this.width);

    let bar: string;
    switch (this.style) {
      case 'ascii':
        bar = '[' + '='.repeat(filled) + ' '.repeat(this.width - filled) + ']';
        break;
      case 'unicode':
        bar = '\u2503' + '\u2588'.repeat(filled) + '\u2591'.repeat(this.width - filled) + '\u2503';
        break;
      case 'block':
        bar = this.renderBlockBar(percent);
        break;
      case 'dots':
        bar = '\u25CF'.repeat(filled) + '\u25CB'.repeat(this.width - filled);
        break;
      default:
        bar = '[' + '#'.repeat(filled) + '-'.repeat(this.width - filled) + ']';
    }

    const percentStr = `${Math.round(percent)}%`.padStart(4);
    return label ? `${label} ${bar} ${percentStr}` : `${bar} ${percentStr}`;
  }

  private renderBlockBar(percent: number): string {
    const blocks = [' ', '\u258F', '\u258E', '\u258D', '\u258C', '\u258B', '\u258A', '\u2589', '\u2588'];
    const totalUnits = this.width * 8;
    const filledUnits = Math.round((percent / 100) * totalUnits);

    let bar = '';
    for (let i = 0; i < this.width; i++) {
      const unitStart = i * 8;
      const unitEnd = (i + 1) * 8;

      if (filledUnits >= unitEnd) {
        bar += '\u2588';
      } else if (filledUnits <= unitStart) {
        bar += ' ';
      } else {
        const partial = filledUnits - unitStart;
        bar += blocks[partial];
      }
    }

    return '\u2503' + bar + '\u2503';
  }
}

// Factory functions
export function createDebateRenderer(config?: Partial<RendererConfig>): DebateRenderer {
  return new DebateRenderer(config);
}

export function createProgressRenderer(
  style?: 'ascii' | 'unicode' | 'block' | 'dots',
  width?: number
): ProgressRenderer {
  return new ProgressRenderer(style, width);
}
