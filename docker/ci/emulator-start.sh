#!/usr/bin/env bash
set -euo pipefail

EMULATOR_NAME="${EMULATOR_NAME:-ci-device}"
EMULATOR_PORT="${EMULATOR_PORT:-5554}"
EMULATOR_DISPLAY="${EMULATOR_DISPLAY:-768x1280}"
EMULATOR_DPI="${EMULATOR_DPI:-320}"
EMULATOR_MEMORY="${EMULATOR_MEMORY:-4096}"
EMULATOR_CORES="${EMULATOR_CORES:-4}"
EMULATOR_NO_WINDOW="${EMULATOR_NO_WINDOW:-true}"
EMULATOR_WIPE_DATA="${EMULATOR_WIPE_DATA:-true}"
EMULATOR_BOOT_TIMEOUT="${EMULATOR_BOOT_TIMEOUT:-300}"

mkdir -p /var/log/emulator

echo "========================================"
echo "Android Emulator Startup"
echo "Name: ${EMULATOR_NAME}"
echo "Port: ${EMULATOR_PORT}"
echo "Display: ${EMULATOR_DISPLAY}"
echo "Memory: ${EMULATOR_MEMORY}MB"
echo "Cores: ${EMULATOR_CORES}"
echo "Boot Timeout: ${EMULATOR_BOOT_TIMEOUT}s"
echo "========================================"

adb kill-server 2>/dev/null || true

ACCEL_OPTS=()
if [ -e /dev/kvm ] && [ -r /dev/kvm ] && [ -w /dev/kvm ]; then
    ACCEL_OPTS=(-accel on)
    echo "[OK] KVM acceleration available and enabled"
else
    ACCEL_OPTS=(-accel on -gpu swiftshader_indirect)
    echo "[WARN] KVM not available, using software rendering"
fi

WINDOW_OPTS=()
if [ "${EMULATOR_NO_WINDOW}" = "true" ]; then
    WINDOW_OPTS=(-no-window -no-boot-anim)
fi

WIPE_OPTS=()
if [ "${EMULATOR_WIPE_DATA}" = "true" ]; then
    WIPE_OPTS=(-no-snapshot -wipe-data)
fi

EMULATOR_OPTS=(
    -avd "${EMULATOR_NAME}"
    -port "${EMULATOR_PORT}"
    -no-audio
    -memory "${EMULATOR_MEMORY}"
    -cores "${EMULATOR_CORES}"
    "${ACCEL_OPTS[@]}"
    "${WINDOW_OPTS[@]}"
    "${WIPE_OPTS[@]}"
    -logcat "*:I"
)

echo "[INFO] Starting emulator with options: ${EMULATOR_OPTS[*]}"

emulator "${EMULATOR_OPTS[@]}" 2>&1 | tee /var/log/emulator/emulator.log &
EMULATOR_PID=$!

echo "[INFO] Waiting for ADB device..."
adb start-server

DEVICE="emulator-${EMULATOR_PORT}"
ELAPSED=0
while ! adb -s "${DEVICE}" get-state 2>/dev/null | grep -q "device"; do
    sleep 2
    ELAPSED=$((ELAPSED + 2))
    if [ "${ELAPSED}" -ge "${EMULATOR_BOOT_TIMEOUT}" ]; then
        echo "[ERROR] ADB device not available within ${EMULATOR_BOOT_TIMEOUT}s"
        kill "${EMULATOR_PID}" 2>/dev/null || true
        exit 1
    fi
    echo "  Waiting for ADB... (${ELAPSED}s)"
done

echo "[OK] ADB device available: ${DEVICE}"

echo "[INFO] Waiting for boot completion..."
ELAPSED=0
while [ "$(adb -s "${DEVICE}" shell getprop sys.boot_completed 2>/dev/null | tr -d '\r\n')" != "1" ]; do
    sleep 3
    ELAPSED=$((ELAPSED + 3))
    if [ "${ELAPSED}" -ge "${EMULATOR_BOOT_TIMEOUT}" ]; then
        echo "[ERROR] Emulator failed to boot within ${EMULATOR_BOOT_TIMEOUT}s"
        kill "${EMULATOR_PID}" 2>/dev/null || true
        exit 1
    fi
    echo "  Waiting for boot... (${ELAPSED}s)"
done

echo "[OK] Emulator booted successfully"

adb -s "${DEVICE}" shell input keyevent KEYCODE_WAKEUP 2>/dev/null || true
adb -s "${DEVICE}" shell input keyevent 82 2>/dev/null || true

TCP_PORT=$((EMULATOR_PORT + 1))
if adb -s "${DEVICE}" tcpip "${TCP_PORT}" 2>/dev/null; then
    echo "[OK] ADB TCP enabled on port ${TCP_PORT}"
    sleep 2
fi

adb -s "${DEVICE}" shell getprop ro.build.version.sdk > /var/log/emulator/api-level.txt 2>/dev/null || true
adb -s "${DEVICE}" shell getprop ro.build.version.release > /var/log/emulator/android-version.txt 2>/dev/null || true
adb -s "${DEVICE}" shell getprop ro.product.model > /var/log/emulator/device-model.txt 2>/dev/null || true

echo "========================================"
echo "Emulator Ready"
echo "ADB Serial: ${DEVICE}"
echo "API Level: $(cat /var/log/emulator/api-level.txt 2>/dev/null || echo 'unknown')"
echo "Android: $(cat /var/log/emulator/android-version.txt 2>/dev/null || echo 'unknown')"
echo "Device: $(cat /var/log/emulator/device-model.txt 2>/dev/null || echo 'unknown')"
echo "========================================"

wait "${EMULATOR_PID}"
