<!-- 
Sync Impact Report
==================
Version Change: 0.0.0 → 1.0.0 (Initial adoption)
Added Sections: All principles newly defined
Templates Requiring Updates: 
  ✅ Updated - constitution.md
  ⚠ Pending - plan-template.md (review for test coverage requirements)
  ⚠ Pending - spec-template.md (review for testing requirements)
  ⚠ Pending - tasks-template.md (review for test task categories)
Follow-up TODOs: 
  - RATIFICATION_DATE needs to be confirmed (currently set to today)
-->

# x402-go Constitution

## Core Principles

### I. No Unnecessary Documentation
No markdown files or documentation shall be created unless explicitly 
prompted by the user. Documentation must serve a clear, requested purpose.
Every piece of documentation must be justified by explicit user need, not 
preemptive assumption.

### II. Test Coverage Preservation
Test coverage must never decrease. Every change to the codebase must maintain
or improve the existing test coverage percentage. This is measured and 
enforced through automated tooling (`go test -cover`). Coverage regressions
block merges without exception.

### III. Test-First Development
All new features must have tests written before implementation. The 
development cycle follows: write test → verify test fails → implement feature
→ verify test passes. This ensures features are testable by design and 
requirements are clearly understood before coding begins.

### IV. Stdlib-First Approach
Prefer Go standard library packages when they reasonably solve the problem.
External dependencies should only be introduced when the stdlib solution would
be significantly more complex or less performant. Each external dependency 
must be justified with clear rationale documenting why stdlib is insufficient.

### V. Code Conciseness
Keep code concise and readable. Avoid unnecessary abstractions, verbose 
naming, or complex hierarchies. Code should be as simple as possible but no
simpler. Favor clarity and directness over clever solutions. Every line should
earn its place through clear value addition.

### VI. Binary Cleanup
Build artifacts and compiled binaries must never be committed to the repository.
After building examples or running compilation tests, all binary files must be
removed before committing changes. This includes executables, .exe files, .out
files, and any other compiled artifacts. The .gitignore must properly exclude
these files, and developers must verify no binaries exist before creating commits.

## Development Standards

### Testing Requirements
- Unit tests required for all packages
- Integration tests required for inter-package communication
- Table-driven tests preferred for comprehensive coverage
- Mock external dependencies to ensure test isolation
- All tests must run with `-race` flag to detect race conditions
- Minimum coverage threshold: maintain or exceed existing levels

### Code Quality Gates
- All code must pass `go fmt` formatting
- All code must pass `go vet` static analysis
- All code must pass `golangci-lint` linting
- All tests must pass before merge
- Coverage reports reviewed on every PR
- No compiled binaries or build artifacts in commits

### Binary Management
- Remove all binaries after building examples: `find examples -type f -executable -delete`
- Verify clean state before commits: `git status` should show no binaries
- Use .gitignore patterns to exclude: `*.exe`, `*.out`, `*.test`, executable files
- Build artifacts belong in `bin/`, `build/`, or `dist/` (all gitignored)

## Governance

The Constitution supersedes all other development practices and guidelines.
Any amendments to these principles require:
1. Clear documentation of the change and rationale
2. Review and approval through standard PR process
3. Version increment following semantic versioning
4. Update of Last Amended date

All pull requests and code reviews must verify compliance with these 
principles. Violations must be corrected before merge. Use AGENTS.md for
Go-specific development guidance and tooling commands.

**Version**: 1.3.0 | **Ratified**: 2025-10-28 | **Last Amended**: 2025-10-28

## Amendment History

### Version 1.3.0 (2025-10-28)
- Added `-race` flag requirement to Testing Requirements
- Mandates race condition detection for all test runs

### Version 1.2.0 (2025-10-28)
- Added `golangci-lint` requirement to Code Quality Gates
- Enforces comprehensive linting beyond basic `go vet` checks

### Version 1.1.0 (2025-10-28)
- Added Principle VI: Binary Cleanup
- Added Binary Management section to Development Standards
- Mandates removal of all compiled binaries before commits