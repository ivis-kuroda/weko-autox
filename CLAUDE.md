# CLAUDE.md

@AGENTS.md

## Notes for Claude Code

- The import above pulls in the root agent context: a quick reference plus a map to the
  detailed docs under `docs/agents/`.
- Detailed context is split by role, not by source package, to avoid loading unrelated
  content: `docs/agents/repository-structure.md` (architecture), `cli-specification.md`
  (CLI behavior), `feature-specification.md` (runtime behavior). Read the one relevant to
  your task from the table in `AGENTS.md` before editing.
