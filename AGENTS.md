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
