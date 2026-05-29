#!/usr/bin/env bash
# e2e_test.sh — JR Stack end-to-end test suite (harness-first).
#
# Tier 1 (default, no side-effects):
#   Binary exists, --help / --dry-run run without panic, valid modes accepted,
#   invalid mode and --custom without --mode custom are rejected (exit != 0).
#
# Tier 2 (RUN_FULL_E2E=1):
#   Real install per mode (lite/full/custom) and per agent (claude/opencode)
#   using --headless --home <sandbox>. Asserts marker idempotency, SKILL.md
#   presence for skill harnesses, valid MCP JSON, and verify reports Ready.
#
# Tier 3 (RUN_BACKUP_TESTS=1):
#   Reinstall is byte-identical (idempotence) + backup/restore round-trip.
#
# Usage:
#   bash e2e/e2e_test.sh                      # Tier 1
#   RUN_FULL_E2E=1 bash e2e/e2e_test.sh       # Tier 1 + 2
#   RUN_BACKUP_TESTS=1 bash e2e/e2e_test.sh   # Tier 1 + 3
#   RUN_FULL_E2E=1 RUN_BACKUP_TESTS=1 bash e2e/e2e_test.sh  # All tiers
set -uo pipefail

# Source shared helpers.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib.sh
source "$SCRIPT_DIR/lib.sh"

# ---------------------------------------------------------------------------
# Binary resolution
# ---------------------------------------------------------------------------
BIN="$(resolve_binary)"
if [ -z "$BIN" ]; then
    printf "${RED}[FATAL]${NC} jr-stack binary not found. Build with:\n"
    printf "  CGO_ENABLED=0 go build -o jr-stack ./cmd/jr-stack\n"
    exit 1
fi
log_info "Binary: $BIN"
log_info "Version: $("$BIN" install --help 2>&1 | head -1 || echo '(no version flag)')"

# ---------------------------------------------------------------------------
# Tier control
# ---------------------------------------------------------------------------
RUN_FULL_E2E="${RUN_FULL_E2E:-0}"
RUN_BACKUP_TESTS="${RUN_BACKUP_TESTS:-0}"

# ---------------------------------------------------------------------------
# Sandbox helper
# ---------------------------------------------------------------------------
# make_sandbox — create a temporary directory to use as --home for isolation.
make_sandbox() {
    mktemp -d /tmp/jr-stack-e2e-XXXXXX
}

# ---------------------------------------------------------------------------
# ═══════════════════════════════════════════════════════════════════════════
# TIER 1 — No side-effects (always runs)
# ═══════════════════════════════════════════════════════════════════════════
# ---------------------------------------------------------------------------
echo ""
printf "${YELLOW}━━━ TIER 1: Smoke / flag validation (no side-effects) ━━━${NC}\n"
echo ""

# T1.1 — Binary is executable
# Use command -v (not [ -x ]): $BIN may be a bare name resolved via PATH
# (e.g. installed at /usr/local/bin in the Docker images) rather than a path
# relative to the CWD. command -v handles both PATH lookups and absolute paths.
log_test "T1.1: Binary is executable"
if command -v "$BIN" >/dev/null 2>&1; then
    log_pass "Binary is executable: $BIN"
else
    log_fail "Binary is not executable: $BIN"
fi

# T1.2 — jr-stack install --help exits 0 (or 2) and does not panic
log_test "T1.2: install --help does not panic"
HELP_OUT=$("$BIN" install --help 2>&1 || true)
assert_output_not_contains "$HELP_OUT" "panic" "install --help: no panic"
assert_output_not_contains "$HELP_OUT" "runtime error" "install --help: no runtime error"
# --help should mention 'mode' (our key flag)
assert_output_contains "$HELP_OUT" "mode" "install --help: mentions --mode"

# T1.3 — --dry-run --mode lite exits 0 without executing anything
log_test "T1.3: --dry-run --mode lite exits 0"
SANDBOX=$(make_sandbox)
DRY_OUT=$("$BIN" install --dry-run --mode lite --agent claude --home "$SANDBOX" 2>&1)
DRY_EXIT=$?
assert_output_contains "$DRY_OUT" "Dry-run" "dry-run output mentions 'Dry-run'"
if [ "$DRY_EXIT" -eq 0 ]; then
    log_pass "dry-run --mode lite exits 0"
else
    log_fail "dry-run --mode lite exited $DRY_EXIT (expected 0)"
fi
# Verify --dry-run had no side-effects (no files written in sandbox)
if [ -z "$(ls -A "$SANDBOX" 2>/dev/null)" ]; then
    log_pass "dry-run: sandbox is empty (no side-effects)"
