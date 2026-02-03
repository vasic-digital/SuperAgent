# HelixAgent Video Course Production Guide

## Overview

This guide provides comprehensive video production guidelines for creating the HelixAgent video course. Following these standards ensures consistent, high-quality educational content across all 74 videos.

---

## Video Format Standards

### Technical Specifications

| Attribute | Standard | Notes |
|-----------|----------|-------|
| Resolution | 1920x1080 (1080p) | Minimum for clarity |
| Frame Rate | 30 fps | Consistent throughout |
| Aspect Ratio | 16:9 | Standard widescreen |
| Video Codec | H.264 or H.265 | For compatibility |
| Audio Codec | AAC | 48kHz sample rate |
| Audio Bitrate | 192 kbps minimum | Clear voice quality |
| Video Bitrate | 8-12 Mbps | Balance quality/size |
| File Format | MP4 | Universal compatibility |

### Recording Setup

**Screen Recording**:
- Use OBS Studio, Camtasia, or ScreenFlow
- Capture at native resolution (1920x1080 or higher)
- Include system audio for demos
- Use consistent screen layout

**Camera (Optional)**:
- 1080p minimum, 4K preferred
- Good lighting (ring light or softbox)
- Neutral background
- Camera at eye level

**Audio**:
- USB condenser microphone or XLR setup
- Pop filter required
- Quiet recording environment
- No background music during instruction

---

## Video Length Guidelines

### Optimal Duration by Content Type

| Content Type | Optimal Length | Maximum |
|--------------|----------------|---------|
| Introduction/Welcome | 5-8 min | 10 min |
| Concept Explanation | 10-15 min | 18 min |
| Code Walkthrough | 12-18 min | 20 min |
| Live Demo | 8-15 min | 18 min |
| Lab Overview | 5-10 min | 12 min |
| Module Summary | 3-5 min | 8 min |

### Pacing Guidelines

- **Speaking pace**: 120-150 words per minute
- **Pause after key points**: 2-3 seconds
- **Demo operations**: Wait for completion before proceeding
- **Code typing**: Real-time or slightly accelerated (1.25x max)
- **Transitions**: 1-2 second fade between sections

---

## Visual Style Guide

### Screen Layout

**Code Demonstration Layout**:
```
+----------------------------------+
|  Terminal/IDE        |  Camera   |
|  (80% width)         |  (20%)    |
|                      |  [opt]    |
+----------------------------------+
|  Lower Third: Title/Topic        |
+----------------------------------+
```

**Slide Presentation Layout**:
```
+----------------------------------+
|                                  |
|         Slide Content            |
|                                  |
+----------------------------------+
|  Lower Third: Speaker Info       |
+----------------------------------+
```

**Split Screen (Demo + Explanation)**:
```
+----------------------------------+
|  Demo/Terminal  |  Slides/Notes  |
|    (60%)        |     (40%)      |
+----------------------------------+
```

### Color Palette

| Element | Color | Hex |
|---------|-------|-----|
| Primary | Blue | #2563EB |
| Secondary | Purple | #7C3AED |
| Accent | Green | #10B981 |
| Background | Dark | #1F2937 |
| Text | White | #F9FAFB |
| Code Background | Darker | #111827 |

### Typography

| Element | Font | Size |
|---------|------|------|
| Titles | Inter Bold | 48-64px |
| Headings | Inter Semi-Bold | 32-40px |
| Body Text | Inter Regular | 24-28px |
| Code | JetBrains Mono | 18-22px |
| Captions | Inter Regular | 18-20px |

### IDE/Terminal Theme

**Recommended**:
- VS Code with "One Dark Pro" or "GitHub Dark" theme
- Terminal with dark background, light text
- Font: JetBrains Mono or Fira Code
- Font size: 16-18px for readability

**Settings**:
```json
{
  "editor.fontSize": 16,
  "editor.lineHeight": 1.6,
  "terminal.integrated.fontSize": 16,
  "editor.minimap.enabled": false,
  "editor.wordWrap": "on"
}
```

---

