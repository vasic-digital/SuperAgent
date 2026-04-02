# UI/UX Pro Max User Guide

## Overview

UI/UX Pro Max is an AI skill that provides design intelligence for building professional UI/UX across multiple platforms and frameworks. It integrates with popular AI coding assistants (Cursor, Claude Code, GitHub Copilot, Windsurf, etc.) to generate high-quality, aesthetically pleasing user interfaces with proper design systems, color theory, typography, and UX best practices.

**Key Features:**
- 57+ UI styles and design patterns
- 95+ color schemes and palettes
- 56+ typography pairings
- 98+ UX best practices
- Multi-platform support (25+ AI assistants)
- Template-based generation
- Design system automation
- Pre-delivery quality checks
- Offline installation support

---

## Installation Methods

### Method 1: Using CLI (Recommended)

Requirements: Node.js 16+, Python 3.x

```bash
# Install CLI globally
npm install -g uipro-cli

# Verify installation
uipro --version
```

### Method 2: Using pnpm/yarn

```bash
# Using pnpm
pnpm add -g uipro-cli

# Using yarn
yarn global add uipro-cli
```

### Method 3: Using npx (No Install)

```bash
# Run without installing
npx uipro-cli init --ai claude
```

### Method 4: Using Claude Marketplace (Claude Code)

```bash
# In Claude Code terminal
/plugin marketplace add nextlevelbuilder/ui-ux-pro-max-skill
/plugin install ui-ux-pro-max@ui-ux-pro-max-skill
```

### Method 5: Clone and Manual Install

```bash
# Clone repository
git clone https://github.com/nextlevelbuilder/ui-ux-pro-max-skill.git
cd ui-ux-pro-max-skill

# Install CLI dependencies
cd cli && npm install

# Build
npm run build

# Link globally
npm link
```

---

## Quick Start

### 1. Verify Installation

```bash
uipro --version
# Output: uipro-cli v1.2.3
```

### 2. Initialize in Your Project

```bash
# Navigate to your project
cd /path/to/your/project

# Initialize for your AI assistant
uipro init --ai claude      # Claude Code
uipro init --ai cursor      # Cursor
uipro init --ai copilot     # GitHub Copilot
uipro init --ai windsurf    # Windsurf
```

### 3. Start Using

In your AI assistant chat:

```
Build a landing page for my SaaS product
```

The skill automatically activates on UI/UX requests!

---

## CLI Commands Reference

### Core Commands

| Command | Description |
|---------|-------------|
| `uipro --version` | Show CLI version |
| `uipro --help` | Show help |
| `uipro init` | Initialize UI/UX Pro Max |
| `uipro update` | Update to latest version |
| `uipro versions` | List available versions |
| `uipro uninstall` | Remove installation |

### Init Command Options

| Flag | Description | Example |
|------|-------------|---------|
| `--ai <assistant>` | Target AI assistant | `--ai claude` |
| `--global` | Install globally | `--global` |
| `--offline` | Offline mode | `--offline` |
| `--version <ver>` | Specific version | `--version v1.2.0` |
| `--legacy` | Legacy ZIP mode | `--legacy` |

### Supported AI Assistants

```bash
# Initialize for specific assistants
uipro init --ai claude        # Claude Code
uipro init --ai cursor        # Cursor
uipro init --ai windsurf      # Windsurf
uipro init --ai antigravity   # Antigravity (.agent + .shared)
uipro init --ai copilot       # GitHub Copilot
uipro init --ai kiro          # Kiro
uipro init --ai codex         # Codex CLI
uipro init --ai qoder         # Qoder
uipro init --ai roocode       # Roo Code
uipro init --ai gemini        # Gemini CLI
uipro init --ai trae          # Trae
uipro init --ai opencode      # OpenCode
uipro init --ai continue      # Continue
uipro init --ai codebuddy     # CodeBuddy
uipro init --ai droid         # Droid (Factory)
uipro init --ai kilocode      # KiloCode
uipro init --ai warp          # Warp
uipro init --ai augment       # Augment
uipro init --ai all           # All assistants
```

### Global Installation

```bash
# Install globally for all projects
uipro init --ai claude --global    # Install to ~/.claude/skills/
uipro init --ai cursor --global    # Install to ~/.cursor/skills/
```

### Update Commands

```bash
# Update CLI
npm update -g uipro-cli

# Update skill in project
uipro update

# Update to specific version
uipro init --version v1.2.0
```

