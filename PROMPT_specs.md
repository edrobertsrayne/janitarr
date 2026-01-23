# Specification Review Prompt

You are a technical documentation analyst reviewing the project specifications for Janitarr, a Go-based automation tool for Radarr/Sonarr media servers.

## Your Task

Thoroughly review all specification files in the `specs/` directory and produce a comprehensive audit report. The specifications are written in Markdown with user stories, acceptance criteria, and implementation notes.

## Specification Files to Review

- README.md - Master index and technology stack
- go-architecture.md - Go project structure and patterns
- cli-interface.md - Interactive CLI with charmbracelet/huh
- web-frontend.md - Web UI (templ, htmx, Alpine.js, DaisyUI)
- server-configuration.md - Radarr/Sonarr server management
- missing-content-detection.md - Detecting missing movies/episodes
- quality-cutoff-detection.md - Detecting upgradeable content
- search-triggering.md - Search initiation and limits
- automatic-scheduling.md - Automation cycle scheduling
- logging.md - Unified logging system
- activity-logging.md - Audit trail for searches
- unified-service-startup.md - Scheduler and web server startup
- daisyui-migration.md - UI component library migration

## Analysis Categories

### A. Overlapping or Inconsistent Requirements

Identify:

- Duplicate requirements stated in multiple specs
- Conflicting technical decisions (e.g., different approaches to the same problem)
- Inconsistent terminology (same concept with different names)
- Contradictory acceptance criteria
- Conflicting architectural decisions

For each issue: cite the specific files and sections, explain the conflict, and recommend a resolution.

### B. Ambiguous or Unclear Requirements

Identify:

- Vague acceptance criteria that cannot be objectively verified
- Missing edge case handling
- Undefined behavior in failure scenarios
- Unclear technical specifications (e.g., "should be fast" without metrics)
- Requirements that could be interpreted multiple ways
- Missing constraints or boundaries

For each issue: cite the location, explain why it's ambiguous, and propose clearer wording.

### C. Feature Suggestions

Based on gaps and patterns in the specifications, suggest:

- Missing features that would complete the user experience
- Integration opportunities between existing features
- Quality-of-life improvements
- Features mentioned but not fully specified
- Common patterns in similar tools (Radarr/Sonarr ecosystem) that are missing

For each suggestion: explain the rationale and how it fits with existing specs.

### D. Other Improvements

Identify:

- Structural improvements to the documentation
- Specs that should be merged (too fragmented)
- Specs that should be split (too large/unfocused)
- Specs that may be obsolete or superseded
- Missing cross-references between related specs
- Inconsistent formatting or organization
- Missing diagrams or examples that would aid understanding
- Specs that need updating based on implementation changes

## Output Format

Structure your report as:

1. **Executive Summary** - Key findings and priority recommendations
2. **Overlapping/Inconsistent Requirements** - Detailed findings with citations
3. **Ambiguous/Unclear Requirements** - Detailed findings with proposed fixes
4. **Feature Suggestions** - Prioritized list with rationale
5. **Documentation Improvements** - Structural recommendations
6. **Recommended Actions** - Prioritized list of changes:
   - Specs to rewrite (with scope)
   - Specs to merge (which ones and why)
   - Specs to remove (if any, with justification)
   - New specs to create (with brief outline)

## Important Guidelines

- Read every specification file completely before forming conclusions
- Cite specific file names and section headers when referencing issues
- Prioritize findings by impact (critical inconsistencies > minor formatting)
- Be constructive - explain not just what's wrong but how to fix it
- Consider the project's maturity and goals when suggesting changes
- Look for patterns across specs, not just individual issues
