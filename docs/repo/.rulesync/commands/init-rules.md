---
description: "Analyze repo and generate new rules with glob patterns"
targets: ["*"]
---

Load the `rulesync` skill, then:

1. Analyze the current repository structure, language, frameworks, and codebase conventions
2. Identify key file types, directories, and patterns that should have specific rules
3. Generate new rule files in `.rulesync/rules/` with:
   - Appropriate frontmatter (`root`, `targets`, `description`, `globs`)
   - Meaningful rule content based on repo analysis
   - Glob patterns that match relevant files for each rule
   - Keep rules small, concise. Keep it to 10-15 lines for each (excluding frontmatter)
4. Follow rulesync rule format from `.rulesync/skills/rulesync/file-formats.md`
5. Suggest running `rulesync generate --targets "*" --features "*"` after review
