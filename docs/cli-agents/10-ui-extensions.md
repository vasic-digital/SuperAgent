# UI Extensions

Guide for implementing rich UI/UX extensions in CLI agent plugins.

## Overview

HelixAgent provides UI extension components for:

- **AI Debate Visualization** - Multi-round debate display with phases
- **Progress Bars** - Task progress with multiple styles
- **Notifications** - Alerts, warnings, and status updates
- **Resource Gauges** - CPU, memory, I/O monitoring

## Render Styles

| Style | Description | Best For |
|-------|-------------|----------|
| `theater` | Full visualization with animations | Rich terminals |
| `novel` | Narrative format with prose | Documentation |
| `screenplay` | Script-like dialogue format | Review mode |
| `minimal` | Compact, essential info only | Limited space |
| `plain` | Text only, no formatting | Piping/logging |

---

## AI Debate Visualization

### Theater Style (Full)

```
┌─────────────────────────────────────────────────────────────┐
│                   🎭 AI Debate in Progress                   │
├─────────────────────────────────────────────────────────────┤
│ Round 2/3                                        [████░░] 67%│
├─────────────────────────────────────────────────────────────┤
│ 🔵 Position 1: Claude (Advocate)                            │
│ ├─ "The proposed solution handles edge cases well and       │
│ │   provides clear error messages for debugging."           │
│ └─ Confidence: 0.92 ████████████████████░░░░                │
├─────────────────────────────────────────────────────────────┤
│ 🔴 Position 2: Gemini (Critic)                              │
│ ├─ "However, performance could be improved by using         │
│ │   a more efficient data structure for lookups."           │
│ └─ Confidence: 0.85 █████████████████░░░░░░░                │
├─────────────────────────────────────────────────────────────┤
│ 🟢 Position 3: DeepSeek (Synthesizer)                       │
│ ├─ "Combining both perspectives: maintain the clear         │
│ │   error handling while optimizing the data structure."    │
│ └─ Confidence: 0.88 █████████████████░░░░░░                 │
├─────────────────────────────────────────────────────────────┤
│ Current Phase: ✨ POLISH & IMPROVE                          │
│ Votes: 12/15                            Consensus: Building │
└─────────────────────────────────────────────────────────────┘
```

### Novel Style (Narrative)

```
═══════════════════════════════════════════════════════════════
                     AI DEBATE: Round 2 of 3
═══════════════════════════════════════════════════════════════

CLAUDE, the Advocate, presents their argument with conviction:

   "The proposed solution demonstrates excellent handling of
    edge cases, particularly in the error boundary scenarios.
    The clear error messages will significantly aid debugging
    and maintenance efforts."

Their confidence in this position: 92%

GEMINI, the Critic, offers a counterpoint:

   "While the error handling is commendable, I observe an
    opportunity for optimization. The current data structure
    could be replaced with a more efficient alternative for
    lookup operations."

Their confidence: 85%

DEEPSEEK, the Synthesizer, bridges the perspectives:

   "An elegant synthesis presents itself: we can preserve
    the valuable error handling mechanisms while introducing
    a more performant data structure. The best of both worlds."

Their confidence: 88%

───────────────────────────────────────────────────────────────
Phase: Polish & Improve | Votes: 12/15 | Consensus: Building
═══════════════════════════════════════════════════════════════
```

### Screenplay Style

```
                    AI DEBATE - ROUND 2
                    ====================

[The debate chamber. Three AI participants at their podiums.]

CLAUDE (Advocate)
     The proposed solution handles edge cases well and
     provides clear error messages for debugging.
     (pauses, confidence: 92%)

GEMINI (Critic)
     (stepping forward)
     However, performance could be improved by using
     a more efficient data structure for lookups.
     (confidence: 85%)

DEEPSEEK (Synthesizer)
     (thoughtfully)
     Combining both perspectives: maintain the clear
     error handling while optimizing the data structure.
     (confidence: 88%)

[PHASE: POLISH & IMPROVE. VOTES: 12/15]

                         SCENE CONTINUES...
```

