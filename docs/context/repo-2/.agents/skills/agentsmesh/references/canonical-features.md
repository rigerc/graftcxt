# AgentsMesh Canonical Feature Files Reference

## Rules (`rules/*.md`)

Rules are markdown files with optional YAML frontmatter. `_root.md` is always applied.

### Frontmatter fields

| Field | Type | Description |
|-------|------|-------------|
| `root` | boolean | Mark as root rule (always applied). Only for `_root.md`. |
| `description` | string | Human-readable name shown in menus |
| `targets` | string[] | Limit generation to specific tool IDs |
| `globs` | string[] | gitignore-style file patterns for context scoping |
| `trigger` | string | Windsurf only: `always_on` \| `model_decision` \| `glob` \| `manual` |
| `codexEmit` | string | Codex CLI only: `advisory` \| `execution` |

### Examples

```markdown
---
description: TypeScript coding standards
targets: [claude-code, cursor, copilot]
globs: ["**/*.ts", "**/*.tsx"]
---

Use strict TypeScript. Never use `any`. Prefer `unknown` for unknown types.
Enforce `exactOptionalPropertyTypes` in tsconfig.
```

```markdown
---
description: Database access patterns
globs: ["src/db/**", "src/repositories/**"]
---

Use parameterized queries. Never concatenate user input into SQL strings.
Always close connections in finally blocks.
```

---

## Agents (`agents/*.md`)

Agents define autonomous subagents with scoped tools, model hints, and lifecycle config.

### Frontmatter fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | **Required.** Agent identifier (used in invocation) |
| `description` | string | Short description shown in menus |
| `tools` | string[] | Permitted tools. Supports wildcard: `Bash(git:*)` |
| `disallowedTools` | string[] | Explicitly denied tools (takes precedence over `tools`) |
| `model` | string | Model hint: `sonnet` \| `opus` \| `haiku` |
| `permissionMode` | string | `ask` \| `default` \| `none` |
| `maxTurns` | number | Maximum conversation turns before stopping |
| `mcpServers` | string[] | MCP server names available to this agent |
| `hooks` | object | Agent-level lifecycle hooks (same schema as hooks.yaml) |
| `skills` | string[] | Skill names available to this agent |
| `memory` | string | Path to external memory file |

### Tool permission syntax

```
Read                     # Specific tool
Bash                     # All Bash commands
Bash(git:*)              # Bash limited to git subcommands
Bash(npm:test,npm:build) # Bash limited to specific npm commands
mcp__server-name__*      # All tools from an MCP server
```

### Example

```markdown
---
name: security-auditor
description: Audits code for security vulnerabilities
tools:
  - Read
  - Grep
  - Bash(git:log,git:diff)
disallowedTools:
  - Write
  - Edit
model: opus
maxTurns: 30
skills: [security-review, owasp-top10]
---

Audit the provided code for OWASP Top 10 vulnerabilities.
Focus on injection flaws, authentication issues, and sensitive data exposure.
Report findings with severity (Critical/High/Medium/Low) and remediation steps.
```

---

## Commands (`commands/*.md`)

Commands define slash-command prompts invokable by name.

### Frontmatter fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | Short description shown in menus |
| `allowed-tools` | string[] | Tools the command can use |

### Example

```markdown
---
description: Review recent git changes for issues
allowed-tools: [Read, Bash, Grep]
---

Review the diff from `git diff HEAD~1` for:
1. Security vulnerabilities
2. Missing error handling
3. Performance regressions
4. Incomplete test coverage

Use $ARGUMENTS to focus on a specific concern if provided.
```

---

## Skills (`skills/{name}/SKILL.md`)

Skills are knowledge packages that provide specialized guidance.

### Frontmatter fields

| Field | Type | Description |
|-------|------|-------------|
| `description` | string | **Required.** Short description used for triggering |

### Supporting files

Reference supporting files with relative paths in SKILL.md:

```markdown
See [schema reference](./references/schema.md) for table definitions.
Use the [query builder script](./scripts/build-query.py) for complex queries.
```

### Directory structure

