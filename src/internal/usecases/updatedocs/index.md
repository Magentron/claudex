# UpdateDocs UseCase

Updates index.md documentation files based on git history changes.

## Key Files

- **updatedocs.go** - UseCase orchestrating git-based documentation updates

## Flow

1. Initialize `sessions/` directory for tracking state
2. Read tracking file for last processed commit SHA
3. Validate SHA reachability (fallback to merge-base if unreachable)
4. Compute changed files via `git diff --name-only base..HEAD`
5. Apply skip rules (docs-only, env var, commit tag)
6. Map changed files to affected index.md files
7. Update each index via Claude (using Haiku model)
8. Write tracking file with new HEAD SHA

## State Management

- Tracking state stored in `sessions/` directory (replaces root-level tracking)
- Keeps documentation state organized with other session data
- Enables concurrent operations via file-based locking

## Dependencies

- `internal/doc/rangeupdater` - Core update orchestration
- `internal/services/git` - Git operations
- `internal/services/lock` - Concurrency control
- `internal/services/doctracking` - State persistence