### Minimal Style

```
[Debate R2/3] Claude:92% | Gemini:85% | DeepSeek:88% | Phase:Polish | Votes:12/15
```

### Plain Style (No Formatting)

```
AI Debate Round 2/3
Position 1 (Claude/Advocate): The proposed solution handles edge cases well... [92%]
Position 2 (Gemini/Critic): However, performance could be improved... [85%]
Position 3 (DeepSeek/Synthesizer): Combining both perspectives... [88%]
Phase: Polish & Improve | Votes: 12/15
```

---

## Implementation

### Debate Renderer (TypeScript)

```typescript
// packages/ui/src/debate_renderer.ts

export interface DebateRenderOptions {
  style: 'theater' | 'novel' | 'screenplay' | 'minimal' | 'plain';
  colorScheme: '256' | '16' | 'none';
  width: number;
  showConfidenceBars: boolean;
  showPhaseIndicators: boolean;
  animate: boolean;
}

export interface DebateState {
  round: number;
  totalRounds: number;
  positions: Position[];
  phase: 'initial' | 'validation' | 'polish' | 'conclusion';
  votes: { for: number; total: number };
  consensus?: string;
}

export interface Position {
  participant: string;
  role: 'advocate' | 'critic' | 'synthesizer' | 'validator' | 'moderator';
  argument: string;
  confidence: number;
}

export class DebateRenderer {
  private options: DebateRenderOptions;
  private chalk: Chalk;

  constructor(options: Partial<DebateRenderOptions> = {}) {
    this.options = {
      style: 'theater',
      colorScheme: '256',
      width: 65,
      showConfidenceBars: true,
      showPhaseIndicators: true,
      animate: true,
      ...options,
    };

    this.chalk = new Chalk({ level: this.getColorLevel() });
  }

  render(state: DebateState): string {
    switch (this.options.style) {
      case 'theater':
        return this.renderTheater(state);
      case 'novel':
        return this.renderNovel(state);
      case 'screenplay':
        return this.renderScreenplay(state);
      case 'minimal':
        return this.renderMinimal(state);
      case 'plain':
        return this.renderPlain(state);
      default:
        return this.renderTheater(state);
    }
  }

  private renderTheater(state: DebateState): string {
    const lines: string[] = [];
    const w = this.options.width;

    // Header
    lines.push(this.box('top', w));
    lines.push(this.boxLine('🎭 AI Debate in Progress', w, 'center'));
    lines.push(this.box('separator', w));

    // Progress
    const progress = Math.round((state.round / state.totalRounds) * 100);
    const progressBar = this.progressBar(progress, 10);
    lines.push(this.boxLine(
      `Round ${state.round}/${state.totalRounds}` +
      ' '.repeat(w - 30) +
      `[${progressBar}] ${progress}%`,
      w
    ));
    lines.push(this.box('separator', w));

    // Positions
    const roleColors: Record<string, string> = {
      advocate: '🔵',
      critic: '🔴',
      synthesizer: '🟢',
      validator: '🟡',
      moderator: '⚪',
    };

    for (let i = 0; i < state.positions.length; i++) {
      const pos = state.positions[i];
      const icon = roleColors[pos.role] || '⚪';

      lines.push(this.boxLine(
        `${icon} Position ${i + 1}: ${pos.participant} (${this.capitalize(pos.role)})`,
        w
      ));

      // Wrap argument
      const wrapped = this.wrapText(pos.argument, w - 6);
      for (let j = 0; j < wrapped.length; j++) {
        const prefix = j === 0 ? '├─ "' : '│   ';
        const suffix = j === wrapped.length - 1 ? '"' : '';
        lines.push(this.boxLine(`${prefix}${wrapped[j]}${suffix}`, w));
      }

      // Confidence bar
      if (this.options.showConfidenceBars) {
        const confBar = this.progressBar(pos.confidence * 100, 24);
        lines.push(this.boxLine(`└─ Confidence: ${(pos.confidence).toFixed(2)} ${confBar}`, w));
      }

      if (i < state.positions.length - 1) {
        lines.push(this.box('separator', w));
      }
    }

    // Phase and votes
    lines.push(this.box('separator', w));
    const phaseIcon = this.getPhaseIcon(state.phase);
    const phaseName = this.getPhaseName(state.phase);
    lines.push(this.boxLine(
      `Current Phase: ${phaseIcon} ${phaseName}`,
      w
    ));
    lines.push(this.boxLine(
      `Votes: ${state.votes.for}/${state.votes.total}` +
      ' '.repeat(w - 45) +
      `Consensus: ${state.consensus || 'Building'}`,
      w
    ));

    lines.push(this.box('bottom', w));

    return lines.join('\n');
  }

  private renderMinimal(state: DebateState): string {
    const positions = state.positions
      .map(p => `${p.participant}:${Math.round(p.confidence * 100)}%`)
      .join(' | ');

    return `[Debate R${state.round}/${state.totalRounds}] ${positions} | Phase:${this.capitalize(state.phase)} | Votes:${state.votes.for}/${state.votes.total}`;
  }

  private renderPlain(state: DebateState): string {
    const lines: string[] = [];
    lines.push(`AI Debate Round ${state.round}/${state.totalRounds}`);

    for (let i = 0; i < state.positions.length; i++) {
      const pos = state.positions[i];
      lines.push(`Position ${i + 1} (${pos.participant}/${this.capitalize(pos.role)}): ${pos.argument.slice(0, 60)}... [${Math.round(pos.confidence * 100)}%]`);
    }

    lines.push(`Phase: ${this.capitalize(state.phase)} | Votes: ${state.votes.for}/${state.votes.total}`);
    return lines.join('\n');
  }

  private getPhaseIcon(phase: string): string {
    const icons: Record<string, string> = {
      initial: '🔍',
      validation: '✓',
      polish: '✨',
      conclusion: '📜',
    };
    return icons[phase] || '●';
  }

  private getPhaseName(phase: string): string {
    const names: Record<string, string> = {
      initial: 'INITIAL RESPONSE',
      validation: 'VALIDATION',
      polish: 'POLISH & IMPROVE',
      conclusion: 'FINAL CONCLUSION',
    };
    return names[phase] || phase.toUpperCase();
  }

  private progressBar(percent: number, width: number): string {
    const filled = Math.round((percent / 100) * width);
    const empty = width - filled;
    return '█'.repeat(filled) + '░'.repeat(empty);
  }

  private box(type: 'top' | 'bottom' | 'separator', width: number): string {
    switch (type) {
      case 'top':
        return '┌' + '─'.repeat(width - 2) + '┐';
      case 'bottom':
        return '└' + '─'.repeat(width - 2) + '┘';
      case 'separator':
        return '├' + '─'.repeat(width - 2) + '┤';
    }
  }

  private boxLine(content: string, width: number, align: 'left' | 'center' | 'right' = 'left'): string {
    const innerWidth = width - 4;
    let text = content.slice(0, innerWidth);

    if (align === 'center') {
      const pad = Math.floor((innerWidth - text.length) / 2);
      text = ' '.repeat(pad) + text + ' '.repeat(innerWidth - pad - text.length);
    } else if (align === 'right') {
      text = text.padStart(innerWidth);
    } else {
      text = text.padEnd(innerWidth);
    }

    return '│ ' + text + ' │';
  }

  private wrapText(text: string, width: number): string[] {
    const words = text.split(' ');
    const lines: string[] = [];
    let current = '';

    for (const word of words) {
      if ((current + ' ' + word).length > width) {
        lines.push(current);
        current = word;
      } else {
        current = current ? current + ' ' + word : word;
      }
    }
    if (current) lines.push(current);

    return lines;
  }

  private capitalize(s: string): string {
    return s.charAt(0).toUpperCase() + s.slice(1);
  }

  private getColorLevel(): number {
    switch (this.options.colorScheme) {
      case '256': return 2;
      case '16': return 1;
      default: return 0;
    }
  }
}
```

