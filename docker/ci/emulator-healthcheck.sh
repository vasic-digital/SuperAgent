#!/usr/bin/env bash

EMULATOR_PORT="${EMULATOR_PORT:-5554}"
DEVICE="emulator-${EMULATOR_PORT}"

adb -s "${DEVICE}" shell getprop sys.boot_completed 2>/dev/null | grep -q "1" || exit 1

adb -s "${DEVICE}" shell getprop init.svc.bootanim 2>/dev/null | grep -q "stopped" || exit 1

adb -s "${DEVICE}" shell pm list packages 2>/dev/null | grep -q "package:android" || exit 1

exit 0
