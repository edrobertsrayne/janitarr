# UI Analysis Prompt for Janitarr Web App

**Objective**: Perform a comprehensive UI analysis of the Janitarr web application to identify issues, improvement opportunities, and necessary E2E tests. Create a plan with recommendations—do not make any code or documentation changes.

## Pre-Analysis

1. Read `specs/README.md` and any relevant specification files to understand the intended design and behavior of the application
2. Note any discrepancies between specs and implementation during analysis

## Setup

1. Start the dev server: `./janitarr dev --host 0.0.0.0`
2. Get host IP: `ip a | grep -oP '(?<=inet\s)\d+\.\d+\.\d+\.\d+' | grep -v '^127' | head -1`
3. Navigate to `http://<HOST_IP>:3434`

**Note**: Test server credentials for Radarr and Sonarr (URLs and API keys) can be found in the `.env` file. Use these when testing server connection features.

## Analysis Instructions

Navigate through every page and interactive element in the application. For each page, capture a snapshot and evaluate:

### 1. Visual & Layout Issues

- Alignment and spacing inconsistencies
- Color contrast and accessibility (WCAG compliance)
- Responsive design problems
- Visual hierarchy and typography
- Loading states and empty states

### 2. Usability Issues

- Unclear navigation or user flows
- Missing feedback on user actions (clicks, form submissions)
- Confusing labels, icons, or terminology
- Missing confirmation dialogs for destructive actions
- Form validation and error messaging

### 3. Functionality Gaps

- Buttons or links that don't work as expected
- Missing keyboard navigation/accessibility
- Broken or missing hover/focus states
- Console errors or warnings

### 4. UX Improvements

- Opportunities to reduce clicks/steps
- Missing helpful tooltips or documentation
- Areas needing progress indicators
- Opportunities for better defaults

---

## Deliverables

**Important**: This analysis is read-only. Do not modify any files. Present findings and recommendations for review.

### 1. Issues List

Categorized by severity (Critical/High/Medium/Low):

- Page/component affected
- Description of the issue
- Suggested fix

### 2. E2E Test Recommendations

List of tests needed to enforce desired/missing behavior:

- Test name and description
- User flow being tested
- Key assertions needed

### 3. UI Improvement Suggestions

Prioritized list:

- Current state vs proposed improvement
- Impact on user experience

### 4. Specification Update Recommendations

List any specification files that need updating to reflect:

- Missing behavior definitions
- Gaps between spec and current implementation
- New features or behaviors that should be documented

### 5. Proposed Implementation Plan

Provide a bullet point task list suitable for adding to `IMPLEMENTATION_PLAN.md`:

```markdown
## Phase XX: UI Improvements & E2E Tests

### UI Fixes

- [ ] Fix: [description of issue and fix]

### UI Enhancements

- [ ] Improve: [description of enhancement]

### E2E Tests

- [ ] Test: [test name] - [what it verifies]

### Specification Updates

- [ ] Update [spec file]: [what needs to be added/changed]
```
