---
name: agentsmesh
description: This skill should be used when the user asks to "set up agentsmesh", "initialize agentsmesh", "configure agentsmesh", "generate AI tool configs", "sync AI coding configs", "import configs to agentsmesh", "add a target to agentsmesh", "create an agentsmesh rule", "create an agentsmesh agent", "create an agentsmesh command", "create an agentsmesh skill", "configure MCP in agentsmesh", "set up hooks in agentsmesh", "share AI configs across tools", "canonical config", or mentions agentsmesh.yaml.
---

# AgentsMesh

AgentsMesh turns fragmented AI tool configurations (Claude Code, Cursor, Copilot, Gemini CLI, Cline, Codex CLI, and 9 others) into one canonical `.agentsmesh/` source of truth. Edit once, generate native config files for every tool.

**Core invariant:** Only edit files inside `.agentsmesh/`. Never edit generated files directly (`.claude/`, `.cursor/`, etc.) — they will be overwritten on the next `generate`.

---

## Canonical Directory Structure

```
.agentsmesh/
├── agentsmesh.yaml        # Project config (targets, features, extends)
├── agentsmesh.local.yaml  # Gitignored per-developer overrides
├── rules/
│   ├── _root.md           # Always-applied root rules (required)
│   └── *.md               # Additional scoped rules
├── commands/
│   └── *.md               # Slash-command prompts
├── agents/
│   └── *.md               # Subagent definitions
├── skills/
│   └── {name}/
│       └── SKILL.md       # Skill definition + supporting files
├── mcp.json               # MCP server definitions
├── hooks.yaml             # Lifecycle hooks
├── permissions.yaml       # Allow/deny lists
├── ignore                 # gitignore-style patterns
├── installs.yaml          # Managed by CLI (installed packs)
└── .lock                  # Managed by CLI (checksums)
```

---

## CLI Commands

### Core workflow

```bash
agentsmesh init              # Scaffold .agentsmesh/ + detect existing configs
agentsmesh generate          # Generate target-specific files from canonical
agentsmesh import            # Import existing tool configs → canonical
agentsmesh diff              # Preview changes without writing
agentsmesh check             # Verify sync status (exit 1 if drift; use in CI)
```

### Utilities

```bash
agentsmesh lint              # Validate canonical files
agentsmesh watch             # Auto-regenerate on file changes
agentsmesh matrix            # Show feature-target compatibility matrix
agentsmesh install <source>  # Install a shared pack (GitHub/GitLab/local)
agentsmesh merge             # Resolve lock file conflicts after git merge
agentsmesh plugin add <pkg>  # Register a third-party target plugin
```

### Global mode

```bash
agentsmesh init --global     # Set up ~/.agentsmesh/
agentsmesh generate --global --targets claude-code,cursor
agentsmesh import --global --from claude-code
```

---

## agentsmesh.yaml Quick Reference

```yaml
version: 1
targets:
  - claude-code
  - cursor
  - copilot
  - gemini-cli
  - codex-cli
features:
  - rules
  - commands
  - agents
  - skills
  - mcp
  - hooks
  - ignore
  - permissions
extends:
  - name: company-rules
    source: github:org/repo@v1.0.0
    features: [rules, commands]
    pick:
      rules: [security, performance]
collaboration:
  strategy: merge          # merge | lock | last-wins
  lock_features: [mcp, permissions]
conversions:
  commands_to_skills:
    codex-cli: true
  agents_to_skills:
    cline: true
overrides:
  cursor:
    features: [rules, commands]
```

See **`./references/configuration.md`** for full field documentation and `agentsmesh.local.yaml` patterns.

---

## Feature Files Quick Reference

### Rules (`rules/*.md`)

```markdown
---
description: Security guidelines
targets: [claude-code, cursor]    # optional: limit to specific tools
globs: ["src/**/*.ts"]            # optional: file-pattern scoping
---

Never store credentials in source code.
Always validate user input at trust boundaries.
```

