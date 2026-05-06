# Issue tracker: td

Issues for this repo live in `td`, the local task-management CLI used by agents.

## Conventions

- Run `td usage -q` to inspect current state after the initial session read.
- Create work with `td create ...` when a new issue is needed.
- Read issue details with `td show <id>` or `td context <id>`.
- Start work with `td start <id>`.
- Log progress with `td log <id> "..."`.
- Handoff before stopping with `td handoff <id> --done ... --remaining ...`.
- Submit work with `td review <id>`; a different session must approve.

## When a skill says "publish to the issue tracker"

Create or update a `td` issue. Prefer rich descriptions via temp markdown files when needed.

## When a skill says "fetch the relevant ticket"

Use `td show <id>` and `td context <id>`.
