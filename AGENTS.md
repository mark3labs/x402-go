# Agent Development Guidelines for x402-go

## Issue Tracking with bd (beads)

**IMPORTANT**: This project uses **bd (beads)** for ALL issue tracking. Do NOT use markdown TODOs, task lists, or other tracking methods.

### Why bd?

- Dependency-aware: Track blockers and relationships between issues
- Git-friendly: Auto-syncs to JSONL for version control
- Agent-optimized: JSON output, ready work detection, discovered-from links
- Prevents duplicate tracking systems and confusion

### Quick Start

**Check for ready work:**
```bash
bd ready --json
```

**Create new issues:**
```bash
bd create "Issue title" -t bug|feature|task -p 0-4 --json
bd create "Issue title" -p 1 --deps discovered-from:bd-123 --json
```

**Claim and update:**
```bash
bd update bd-42 --status in_progress --json
bd update bd-42 --priority 1 --json
```

**Complete work:**
```bash
bd close bd-42 --reason "Completed" --json
```

### Issue Types

- `bug` - Something broken
- `feature` - New functionality
- `task` - Work item (tests, docs, refactoring)
- `epic` - Large feature with subtasks
- `chore` - Maintenance (dependencies, tooling)

### Priorities

- `0` - Critical (security, data loss, broken builds)
- `1` - High (major features, important bugs)
- `2` - Medium (default, nice-to-have)
- `3` - Low (polish, optimization)
- `4` - Backlog (future ideas)

### Workflow for AI Agents

1. **Check ready work**: `bd ready` shows unblocked issues
2. **Claim your task**: `bd update <id> --status in_progress`
3. **Work on it**: Implement, test, document
4. **Discover new work?** Create linked issue:
   - `bd create "Found bug" -p 1 --deps discovered-from:<parent-id>`
5. **Complete**: `bd close <id> --reason "Done"`

### Auto-Sync

bd automatically syncs with git:
- Exports to `.beads/issues.jsonl` after changes (5s debounce)
- Imports from JSONL when newer (e.g., after `git pull`)
- No manual export/import needed!

### MCP Server (Recommended)

If using Claude or MCP-compatible clients, install the beads MCP server:

```bash
pip install beads-mcp
```

Add to MCP config (e.g., `~/.config/claude/config.json`):
```json
{
  "beads": {
    "command": "beads-mcp",
    "args": []
  }
}
```

Then use `mcp__beads__*` functions instead of CLI commands.

### Important Rules

- ✅ Use bd for ALL task tracking
- ✅ Always use `--json` flag for programmatic use
- ✅ Link discovered work with `discovered-from` dependencies
- ✅ Check `bd ready` before asking "what should I work on?"
- ❌ Do NOT create markdown TODO lists
- ❌ Do NOT use external issue trackers
- ❌ Do NOT duplicate tracking systems

For more details, see README.md and QUICKSTART.md.

## Build & Test Commands
- **Build**: `go build ./...` - Builds all packages
- **Test All**: `go test -race ./...` - Runs all tests with race detection
- **Test Single**: `go test -race -run TestName ./path/to/package` - Run specific test with race detection
- **Test Coverage**: `go test -race -cover ./...` - Run tests with coverage and race detection
- **Format**: `go fmt ./...` - Format all Go files
- **Lint (Go)**: `go vet ./...` - Run Go static analysis
- **Lint (Full)**: `golangci-lint run` - Run comprehensive linting

## Code Style & Conventions
- **Package Structure**: Use `cmd/` for executables, `internal/` for private packages, `pkg/` for public libraries
- **Imports**: Group as stdlib, external deps, then internal packages with blank lines between
- **Naming**: Use camelCase for variables/functions, PascalCase for exported items, avoid abbreviations
- **Error Handling**: Always check errors; wrap with context using `fmt.Errorf("context: %w", err)`
- **Comments**: Start with function name for exported functions; use `//` for inline, `/* */` for blocks
- **Testing**: Test files end with `_test.go`; use table-driven tests; mock external dependencies
- **Concurrency**: Prefer channels over mutexes; always handle goroutine lifecycles properly
- **Dependencies**: Use go.mod; run `go mod tidy` after adding/removing deps
- **Project Scripts**: Use `.specify/scripts/bash/` for automation scripts (check-prerequisites.sh, create-new-feature.sh, etc.)

## Module: github.com/mark3labs/x402-go | Go Version: 1.25.1

## Active Technologies
- Go 1.25.1 + Go standard library (net/http, encoding/json, encoding/base64, context) (001-x402-payment-middleware)
- N/A (stateless middleware, nonce tracking delegated to facilitator) (001-x402-payment-middleware)
- File-based persistence for budget tracking (JSON files in user config directory) (002-x402-client)

## Recent Changes
- 001-x402-payment-middleware: Added Go 1.25.1 + Go standard library (net/http, encoding/json, encoding/base64, context)