### Debate Renderer (Go)

```go
// packages/ui/debate_renderer.go
package ui

import (
    "fmt"
    "strings"
)

type RenderStyle string

const (
    StyleTheater    RenderStyle = "theater"
    StyleNovel      RenderStyle = "novel"
    StyleScreenplay RenderStyle = "screenplay"
    StyleMinimal    RenderStyle = "minimal"
    StylePlain      RenderStyle = "plain"
)

type DebateRendererOptions struct {
    Style              RenderStyle
    ColorScheme        string
    Width              int
    ShowConfidenceBars bool
    ShowPhaseIndicators bool
    Animate            bool
}

type DebateState struct {
    Round       int
    TotalRounds int
    Positions   []Position
    Phase       string
    Votes       struct {
        For   int
        Total int
    }
    Consensus string
}

type Position struct {
    Participant string
    Role        string
    Argument    string
    Confidence  float64
}

type DebateRenderer struct {
    options DebateRendererOptions
}

func NewDebateRenderer(opts DebateRendererOptions) *DebateRenderer {
    if opts.Width == 0 {
        opts.Width = 65
    }
    if opts.Style == "" {
        opts.Style = StyleTheater
    }
    return &DebateRenderer{options: opts}
}

func (r *DebateRenderer) Render(state DebateState) string {
    switch r.options.Style {
    case StyleTheater:
        return r.renderTheater(state)
    case StyleNovel:
        return r.renderNovel(state)
    case StyleScreenplay:
        return r.renderScreenplay(state)
    case StyleMinimal:
        return r.renderMinimal(state)
    case StylePlain:
        return r.renderPlain(state)
    default:
        return r.renderTheater(state)
    }
}

func (r *DebateRenderer) renderTheater(state DebateState) string {
    var sb strings.Builder
    w := r.options.Width

    // Header
    sb.WriteString(r.boxTop(w))
    sb.WriteString(r.boxLine("🎭 AI Debate in Progress", w, "center"))
    sb.WriteString(r.boxSep(w))

    // Progress
    progress := state.Round * 100 / state.TotalRounds
    progressBar := r.progressBar(progress, 10)
    sb.WriteString(r.boxLine(
        fmt.Sprintf("Round %d/%d%s[%s] %d%%",
            state.Round, state.TotalRounds,
            strings.Repeat(" ", w-30),
            progressBar, progress),
        w, "left"))
    sb.WriteString(r.boxSep(w))

    // Positions
    roleIcons := map[string]string{
        "advocate":    "🔵",
        "critic":      "🔴",
        "synthesizer": "🟢",
        "validator":   "🟡",
        "moderator":   "⚪",
    }

    for i, pos := range state.Positions {
        icon := roleIcons[pos.Role]
        if icon == "" {
            icon = "⚪"
        }

        sb.WriteString(r.boxLine(
            fmt.Sprintf("%s Position %d: %s (%s)",
                icon, i+1, pos.Participant, strings.Title(pos.Role)),
            w, "left"))

        // Argument
        wrapped := r.wrapText(pos.Argument, w-6)
        for j, line := range wrapped {
            prefix := "├─ \""
            suffix := ""
            if j > 0 {
                prefix = "│   "
            }
            if j == len(wrapped)-1 {
                suffix = "\""
            }
            sb.WriteString(r.boxLine(prefix+line+suffix, w, "left"))
        }

        // Confidence
        if r.options.ShowConfidenceBars {
            confBar := r.progressBar(int(pos.Confidence*100), 24)
            sb.WriteString(r.boxLine(
                fmt.Sprintf("└─ Confidence: %.2f %s", pos.Confidence, confBar),
                w, "left"))
        }

        if i < len(state.Positions)-1 {
            sb.WriteString(r.boxSep(w))
        }
    }

    // Phase and votes
    sb.WriteString(r.boxSep(w))
    phaseIcon := r.getPhaseIcon(state.Phase)
    phaseName := r.getPhaseName(state.Phase)
    sb.WriteString(r.boxLine(
        fmt.Sprintf("Current Phase: %s %s", phaseIcon, phaseName),
        w, "left"))

    consensus := state.Consensus
    if consensus == "" {
        consensus = "Building"
    }
    sb.WriteString(r.boxLine(
        fmt.Sprintf("Votes: %d/%d%sConsensus: %s",
            state.Votes.For, state.Votes.Total,
            strings.Repeat(" ", w-45),
            consensus),
        w, "left"))

    sb.WriteString(r.boxBottom(w))

    return sb.String()
}

func (r *DebateRenderer) renderMinimal(state DebateState) string {
    var positions []string
    for _, p := range state.Positions {
        positions = append(positions, fmt.Sprintf("%s:%d%%", p.Participant, int(p.Confidence*100)))
    }

    return fmt.Sprintf("[Debate R%d/%d] %s | Phase:%s | Votes:%d/%d",
        state.Round, state.TotalRounds,
        strings.Join(positions, " | "),
        strings.Title(state.Phase),
        state.Votes.For, state.Votes.Total)
}

func (r *DebateRenderer) progressBar(percent, width int) string {
    filled := percent * width / 100
    empty := width - filled
    return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}

func (r *DebateRenderer) boxTop(w int) string {
    return "┌" + strings.Repeat("─", w-2) + "┐\n"
}

func (r *DebateRenderer) boxBottom(w int) string {
    return "└" + strings.Repeat("─", w-2) + "┘\n"
}

func (r *DebateRenderer) boxSep(w int) string {
    return "├" + strings.Repeat("─", w-2) + "┤\n"
}

func (r *DebateRenderer) boxLine(content string, w int, align string) string {
    inner := w - 4
    if len(content) > inner {
        content = content[:inner]
    }

    var text string
    switch align {
    case "center":
        pad := (inner - len(content)) / 2
        text = strings.Repeat(" ", pad) + content + strings.Repeat(" ", inner-pad-len(content))
    case "right":
        text = strings.Repeat(" ", inner-len(content)) + content
    default:
        text = content + strings.Repeat(" ", inner-len(content))
    }

    return "│ " + text + " │\n"
}

func (r *DebateRenderer) wrapText(text string, width int) []string {
    words := strings.Fields(text)
    var lines []string
    var current string

    for _, word := range words {
        if len(current)+1+len(word) > width {
            lines = append(lines, current)
            current = word
        } else if current == "" {
            current = word
        } else {
            current += " " + word
        }
    }
    if current != "" {
        lines = append(lines, current)
    }

    return lines
}

func (r *DebateRenderer) getPhaseIcon(phase string) string {
    icons := map[string]string{
        "initial":    "🔍",
        "validation": "✓",
        "polish":     "✨",
        "conclusion": "📜",
    }
    if icon, ok := icons[phase]; ok {
        return icon
    }
    return "●"
}

func (r *DebateRenderer) getPhaseName(phase string) string {
    names := map[string]string{
        "initial":    "INITIAL RESPONSE",
        "validation": "VALIDATION",
        "polish":     "POLISH & IMPROVE",
        "conclusion": "FINAL CONCLUSION",
    }
    if name, ok := names[phase]; ok {
        return name
    }
    return strings.ToUpper(phase)
}
```

