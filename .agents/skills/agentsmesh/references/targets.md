# AgentsMesh Targets Reference

## Feature Support Matrix

Run `agentsmesh matrix` to see this live. Legend: **N** = Native, **E** = Embedded, **P** = Partial, **—** = Not supported.

| Target | Rules | Add'l Rules | Commands | Agents | Skills | MCP | Hooks | Ignore | Perms |
|--------|-------|-------------|----------|--------|--------|-----|-------|--------|-------|
| `claude-code` | N | N | N | N | N | N | N | N | N |
| `cursor` | N | N | N | — | N | N | — | N | — |
| `copilot` | N | E | E | — | — | — | — | N | — |
| `gemini-cli` | N | N | N | N | N | N | N | N | N |
| `cline` | N | N | N | N | N | N | N | N | N |
| `codex-cli` | N | N | N | N | N | N | N | N | N |
| `windsurf` | N | N | N | N | N | N | N | N | — |
| `continue` | N | N | N | N | N | N | N | N | — |
| `junie` | N | N | N | N | N | N | N | N | — |
| `kiro` | N | N | N | N | N | N | N | N | N |
| `roo-code` | N | N | N | N | N | N | N | N | N |
| `antigravity` | N | N | E | N | N | N | N | N | — |
| `goose` | N | N | N | N | N | N | N | N | — |
| `kilo-code` | N | N | N | N | N | N | N | N | N |
| `opencode` | N | N | N | N | N | N | N | N | — |

**Native** — Direct format mapping, full fidelity, full round-trip support.
**Embedded** — Projected with metadata; can be imported back but may lose some structure.
**Partial** — Limited support; may not fully round-trip.
**—** — Feature skipped for this target.

---

## Target-Specific Output Paths

### `claude-code`
```
.claude/CLAUDE.md              # root rules
.claude/rules/*.md             # additional rules
.claude/commands/*.md          # commands
.agents/agents/*.md            # agents
.agents/skills/{name}/SKILL.md # skills
.mcp.json                      # MCP servers
.claude/settings.json          # hooks + permissions
.claudeignore                  # ignore
```

### `cursor`
```
.cursor/rules/*.mdc            # rules (native MDC format)
.cursor/mcp.json               # MCP servers
.cursorignore                  # ignore
```

### `copilot`
```
.github/copilot-instructions.md  # root + additional rules (embedded)
```

### `gemini-cli`
```
GEMINI.md                      # root rules
.gemini/rules/*.md             # additional rules
.gemini/commands/*.md          # commands
.gemini/agents/*.md            # agents
.gemini/skills/*/SKILL.md      # skills
.gemini/mcp.json               # MCP servers
.gemini/hooks.yaml             # hooks
.gemini/permissions.yaml       # permissions
.geminiignore                  # ignore
```

### `codex-cli`
```
AGENTS.md                      # root rules
.codex/rules/*.md              # additional rules
.codex/commands/*.md           # commands
.codex/agents/*.md             # agents
.codex/skills/*/SKILL.md       # skills
.codex/mcp.json                # MCP servers
.codex/hooks.yaml              # hooks
.codex/permissions.yaml        # permissions
.codexignore                   # ignore
```

---

## Plugin Development

Build a plugin to add support for a custom or third-party AI tool.

### Package structure

```
agentsmesh-target-foo-ide/
├── package.json
└── index.js (or index.ts)
```

### Minimal plugin descriptor

```js
// index.js
export const descriptor = {
  id: 'foo-ide',                    // Unique target ID
  generators: {
    generateRules(canonical) {
      // Return array of {path, content} objects
      return canonical.rules.map(rule => ({
        path: `.foo-ide/rules/${rule.slug}.md`,
        content: rule.body,
      }));
    },
    async importFrom(projectRoot) {
      // Return ImportResult[] from existing files
      return [];
    },
  },
  capabilities: {
    rules: 'native',
    additionalRules: 'native',
    commands: 'none',
    agents: 'none',
    skills: 'none',
    mcp: 'none',
    hooks: 'none',
    ignore: 'none',
    permissions: 'none',
  },
  project: {
    managedOutputs: {
      dirs: ['.foo-ide/rules'],
      files: [],
    },
    paths: {
      rulePath: (slug) => `.foo-ide/rules/${slug}.md`,
      commandPath: () => null,
      agentPath: () => null,
    },
  },
  emptyImportMessage: 'No Foo IDE config found.',
  lintRules: null,
  buildImportPaths: async () => {},
  detectionPaths: ['.foo-ide'],
};
```

### Register and use

```bash
# Publish to npm as agentsmesh-target-foo-ide, then:
agentsmesh plugin add agentsmesh-target-foo-ide

# Add to agentsmesh.yaml
# targets: [foo-ide]

agentsmesh generate
```

---

## Programmatic API

```ts
import {
  loadProjectContext,
  generate,
  diff,
  check,
  lint,
  importFrom,
  registerTargetDescriptor,
} from 'agentsmesh';

// Load canonical state
const project = await loadProjectContext(process.cwd());

// Generate (dry-run — no files written)
const results = await generate(project);
// results: Array<{target, path, content, action: 'write'|'delete'|'skip'}>

// Check sync status
const status = await check(project);
// status.drifted: boolean

// Lint canonical files
const issues = await lint(project);

// Import from a tool
const imported = await importFrom(project, 'claude-code');

// Register custom target at runtime
registerTargetDescriptor(myDescriptor);
```

### Subpackage imports

```ts
import { loadConfig } from 'agentsmesh/engine';
import { parseRule, parseAgent } from 'agentsmesh/canonical';
import { getTargetDescriptor } from 'agentsmesh/targets';
```

### Error handling

```ts
import { AgentsMeshError } from 'agentsmesh';

try {
  await generate(project);
} catch (err) {
  if (err instanceof AgentsMeshError) {
    // err.code: 'AM_CONFIG_NOT_FOUND' | 'AM_GENERATION_FAILED' | ...
    console.error(err.code, err.message);
  }
}
```

Error codes: `AM_CONFIG_NOT_FOUND`, `AM_GENERATION_FAILED`, `AM_LINT_ERROR`, `AM_IMPORT_FAILED`, `AM_LOCK_CONFLICT`

---

## CI Drift Detection

Run `agentsmesh check` in CI to fail builds when generated files are out of sync with canonical sources:

```yaml
# .github/workflows/agentsmesh.yml
- name: Check AgentsMesh sync
  run: agentsmesh check
```

Exit codes:
- `0` — All generated files are in sync
- `1` — Drift detected (generated files differ from what `generate` would produce)

Commit `../../../../.agentsmesh/.lock` to version-control the expected state.
