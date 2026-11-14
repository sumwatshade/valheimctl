#!/usr/bin/env bash
# test-valheimctl.sh - End-to-end smoke test for valheimctl.
# Run this inside the VM as a user with sudo.
# Usage: ./scripts/test-valheimctl.sh [ROOT_DIR]
#
# ROOT_DIR defaults to ~/valheim-test-root. It is wiped and recreated each run
# so you can iterate repeatedly.
set -euo pipefail

ROOT_DIR="${1:-$HOME/valheim-test-root}"
VALHEIMCTL="${VALHEIMCTL:-valheimctl}"
SUDO="${SUDO:-sudo}"

echo "==> Resetting test root: $ROOT_DIR"
rm -rf "$ROOT_DIR"
mkdir -p "$ROOT_DIR/worlds_local"
mkdir -p "$ROOT_DIR/etc/systemd/system"

# Create dummy world folders so backup create has something to copy
mkdir -p "$ROOT_DIR/worlds_local/Dedicated"
echo "level.dat" > "$ROOT_DIR/worlds_local/Dedicated/level.dat"
echo "world.json" > "$ROOT_DIR/worlds_local/Dedicated/world.json"

PASS=0
FAIL=0

assert_contains() {
  local desc="$1" output="$2" needle="$3"
  if echo "$output" | grep -q "$needle"; then
    echo "PASS: $desc"
    PASS=$((PASS + 1))
  else
    echo "FAIL: $desc (expected '$needle' in output)"
    echo "---- output ----"
    echo "$output"
    echo "----------------"
    FAIL=$((FAIL + 1))
  fi
}

assert_exit() {
  local desc="$1" expected="$2" actual="$3"
  if [ "$actual" -eq "$expected" ]; then
    echo "PASS: $desc (exit $actual)"
    PASS=$((PASS + 1))
  else
    echo "FAIL: $desc (expected exit $expected, got $actual)"
    FAIL=$((FAIL + 1))
  fi
}

run() {
  # Run valheimctl with ROOT_DIR as the working directory so relative paths resolve
  (cd "$ROOT_DIR" && "$VALHEIMCTL" "$@")
}

echo ""
echo "==> Test: init"
out=$(run init 2>&1) || true
assert_contains "init prints initialized" "$out" "initialized"

echo ""
echo "==> Test: status (after init - stopped, not running)"
out=$(run status 2>&1) || true
assert_contains "status reports stopped after init" "$out" "stopped"

echo ""
echo "==> Test: start"
out=$(run start 2>&1) || true
assert_contains "start reports started" "$out" "started"

echo ""
echo "==> Test: status (running)"
out=$(run status 2>&1) || true
assert_contains "status reports running" "$out" "running"

echo ""
echo "==> Test: stop"
out=$(run stop 2>&1) || true
assert_contains "stop reports stopped" "$out" "stopped"

echo ""
echo "==> Test: status (stopped)"
out=$(run status 2>&1) || true
assert_contains "status reports stopped" "$out" "stopped"

echo ""
echo "==> Test: register (requires sudo for systemctl)"
if command -v systemctl >/dev/null 2>&1; then
  out=$($SUDO "$VALHEIMCTL" register myvalheim "$ROOT_DIR" "My Valheim Server" 2>&1) || true
  assert_contains "register reports registered" "$out" "registered"
else
  echo "SKIP: register (systemctl not available in this environment)"
fi

echo ""
echo "==> Test: backup create"
out=$(run backup create mybackup 2>&1) || true
assert_contains "backup create reports created" "$out" "created mybackup"

echo ""
echo "==> Test: backup list"
out=$(run backup list 2>&1) || true
assert_contains "backup list shows mybackup" "$out" "mybackup"

echo ""
echo "==> Test: backup apply"
out=$(run backup apply mybackup 2>&1) || true
assert_contains "backup apply reports restored" "$out" "restored mybackup"

echo ""
echo "==> Test: backup (no args = describe/list)"
out=$(run backup 2>&1) || true
assert_contains "backup lists mybackup" "$out" "mybackup"

echo ""
echo "========================================"
echo "Results: $PASS passed, $FAIL failed"
echo "========================================"

if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