---

## Progress Bars

### Styles

| Style | Example |
|-------|---------|
| `ascii` | `[========>   ] 75%` |
| `unicode` | `[████████░░░] 75%` |
| `block` | `▓▓▓▓▓▓▓▓░░░ 75%` |
| `dots` | `●●●●●●●●○○○ 75%` |
| `braille` | `⣿⣿⣿⣿⣿⣿⣿⣿⣀⣀⣀ 75%` |

### Implementation

```typescript
// packages/ui/src/progress_bar.ts

export type ProgressStyle = 'ascii' | 'unicode' | 'block' | 'dots' | 'braille';

export interface ProgressBarOptions {
  style: ProgressStyle;
  width: number;
  showPercentage: boolean;
  showLabel: boolean;
  color: boolean;
}

export class ProgressBar {
  private options: ProgressBarOptions;

  constructor(options: Partial<ProgressBarOptions> = {}) {
    this.options = {
      style: 'unicode',
      width: 20,
      showPercentage: true,
      showLabel: false,
      color: true,
      ...options,
    };
  }

  render(percent: number, label?: string): string {
    const bar = this.renderBar(percent);
    let result = bar;

    if (this.options.showPercentage) {
      result += ` ${Math.round(percent)}%`;
    }

    if (this.options.showLabel && label) {
      result = `${label}: ${result}`;
    }

    return result;
  }

  private renderBar(percent: number): string {
    const filled = Math.round((percent / 100) * this.options.width);
    const empty = this.options.width - filled;

    switch (this.options.style) {
      case 'ascii':
        return '[' + '='.repeat(Math.max(0, filled - 1)) +
               (filled > 0 ? '>' : '') +
               ' '.repeat(empty) + ']';

      case 'unicode':
        return '[' + '█'.repeat(filled) + '░'.repeat(empty) + ']';

      case 'block':
        return '▓'.repeat(filled) + '░'.repeat(empty);

      case 'dots':
        return '●'.repeat(filled) + '○'.repeat(empty);

      case 'braille':
        return '⣿'.repeat(filled) + '⣀'.repeat(empty);

      default:
        return '[' + '█'.repeat(filled) + '░'.repeat(empty) + ']';
    }
  }
}

// Animated progress bar
export class AnimatedProgressBar {
  private bar: ProgressBar;
  private spinner = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];
  private spinnerIndex = 0;
  private interval: NodeJS.Timer | null = null;

  constructor(options?: Partial<ProgressBarOptions>) {
    this.bar = new ProgressBar(options);
  }

  start(callback: (update: (percent: number, label?: string) => void) => void): void {
    this.interval = setInterval(() => {
      this.spinnerIndex = (this.spinnerIndex + 1) % this.spinner.length;
    }, 100);

    callback((percent, label) => {
      process.stdout.write('\r' + this.render(percent, label));
    });
  }

  render(percent: number, label?: string): string {
    const spinner = percent < 100 ? this.spinner[this.spinnerIndex] + ' ' : '✓ ';
    return spinner + this.bar.render(percent, label);
  }

  stop(): void {
    if (this.interval) {
      clearInterval(this.interval);
      this.interval = null;
    }
    console.log(); // New line after progress bar
  }
}
```

