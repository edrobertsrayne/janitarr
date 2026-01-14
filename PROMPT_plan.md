# Project Analysis & Implementation Planning

## Phase 1: Understand the Project

- Study `specs/*` to learn application requirements and design
- Review @IMPLEMENTATION_PLAN.md (if present) for current status
- Analyze `src/lib/*` to understand shared utilities and components
- Reference application source in `src/*`

## Phase 2: Gap Analysis & Planning

Systematically analyze the codebase against specifications:

- Compare existing `src/*` against `specs/*` to identify gaps
- Search for TODO comments, placeholder implementations, and minimal stubs
- Look for skipped/flaky tests and inconsistent patterns
- Verify assumptions with code search before flagging items as missing

Ultrathink through the findings to create or update @IMPLEMENTATION_PLAN.md as a
prioritized task list:

- Items yet to be implemented (highest priority first)
- Mark items complete/incomplete as you validate
- Break down complex tasks into actionable steps

## Critical Constraints

- **PLAN ONLY** - Do not implement anything yet
- Confirm missing functionality through code search first
- Treat `src/lib` as the project standard library - prefer consolidated
  implementations there
- For missing elements: search first, then document at `specs/FILENAME.md` if
  needed

## Project Goal

[Replace with specific goal]

Work systematically through the codebase, documenting findings and maintaining
an accurate implementation roadmap.
