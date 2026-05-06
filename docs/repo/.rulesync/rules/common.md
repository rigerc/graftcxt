# Common Rules

## Code Style

### Organization
- **One responsibility per file** - Split large files (>300 lines) into focused modules
- **Consistent structure** - Imports -> Constants -> Types -> Functions -> Exports
- **Avoid deep nesting** - Maximum 3 levels; extract helper functions
- **No orphaned code** - Delete unused functions, imports, and variables

### Naming

| Element | Convention | Example |
|---------|------------|---------|
| Files | kebab-case | `user-service.ts` |
| Classes | PascalCase | `UserService` |
| Functions | camelCase (JS/TS), snake_case (Python/Kotlin/Rust) | `getUser`, `get_user` |
| Constants | UPPER_SNAKE_CASE | `MAX_RETRIES` |

### Quality
- **Immutability first** - Prefer `const`, `val`, immutable data structures
- **Pure functions** - Minimize side effects; isolate I/O at boundaries
- **Explicit over implicit** - Clear types, explicit returns, no magic
- **Early returns** - Reduce nesting with guard clauses
- **No magic numbers** - Extract constants with descriptive names

### Error Handling
- Catch specific exceptions, not generic ones
- Fail fast with clear messages
- Don't swallow errors - log or re-raise
- Clean up resources with try/finally or context managers

### Comments
- Self-documenting code over comments
- Explain **why**, not what
- No commented-out code - use version control

## Testing

### TDD Workflow
1. Write failing test - Define expected behavior
2. Make it pass - Minimal implementation
3. Refactor - Improve while green

### Coverage

| Scope | Minimum |
|-------|---------|
| Overall | 80% |
| New code | 90% |

### Naming Convention
```
test_<unit>_<situation>_<expected>
```
Examples: `test_login_valid_credentials_returns_token`, `test_payment_insufficient_funds_raises_error`

### Test Structure (Situation / Expected)
```python
def test_user_creation_with_valid_data_succeeds():
    # Situation - valid user data provided
    user_data = {"email": "test@example.com", "name": "Test"}

    # Expected - user is created with correct fields
    user = create_user(user_data)
    assert user.email == "test@example.com"
    assert user.id is not None
```

### Test Categories

| Type | Purpose | Speed |
|------|---------|-------|
| Unit | Single function | <100ms |
| Integration | Module interaction | <1s |
| E2E | Full user flows | >1s |

### Test Quality
- Test behavior, not implementation
- Keep tests independent - no shared state
- Make tests deterministic
- Cover edge cases
- Mock at boundaries only

## Git

### Branch Naming
```
<type>/<description>
```
Types: `feature`, `fix`, `refactor`, `docs`, `test`, `chore`

### Commit Messages
```
type(scope): concise description

- Detail 1
- Detail 2

Co-Authored-By: Claude <noreply@anthropic.com>
```
Types: feat, fix, docs, style, refactor, test, chore

### Pre-Commit Checklist
- `git status` - Verify expected files
- `git diff --cached` - Review changes
- No sensitive files (.env, credentials)
- No debug statements (console.log, print)
- Tests pass

### PR Process
- Same format as commit messages for title
- Include summary, changes, and test plan
- At least 1 approval, all CI checks pass

### Rules
- Never force push to main/master
- Never commit secrets
- Keep commits atomic (one logical change)
- Run tests before pushing

## Performance

### Efficiency
- Only load files you need
- Use parallel tool calls for independent tasks
- Batch git operations into single commands
- Use specific search patterns over broad scans
