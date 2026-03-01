# Spec Documentation Style Guide

This document defines how we write requirements, design documents, and implementation task lists for this project. Follow this guide when creating specs for any new feature or bugfix so the documentation is consistent and machine-readable by LLMs.

---

## File Structure

Every feature lives in its own directory under `.kiro/specs/`:

```
.kiro/specs/{feature-name}/
├── requirements.md
├── design.md
└── tasks.md
```

Use kebab-case for `{feature-name}` (e.g., `conversational-document-creation`, `svelte5-ecosystem-upgrade`).

---

## 1. Requirements Document (`requirements.md`)

### Structure

```markdown
# Requirements Document

## Introduction
One or two paragraphs explaining what this feature does and why it exists.
Be specific about what changes and what stays the same.

## Glossary
Define every domain term, abbreviation, or concept that appears in the
requirements. Use bold for the term name. Each entry is a single line.

- **Term_Name**: Definition of the term in context of this feature.

## Requirements

### Requirement N: Short Descriptive Title

**User Story:** As a [role], I want [goal], so that [benefit].

#### Acceptance Criteria

1. WHEN [trigger], THE [component] SHALL [behavior].
2. WHEN [condition], THE [component] SHALL [behavior].
3. IF [condition], THEN THE [component] SHALL [behavior].
```

### Rules

- Number requirements sequentially: Requirement 1, Requirement 2, etc.
- Number acceptance criteria sequentially within each requirement: 1, 2, 3, etc.
- Use the WHEN/SHALL/IF/THEN pattern for all acceptance criteria. This makes them unambiguous and testable.
- Reference glossary terms using their exact defined names (e.g., `Scoped_Chat`, `Session_Mode`).
- Each requirement has exactly one user story.
- Acceptance criteria are exhaustive — if a behavior isn't listed, it's not required.
- Keep requirements focused on what the system does, not how it does it. Implementation details belong in the design doc.

### Glossary Guidelines

- Define every term that a new team member wouldn't immediately understand.
- Use `Snake_Case` for multi-word terms to make them visually distinct in the document.
- Include technical terms (e.g., `SSE`, `ORM`) and domain terms (e.g., `Estimación`, `Schedule_of_Values`).
- Keep definitions concise — one or two sentences max.

---

## 2. Design Document (`design.md`)

### Structure

```markdown
# Design Document: Feature Name

## Overview
2-3 paragraphs summarizing the feature, the approach, and key decisions.
State what changes and what doesn't.

### Key Design Decisions
Numbered list of the most important architectural choices and why they
were made. These are the decisions a reviewer would question first.

## Architecture
Mermaid diagrams showing high-level flow, backend architecture, and
frontend architecture. Use flowchart or graph syntax.

## Components and Interfaces
Detailed description of every new or modified file. Include:
- File path
- Class/function signatures with types
- Props/interfaces for frontend components
- Method signatures for backend services

## Data Models
New or modified database models, API request/response types, and
frontend TypeScript interfaces. Include full field definitions with types.

## Correctness Properties
Formal properties that must hold true across all valid executions.
These drive property-based tests.

## Error Handling
Tables listing error scenarios and expected behavior, split by
frontend and backend.

## Testing Strategy
Description of the testing approach: which properties get PBT,
which behaviors get unit tests, which edge cases to cover.
```

### Key Design Decisions Section

This is one of the most important sections. List 3-7 decisions as a numbered list. Each decision should explain the choice and the reasoning. Example:

```markdown
1. **Scoped chat sessions are separate from general chat** — they use the same
   SSE streaming infrastructure but have their own conversation records,
   mode-specific prompts, and are linked to the resulting document for audit.
2. **Photos are uploaded AFTER context gathering** — in conversational mode,
   the agent collects data first, then requests photos. This allows intelligent
   photo-to-line-item association.
```

### Architecture Diagrams

Use Mermaid syntax. Include at minimum:

- A high-level flow diagram showing the user journey through the feature.
- A backend architecture diagram showing API layer, service layer, and persistence.
- A frontend architecture diagram showing components, stores, and data flow.

Keep diagrams focused — don't try to show everything in one diagram.

### Components and Interfaces

For every new or modified file, include:

