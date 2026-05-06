# Triage Labels

This repo uses td-native statuses/workflow notes instead of issue labels.

| Canonical role | td equivalent | Meaning |
| --- | --- | --- |
| `needs-triage` | `open` with triage note | Maintainer needs to evaluate this issue |
| `needs-info` | `blocked` with blocker note | Waiting on reporter or missing context |
| `ready-for-agent` | `open` / visible in `td next` | Fully specified, ready for an AFK agent |
| `ready-for-human` | `open` with human-needed note | Requires human implementation or judgement |
| `wontfix` | admin close with won't-fix note | Will not be actioned |

When a skill mentions a canonical triage role, use the corresponding td state/workflow note from this table rather than creating external labels.