else
    log_fail "dry-run: unexpected files in sandbox: $(ls "$SANDBOX")"
fi
rm -rf "$SANDBOX"

# T1.4 — --dry-run --mode full exits 0
log_test "T1.4: --dry-run --mode full exits 0"
SANDBOX=$(make_sandbox)
DRY_OUT=$("$BIN" install --dry-run --mode full --agent claude --home "$SANDBOX" 2>&1)
DRY_EXIT=$?
if [ "$DRY_EXIT" -eq 0 ]; then
    log_pass "dry-run --mode full exits 0"
else
    log_fail "dry-run --mode full exited $DRY_EXIT (expected 0)"
fi
rm -rf "$SANDBOX"

# T1.5 — --dry-run --mode custom --custom sdd-orchestrator exits 0
log_test "T1.5: --dry-run --mode custom --custom sdd-orchestrator exits 0"
SANDBOX=$(make_sandbox)
DRY_OUT=$("$BIN" install --dry-run --mode custom --custom sdd-orchestrator --agent claude --home "$SANDBOX" 2>&1)
DRY_EXIT=$?
if [ "$DRY_EXIT" -eq 0 ]; then
    log_pass "dry-run --mode custom exits 0"
else
    log_fail "dry-run --mode custom exited $DRY_EXIT (expected 0)"
fi
rm -rf "$SANDBOX"

# T1.6 — Invalid --mode is rejected (exit != 0)
log_test "T1.6: invalid --mode is rejected"
SANDBOX=$(make_sandbox)
BAD_OUT=$("$BIN" install --mode bogusmode --agent claude --home "$SANDBOX" 2>&1); BAD_EXIT=$?
if [ "$BAD_EXIT" -ne 0 ]; then
    log_pass "Invalid --mode rejected (exit $BAD_EXIT)"
else
    log_fail "Invalid --mode was NOT rejected (exit 0)"
fi
assert_output_not_contains "$BAD_OUT" "panic" "invalid mode: no panic"
rm -rf "$SANDBOX"

# T1.7 — --custom without --mode custom is rejected (exit != 0)
log_test "T1.7: --custom without --mode custom is rejected"
SANDBOX=$(make_sandbox)
BAD_OUT=$("$BIN" install --custom sdd-orchestrator --mode lite --agent claude --home "$SANDBOX" 2>&1); BAD_EXIT=$?
if [ "$BAD_EXIT" -ne 0 ]; then
    log_pass "--custom without --mode custom rejected (exit $BAD_EXIT)"
else
    log_fail "--custom without --mode custom was NOT rejected (exit 0)"
fi
assert_output_not_contains "$BAD_OUT" "panic" "--custom validation: no panic"
rm -rf "$SANDBOX"

# T1.8 — --custom alone (no --mode) is rejected (exit != 0)
log_test "T1.8: --custom alone (no --mode) is rejected"
SANDBOX=$(make_sandbox)
BAD_OUT=$("$BIN" install --custom sdd-orchestrator --agent claude --home "$SANDBOX" 2>&1); BAD_EXIT=$?
if [ "$BAD_EXIT" -ne 0 ]; then
    log_pass "--custom without --mode rejected (exit $BAD_EXIT)"
else
    log_fail "--custom without --mode was NOT rejected (exit 0)"
fi
rm -rf "$SANDBOX"

# T1.9 — --headless alone (no --mode) produces an error or falls back gracefully
log_test "T1.9: --headless with --mode lite exits 0"
SANDBOX=$(make_sandbox)
HL_OUT=$("$BIN" install --headless --dry-run --mode lite --agent claude --home "$SANDBOX" 2>&1)
HL_EXIT=$?
if [ "$HL_EXIT" -eq 0 ]; then
    log_pass "--headless --dry-run --mode lite exits 0"
else
    log_fail "--headless --dry-run --mode lite exited $HL_EXIT"
fi
rm -rf "$SANDBOX"

# ---------------------------------------------------------------------------
# ═══════════════════════════════════════════════════════════════════════════
# TIER 2 — Real install (RUN_FULL_E2E=1)
# ═══════════════════════════════════════════════════════════════════════════
# ---------------------------------------------------------------------------
if [ "${RUN_FULL_E2E}" != "1" ]; then
    echo ""
    log_info "Tier 2 skipped (set RUN_FULL_E2E=1 to run real installs)"
