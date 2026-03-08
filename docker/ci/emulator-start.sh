#!/usr/bin/env bash
set -euo pipefail

echo "Starting Android emulator (headless)..."

# Start emulator in background
emulator -avd ci-device \
  -no-window \
  -no-audio \
  -no-boot-anim \
  -gpu swiftshader_indirect \
  -no-snapshot \
  -wipe-data \
  -port 5554 &

EMULATOR_PID=$!

# Wait for ADB device
echo "Waiting for ADB device..."
adb wait-for-device

# Wait for boot_completed
TIMEOUT=300
ELAPSED=0
while [ "$(adb shell getprop sys.boot_completed 2>/dev/null | tr -d '\r')" != "1" ]; do
  sleep 5
  ELAPSED=$((ELAPSED + 5))
  if [ "${ELAPSED}" -ge "${TIMEOUT}" ]; then
    echo "ERROR: Emulator failed to boot within ${TIMEOUT}s"
    exit 1
  fi
  echo "  Waiting for boot... (${ELAPSED}s)"
done

echo "Emulator booted successfully"

# Enable ADB over TCP for network access from other containers
adb tcpip 5555

echo "ADB TCP enabled on port 5555"

# Keep alive
wait "${EMULATOR_PID}"
