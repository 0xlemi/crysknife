# Principles

Shared guardrails for all agents. Every agent reads this file on startup.

## Area Boundaries (CRITICAL)
- Each worker is assigned a specific area of the codebase in their task file.
- NEVER modify files outside your assigned area.
- If you need a change outside your area, write it in your Feedback section and the mayor will assign it to the correct worker.
- If two workers need to change the same file, that is a planning error. Flag it immediately.

## Code Consistency
- Follow patterns established in design.md.
- When in doubt about a pattern, check existing code in the same directory first.
- Don't introduce new dependencies without noting it in your Feedback section.

## Git Discipline
- Commit frequently with descriptive messages.
- Only commit to your assigned branch.
- Never push to main directly. The merger handles that.
- Never merge other branches into yours. Work from your branch only.

## Communication
- Write discoveries, blockers, and questions in your task file's Feedback section.
- Don't guess at product requirements. Flag them for the mayor.
- If design.md contradicts what you see in the code, flag it — don't silently "fix" it.

## Anti-Patterns (DO NOT do these)
- Don't create files outside your assigned area
- Don't refactor code unrelated to your task
- Don't change shared configuration files without mayor approval
- Don't assume how another worker's code works — read design.md instead