## Content Structure

### Standard Video Structure

1. **Hook** (10-20 seconds)
   - Engaging question or statement
   - What learners will achieve

2. **Introduction** (30-60 seconds)
   - Video title and module context
   - Prerequisites reminder
   - Learning objectives (3-5 bullet points)

3. **Main Content** (80% of video)
   - Logical progression of topics
   - Mix of explanation and demonstration
   - Regular checkpoints and summaries

4. **Summary** (30-60 seconds)
   - Recap key points
   - Preview next video
   - Call to action (lab, quiz, etc.)

5. **Outro** (10-15 seconds)
   - Consistent end screen
   - Links to resources

### Module Structure

Each module follows this structure:

1. **Module Introduction Video**
   - Overview of entire module
   - Learning objectives
   - Prerequisites
   - What to expect

2. **Concept Videos** (2-4 per module)
   - Theory and explanation
   - Architecture diagrams
   - Code examples

3. **Demo Videos** (1-3 per module)
   - Live coding
   - Configuration walkthrough
   - Real-world scenarios

4. **Lab Introduction Video**
   - Lab objectives
   - Setup requirements
   - Expected outcomes

5. **Module Summary Video** (optional)
   - Key takeaways
   - Common pitfalls
   - Next steps

---

## Recording Guidelines

### Pre-Recording Checklist

**Environment**:
- [ ] Close unnecessary applications
- [ ] Disable notifications
- [ ] Clean desktop (remove personal files)
- [ ] Terminal history cleared
- [ ] Browser bookmarks hidden
- [ ] Set Do Not Disturb mode

**Technical**:
- [ ] Test microphone levels (-12dB to -6dB peaks)
- [ ] Verify screen recording settings
- [ ] Check storage space (min 10GB free)
- [ ] Test demo environment is working
- [ ] HelixAgent running and healthy

**Content**:
- [ ] Review script/outline
- [ ] Open all needed files/tabs
- [ ] Practice demo commands
- [ ] Prepare fallback examples

### Recording Best Practices

**Voice**:
- Speak clearly and at moderate pace
- Use natural, conversational tone
- Avoid filler words ("um", "uh", "like")
- Pause before important points
- Maintain consistent energy level

**Screen Actions**:
- Move cursor slowly and deliberately
- Highlight what you're discussing
- Zoom in for small text/details
- Wait for commands to complete
- Use keyboard shortcuts with explanations

**Code Demonstrations**:
- Type at readable pace (or use pre-typed snippets)
- Explain as you type
- Highlight relevant code sections
- Show both input and output
- Handle errors gracefully

### Common Mistakes to Avoid

- Rushing through complex topics
- Forgetting to explain prerequisites
- Mouse moving erratically
- Typing too fast to follow
- Not testing demos before recording
- Inconsistent audio levels
- Using jargon without explanation
- Skipping error handling examples

---

## Post-Production

### Editing Guidelines

**Required Edits**:
- Remove long pauses (>3 seconds)
- Cut mistakes and retakes
- Add intro/outro animations
- Include lower thirds for sections
- Add callouts for important points

**Optional Enhancements**:
- Zoom effects on important areas
- Picture-in-picture for camera
- Animated transitions
- Chapter markers

**Audio Processing**:
- Normalize audio levels
- Remove background noise
- Apply compression if needed
- Ensure consistent volume across videos

### Quality Checklist

Before Publishing:

- [ ] Audio clear throughout
- [ ] No dead air >3 seconds
- [ ] All code visible and readable
- [ ] Demo commands work correctly
- [ ] Captions synchronized (if added)
- [ ] Chapter markers accurate
- [ ] Thumbnail created
- [ ] Metadata complete

---

## Accessibility

### Requirements

**Captions**:
- Auto-generate initial captions
- Review and correct errors
- Include speaker identification
- Describe visual elements

**Visual**:
- Sufficient color contrast (WCAG AA)
- Don't rely solely on color for meaning
- Include alt text for diagrams
- Describe what's on screen

**Audio**:
- Clear narration throughout
- Describe visual-only content
- Announce section transitions