---

## Notifications

### Types

| Type | Icon | Use Case |
|------|------|----------|
| `info` | ℹ️ | General information |
| `success` | ✓ | Completed operations |
| `warning` | ⚠️ | Non-critical issues |
| `error` | ✗ | Failures |
| `debate` | 🎭 | Debate events |
| `task` | ⚡ | Task events |

### Implementation

```typescript
// packages/ui/src/notifications.ts

export type NotificationType = 'info' | 'success' | 'warning' | 'error' | 'debate' | 'task';

export interface Notification {
  type: NotificationType;
  title: string;
  body?: string;
  timestamp?: Date;
  duration?: number;
}

export class NotificationRenderer {
  private icons: Record<NotificationType, string> = {
    info: 'ℹ️',
    success: '✓',
    warning: '⚠️',
    error: '✗',
    debate: '🎭',
    task: '⚡',
  };

  private colors: Record<NotificationType, string> = {
    info: '\x1b[36m',    // Cyan
    success: '\x1b[32m', // Green
    warning: '\x1b[33m', // Yellow
    error: '\x1b[31m',   // Red
    debate: '\x1b[35m',  // Magenta
    task: '\x1b[34m',    // Blue
  };

  private reset = '\x1b[0m';

  render(notification: Notification, useColor = true): string {
    const icon = this.icons[notification.type];
    const color = useColor ? this.colors[notification.type] : '';
    const reset = useColor ? this.reset : '';

    let output = `${color}${icon} ${notification.title}${reset}`;

    if (notification.body) {
      output += `\n   ${notification.body}`;
    }

    if (notification.timestamp) {
      const time = notification.timestamp.toLocaleTimeString();
      output += ` ${color}(${time})${reset}`;
    }

    return output;
  }

  toast(notification: Notification): void {
    console.log(this.render(notification));

    if (notification.duration) {
      setTimeout(() => {
        // Clear the notification (move cursor up and clear line)
        process.stdout.write('\x1b[1A\x1b[2K');
      }, notification.duration);
    }
  }
}
```

