# Implementation Workflow

## Phase 1: Preparation

- Study `specs/*` to understand application specifications
- Study @IMPLEMENTATION_PLAN.md for current priorities
- Study application source in `src/*`

## Phase 2: Implementation

1. Select the highest priority item from @IMPLEMENTATION_PLAN.md
2. **Search first** - verify functionality isn't already implemented before
   coding
3. Implement functionality completely per specifications
4. Run tests for the affected code unit
5. Resolve any test failures (including unrelated failures discovered during
   your work)

Ultrathink when debugging or making architectural decisions.

## Phase 3: Validation & Documentation

After successful tests:

- Update @IMPLEMENTATION_PLAN.md to reflect completion
- `git add -A`
- `git commit` with descriptive message
- `git push`
- If no build/test errors remain, create git tag (start at 0.0.0, increment
  patch)

## Critical Practices

**Documentation:**

- Capture the _why_ in documentation and tests, not just the _what_
- Keep @IMPLEMENTATION_PLAN.md current - update immediately when discovering
  issues or completing work
- Update @CLAUDE.md with operational learnings only (correct commands, run
  procedures)
- **Do NOT** put status updates or progress notes in @CLAUDE.md - those belong
  in @IMPLEMENTATION_PLAN.md
- Periodically clean completed items from @IMPLEMENTATION_PLAN.md

**Code Quality:**

- Single sources of truth - no migrations or adapters
- Complete implementations only - no placeholders or stubs
- Add logging if needed for debugging
- Document or resolve any bugs discovered, even if unrelated to current work
- If specs have inconsistencies, ultrathink and update `specs/*` accordingly

**Workflow Efficiency:**

- Resolve unrelated test failures as part of your increment
- Keep documentation concise to avoid context pollution
- Update plans immediately to prevent duplicate efforts

Work systematically: search → implement → test → document → commit.