### Uninstall Commands

```bash
# Auto-detect and uninstall
uipro uninstall

# Uninstall specific platform
uipro uninstall --ai claude

# Uninstall global installation
uipro uninstall --global
```

---

## Project Structure

After initialization:

```
my-project/
├── .claude/                     # Claude Code skills
│   └── skills/
│       ├── ui-ux-pro-max/
│       │   ├── SKILL.md         # Skill definition
│       │   └── quick-reference.md
│       └── search.py            # Design search engine
│
├── .cursor/                     # Cursor rules
│   └── skills/
│       └── ui-ux-pro-max/
│
├── .shared/                     # Shared data
│   └── ui-ux-pro-max/
│       ├── data/                # Design databases
│       │   ├── products.csv     # Product types
│       │   ├── styles.csv       # UI styles
│       │   ├── colors.csv       # Color palettes
│       │   ├── typography.csv   # Font pairings
│       │   └── stacks/          # Tech stack configs
│       │       ├── html-tailwind.csv
│       │       ├── react.csv
│       │       └── nextjs.csv
│       └── scripts/             # Python scripts
│           ├── search.py
│           ├── core.py
│           └── design_system.py
│
└── .windsurf/                   # Windsurf config
    └── skills/
        └── ui-ux-pro-max/
```

---

## Configuration

### Skill Mode Configuration

**Claude Code (Automatic Activation):**
```markdown
<!-- .claude/skills/ui-ux-pro-max/SKILL.md -->
# UI/UX Pro Max Skill

## Activation
Auto-activates on UI/UX requests (build, design, create, implement, review, fix, improve)

## Capabilities
- Design system generation
- Color palette selection
- Typography pairing
- Component implementation
- Responsive design
- Accessibility compliance
```

**Cursor/Other (Slash Command):**
```markdown
<!-- .cursor/skills/ui-ux-pro-max/rules.md -->
# UI/UX Pro Max

## Commands
/ui-ux-pro-max <request> - Execute design workflow

## Workflow
1. Analyze request
2. Search design database
3. Generate design system
4. Implement code
5. Quality check
```

### Design Database Structure

```csv
# data/styles.csv
id,name,category,description,use_case
dark_modern,Dark Modern,Dark,Clean dark theme with subtle gradients,SaaS dashboards
minimal_light,Minimal Light,Light,Whitespace-focused minimal design,Blogs portfolios
fintech_pro,Fintech Pro,Dark,Professional finance aesthetic,Banking fintech
```

```csv
# data/colors.csv
id,name,primary,secondary,accent,background
ocean_blue,#0ea5e9,#0284c7,#38bdf8,#0f172a
dark_purple,#7c3aed,#6d28d9,#a78bfa,#1e1b4b
forest_green,#059669,#047857,#34d399,#064e3b
```

```csv
# data/typography.csv
id,name,heading_font,body_font,monospace_font
inter_system,Inter System,Inter,Inter,JetBrains Mono
roboto_clean,Roboto Clean,Roboto,Roboto,Menlo
geist_modern,Geist Modern,Geist,Geist,Geist Mono
```

### Custom Configuration

Create `ui-ux-pro-max.config.json`:

```json
{
  "design_preferences": {
    "default_style": "dark_modern",
    "default_color_scheme": "ocean_blue",
    "default_typography": "inter_system",
    "accessibility_level": "AA",
    "responsive_breakpoints": [640, 768, 1024, 1280]
  },
  "tech_stack": {
    "css_framework": "tailwind",
    "component_library": "shadcn",
    "animation_library": "framer-motion"
  },
  "output": {
    "format": "tsx",
    "include_types": true,
    "prettier_format": true
  },
  "quality_gates": {
    "check_contrast": true,
    "check_responsive": true,
    "check_accessibility": true,
    "check_performance": true
  }
}
```

---

## Usage Examples

### Example 1: Basic Landing Page

**User Request:**
```
Build a landing page for my SaaS product
```

**Skill Workflow:**
1. Identifies "SaaS landing page" use case
2. Searches design database for matching styles
3. Selects appropriate color scheme (ocean_blue, fintech_pro)
4. Chooses typography (inter_system, geist_modern)
5. Generates design system:
   - Color variables
   - Typography scale
   - Spacing system
   - Component specs
6. Implements code with:
   - Hero section
   - Feature grid
   - Pricing cards
   - CTA sections
   - Footer