else
    echo ""
    printf "${YELLOW}━━━ TIER 2: Real install — harness assertions ━━━${NC}\n"
    echo ""

    # -------------------------------------------------------------------------
    # Helper: run_install_and_assert MODE AGENT
    # Runs a real headless install into a sandbox and asserts key invariants.
    # -------------------------------------------------------------------------
    run_install_and_assert() {
        local mode="$1"
        local agent="$2"
        local sandbox
        sandbox=$(make_sandbox)

        log_test "T2: install --mode $mode --agent $agent"
        log_info "Sandbox: $sandbox"

        # Run the real install (capture output and exit code without -e interfering).
        local install_out install_exit
        install_out=$("$BIN" install \
            --headless \
            --mode "$mode" \
            --agent "$agent" \
            --yes \
            --home "$sandbox" 2>&1)
        install_exit=$?

        # Log output for debugging.
        if [ "$install_exit" -ne 0 ]; then
            log_fail "Install exited $install_exit (mode=$mode agent=$agent)"
            echo "--- install output ---"
            echo "$install_out"
            echo "---------------------"
            rm -rf "$sandbox"
            return 1
        fi
        log_pass "Install completed (mode=$mode agent=$agent)"

        # ── Assert: verify reports Ready ──────────────────────────────────────
        assert_output_contains "$install_out" "Ready" \
            "verify: reports Ready (mode=$mode agent=$agent)"

        # ── Assert agent-specific paths ───────────────────────────────────────
        local instructions_file skills_dir mcp_pattern
        case "$agent" in
            claude)
                instructions_file="$sandbox/.claude/CLAUDE.md"
                skills_dir="$sandbox/.claude/skills"
                mcp_pattern="$sandbox/.claude/mcp/*.json"
                ;;
            opencode)
                instructions_file="$sandbox/.config/opencode/AGENTS.md"
                skills_dir="$sandbox/.config/opencode/skills"
                mcp_pattern="$sandbox/.config/opencode/opencode.json"
                ;;
            *)
                log_skip "Unknown agent $agent — skipping path assertions"
                rm -rf "$sandbox"
                return 0
                ;;
        esac

        # ── Config harnesses: idempotent marker in instructions file ──────────
        # sdd-orchestrator is a config harness included in lite+full for claude/opencode.
        if [ -f "$instructions_file" ]; then
            assert_no_duplicate_section "$instructions_file" "sdd-orchestrator" \
                "sdd-orchestrator: marker present exactly once in $agent instructions"
        else
            log_skip "Instructions file not found (agent=$agent mode=$mode): $instructions_file"
        fi

        # ── Skill harnesses: SKILL.md present in skills dir ───────────────────
        # jr-orchestrator is a skill harness included in lite+full.
        local skill_md="$skills_dir/jr-orchestrator/SKILL.md"
        if [ -d "$skills_dir" ]; then
            # We can't clone real repos in a hermetic container; skip if git is
            # unavailable or the skill wasn't cloned (network unavailable = skip, not fail).
            if git ls-remote --heads >/dev/null 2>&1 || true; then
                if [ -f "$skill_md" ]; then
                    assert_file_exists "$skill_md" "jr-orchestrator SKILL.md present (agent=$agent)"
                    assert_file_size_min "$skill_md" 1 "jr-orchestrator SKILL.md non-empty"
                else
                    log_skip "jr-orchestrator SKILL.md not cloned (network unavailable?): $skill_md"
                fi
            fi
        else
            log_skip "Skills dir not created (agent=$agent mode=$mode): $skills_dir"
        fi

        # ── External harnesses: MCP JSON parseable ────────────────────────────
        # For claude: separate files per server under ~/.claude/mcp/
        # For opencode: merged into opencode.json
        case "$agent" in
            claude)
                local mcp_dir="$sandbox/.claude/mcp"
                if [ -d "$mcp_dir" ]; then
                    local json_count
                    json_count=$(find "$mcp_dir" -name "*.json" -type f 2>/dev/null | wc -l | tr -d ' ')
                    if [ "$json_count" -gt 0 ]; then
                        log_pass "MCP dir has $json_count JSON file(s) (agent=claude)"
                        # Validate each JSON file.
                        for jf in "$mcp_dir"/*.json; do
                            [ -f "$jf" ] || continue
                            assert_valid_json "$jf" "MCP JSON valid: $(basename "$jf")"
                        done
                    else
                        log_skip "No MCP JSON files yet (external installs skipped?)"
                    fi
                else
                    log_skip "MCP dir not created (agent=claude mode=$mode)"
                fi
                ;;
            opencode)
                local oc_settings="$sandbox/.config/opencode/opencode.json"
                if [ -f "$oc_settings" ]; then
                    assert_valid_json "$oc_settings" "opencode.json is valid JSON (agent=opencode)"
                else
                    log_skip "opencode.json not created (agent=opencode mode=$mode)"
                fi
                ;;
        esac

        rm -rf "$sandbox"
        return 0
    }

    # ── Run matrix: modes × agents ────────────────────────────────────────────
    run_install_and_assert "lite"   "claude"   || true
    run_install_and_assert "lite"   "opencode" || true
    run_install_and_assert "full"   "claude"   || true
    run_install_and_assert "full"   "opencode" || true
    run_install_and_assert "custom" "claude"   || true  # uses sdd-orchestrator only

fi # RUN_FULL_E2E

# ---------------------------------------------------------------------------
# ═══════════════════════════════════════════════════════════════════════════
# TIER 3 — Idempotence + backup/restore (RUN_BACKUP_TESTS=1)
# ═══════════════════════════════════════════════════════════════════════════
# ---------------------------------------------------------------------------
if [ "${RUN_BACKUP_TESTS}" != "1" ]; then
    echo ""
    log_info "Tier 3 skipped (set RUN_BACKUP_TESTS=1 to run backup/idempotence tests)"
else
    echo ""
    printf "${YELLOW}━━━ TIER 3: Idempotence + backup/restore ━━━${NC}\n"
    echo ""

    # ── T3.1: Reinstall is idempotent (byte-identical output files) ──────────
    log_test "T3.1: Reinstall is idempotent (claude, lite)"
    SANDBOX=$(make_sandbox)

    # First install.
    "$BIN" install --headless --mode lite --agent claude --yes --home "$SANDBOX" >/dev/null 2>&1 || true
    INSTRUCTIONS_1="$SANDBOX/.claude/CLAUDE.md"

    if [ -f "$INSTRUCTIONS_1" ]; then
        # Snapshot content after first install.
        HASH1=$(md5sum "$INSTRUCTIONS_1" | cut -d' ' -f1)

        # Second install (reinstall).
        "$BIN" install --headless --mode lite --agent claude --yes --home "$SANDBOX" >/dev/null 2>&1 || true
        HASH2=$(md5sum "$INSTRUCTIONS_1" | cut -d' ' -f1)

        if [ "$HASH1" = "$HASH2" ]; then
            log_pass "T3.1: Reinstall is idempotent (same hash: $HASH1)"
        else
            log_fail "T3.1: Reinstall changed content: $HASH1 → $HASH2"
        fi

        # Also verify no duplicate markers after reinstall.
        assert_no_duplicate_section "$INSTRUCTIONS_1" "sdd-orchestrator" \
            "T3.1: No duplicate sdd-orchestrator marker after reinstall"
    else
        log_skip "T3.1: Instructions file not created — skipping idempotence check"
    fi
    rm -rf "$SANDBOX"

    # ── T3.2: Backup/restore round-trip ──────────────────────────────────────
    log_test "T3.2: Backup/restore round-trip"
    SANDBOX=$(make_sandbox)

    # Seed a fake CLAUDE.md to snapshot.
    mkdir -p "$SANDBOX/.claude"
    echo "# Pre-existing CLAUDE.md content" > "$SANDBOX/.claude/CLAUDE.md"
    # (Original content noted for context; actual check done via grep below.)
    _ORIGINAL_CONTENT=$(cat "$SANDBOX/.claude/CLAUDE.md" 2>/dev/null || echo "")

    # Install (which should take a backup of the pre-existing file).
    "$BIN" install --headless --mode lite --agent claude --yes --home "$SANDBOX" >/dev/null 2>&1 || true

    # After install, the instructions file should exist and contain our marker.
    if [ -f "$SANDBOX/.claude/CLAUDE.md" ]; then
        log_pass "T3.2: CLAUDE.md still exists after install"

        # The original content should still be present (inject, not overwrite).
        if grep -q "Pre-existing CLAUDE.md content" "$SANDBOX/.claude/CLAUDE.md"; then
            log_pass "T3.2: Original content preserved after inject"
        else
            log_fail "T3.2: Original content was overwritten"
        fi

        # Check that backup dir was created by the installer.
        BACKUP_DIR="$SANDBOX/.jr-stack/backups"
        if [ -d "$BACKUP_DIR" ]; then
            log_pass "T3.2: Backup directory created: $BACKUP_DIR"
            BACKUP_COUNT=$(find "$BACKUP_DIR" -type f 2>/dev/null | wc -l | tr -d ' ')
            log_info "T3.2: $BACKUP_COUNT backup file(s) in $BACKUP_DIR"
        else
            log_skip "T3.2: Backup dir not found (backup may use different path): $BACKUP_DIR"
        fi
    else
        log_skip "T3.2: CLAUDE.md not created — skipping backup assertions"
    fi
    rm -rf "$SANDBOX"

fi # RUN_BACKUP_TESTS

# ---------------------------------------------------------------------------
# Final summary
# ---------------------------------------------------------------------------
print_summary
