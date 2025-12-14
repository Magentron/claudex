# UpdateDocs UseCase

Updates index.md documentation files based on git history changes.

## Key Files

- **updatedocs.go** - UseCase orchestrating git-based documentation updates

## Flow

1. Read tracking file for last processed commit SHA
2. Validate SHA reachability (fallback to merge-base if unreachable)
3. Compute changed files via `git diff --name-only base..HEAD`
4. Apply skip rules (docs-only, env var, commit tag)
5. Map changed files to affected index.md files
6. Update each index via Claude (using Haiku model)
7. Write tracking file with new HEAD SHA

## Dependencies

- `internal/doc/rangeupdater` - Core update orchestration
- `internal/services/git` - Git operations
- `internal/services/lock` - Concurrency control
- `internal/services/doctracking` - State persistence