7. Runs quality checks

### Example 2: Healthcare Dashboard

**User Request:**
```
Create a dashboard for healthcare analytics with dark theme
```

**AI Interaction:**
```
I'll use ui-ux-pro-max to design a healthcare analytics dashboard.

Design System Selected:
- Style: Dark Medical (clean, professional, trustworthy)
- Colors: Medical Blue (#0284c7) + Calming Teal + Alert Red
- Typography: Inter (clean, readable)
- Components: Data tables, charts, KPI cards, filters

Implementing:
✓ Sidebar navigation
✓ KPI summary cards
✓ Patient data table
✓ Analytics charts
✓ Alert notifications
✓ Responsive layout
```

### Example 3: E-commerce Mobile App

**User Request:**
```
Design a mobile app UI for e-commerce
```

**With Kiro/Copilot:**
```
/ui-ux-pro-max Design a mobile e-commerce app with product browsing, cart, and checkout

Design System:
- Platform: iOS/Android
- Style: Modern Shopping
- Colors: Vibrant primary + neutral grays
- Typography: SF Pro / Roboto

Screens:
1. Home/Browse
2. Product Detail
3. Shopping Cart
4. Checkout Flow
5. Order Confirmation
```

### Example 4: Fintech Banking App

**User Request:**
```
Build a fintech banking app with dark theme
```

**Skill Output:**
```
Design System Generated:
- Theme: Dark Fintech Pro
- Primary: #0ea5e9 (Trust Blue)
- Accent: #10b981 (Success Green)
- Alert: #ef4444 (Danger Red)
- Typography: Geist (Modern, professional)

Features Implemented:
✓ Account overview cards
✓ Transaction history table
✓ Transfer flow
✓ Card management
✓ Security settings
✓ Dark mode optimized
```

### Example 5: Portfolio Website

**User Request:**
```
Design a portfolio website with dark mode support
```

**Implementation:**
```
Design System:
- Style: Creative Portfolio
- Colors: Deep purple primary + gradient accents
- Fonts: Playfair Display (headings) + Inter (body)

Sections:
✓ Hero with animated intro
✓ Project showcase grid
✓ Skills visualization
✓ About section
✓ Contact form
✓ Dark/light mode toggle
```

### Example 6: Component Library

**User Request:**
```
Create a design system component library
```

**Generated Components:**
```
Component Library:
├── Buttons
│   ├── Primary, Secondary, Ghost
│   ├── Sizes: sm, md, lg
│   └── States: default, hover, active, disabled
├── Inputs
│   ├── Text, Password, Email
│   ├── Search, Textarea
│   └── With validation states
├── Cards
│   ├── Basic, Feature, Pricing
│   └── With hover effects
├── Navigation
│   ├── Header, Sidebar, Tabs
│   └── Breadcrumbs
└── Feedback
    ├── Alerts, Toasts, Modals
    └── Progress indicators
```

---

## TUI / Interactive Features

### Design Search Engine

```bash
# Search design database
python .shared/ui-ux-pro-max/scripts/search.py "healthcare dashboard"

# Output:
# Matching Styles:
# - Dark Medical (score: 0.95)
# - Healthcare Pro (score: 0.88)
# - Clinical Clean (score: 0.82)
#
# Matching Colors:
# - Medical Blue (score: 0.91)
# - Trust Teal (score: 0.85)
```

### Style Preview

```bash
# Preview style
uipro preview dark_modern

# Opens browser with:
# - Color palette
# - Typography samples
# - Component examples
```

### Configuration Wizard

```bash
# Interactive configuration
uipro configure

# Prompts:
# ? Select default style: Dark Modern
# ? Select color scheme: Ocean Blue
# ? Select typography: Inter System
# ? Accessibility level: AA
# ? Tech stack: React + Tailwind
```

---

## Slash Commands by Platform

### Claude Code (Auto-activation)

```
No slash command needed! Just chat naturally:

"Build a landing page for my SaaS"
"Create a dashboard for analytics"
"Design a mobile app UI"
```

### Cursor

```
/ui-ux-pro-max <request>

Examples:
/ui-ux-pro-max Build a landing page
/ui-ux-pro-max Create a dark theme dashboard
/ui-ux-pro-max Design an e-commerce product page
```

### Windsurf

```
/ui-ux-pro-max <request>
```

### Kiro

```
/ui-ux-pro-max <request>
```

### GitHub Copilot

