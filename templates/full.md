# {{AGENT_ID}} — {{TASK_NAME}}

## Context
Read: .kiro/specs/requirements.md, .kiro/specs/design.md, .kiro/specs/principles.md
Work in: {{AREA}}
Branch: {{BRANCH}}
Workflow tier: full

## Tasks
- [ ] Review requirements and design docs for relevant context
- [ ] Design approach for {{TASK_NAME}}
- [ ] Implement {{TASK_NAME}}
- [ ] Self-review against design.md and principles.md
- [ ] Write unit tests
- [ ] Run full test suite, fix any regressions
- [ ] Write summary of what you did at the top of this file
- [ ] Run: crys done {{AGENT_ID}}

## Rules
- Follow patterns in design.md
- Respect guardrails in principles.md
- Don't touch files outside your area ({{AREA}})
- Commit to branch {{BRANCH}}
- If you discover something unexpected, write it under Feedback below
- When fully done, run `crys done {{AGENT_ID}}`

## Feedback
(write any issues, questions, or discoveries here for the mayor)
