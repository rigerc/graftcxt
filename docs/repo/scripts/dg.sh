#!/bin/bash
#
# Manage a docs context file and sync repositories into ./docs with tiged.

set -euo pipefail

SCRIPT_DIR="$(
  cd "$(dirname "${BASH_SOURCE[0]}")" && pwd
)"
readonly SCRIPT_DIR
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly ROOT_DIR
readonly CONTEXT_FILE="${ROOT_DIR}/.context.json"
readonly DOCS_DIR="${ROOT_DIR}/docs"
readonly DATE_FORMAT="+%Y-%m-%dT%H:%M:%S%z"

err() {
  echo "[$(date "${DATE_FORMAT}")] Error: $*" >&2
}

usage() {
  cat <<EOF
Usage:
  dg add <repo>
  dg rem <repo>
  dg ls
  dg sync

Commands:
  add   Clone a repo into ./docs/<repo-name>/ and record it in .context.json
  rem   Remove a repo entry from .context.json by stored repo identifier
  ls    List tracked repos from .context.json
  sync  Re-sync all tracked repos into ./docs/<repo-name>/ with force enabled
EOF
}

require_command() {
  local command_name="$1"

  if ! command -v "${command_name}" >/dev/null 2>&1; then
    err "Required command not found: ${command_name}"
    exit 1
  fi
}

resolve_tiged_command() {
  if command -v tiged >/dev/null 2>&1; then
    echo "tiged"
    return
  fi

  if command -v degit >/dev/null 2>&1; then
    echo "degit"
    return
  fi

  err "Neither tiged nor degit is installed"
  exit 1
}

ensure_layout() {
  mkdir -p "${DOCS_DIR}"

  if [[ ! -f "${CONTEXT_FILE}" ]]; then
    jq -n '{version: 1, repos: []}' > "${CONTEXT_FILE}"
  fi
}

validate_context() {
  if ! jq -e '
    .version == 1 and
    (.repos | type == "array") and
    all(
      .repos[];
      (.repo | type == "string" and length > 0) and
      (.name | type == "string" and length > 0) and
      (
        (.last_sync == null) or
        (.last_sync | type == "string")
      )
    )
  ' "${CONTEXT_FILE}" >/dev/null; then
    err "Invalid context file schema: ${CONTEXT_FILE}"
    exit 1
  fi
}

parse_repo_name() {
  local repo_id="$1"
  local trimmed="${repo_id%/}"
  local without_ref="${trimmed%%#*}"
  local name

  name="$(basename "${without_ref}")"

  if [[ -z "${name}" || "${name}" == "." || "${name}" == "/" ]]; then
    err "Unable to derive repo directory name from: ${repo_id}"
    exit 1
  fi

  printf '%s\n' "${name}"
}

context_has_repo() {
  local repo_id="$1"

  jq -e --arg repo "${repo_id}" '
    any(.repos[]?; .repo == $repo)
  ' "${CONTEXT_FILE}" >/dev/null
}

sync_single_repo() {
  local repo_id="$1"
  local repo_name="$2"
  local tiged_command="$3"
  local destination="${DOCS_DIR}/${repo_name}"
  local timestamp

  mkdir -p "${destination}"

  echo "Syncing ${repo_id} -> ${destination}"
  "${tiged_command}" --force "${repo_id}" "${destination}"

  timestamp="$(date "${DATE_FORMAT}")"
  update_last_sync "${repo_id}" "${timestamp}"
}

update_last_sync() {
  local repo_id="$1"
  local timestamp="$2"
  local tmp_file

  tmp_file="$(mktemp)"

  jq --arg repo "${repo_id}" --arg timestamp "${timestamp}" '
    .repos |= map(
      if .repo == $repo then
        .last_sync = $timestamp
      else
        .
      end
    )
  ' "${CONTEXT_FILE}" > "${tmp_file}"

  mv "${tmp_file}" "${CONTEXT_FILE}"
}