- The file path (new or modified)
- TypeScript/Python interfaces or class signatures
- For frontend components: props interface with types
- For backend services: method signatures with parameter and return types
- For API endpoints: HTTP method, path, request/response shapes

Use code blocks for signatures. Don't include full implementations — just the contract.

### Correctness Properties

These are formal statements about system behavior that must hold for all valid inputs. They drive property-based testing.

Format:

```markdown
### Property N: Short descriptive name

*For any* [input domain], [condition], [expected behavior].

**Validates: Requirements X.Y, Z.W**
```

Guidelines:

- Start with "*For any*" to emphasize universality.
- Reference the specific acceptance criteria each property validates.
- Aim for 7-21 properties per feature depending on complexity.
- Properties should be testable with randomized inputs (PBT) or exhaustive checks.
- Each property maps to a single property-based test.

### Error Handling

Use tables with two columns: Scenario and Behavior. Split into frontend and backend sections. Be specific about HTTP status codes, UI feedback, and recovery actions.

### Testing Strategy

Describe:

- Which properties get property-based tests (PBT) and with which library (fast-check for frontend, Hypothesis for backend).
- PBT configuration: minimum iterations (100 per test).
- Which behaviors get unit tests and what they cover.
- Edge cases to test explicitly.
- Whether E2E tests are needed or if existing ones suffice.

---

## 3. Implementation Tasks (`tasks.md`)

### Structure

```markdown
# Implementation Plan: Feature Name

## Overview
Brief summary of what's being built and the implementation strategy.

**Testing Strategy:**
- Bullet points summarizing the testing approach for this implementation.

## Tasks

- [ ] 1. Task Group Title
  - [ ] 1.1 Subtask title
    - Implementation details as bullet points
    - Specific files to create or modify
    - _Requirements: X.Y, Z.W_

  - [ ] 1.2 Subtask title (unit tests for this group)
    - Test case descriptions
    - Run command: `command to run tests`
    - _Requirements: X.Y_

- [ ] 2. Property Test Checkpoint N
  - [ ] 2.1 Write property-based tests for Properties X, Y, Z
    - **Property X: Name** — description of what to generate and verify
    - **Property Y: Name** — description
    - Run: `command to run PBT`
    - _Validates: Requirements A.B, C.D_

  - [ ] 2.2 Run all tests and verify no regressions
    - Run: `full test suite command`
    - Fix any failures from previous tasks

## Notes
Bullet points with important context, constraints, known issues,
and future considerations.
```

### Task Numbering and Grouping

- Top-level tasks are numbered sequentially: 1, 2, 3, etc.
- Subtasks use dot notation: 1.1, 1.2, 1.3, etc.
- Group related work into a single top-level task (e.g., "Backend: Conversation Persistence Models" contains the ORM, migration, and tests).
- The last subtask in every group is the unit tests for that group.

### The Unit Test Subtask Pattern

Every task group ends with a subtask that writes unit tests for the work done in that group. This subtask:

- Lists specific test cases as bullet points.
- Includes the exact command to run the tests.
- References the requirements being validated.

Example:

```markdown
  - [ ] 1.3 Write unit tests for conversation models
    - Test ConversationModel creation with all fields
    - Test ConversationMessageModel creation and FK relationship
    - Test session_mode accepts valid values
    - Test document_id linking (nullable → set after creation)
    - Run: `backend/.venv/bin/pytest backend/tests/test_conversation_models.py -v`
    - _Requirements: 2.1, 13.4, 13.5_
```

### Property Test Checkpoints

Insert a property test checkpoint every 3-4 task groups. These checkpoints:

- Write PBT for the correctness properties relevant to the work completed so far.
- Run the full test suite to catch regressions.
- Fix any failures before proceeding.

Format:

```markdown
- [ ] N. Property Test Checkpoint X (Backend/Frontend/Both)
  - [ ] N.1 Write property-based tests for Properties A, B, C
    - **Property A: Name** — generate [inputs]; verify [assertion]
    - **Property B: Name** — generate [inputs]; verify [assertion]
    - Library config: `fast-check { numRuns: 100 }` or `@settings(max_examples=100)`
    - Run: `test command`
    - _Validates: Requirements X.Y, Z.W_

  - [ ] N.2 Run all tests and verify no regressions
    - Run: `full test suite command`
    - Fix any failures from tasks A-B
    - Verify existing tests still pass
```

