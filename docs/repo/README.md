> **Reminder:** run `npx skillgrab` to scan your workspace and fetch recommended skills.

# repo

A scaffold for AI-assisted development across multiple coding agents (Claude Code, Codex CLI, OpenCode, Zed, VS Code Copilot). It wires together four tools:

| Tool | Purpose |
|------|---------|
| **rulesync** | Sync rules, skills, hooks, MCP config, and commands across all agent configs from a single source |
| **skills** | Install and update shared agent skills (e.g. `td-task-management`) |
| **td** | Task management across context windows — tracks what to work on and hands off state between sessions |
| **dg** (`scripts/dg.sh`) | Pull external repos into `./docs/` as read-only context using tiged/degit |

## Quick Start

```bash
# 1. Clone and enter
git clone <this-repo> && cd repo

# 2. Run one-time setup (creates .env, .envrc, sets up td, makes scripts executable)
cat .repo.json   # shows the exec steps — run them manually or via your agent harness

# 3. Sync agent configs from the rulesync source
rulesync generate -s

# 4. Install/update skills
skills update -y

# 5. Start a task session
td usage --new-session
```

## How It Works

### rulesync — Single source of truth for agent config

All agent configuration lives under `.rulesync/`:

```
.rulesync/
  rules/         # Shared rules (become CLAUDE.md, AGENTS.md, etc.)
  skills/        # Skills distributed to each agent's config directory
  commands/      # Slash commands for agents that support them
  hooks/         # Lifecycle hooks (init.sh runs at session start)
  mcp.json       # MCP server config
  .aiignore      # Ignore patterns
```

Running `rulesync generate -s` reads `rulesync.jsonc` and writes the appropriate config files for every target (Claude Code → `.claude/`, Codex → `.codex/`, OpenCode → `.opencode/`, etc.).

Targets are defined in `rulesync.jsonc`. Add or remove tools there.

### skills — Install agent skills from GitHub

Skills are versioned in `skills-lock.json`. Run `skills update -y` to pull the latest versions into each agent's skills directory. The lock file pins hashes so installs are reproducible.

To add a skill:
```bash
skills add <github-user>/<repo>
skills update -y
rulesync generate -s   # propagates to all agent configs
```

### td — Task management across sessions

`td` persists tasks across context-window resets. Every session starts with:

```bash
td usage --new-session   # prints current tasks and picks up where you left off
```

Tasks are stored locally and survive agent restarts. The `td-task-management` skill teaches agents the full workflow.

### dg — Pull external repos as context

`scripts/dg.sh` uses tiged/degit to shallow-clone external repos into `./docs/`. Tracked repos are recorded in `.context.json`.

```bash
./scripts/dg.sh add owner/repo        # clone and track
./scripts/dg.sh sync                  # re-sync all tracked repos
./scripts/dg.sh ls                    # list tracked repos
./scripts/dg.sh rem owner/repo        # remove
```

Docs are excluded from git (see `.gitignore`) — they're always re-fetched.

## Session Init Hook

`.rulesync/hooks/init.sh` runs automatically at the start of each agent session (via hooks in each agent's config). It:

1. Updates skills (`skills update -y`)
2. Imports skills into rulesync (`rulesync import -t agentsskills -f skills`)
3. Regenerates agent configs (`rulesync generate -s`)
4. Syncs external docs (`./scripts/dg.sh sync`)
5. Initialises the task session (`td init`)

## Directory Reference

```
.rulesync/          # Source of truth — edit rules/skills/hooks here
.claude/            # Generated: Claude Code config
.codex/             # Generated: Codex CLI config
.opencode/          # Generated: OpenCode config
.agents/            # Generated: Agents (Amp, etc.) config
.zed/               # Generated: Zed AI config
.vscode/            # Generated: VS Code Copilot config
docs/               # External repos pulled by dg.sh (gitignored)
scripts/dg.sh       # External context manager
skills-lock.json    # Pinned skill versions
rulesync.jsonc      # rulesync target + feature config
.context.json       # dg.sh tracked repo list
```
