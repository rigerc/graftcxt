# AgentsMesh Configuration Reference

## agentsmesh.yaml — Full Spec

```yaml
version: 1                    # Required. Must be 1.

targets:                      # Required. List of tool IDs to generate for.
  - claude-code
  - cursor
  - copilot
  - gemini-cli
  - cline
  - codex-cli
  - windsurf
  - continue
  - junie
  - kiro
  - roo-code
  - antigravity
  - goose
  - kilo-code
  - opencode

features:                     # Optional. Defaults to all supported features.
  - rules
  - commands
  - agents
  - skills
  - mcp
  - hooks
  - ignore
  - permissions

extends:                      # Optional. Pull shared config from remote sources.
  - name: company-rules       # Arbitrary identifier
    source: github:org/repo@v1.0.0   # or gitlab:org/repo, or local path
    features: [rules, commands]      # Which features to pull
    pick:                            # Optional: select specific items
      rules: [security, performance]
      commands: [deploy]

collaboration:                # Optional. Team sync strategy.
  strategy: merge             # merge | lock | last-wins
  lock_features:              # Optional: features requiring --force to overwrite
    - mcp
    - permissions

conversions:                  # Optional. Project features for targets without native support.
  commands_to_skills:         # Emit commands as skills for these targets
    codex-cli: true
  agents_to_skills:           # Emit agents as skills for these targets
    cline: true
    windsurf:
      project: true
      global: false

overrides:                    # Optional. Per-target feature restrictions.
  cursor:
    features: [rules, commands]   # Only generate these features for cursor
  copilot:
    features: [rules]
```

---

## agentsmesh.local.yaml — Per-Developer Overrides

`agentsmesh.local.yaml` is gitignored. It narrows configuration for the local developer only.

**Rules:**
- Can narrow `targets` (remove items, cannot add new ones)
- Can narrow `features` (remove items, cannot add new ones)
- Cannot introduce new targets or features not in `agentsmesh.yaml`

```yaml
# agentsmesh.local.yaml

targets:
  - claude-code               # Only generate for this tool locally
features:
  - rules
  - commands
```

Use case: A developer only uses Claude Code and doesn't want to generate Cursor/Copilot files locally.

---

## Extends — Shared Configuration

The `extends` field pulls canonical config from external sources and merges it into the local project.

### Source formats

```yaml
extends:
  - source: github:org/repo           # Latest default branch
  - source: github:org/repo@v1.2.0    # Tag
  - source: github:org/repo@main      # Branch
  - source: gitlab:org/repo@abc1234   # Commit SHA
  - source: ./shared/config           # Local path (relative to project root)
```

### Merge precedence (highest to lowest)

1. Local `.agentsmesh/` files
2. Installed packs (`.agentsmesh/packs/`)
3. Extended sources (listed sources, first wins on conflict)

### Selective import with `pick`

```yaml
extends:
  - name: company-base
    source: github:acme/ai-configs@v2.0.0
    features: [rules, agents]
    pick:
      rules: [security, code-style]   # Only import these rule slugs
      agents: [code-reviewer]          # Only import this agent
```

### Pack installation (managed extends)

```bash
agentsmesh install github:acme/ai-configs@v2.0.0
agentsmesh install ./local/shared-config
```

Installed packs are recorded in `.agentsmesh/installs.yaml` and cached under `.agentsmesh/packs/`. Commit both for reproducible installs.

---

## Collaboration Strategies

### `merge` (default)

- 3-way merge when regenerating
- Conflicts resolved with `agentsmesh merge`
- Best for teams where multiple developers contribute to `.agentsmesh/`

### `lock`

- Locked features cannot be regenerated without `--force`
- Protects stable features from accidental overwrites
- `lock_features` specifies which features are locked

```yaml
collaboration:
  strategy: lock
  lock_features: [mcp, permissions, hooks]
```

```bash
agentsmesh generate --force   # Override lock
```

### `last-wins`

- Always overwrites on conflict
- Best for single-developer projects or when canonical is the only source of truth

---

## Conversions

Conversions project features for tools that lack native support.

### `commands_to_skills`

Emit command prompts as skill files for targets that support skills but not commands:

```yaml
conversions:
  commands_to_skills:
    codex-cli: true
    antigravity: true
```

### `agents_to_skills`

Emit agent definitions as skill files. Useful for tools that support skills but not agents:

```yaml
conversions:
  agents_to_skills:
    cline: true
    windsurf:
      project: true    # Generate as project-level skill
      global: false    # Don't generate as global skill
```

---

## Global Mode

Set up user-level AI config at `~/.agentsmesh/` for configs that apply across all projects.

```bash
agentsmesh init --global
agentsmesh import --global --from claude-code
agentsmesh generate --global --targets claude-code,cursor,gemini-cli
```

Global config writes to user-level paths:
- `~/.claude/` for Claude Code
- `~/.cursor/` for Cursor
- `~/.agents/skills/` for skill output
- etc.

Global and project configs are independent. Project config takes precedence for project-level features.