### Requirement Traceability

Every subtask includes a `_Requirements: X.Y_` or `_Validates: Requirements X.Y_` line linking back to the acceptance criteria it implements or validates. This creates a traceable chain from requirements → design → tasks.

### Storybook Stories (Frontend Tasks)

For frontend components, include a Storybook story as part of the subtask (not a separate subtask). Mention the story file name and what states it shows:

```markdown
  - [ ] 6.1 Create DocumentCreationShell component
    - New file: `frontend/src/lib/components/creation/DocumentCreationShell.svelte`
    - [implementation details...]
    - Create Storybook story (`DocumentCreationShell.stories.ts` + `DocumentCreationShellDemo.svelte`)
      showing entry point selector for each document type, conversational vs manual states
    - _Requirements: 1.1, 1.2, 1.3_
```

### Optional Tasks

Mark optional tasks with an asterisk after the checkbox:

```markdown
- [ ]* Optional task description
```

### Checkbox Status

- `- [ ]` — Not started
- `- [-]` — In progress
- `- [x]` — Completed
- `- [~]` — Queued

### Notes Section

End the tasks document with a `## Notes` section containing:

- Important context that affects implementation (e.g., "This builds on top of the UI overhaul").
- Known constraints or limitations (e.g., "Svelte 4 constraints still apply").
- Decisions about what's included vs. excluded.
- References to related specs or future work.
- Any test compatibility notes or known issues to watch for.

This section is free-form bullet points. It's the place for anything that doesn't fit in a task but is important for the implementer to know.

---

## 4. Future Features Notes

When completing a spec, add a section at the end of the Notes (or in a separate `docs/FUTURE_FEATURES.md`) capturing potential follow-up work. For each future feature, include:

- A summary of what it does and why.
- Enough implementation detail (data model changes, API endpoints, frontend components) that it can be turned into its own spec later.
- Dependencies on other specs.
- Risks or open questions.

This prevents good ideas from getting lost and gives future specs a head start.

---

## General Writing Guidelines

- Be specific. "Update the component" is bad. "Replace `export let` with `$props()` in `Header.svelte`" is good.
- Use file paths. Always reference the exact file being created or modified.
- Use code blocks for interfaces, types, and signatures.
- Use Mermaid for diagrams — they render in most markdown viewers and are version-controllable.
- Use tables for structured comparisons (package versions, error handling, migration patterns).
- Keep sentences short and direct. This is technical documentation, not prose.
- Reference requirements by number everywhere — in design properties, in task subtasks, in test descriptions. Traceability is the point.

---

## Quick Reference: Document Flow

```
requirements.md          design.md                tasks.md
┌──────────────┐    ┌──────────────────┐    ┌──────────────────┐
│ Introduction │    │ Overview         │    │ Overview         │
│ Glossary     │    │ Key Decisions    │    │ Testing Strategy │
│              │    │ Architecture     │    │                  │
│ Requirement 1│───▶│ Components       │───▶│ Task Group 1     │
│  User Story  │    │ Data Models      │    │  Subtask 1.1     │
│  AC 1.1      │    │                  │    │  Subtask 1.2     │
│  AC 1.2      │    │ Properties ──────│───▶│  Unit Tests 1.3  │
│              │    │  Property 1      │    │                  │
│ Requirement 2│    │  Property 2      │    │ PBT Checkpoint   │
│  ...         │    │                  │    │  Property Tests  │
│              │    │ Error Handling   │    │  Regression Run  │
│              │    │ Testing Strategy │    │                  │
│              │    │                  │    │ Task Group N     │
│              │    │                  │    │  ...             │
│              │    │                  │    │  Unit Tests      │
│              │    │                  │    │                  │
│              │    │                  │    │ Final Checkpoint │
│              │    │                  │    │                  │
│              │    │                  │    │ Notes            │
│              │    │                  │    │  Future Features │
└──────────────┘    └──────────────────┘    └──────────────────┘

Traceability chain:
  AC 1.2 ──▶ Property 3 ──▶ Task 4.1 (PBT for Property 3)
  AC 1.2 ──▶ Task 2.1 (_Requirements: 1.2_)
```