`../../../AGENTS.md` is always applied to every tool (no frontmatter needed, or add `root: true`).

### Agents (`agents/*.md`)

```markdown
---
name: code-reviewer
description: Reviews code for quality issues
tools: [Read, Grep, Bash(git:*)]
disallowedTools: [Write]
model: sonnet
maxTurns: 20
skills: [security-review]
---

Review the provided code for correctness, security, and style.
```

### Commands (`commands/*.md`)

```markdown
---
description: Generate unit tests for the current file
allowed-tools: [Read, Write, Bash]
---

Generate comprehensive unit tests for $ARGUMENTS.
```

### Skills (`skills/{name}/SKILL.md`)

```markdown
---
description: Analyzes database query performance
---

To analyze a slow query, run EXPLAIN ANALYZE and inspect the output...
```

Skills can reference supporting files with relative paths:
```markdown
See [schema reference](./references/schema.md) for table definitions.
```

### Hooks (`hooks.yaml`)

```yaml
hooks:
  - event: PreToolUse
    matcher: "Bash"
    type: command
    command: "./scripts/validate-bash.sh"
    timeout: 10
  - event: PostToolUse
    matcher: "Write"
    type: prompt
    prompt: "Verify the written file compiles without errors."
```

Events: `PreToolUse`, `PostToolUse`, `Notification`, `UserPromptSubmit`, `SubagentStart`, `SubagentStop`

### MCP Servers (`mcp.json`)

```json
{
  "servers": {
    "my-api": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@company/mcp-server"],
      "env": { "API_KEY": "${MY_API_KEY}" }
    }
  }
}
```

### Permissions (`permissions.yaml`)

```yaml
allow:
  - "Bash(npm:*)"
  - "Bash(git:*)"
deny:
  - "Bash(rm:-rf *)"
```

### Ignore (`../../../.agentsmesh/ignore`)

```
node_modules/
dist/
*.env
*.secret
```

---

## Common Workflows

### New project setup

```bash
cd my-project
agentsmesh init
# Edit .agentsmesh/agentsmesh.yaml to set targets and features
# Edit .agentsmesh/rules/_root.md with project guidelines
agentsmesh generate
```

### Import existing configs then unify

```bash
agentsmesh import --from claude-code
agentsmesh import --from cursor
# Review .agentsmesh/ for duplicates, consolidate
agentsmesh generate
```

### Add a new target mid-project

1. Add the target ID to `targets:` in `agentsmesh.yaml`
2. Run `agentsmesh generate`

### CI drift detection

```bash
agentsmesh check  # exits 1 if generated files are out of sync
```

### Watch mode during active development

```bash
agentsmesh watch  # regenerates on every .agentsmesh/ save
```

---

## Generation Rules

- Canonical sources: `.agentsmesh/` — **you edit these**
- Generated artifacts: `.claude/`, `.cursor/`, `../../../AGENTS.md`, etc. — **CLI writes these**
- Lock file `.agentsmesh/.lock` records checksums; commit it to detect drift
- `agentsmesh.local.yaml` is gitignored; can narrow but not expand targets/features
- Extends merge precedence: local `.agentsmesh/` > installed packs > extended sources

---

## Supported Targets (15 tools)

`claude-code`, `cursor`, `copilot`, `gemini-cli`, `cline`, `codex-cli`, `windsurf`, `continue`, `junie`, `kiro`, `roo-code`, `antigravity`, `goose`, `kilo-code`, `opencode`

Not every target supports every feature. Run `agentsmesh matrix` to see the full compatibility table, or see **`./references/targets.md`** for the detailed support matrix.

---

## Additional Resources

### Reference Files

- **`./references/configuration.md`** — Full `agentsmesh.yaml` spec, `agentsmesh.local.yaml`, extends, collaboration strategies, conversion options
- **`./references/canonical-features.md`** — Complete frontmatter fields for rules, agents, commands, skills, hooks, mcp, permissions
- **`./references/targets.md`** — Feature support matrix for all 15 tools, plugin development guide, programmatic API