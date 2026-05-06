#!/bin/bash
# .rulesync/hooks/init.sh - Repository initialization hook
# Runs setup commands only when tools are available and tracks first-run state

set -uo pipefail

REPO_JSON=".repo.json"
TODOS_FLAG="todosExist"
TODOS_DIR=".todos"
DG_SYNC_SCRIPT="./scripts/dg.sh"
HOOK_LOG_FILE=".rulesync/init-hook.log"
HOOK_SUMMARY_LINES=()
HOOK_FAILURE_COUNT=0
HOOK_FAILED_STEPS=()

# Helper: check if command exists
has_cmd() {
    command -v "$1" >/dev/null 2>&1
}

# Hooks should stay quiet on stdout; emit diagnostics to stderr instead.
log() {
    printf '%s\n' "$*" >&2
}

log_step() {
    local status="$1"
    shift
    log "[init] ${status}: $*"
}

add_summary() {
    HOOK_SUMMARY_LINES+=("$*")
}

record_failure() {
    local command_text="$1"
    local exit_code="$2"
    HOOK_FAILURE_COUNT=$((HOOK_FAILURE_COUNT + 1))
    HOOK_FAILED_STEPS+=("${command_text} (exit ${exit_code})")
    add_summary "failed: ${command_text} (exit ${exit_code})"
    log_step "error" "${command_text} failed with exit code ${exit_code}"
}

emit_hook_output() {
    local summary="Init hook completed"
    local additional_context=""
    local joined_summary=""
    if (( HOOK_FAILURE_COUNT > 0 )); then
        summary="Init hook completed with ${HOOK_FAILURE_COUNT} failure(s). Details: ${HOOK_LOG_FILE}"
    elif [[ ${#HOOK_SUMMARY_LINES[@]} -gt 0 ]]; then
        summary="Init hook completed. Details: ${HOOK_LOG_FILE}"
    fi

    {
        printf 'Init hook\n'
        printf '=========\n'
        printf 'Failures: %s\n' "$HOOK_FAILURE_COUNT"
        if (( HOOK_FAILURE_COUNT > 0 )); then
            printf '\nFailed steps\n'
            printf '------------\n'
            local failed_step
            for failed_step in "${HOOK_FAILED_STEPS[@]}"; do
                printf '%s\n' "$failed_step"
            done
        fi
        printf '\nAll steps\n'
        printf '---------\n'
        if [[ ${#HOOK_SUMMARY_LINES[@]} -eq 0 ]]; then
            printf 'No recorded steps.\n'
        else
            local line
            for line in "${HOOK_SUMMARY_LINES[@]}"; do
                printf '%s\n' "$line"
            done
        fi
    } > "$HOOK_LOG_FILE"

    if [[ ${#HOOK_SUMMARY_LINES[@]} -eq 0 ]]; then
        additional_context="Init hook completed with no recorded steps."
    else
        local line
        for line in "${HOOK_SUMMARY_LINES[@]}"; do
            if [[ -n "$joined_summary" ]]; then
                joined_summary="${joined_summary} | "
            fi
            joined_summary="${joined_summary}${line}"
        done
        if (( HOOK_FAILURE_COUNT > 0 )); then
            additional_context="Init hook encountered ${HOOK_FAILURE_COUNT} failure(s). Steps: ${joined_summary}"
        else
            additional_context="Init hook steps: ${joined_summary}"
        fi
    fi

    jq -nc \
      --arg system_message "$summary" \
      --arg additional_context "$additional_context" \
      '{
      continue: true,
      suppressOutput: false,
      systemMessage: $system_message,
      hookSpecificOutput: {
        hookEventName: "SessionStart",
        additionalContext: $additional_context
      }
    }'
}

run_step() {
    local command_text="$*"
    local exit_code=0

    log_step "run" "$command_text"
    if "$@" >&2; then
        log_step "done" "$command_text"
        add_summary "ran: $command_text"
        return 0
    fi

    exit_code=$?
    record_failure "$command_text" "$exit_code"
    return "$exit_code"
}

# Ensure jq is available
if ! has_cmd jq; then
    log_step "error" "jq is required but not installed. Please install jq first."
    exit 1
fi

# Helper: read JSON field using jq
get_json_field() {
    local file="$1" field="$2" default="${3:-false}"
    if [[ -f "$file" ]]; then
        jq -r ".$field // \"$default\"" "$file" 2>/dev/null || echo "$default"
    else
        echo "$default"
    fi
}

# Helper: set JSON field using jq
set_json_field() {
    local file="$1" field="$2" value="$3"
    if [[ -f "$file" ]]; then
        jq ".$field = $value" "$file" > "${file}.tmp" && mv "${file}.tmp" "$file"
    else
        echo "{\"$field\": $value}" > "$file"
    fi
}

# Helper: keep .repo.json aligned with whether td has been initialized
sync_todos_state() {
    if [[ -d "$TODOS_DIR" ]]; then
        set_json_field "$REPO_JSON" "$TODOS_FLAG" true
        log_step "state" "$REPO_JSON => ${TODOS_FLAG}=true ($TODOS_DIR exists)"
        add_summary "state: ${TODOS_FLAG}=true"
    else
        set_json_field "$REPO_JSON" "$TODOS_FLAG" false
        log_step "state" "$REPO_JSON => ${TODOS_FLAG}=false ($TODOS_DIR missing)"
        add_summary "state: ${TODOS_FLAG}=false"
    fi
}

# Run 'skills update' if installed
if has_cmd skills; then
    log_step "start" "skills update -y"
    run_step skills update -y || true
else
    log_step "skip" "skills update -y (skills not installed)"
    add_summary "skipped: skills update -y (skills not installed)"
fi

# Run 'rulesync sync' if installed
if has_cmd rulesync; then
    log_step "start" "rulesync import -t agentsskills -f skills"
    run_step rulesync import -t agentsskills -f skills || true
    log_step "start" "rulesync generate -s"
    run_step rulesync generate -s || true
else
    log_step "skip" "rulesync import/generate (rulesync not installed)"
    add_summary "skipped: rulesync import/generate (rulesync not installed)"
fi

# Sync docs context repositories when the helper is available.
if [[ -x "$DG_SYNC_SCRIPT" ]]; then
    log_step "start" "$DG_SYNC_SCRIPT sync"
    run_step "$DG_SYNC_SCRIPT" sync || true
else
    log_step "skip" "$DG_SYNC_SCRIPT sync (script not executable)"
    add_summary "skipped: $DG_SYNC_SCRIPT sync (script not executable)"
fi

# Keep repo state aligned with the actual td initialization state on every run.
sync_todos_state

# Initialize td whenever .todos is missing.
if [[ ! -d "$TODOS_DIR" ]]; then
    log_step "start" "td init ($TODOS_DIR missing)"
    if has_cmd td; then
        run_step td init || true
        sync_todos_state
    else
        log_step "skip" "td init (td not installed)"
        add_summary "skipped: td init (td not installed)"
    fi
else
    log_step "skip" "td init ($TODOS_DIR already exists)"
    add_summary "skipped: td init ($TODOS_DIR already exists)"
fi

log_step "done" "Initialization complete"
emit_hook_output
exit 0
