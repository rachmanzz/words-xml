# AGENTS.md

## Versioning Rule

**NEVER auto-update version numbers without asking the user first.**

Version changes (e.g., `1.0.1` → `1.1.0`) are the user's decision. Always ask before bumping any version string in:
- `preprocessor.go` (emitted version)
- `verify.go` (expected version)
- `go.mod` (module version)
- Documentation files (`*.md`)
- Test files

When a feature change might warrant a version bump, suggest it to the user and let them decide the version number.

## Knowledge Graph (graphify)

Whenever the codebase changes (edits to source files, docs, tests, or config), run:

```
graphify update .
```

This re-extracts code files and keeps the knowledge graph in `graphify-out/` in sync. Run it after each meaningful change, before wrapping up a task.

## Reference Projects

The following external projects may be referenced as **read-only examples** of words-xml consumers:

- `/home/natadana/Projects/dcdtunning` — Example consumer using words-xml for DOCX rendering

**Rules:**
- NEVER edit another project (external project), for any reason, even with user approval
- words-xml is the source of truth: external consumers must adapt to words-xml, not the other way around
- All changes to words-xml must be general-purpose, not specific to any consumer
- Reference projects are for learning context only, not absolute requirements
- Do not create documentation files, README, or project-specific guides for consumers