add_repo() {
  local repo_id="$1"
  local repo_name
  local tiged_command
  local tmp_file

  if context_has_repo "${repo_id}"; then
    err "Repo is already tracked: ${repo_id}"
    exit 1
  fi

  repo_name="$(parse_repo_name "${repo_id}")"
  tiged_command="$(resolve_tiged_command)"

  sync_single_repo "${repo_id}" "${repo_name}" "${tiged_command}"

  tmp_file="$(mktemp)"
  jq --arg repo "${repo_id}" --arg name "${repo_name}" '
    .repos += [{
      repo: $repo,
      name: $name,
      last_sync: null
    }]
  ' "${CONTEXT_FILE}" > "${tmp_file}"
  mv "${tmp_file}" "${CONTEXT_FILE}"

  update_last_sync "${repo_id}" "$(date "${DATE_FORMAT}")"
  echo "Added ${repo_id}"
}

remove_repo() {
  local repo_id="$1"
  local repo_name
  local repo_path
  local tmp_file

  if ! context_has_repo "${repo_id}"; then
    err "Repo is not tracked: ${repo_id}"
    exit 1
  fi

  repo_name="$(jq -r --arg repo "${repo_id}" '
    .repos[] | select(.repo == $repo) | .name
  ' "${CONTEXT_FILE}")"
  repo_path="${DOCS_DIR}/${repo_name}"

  if [[ -z "${repo_name}" || "${repo_name}" == "null" ]]; then
    err "Tracked repo is missing a directory name: ${repo_id}"
    exit 1
  fi

  if [[ "${repo_path}" != "${DOCS_DIR}/"* ]]; then
    err "Refusing to remove path outside docs directory: ${repo_path}"
    exit 1
  fi

  if [[ -e "${repo_path}" ]]; then
    rm -rf "${repo_path}"
  fi

  tmp_file="$(mktemp)"
  jq --arg repo "${repo_id}" '
    .repos |= map(select(.repo != $repo))
  ' "${CONTEXT_FILE}" > "${tmp_file}"
  mv "${tmp_file}" "${CONTEXT_FILE}"

  echo "Removed ${repo_id}"
}

list_repos() {
  jq -r '
    if (.repos | length) == 0 then
      "No tracked repos."
    else
      (.repos[] |
        "\(.repo)\t\(.name)\t\(.last_sync // "never")")
    end
  ' "${CONTEXT_FILE}"
}

sync_repos() {
  local tiged_command
  local item_count
  local repo_id
  local repo_name

  item_count="$(jq '.repos | length' "${CONTEXT_FILE}")"
  if [[ "${item_count}" == "0" ]]; then
    echo "No tracked repos to sync."
    return
  fi

  tiged_command="$(resolve_tiged_command)"

  while IFS=$'\t' read -r repo_id repo_name; do
    sync_single_repo "${repo_id}" "${repo_name}" "${tiged_command}"
  done < <(
    jq -r '.repos[] | [.repo, .name] | @tsv' "${CONTEXT_FILE}"
  )
}

main() {
  local command_name="${1:-}"

  require_command "jq"
  ensure_layout
  validate_context

  case "${command_name}" in
    add)
      if [[ $# -ne 2 ]]; then
        err "add requires exactly one repo identifier"
        usage
        exit 1
      fi
      add_repo "$2"
      ;;
    rem)
      if [[ $# -ne 2 ]]; then
        err "rem requires exactly one repo identifier"
        usage
        exit 1
      fi
      remove_repo "$2"
      ;;
    ls)
      if [[ $# -ne 1 ]]; then
        err "ls does not accept extra arguments"
        usage
        exit 1
      fi
      list_repos
      ;;
    sync)
      if [[ $# -ne 1 ]]; then
        err "sync does not accept extra arguments"
        usage
        exit 1
      fi
      sync_repos
      ;;
    ""|-h|--help|help)
      usage
      ;;
    *)
      err "Unknown command: ${command_name}"
      usage
      exit 1
      ;;
  esac
}

main "$@"