```
skills/my-skill/
├── SKILL.md               # Required
├── references/            # Optional: documentation loaded as needed
│   └── schema.md
├── examples/              # Optional: working examples
│   └── sample-query.sql
└── scripts/               # Optional: utility scripts
    └── build-query.py
```

---

## Hooks (`hooks.yaml`)

### Structure

```yaml
hooks:
  - event: <EventName>
    matcher: "<regex>"        # Matches against tool name or *
    type: command             # command | prompt
    command: "<shell cmd>"    # Used when type: command
    prompt: "<injected text>" # Used when type: prompt
    timeout: 30               # Seconds (default: 30)
```

### Events

| Event | Fires when |
|-------|-----------|
| `PreToolUse` | Before any tool call |
| `PostToolUse` | After any tool call completes |
| `Notification` | Agent sends a notification |
| `UserPromptSubmit` | User submits a prompt |
| `SubagentStart` | A subagent begins |
| `SubagentStop` | A subagent finishes |

### Hook types

**Command hooks** run a shell script. Exit code 0 = allow, non-zero = block.

```yaml
- event: PreToolUse
  matcher: "Bash"
  type: command
  command: "./scripts/check-dangerous-commands.sh"
  timeout: 5
```

**Prompt hooks** inject text into the agent's context.

```yaml
- event: PostToolUse
  matcher: "Write"
  type: prompt
  prompt: "Confirm the file was written correctly and compiles without errors."
```

### Matcher patterns

```yaml
matcher: "*"             # All tools
matcher: "Bash"          # Exact tool name
matcher: "Bash|Write"    # Either tool (regex alternation)
matcher: "mcp__.*"       # All MCP tools (regex)
```

### Example: full hooks.yaml

```yaml
hooks:
  - event: PreToolUse
    matcher: "Bash"
    type: command
    command: "scripts/validate-bash.sh"
    timeout: 10

  - event: PostToolUse
    matcher: "Write|Edit"
    type: prompt
    prompt: "Verify the changed file has no syntax errors."

  - event: UserPromptSubmit
    matcher: "*"
    type: prompt
    prompt: "Always check for existing tests before modifying code."
```

---

## MCP Servers (`mcp.json`)

### Structure

```json
{
  "servers": {
    "<server-name>": {
      "type": "stdio",
      "command": "<executable>",
      "args": ["<arg1>", "<arg2>"],
      "env": {
        "KEY": "${ENV_VAR}"
      }
    }
  }
}
```

### Server types

**stdio** — subprocess communicating via stdin/stdout:

```json
{
  "servers": {
    "my-tool": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@company/mcp-my-tool"],
      "env": { "TOOL_API_KEY": "${TOOL_API_KEY}" }
    }
  }
}
```

**HTTP** — remote server over HTTP:

```json
{
  "servers": {
    "remote-api": {
      "type": "http",
      "url": "https://mcp.company.internal/api",
      "headers": { "Authorization": "Bearer ${MCP_TOKEN}" }
    }
  }
}
```

Reference MCP servers from agents using the `mcpServers` field:

```markdown
---
name: data-analyst
mcpServers: [my-tool, remote-api]
tools: [mcp__my-tool__*, mcp__remote-api__query]
---
```

---

## Permissions (`permissions.yaml`)

### Structure

```yaml
allow:
  - "<tool-expression>"
deny:
  - "<tool-expression>"
```

### Tool expression syntax

```yaml
allow:
  - "Bash(npm:*)"           # All npm subcommands
  - "Bash(git:commit,git:push)"   # Specific git subcommands
  - "Read"                  # Specific tool
  - "mcp__my-server__*"    # All tools from an MCP server

deny:
  - "Bash(rm:-rf *)"        # Specific dangerous pattern
  - "Bash(sudo:*)"          # All sudo commands
```

---

## Ignore (`../../../../.agentsmesh/ignore`)

gitignore-style patterns specifying files AI tools should skip:

```
# Build artifacts
dist/
build/
*.min.js

# Dependencies
node_modules/
vendor/

# Secrets
.env
.env.*
*.pem
*.key
credentials.json

# Generated files
*.generated.ts
__snapshots__/
```
