# Mobile Agent - User Guide

> Mobile automation CLI for AI agents - control iOS simulators and Android devices with simple commands.

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Commands](#cli-commands)
- [Interactive Commands](#interactive-commands)
- [Configuration](#configuration)
- [Usage Examples](#usage-examples)
- [Troubleshooting](#troubleshooting)

---

## Overview

Mobile Agent (also known as `agent-mobile` or `agent-device`) is a lightweight CLI for AI-driven mobile automation on iOS and Android. It provides structured snapshots and deterministic replay capabilities, making it perfect for AI agents to understand and interact with mobile UIs.

### Key Features

- **Simple Command Model**: 7 core commands cover all basic interactions
- **Refs Pattern**: Elements assigned stable refs (@e1, @e2) for easy targeting
- **Zero Config**: Appium bundled and auto-starts
- **AI Compatible**: Designed for Claude Code, Cursor, and other AI assistants
- **Cross-Platform**: Supports iOS, Android, tvOS, macOS, and AndroidTV
- **Session-Based**: Open once, interact within session, close cleanly
- **Replay Scripts**: Save and replay automation scripts

### Available Tools

| Tool | Description |
|------|-------------|
| `agent-mobile` | iOS simulator automation (npm package) |
| `agent-device` | Full mobile automation CLI (Callstack) |

---

## Installation

### Prerequisites

- **Node.js 16+** (for agent-mobile)
- **Xcode** with iOS Simulator (for iOS)
- **Android Studio** with SDK (for Android)
- **Appium** (auto-installed with agent-mobile)

### agent-mobile Installation

```bash
# Install globally
npm install -g agent-mobile

# Or use npx
npx agent-mobile
```

### agent-device Installation

```bash
# Install globally
npm install -g agent-device

# Or use npx
npx agent-device
```

### macOS Helper (agent-device)

On macOS, `agent-device` includes a local helper package:

```bash
# Helper is built on demand for desktop permission checks
# For release distribution, use a signed/notarized helper build
# Source checkouts fall back to local Swift build
```

---

## Quick Start

### 1. Setup (agent-mobile)

```bash
# First-time setup
agent-mobile setup
```

This configures:
- Xcode command line tools
- iOS simulators
- WebDriverAgent (WDA)

### 2. Check Status

```bash
# Diagnose setup issues
agent-mobile doctor
```

### 3. Basic Workflow

```bash
# Open an app
agent-mobile open com.apple.Preferences

# Get UI snapshot
agent-mobile snapshot

# Interact with element
agent-mobile tap @e3

# Capture screenshot
agent-mobile screenshot /tmp/screen.png

# Close session
agent-mobile close
```

---

## CLI Commands

### agent-mobile Commands

#### Setup Commands

| Command | Description |
|---------|-------------|
| `setup` | First-time setup (Xcode, simulators) |
| `doctor` | Diagnose setup issues |

#### Session Commands

| Command | Description |
|---------|-------------|
| `open <bundle-id>` | Launch an iOS app by bundle ID |
| `close` | End session and close app |

#### Inspection Commands

| Command | Description |
|---------|-------------|
| `snapshot` | Get UI elements with refs |
| `screenshot <path>` | Capture the screen |

#### Interaction Commands

| Command | Description |
|---------|-------------|
| `tap <ref>` | Tap element by ref or coordinates |
| `fill <ref> <text>` | Fill text into an input |
| `swipe <direction>` | Swipe up/down/left/right |

### agent-device Commands

#### App Lifecycle

| Command | Description |
|---------|-------------|
| `open <target>` | Open app or URL |
| `close` | Close session |

#### UI Inspection

| Command | Description |
|---------|-------------|
| `snapshot` | Get accessibility tree with refs |
| `snapshot -i` | Interactive snapshot mode |
| `diff snapshot` | Compare with previous snapshot |

#### Interactions

| Command | Description |
|---------|-------------|
| `press <ref>` | Press/tap element |
| `fill <ref> <text>` | Fill text input |
| `scroll <direction>` | Scroll view |
| `type <text>` | Type text |
| `get <ref>` | Get element value |
| `wait <ms>` | Wait for milliseconds |
| `find <query>` | Find element by text/selector |

#### Replay & Testing

| Command | Description |
|---------|-------------|
| `replay <script.ad>` | Replay saved script |
| `test <glob>` | Run test suite |
| `--save-script <name>` | Save current flow as script |

### Global Flags

| Flag | Description |
|------|-------------|
| `--platform <ios/android>` | Specify platform |
| `--device <name>` | Select specific device |
| `--verbose` | Enable verbose output |
| `--json` | Output in JSON format |

---

## Interactive Commands

### Command Flow (agent-device)

The canonical automation loop:

```bash
# 1. Open app
agent-device open SampleApp --platform ios

# 2. Inspect current screen
agent-device snapshot -i

# 3. Perform action
agent-device press @e3

# 4. Check for changes
agent-device diff snapshot -i

# 5. Fill input
agent-device fill @e5 "test"

# 6. Close session
agent-device close
```

### Understanding Snapshots

Snapshot output example:
```
@e1 button "General"
@e2 button "Display & Brightness"
@e3 switch "Airplane Mode" [off]
@e4 text "Wi-Fi"
@e5 textfield "Search"
```

Format: `@<ref> <type> "<label/placeholder>" [<state>]`

### Saving Replay Scripts

```bash
# Save automation flow
agent-device open MyApp
agent-device snapshot -i
agent-device press @e1
agent-device fill @e2 "username"
agent-device fill @e3 "password"
agent-device press @e4 --save-script login.ad
agent-device close

# Replay later
agent-device replay login.ad
```

### Test Execution

```bash
# Run single test
agent-device test login.ad

# Run test suite
agent-device test "tests/*.ad"

# With options
agent-device test features/ --retries 3 --timeout 30000
```

---

## Configuration

### agent-mobile Configuration

Configuration is handled automatically during `setup`:

```bash
# Config stored in:
~/.agent-mobile/config.json
```

### agent-device Configuration

```bash
# Config location
~/.config/agent-device/config.json
```

Example configuration:
```json
{
  "defaultPlatform": "ios",
  "ios": {
    "simulator": "iPhone 15 Pro",
    "osVersion": "17.0",
    "wdaPath": "/path/to/wda"
  },
  "android": {
    "avd": "Pixel_7_API_34",
    "adbPath": "/path/to/adb"
  },
  "timeouts": {
    "implicit": 10000,
    "explicit": 30000
  }
}
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `AGENT_MOBILE_PLATFORM` | Default platform (ios/android) |
| `AGENT_MOBILE_DEVICE` | Default device name |
| `APPIUM_PATH` | Custom Appium path |
| `WDA_PATH` | WebDriverAgent path |
| `ADB_PATH` | Android Debug Bridge path |

### iOS Simulator Configuration

```bash
# List available simulators
xcrun simctl list devices

# Boot specific simulator
xcrun simctl boot "iPhone 15 Pro"

# Set as default
export AGENT_MOBILE_DEVICE="iPhone 15 Pro"
```

### Android Emulator Configuration

```bash
# List available AVDs
emulator -list-avds

# Start specific emulator
emulator -avd Pixel_7_API_34

# Set as default
export AGENT_MOBILE_DEVICE="Pixel_7_API_34"
```

---

## Usage Examples

### Example 1: Basic iOS Automation

```bash
# Setup (first time only)
agent-mobile setup

# Open Settings app
agent-mobile open com.apple.Preferences

# Get UI elements
agent-mobile snapshot
# Output:
# @e1 button "General"
# @e2 button "Display & Brightness"
# @e3 switch "Airplane Mode" [off]

# Toggle airplane mode
agent-mobile tap @e3

# Take screenshot
agent-mobile screenshot /tmp/airplane_mode.png

# Close
agent-mobile close
```

### Example 2: Login Flow Automation

```bash
# Open your app
agent-device open com.example.myapp --platform ios

# Get snapshot
agent-device snapshot -i

# Output:
# @e1 textfield "Username"
# @e2 textfield "Password"
# @e3 button "Login"
# @e4 button "Forgot Password"

# Fill login form
agent-device fill @e1 "testuser"
agent-device fill @e2 "testpass123"

# Tap login
agent-device press @e3

# Wait for home screen
agent-device wait 2000

# Verify login success
agent-device snapshot -i

# Close
agent-device close
```

### Example 3: Form Testing with AI

Share this with your AI assistant:

```
I want to automate iOS simulators. Use agent-mobile - a CLI tool 
that lets you control simulators with simple commands.

Install: npm install -g agent-mobile
Setup: agent-mobile setup (first-time: configures Xcode, simulators, WDA)

Basic workflow:
1. agent-mobile open com.apple.Preferences (launch app)
2. agent-mobile snapshot (get elements with refs like @e1, @e2)
3. agent-mobile tap @e1 (interact by ref)
4. agent-mobile screenshot /tmp/s.png (capture screen)
5. agent-mobile close (end session)

For full command reference: agent-mobile.dev/SKILL.md
```

### Example 4: Deterministic Replay Testing

```bash
# Create test script: login.ad
# Content:
# open com.example.myapp
# snapshot -i
# fill @e1 "${USERNAME}"
# fill @e2 "${PASSWORD}"
# press @e3
# wait 2000
# snapshot -i
# close

# Run with variables
USERNAME=testuser PASSWORD=testpass agent-device replay login.ad

# Or run as test
agent-device test login.ad
```

### Example 5: Android Automation

```bash
# Android with agent-device
agent-device open com.android.settings --platform android

# Get snapshot
agent-device snapshot

# Interact
agent-device press @e5

# Scroll
agent-device scroll down

# Find and tap
agent-device find "Display" | xargs agent-device press

# Close
agent-device close
```

### Example 6: Multi-Step Workflow

```bash
#!/bin/bash
# test-e2e.sh

APP_BUNDLE="com.example.myapp"
TEST_USERNAME="user@example.com"
TEST_PASSWORD="password123"

# Open app
agent-device open $APP_BUNDLE

# Onboarding
agent-device snapshot -i
agent-device press @e1  # "Get Started"

# Login
agent-device fill @e1 $TEST_USERNAME
agent-device fill @e2 $TEST_PASSWORD
agent-device press @e3  # Login button
agent-device wait 3000

# Navigate
agent-device press @e5  # Profile tab
agent-device snapshot -i

# Verify
if agent-device find "Welcome"; then
    echo "Login successful!"
else
    echo "Login failed!"
    agent-device screenshot /tmp/failure.png
fi

# Cleanup
agent-device close
```

---

## Troubleshooting

### Setup Issues

#### "Xcode not found" (agent-mobile setup)

```bash
# Install Xcode from App Store
# Then install command line tools
xcode-select --install

# Accept license
sudo xcodebuild -license accept
```

#### "WebDriverAgent build failed"

```bash
# Clean and rebuild
rm -rf ~/Library/Developer/Xcode/DerivedData/WebDriverAgent-*
agent-mobile setup

# Or manually build WDA
cd /path/to/wda
xcodebuild -project WebDriverAgent.xcodeproj -scheme WebDriverAgentRunner -destination 'platform=iOS Simulator,name=iPhone 15 Pro' build
```

#### "Android SDK not found"

```bash
# Set ANDROID_HOME
export ANDROID_HOME=$HOME/Library/Android/sdk
export PATH=$PATH:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools

# Or install Android Studio
```

### Runtime Issues

#### "Simulator not booted"

```bash
# List devices
xcrun simctl list devices

# Boot device
xcrun simctl boot "iPhone 15 Pro"

# Or open Xcode and start simulator
```

#### "App not found"

```bash
# List installed apps
xcrun simctl listapps "iPhone 15 Pro"

# Install app first
xcrun simctl install "iPhone 15 Pro" /path/to/app.app
```

#### "Element not found"

```bash
# Get fresh snapshot
agent-device snapshot -i

# Use find with partial match
agent-device find "partial text"

# Check if element is visible (may need scroll)
agent-device scroll down
```

#### "Connection refused"

```bash
# Restart Appium
pkill -f appium
agent-mobile setup

# Or check ports
lsof -i :4723
```

### Common Error Messages

#### "Session does not exist"

```bash
# Start new session
agent-mobile open com.apple.Preferences
```

#### "Stale element reference"

```bash
# UI changed, get fresh snapshot
agent-device snapshot -i
```

#### "Element not interactable"

```bash
# Element may be behind another
# Try scrolling to it
agent-device scroll down
agent-device press @e5
```

### Platform-Specific Issues

#### iOS

```bash
# Reset simulator
xcrun simctl erase "iPhone 15 Pro"

# Check WebDriverAgent logs
tail -f ~/Library/Logs/WebDriverAgent.log
```

#### Android

```bash
# Check ADB connection
adb devices

# Restart ADB server
adb kill-server
adb start-server

# Check emulator status
adb shell getprop sys.boot_completed
```

### Debug Mode

```bash
# Verbose output
agent-mobile --verbose open com.apple.Preferences

# Debug logging
DEBUG=1 agent-device snapshot

# JSON output for parsing
agent-mobile --json snapshot
```

### Getting Help

```bash
# Show help
agent-mobile --help
agent-device --help

# Check version
agent-mobile --version

# Visit documentation
# https://agent-mobile.dev
# https://github.com/callstackincubator/agent-device
```

### Reporting Issues

1. Check existing issues on GitHub
2. Include:
   - Tool version (agent-mobile or agent-device)
   - Platform (iOS/Android)
   - OS version
   - Xcode/Android Studio version
   - Error messages
   - Steps to reproduce

---

## Best Practices

### 1. Use Descriptive Refs

Always get a fresh snapshot before interacting:
```bash
agent-device snapshot -i
agent-device press @e3  # Now we know what @e3 is
```

### 2. Handle Async Operations

Add waits after actions that trigger navigation:
```bash
agent-device press @e1  # Login button
agent-device wait 2000   # Wait for navigation
agent-device snapshot -i # Now on new screen
```

### 3. Save Replay Scripts

For repeatable tests, save scripts:
```bash
agent-device --save-script critical-path.ad
```

### 4. Use Screenshots for Debugging

```bash
agent-device screenshot /tmp/debug-$(date +%s).png
```

### 5. Clean Up Sessions

Always close sessions:
```bash
trap 'agent-device close' EXIT
```

---

## Resources

- **agent-mobile**: https://agent-mobile.dev
- **agent-device**: https://github.com/callstackincubator/agent-device
- **Appium Docs**: https://appium.io/docs
- **WebDriverAgent**: https://github.com/appium/WebDriverAgent

---

*Last updated: 2026-04-02*
