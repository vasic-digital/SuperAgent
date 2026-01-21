# HelixAgent Course Labs

Hands-on lab exercises for the HelixAgent training course.

## Lab Overview

| Lab | Module | Duration | Difficulty | NEW |
|-----|--------|----------|------------|-----|
| [Lab 1: Getting Started](LAB_01_GETTING_STARTED.md) | 1-2 | 45 min | Beginner | |
| [Lab 2: Provider Setup](LAB_02_PROVIDER_SETUP.md) | 4 | 60 min | Intermediate | |
| [Lab 3: AI Debate](LAB_03_AI_DEBATE.md) | 6 | 75 min | Intermediate | |
| [Lab 4: MCP Integration](LAB_04_MCP_INTEGRATION.md) | 8 | 60 min | Intermediate | |
| [Lab 5: Production Deployment](LAB_05_PRODUCTION_DEPLOYMENT.md) | 10-11 | 120 min | Advanced | |
| [Lab 6: Challenge Scripts](LAB_06_CHALLENGE_SCRIPTS.md) | 12 | 90 min | Intermediate | NEW |
| [Lab 7: MCP Tool Search](LAB_07_MCP_TOOL_SEARCH.md) | 13 | 60 min | Intermediate | NEW |
| [Lab 8: Multi-Pass Validation](LAB_08_MULTIPASS_VALIDATION.md) | 14 | 75 min | Advanced | NEW |

## Prerequisites

Before starting the labs, ensure you have:

- [ ] Git installed
- [ ] Go 1.24+ installed (for source builds)
- [ ] Docker and Docker Compose (recommended)
- [ ] Text editor (VS Code recommended)
- [ ] Terminal access
- [ ] Internet connection
- [ ] At least one LLM API key

## Getting Started

1. **Clone the repository**:
   ```bash
   git clone https://github.com/your-org/helix-agent.git
   cd helix-agent
   ```

2. **Set up environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your API keys
   ```

3. **Start the labs**:
   Begin with [Lab 1: Getting Started](LAB_01_GETTING_STARTED.md)

## Lab Completion Tracking

Track your progress through the labs:

- [ ] Lab 1: Getting Started
  - [ ] Repository cloned
  - [ ] Server running
  - [ ] Health check passing

- [ ] Lab 2: Provider Setup
  - [ ] Multiple providers configured
  - [ ] Provider health verified
  - [ ] Fallback chain working

- [ ] Lab 3: AI Debate
  - [ ] Created first debate
  - [ ] Tested different styles
  - [ ] Analyzed consensus results

- [ ] Lab 4: MCP Integration
  - [ ] MCP tools listed
  - [ ] Tool execution working
  - [ ] Resource access verified

- [ ] Lab 5: Production Deployment
  - [ ] Docker stack running
  - [ ] Monitoring configured
  - [ ] Security hardened

- [ ] Lab 6: Challenge Scripts (NEW)
  - [ ] RAGS challenge passed (100%)
  - [ ] MCPS challenge passed (100%)
  - [ ] SKILLS challenge passed (100%)
  - [ ] Understood strict validation

- [ ] Lab 7: MCP Tool Search (NEW)
  - [ ] Tool search working
  - [ ] Suggestions tested
  - [ ] Adapter search working
  - [ ] Discovery workflow created

- [ ] Lab 8: Multi-Pass Validation (NEW)
  - [ ] 4-phase debate completed
  - [ ] Phase indicators observed
  - [ ] Confidence >0.8 achieved
  - [ ] Configuration tuned

## Lab Files

Each lab contains:
- **Objectives**: What you'll learn
- **Prerequisites**: What you need
- **Exercises**: Step-by-step tasks
- **Checkpoints**: Verification points
- **Troubleshooting**: Common issues and solutions
- **Challenge**: Optional advanced exercise

## Certification Requirements

| Level | Required Labs |
|-------|---------------|
| Level 1: Fundamentals | Lab 1 |
| Level 2: Provider Expert | Labs 1-3 |
| Level 3: Advanced | Labs 1-4 |
| Level 4: Master | Labs 1-5 |
| Level 5: Challenge Expert | All labs (1-8) |

### Level 5 Special Requirements
- 100% pass rate on RAGS challenge
- 100% pass rate on MCPS challenge
- 100% pass rate on SKILLS challenge
- MCP Tool Search integration demonstration
- Multi-pass validation debate with >0.8 confidence

## Support

If you encounter issues during labs:

1. Check the troubleshooting section in each lab
2. Review the [Quick Reference](../reference/QUICK_REFERENCE.md)
3. Ask in the course discussion forum
4. Open a GitHub issue

## Contributing

To improve the labs:

1. Fork the repository
2. Make your changes
3. Test all exercises
4. Submit a pull request

---

*Labs Version: 2.0.0*
*Last Updated: January 2026*
*New Labs: 6 (Challenge Scripts), 7 (MCP Tool Search), 8 (Multi-Pass Validation)*