### Caption Format

```
00:00:00,000 --> 00:00:05,000
Welcome to Module 1 of the HelixAgent
video course.

00:00:05,500 --> 00:00:10,000
Today we'll explore the architecture
and core concepts.
```

---

## Demo Environment Setup

### Standard Demo Configuration

**HelixAgent Instance**:
```bash
# Start with demo profile
PORT=7061 GIN_MODE=debug make run-dev

# Verify health
curl http://localhost:7061/health
```

**Required Services**:
```bash
# Start infrastructure
make infra-start

# Verify services
docker-compose ps
```

**Demo API Keys** (use test accounts):
- Set up dedicated demo API keys
- Never show real production keys
- Use environment variables

### Demo Data Preparation

**Pre-loaded Data**:
- Sample prompts for each topic
- Test configurations ready
- Example plugin code available
- Challenge scripts tested

**Backup Plans**:
- Screenshots of expected output
- Pre-recorded fallback clips
- Alternative demo scenarios

---

## File Naming Convention

### Video Files

Format: `M{module}_{video}_title.mp4`

Examples:
```
M01_01_course_welcome.mp4
M01_02_what_is_helixagent.mp4
M06_03_debate_strategies.mp4
M12_05_strict_validation.mp4
```

### Supporting Files

```
M01_01_course_welcome_thumbnail.png
M01_01_course_welcome_captions.srt
M01_01_course_welcome_script.md
```

---

## Production Schedule

### Per-Video Timeline

| Phase | Duration | Activities |
|-------|----------|------------|
| Preparation | 1-2 hours | Script review, demo setup |
| Recording | 1.5x video length | Raw recording |
| Editing | 2-3x video length | Post-production |
| Review | 30-60 min | Quality check |
| Revisions | 1-2 hours | Address feedback |

### Module Production Timeline

| Phase | Duration |
|-------|----------|
| Module Planning | 1 day |
| Script Finalization | 2 days |
| Recording All Videos | 2-3 days |
| Editing | 3-4 days |
| Review and QA | 1-2 days |
| Total per Module | 10-12 days |

---

## Equipment Recommendations

### Minimum Setup

| Equipment | Recommendation | Budget |
|-----------|----------------|--------|
| Microphone | Blue Yeti or similar | $100-150 |
| Screen Recording | OBS Studio (free) | $0 |
| Editing | DaVinci Resolve (free) | $0 |
| Lighting | Ring light | $30-50 |

### Professional Setup

| Equipment | Recommendation | Budget |
|-----------|----------------|--------|
| Microphone | Shure SM7B + interface | $400-500 |
| Camera | Sony ZV-1 or similar | $600-800 |
| Screen Recording | Camtasia or ScreenFlow | $250-300 |
| Editing | Adobe Premiere Pro | $20/month |
| Lighting | Elgato Key Light | $200 |

---

## Review Process

### Internal Review

1. **Technical Review**
   - Code accuracy
   - Demo correctness
   - Command verification

2. **Content Review**
   - Accuracy of information
   - Completeness of coverage
   - Appropriate difficulty level

3. **Quality Review**
   - Audio/video quality
   - Pacing and flow
   - Accessibility compliance

### Feedback Integration

- Collect reviewer comments
- Prioritize critical fixes
- Document minor improvements
- Track changes made

---

## Resources

### Tools

- **Recording**: OBS Studio, Camtasia, ScreenFlow
- **Editing**: DaVinci Resolve, Adobe Premiere, Final Cut Pro
- **Thumbnails**: Canva, Figma, Adobe Photoshop
- **Captions**: YouTube auto-generate, Rev.com, Descript

### Templates

- Thumbnail templates in `/docs/marketing/templates/`
- Lower third designs in brand guidelines
- End screen template for all videos

### Support

- Production questions: video@helixagent.ai
- Technical accuracy: training@helixagent.ai
- Accessibility requirements: accessibility@helixagent.ai

---

*Production Guide Version: 1.0.0*
*Last Updated: February 2026*