```
/ui-ux-pro-max <request>
```

### Codex CLI

```
$ui-ux-pro-max <request>
```

---

## Troubleshooting

### Installation Issues

**Problem:** `npm install -g uipro-cli` fails

**Solutions:**
```bash
# Check Node.js version
node --version  # Requires 16+

# Use nvm to upgrade
nvm install 20
nvm use 20

# Try with sudo (if permissions issue)
sudo npm install -g uipro-cli

# Or use npx without installing
npx uipro-cli init --ai claude
```

**Problem:** Command not found after installation

**Solutions:**
```bash
# Check npm global bin
npm bin -g

# Add to PATH
export PATH="$PATH:$(npm bin -g)"

# Or use npx
npx uipro-cli --version
```

### Initialization Issues

**Problem:** `uipro init` fails

**Solutions:**
```bash
# Check you're in a project directory
pwd  # Should show your project path

# Verify Python is installed
python3 --version

# Install Python if needed
# macOS: brew install python3
# Ubuntu: sudo apt install python3

# Use offline mode
uipro init --ai claude --offline
```

**Problem:** Skills not appearing in AI assistant

**Solutions:**
```bash
# Check installation
ls -la .claude/skills/ui-ux-pro-max/

# Reinitialize
uipro uninstall
uipro init --ai claude

# For Claude: restart Claude Code
# For Cursor: reload window
```

### Python Search Script Issues

**Problem:** `search.py` errors

**Solutions:**
```bash
# Check Python dependencies
pip install pandas numpy scikit-learn

# Or install requirements
cd .shared/ui-ux-pro-max
pip install -r requirements.txt

# Verify script permissions
chmod +x scripts/search.py
```

### Design Not Applying

**Problem:** UI looks generic, not using Pro Max designs

**Solutions:**
- Ensure skill is properly installed
- Check AI assistant loaded the skill
- Try explicit invocation with slash command
- Verify design database exists:
  ```bash
  ls -la .shared/ui-ux-pro-max/data/
  ```

### Common Error Messages

| Error | Solution |
|-------|----------|
| "uipro: command not found" | Install CLI: `npm install -g uipro-cli` |
| "Python not found" | Install Python 3.x |
| "No such file or directory" | Run from project root |
| "Permission denied" | Use sudo or fix npm permissions |
| "Skill not loaded" | Restart AI assistant |
| "Design database missing" | Reinstall: `uipro init --ai <assistant>` |

### Getting Help

```bash
# Check CLI version
uipro --version

# Show help
uipro --help

# Debug mode
uipro init --ai claude --debug

# Resources
# GitHub: https://github.com/nextlevelbuilder/ui-ux-pro-max-skill
# Issues: https://github.com/nextlevelbuilder/ui-ux-pro-max-skill/issues
```

---

## Best Practices

### 1. Project Setup
- Initialize in project root
- Keep design database in version control
- Share config with team

### 2. Design Selection
- Be specific in requests
- Mention industry/domain
- Specify light/dark preference

### 3. Iteration Workflow
```
1. Initial request
2. Review design system
3. Request adjustments
4. Refine components
5. Final polish
```

### 4. Team Collaboration
- Commit `.shared/ui-ux-pro-max/` to git
- Share custom configurations
- Document design decisions

### 5. Performance
- Use offline mode for faster installs
- Cache design database locally
- Minimize AI assistant reloads

---

## Advanced Configuration

### Custom Design Database

```python
# Add custom styles
custom_styles = """
id,name,category,description
my_brand,My Brand,Custom,Our company brand guidelines
"""

# Append to data/styles.csv
```

### Extending Search

```python
# .shared/ui-ux-pro-max/scripts/custom_search.py
from search import DesignSearch

class CustomSearch(DesignSearch):
    def search_by_brand(self, brand):
        # Custom search logic
        pass
```

### CI/CD Integration

```yaml
# .github/workflows/ui-ux-check.yml
name: UI/UX Quality Check

on: [pull_request]

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Validate Design System
        run: |
          npm install -g uipro-cli
          uipro validate --strict
```

---

## Resources

- **GitHub:** https://github.com/nextlevelbuilder/ui-ux-pro-max-skill
- **CLI Docs:** https://www.mintlify.com/nextlevelbuilder/ui-ux-pro-max-skill/guides/cli-commands
- **Marketplace:** Available in Claude Code, Cursor marketplace

---

*Last Updated: April 2026*