---

## Resource Gauges

### Display

```
┌─────────────────────────────────────────┐
│ Resources                               │
├─────────────────────────────────────────┤
│ CPU:  [████████░░░░░░░░] 52%            │
│ MEM:  [██████████░░░░░░] 64%  2.1GB     │
│ I/O:  [███░░░░░░░░░░░░░] 18%  45MB/s    │
│ NET:  [█████░░░░░░░░░░░] 32%  12MB/s    │
└─────────────────────────────────────────┘
```

### Implementation

```typescript
// packages/ui/src/resource_gauge.ts

export interface ResourceStats {
  cpu: number;        // 0-100
  memory: number;     // 0-100
  memoryUsed: string; // e.g., "2.1GB"
  io: number;         // 0-100
  ioRate: string;     // e.g., "45MB/s"
  network: number;    // 0-100
  networkRate: string; // e.g., "12MB/s"
}

export class ResourceGauge {
  private width = 16;

  render(stats: ResourceStats): string {
    const lines: string[] = [];

    lines.push('┌─────────────────────────────────────────┐');
    lines.push('│ Resources                               │');
    lines.push('├─────────────────────────────────────────┤');

    lines.push(this.renderRow('CPU', stats.cpu));
    lines.push(this.renderRow('MEM', stats.memory, stats.memoryUsed));
    lines.push(this.renderRow('I/O', stats.io, stats.ioRate));
    lines.push(this.renderRow('NET', stats.network, stats.networkRate));

    lines.push('└─────────────────────────────────────────┘');

    return lines.join('\n');
  }

  private renderRow(label: string, percent: number, extra?: string): string {
    const bar = this.progressBar(percent);
    const pct = `${percent}%`.padStart(3);
    const extraStr = extra ? `  ${extra}` : '';

    const content = `${label.padEnd(4)} [${bar}] ${pct}${extraStr}`;
    return `│ ${content.padEnd(39)} │`;
  }

  private progressBar(percent: number): string {
    const filled = Math.round((percent / 100) * this.width);
    const empty = this.width - filled;
    return '█'.repeat(filled) + '░'.repeat(empty);
  }
}
```

---

## Configuration

### Full UI Configuration

```json
{
  "ui": {
    "renderStyle": "theater",
    "progressStyle": "unicode",
    "colorScheme": "256",
    "width": 65,

    "debate": {
      "showConfidenceBars": true,
      "showPhaseIndicators": true,
      "showRoundProgress": true,
      "animate": true
    },

    "progress": {
      "showPercentage": true,
      "showLabel": true,
      "showSpinner": true
    },

    "notifications": {
      "showTimestamp": true,
      "duration": 5000,
      "position": "bottom"
    },

    "resources": {
      "showCPU": true,
      "showMemory": true,
      "showIO": true,
      "showNetwork": true,
      "refreshInterval": 1000
    }
  }
}
```
