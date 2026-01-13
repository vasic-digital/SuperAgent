# HelixAgent Video Course Materials

This directory contains comprehensive video course materials for HelixAgent training.

## Directory Structure

```
docs/courses/
|-- COURSE_OUTLINE.md          # Complete 11-module course outline
|-- INSTRUCTOR_GUIDE.md        # Guide for course instructors
|-- README.md                  # This file
|-- slides/                    # Presentation slides for each module
|   |-- MODULE_01_INTRODUCTION.md
|   |-- MODULE_02_INSTALLATION.md
|   |-- MODULE_03_CONFIGURATION.md
|   |-- MODULE_04_PROVIDERS.md
|   |-- MODULE_05_ENSEMBLE.md
|   |-- MODULE_06_AI_DEBATE.md
|   |-- MODULE_07_PLUGINS.md
|   |-- MODULE_08_PROTOCOLS.md
|   |-- MODULE_09_OPTIMIZATION.md
|   |-- MODULE_10_SECURITY.md
|   |-- MODULE_11_TESTING_CICD.md
|-- labs/                      # Hands-on lab exercises
|   |-- README.md
|   |-- LAB_01_GETTING_STARTED.md
|   |-- LAB_02_PROVIDER_SETUP.md
|   |-- LAB_03_AI_DEBATE.md
|-- reference/                 # Quick reference materials
|   |-- QUICK_REFERENCE.md
|-- assessments/               # Quizzes and certifications
    |-- QUIZ_MODULE_1_3.md
    |-- QUIZ_MODULE_4_6.md
```

## Course Overview

**Title**: Mastering HelixAgent: Multi-Provider AI Orchestration
**Duration**: 10+ hours across 11 comprehensive modules
**Target Audience**: Developers, DevOps engineers, AI engineers, technical decision-makers
**Skill Level**: Beginner to Advanced

## Modules

| Module | Title | Duration |
|--------|-------|----------|
| 1 | Introduction to HelixAgent | 45 min |
| 2 | Installation and Setup | 60 min |
| 3 | Configuration | 60 min |
| 4 | LLM Provider Integration | 75 min |
| 5 | Ensemble Strategies | 60 min |
| 6 | AI Debate System | 90 min |
| 7 | Plugin Development | 75 min |
| 8 | MCP/LSP Integration | 60 min |
| 9 | Optimization Features | 75 min |
| 10 | Security Best Practices | 60 min |
| 11 | Testing and CI/CD | 75 min |

## Using the Materials

### Presentation Slides

Each module has a corresponding slide deck in the `slides/` directory. The slides are written in Markdown format and include:

- Title slides
- Learning objectives
- Content slides with code examples
- Visual diagrams (described in text)
- Hands-on lab exercises
- Speaker notes

### Converting to Presentation Format

The Markdown slides can be converted to various presentation formats:

**Using Marp (recommended):**
```bash
# Install Marp CLI
npm install -g @marp-team/marp-cli

# Convert to HTML
marp slides/MODULE_01_INTRODUCTION.md -o MODULE_01.html

# Convert to PDF
marp slides/MODULE_01_INTRODUCTION.md -o MODULE_01.pdf

# Convert to PowerPoint
marp slides/MODULE_01_INTRODUCTION.md -o MODULE_01.pptx
```

**Using Pandoc:**
```bash
pandoc slides/MODULE_01_INTRODUCTION.md -o MODULE_01.pptx
```

### Recording Guidelines

See the video production setup guide at:
- `/docs/marketing/VIDEO_PRODUCTION_SETUP.md`

### Related Resources

- **API Documentation**: `/docs/api/`
- **Feature Documentation**: `/docs/features/`
- **Deployment Guides**: `/docs/deployment/`
- **Optimization Guides**: `/docs/optimization/`

## Certification Path

The course supports a 4-level certification path:

1. **Level 1: HelixAgent Fundamentals** - Modules 1-3
2. **Level 2: Provider Expert** - Modules 4-6
3. **Level 3: Advanced Practitioner** - Modules 7-9
4. **Level 4: Master Engineer** - Modules 10-11

## Hands-On Labs

The course includes 5 comprehensive hands-on labs:

| Lab | Title | Duration | Difficulty |
|-----|-------|----------|------------|
| 1 | [Getting Started](labs/LAB_01_GETTING_STARTED.md) | 45 min | Beginner |
| 2 | [Provider Setup](labs/LAB_02_PROVIDER_SETUP.md) | 60 min | Intermediate |
| 3 | [AI Debate](labs/LAB_03_AI_DEBATE.md) | 75 min | Intermediate |
| 4 | MCP Integration | 60 min | Intermediate |
| 5 | Production Deployment | 120 min | Advanced |

## Reference Materials

- **[Quick Reference Card](reference/QUICK_REFERENCE.md)** - Essential commands and API endpoints
- **[Instructor Guide](INSTRUCTOR_GUIDE.md)** - Delivery guidelines for trainers

## Assessments

Certification assessments are provided for each level:

| Assessment | Modules | Questions | Passing |
|------------|---------|-----------|---------|
| [Level 1 Quiz](assessments/QUIZ_MODULE_1_3.md) | 1-3 | 25 | 80% |
| [Level 2 Quiz](assessments/QUIZ_MODULE_4_6.md) | 4-6 | 30 | 80% |
| Level 3 Quiz | 7-9 | 30 | 80% |
| Level 4 Quiz | 10-11 | 25 | 80% |

## Contributing

To update or improve course materials:

1. Edit the corresponding Markdown file
2. Test slide rendering with Marp
3. Update COURSE_OUTLINE.md if adding new content
4. Submit a pull request

## Version History

- **v2.1.0** (January 2026) - Added labs, assessments, quick reference, instructor guide
- **v2.0.0** (January 2026) - Complete 11-module curriculum
- **v1.0.0** (December 2024) - Initial 6-module course

---

*For questions or feedback, please open an issue in the repository.*
