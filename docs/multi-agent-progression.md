# Multi-Agent Development Progression

A personal roadmap for evolving from single-agent development to multi-agent orchestration, inspired by Steve Yegge's Gas Town article and adapted to a tmux + nvim + kiro-cli + lazygit workflow.

---

## Stage Overview

```
Stage 5    Single agent, sequential work
Stage 6    3-5 agents in parallel, manual coordination
Stage 6.5  Conventions, area ownership, workflow tiers
Stage 7    6-15 agents, Crysknife orchestrator (with upgrades)
Stage 8    20+ agents, Gas Town or equivalent
```

---

## Concepts: Gas Town vs Our Workflow

Before diving into stages, here's how Gas Town's concepts map to what we already have and what's new to learn from.

### What We Already Have

```
OUR WORKFLOW                               GAS TOWN EQUIVALENT
────────────────────────────────────────   ────────────────────────────────
.kiro/specs/requirements.md                Role Bead (shared context that
.kiro/specs/design.md                      defines how agents should behave
                                           and what standards to follow)

.kiro/specs/tasks.md                       Molecule (ordered workflow chain
  - [ ] Task 1                             with checkboxes = bead status)
  - [x] Task 2

SPEC_STYLE_GUIDE.md                        Formula / Protomolecule
  (every feature follows the same          (template that generates
   requirements -> design -> tasks          consistent workflows)
   structure)

Acceptance criteria traceability           NO EQUIVALENT in Gas Town.
  AC 1.2 -> Property 3 -> Task 4.1        We're ahead here.

tmux sessions + window switching           tmux sessions (same)

lazygit                                    Merger agent (we do it manually
                                           now, merger agent at Stage 7)

You deciding what to work on next          Mayor agent (you are the mayor
                                           now, mayor agent at Stage 7)

You checking if things are done            crys watch (you are the monitor
                                           now, automated at Stage 7)

/compact or new session                    Task files + state.json
                                           (Gas Town: GUPP + Hooks.
                                           Ours is file-based, not
                                           automatic yet)
```

### What Gas Town Has That We Don't (Yet)

These concepts from Gas Town were adopted into Crysknife (see "What We Adopted" below and `crysknife-design.md` for full details):

- **Workflow tiers** (full/standard/quick) — match effort to task size
- **Per-agent task files** — split tasks.md so agents don't collide
- **Convoys** — feature-level tracking across multiple agents
- **Task files as hooks** — persistent assignment that survives session death
- **Automated monitoring** (crys watch) — detect idle/dead agents, auto-nudge/restart
- **Shared files** — agents communicate through task files, plan.md, state.json (not direct messaging)
- **crys sling / crys nudge** — dispatch and poke commands

What we skipped (not needed at our scale):
- **Wisps** — ephemeral beads, only useful at 20+ agents
- **gt seance** — session recovery, `/chat save` covers most of this

### The MEOW Stack — Gas Town's Work Layers

Gas Town organizes work in 5 layers (Beads → Epics → Molecules → Protomolecules → Formulas). Our equivalents:

```
MEOW LAYER          OUR EQUIVALENT
────────────────    ──────────────────────────────────────────
Beads               Individual acceptance criteria + subtasks
Epics               A feature's full spec folder
Molecules           tasks.md (ordered checklist)
Protomolecules      SPEC_STYLE_GUIDE.md templates
Formulas            The requirements -> design -> tasks flow
```

### Workflow Tiers — Matching Effort to Task Size

Gas Town uses named workflow variants (called "shiny/chrome" for full, "quick" for minimal). We use three tiers:

```
"Full" tier (our current spec guide):
  requirements.md -> design.md -> tasks.md (with PBT checkpoints)
  For: new features, complex changes, architectural work

"Standard" tier:
  brief requirements -> tasks.md (with unit tests, no PBT)
  For: medium features, refactors, multi-file changes

"Quick" tier:
  just a task description -> implement -> basic test
  For: bug fixes, small tweaks, docs, one-file changes
```

### Patrols — Automated Monitoring Loops (crys watch)

`crys watch` is Crysknife's patrol loop. It uses two detection methods:

```
PRIMARY: Heartbeat timestamps (from stop hooks)
  - Every agent has a stop hook that runs crys heartbeat after each response
  - crys watch reads last_activity timestamps from state.json
  - No heartbeat for 2 min → IDLE (auto-nudge)
  - No heartbeat for 5 min → DEAD (auto-restart)

FALLBACK: tmux pane content diffing
  - Capture pane content, hash it, compare to previous
  - Used when heartbeat hooks fail or aren't configured
```

### What We Adopted from Gas Town

All of these are now part of the Crysknife design (see `crysknife-design.md`):

1. **Workflow tiers** (full/standard/quick) — template-based task assignment
2. **Per-agent task files** — split tasks.md, one per worker
3. **Convoys** — feature-level tracking via `crys convoy`
4. **State file** — .crysknife/state.json as single source of truth
5. **Task files as hooks** — persistent assignment that survives session death
6. **Automated monitoring** — `crys watch` with heartbeat + pane diffing
7. **Git worktrees** — per-worker code isolation with symlinked shared files
8. **Three roles** — mayor (plans), workers (code), merger (merges)
9. **Staging branch** — merge/staging for safe merges before main

What we added that Gas Town doesn't have:
- **preToolUse area enforcement** — hard block on writes outside assigned area
- **kiro-cli native integration** — agent configs, hooks, preToolUse guards instead of fighting the tool
- **Traceability** — acceptance criteria chain that Gas Town explicitly skips

---

## Stage 5 — Where We Started

One kiro-cli session, one task at a time. Work is sequential.

### tmux Layout

```
Window 1: nvim        (editor)
Window 2: kiro-cli    (single agent)
Window 3: lazygit     (git management)
Window 4: terminal    (builds, tests, misc)
```

### Context Files

```
.kiro/specs/
├── requirements.md    <- what to build
├── design.md          <- how to build it
└── tasks.md           <- sequential checklist, one agent works top to bottom
```

### How It Works

- You switch between tmux windows with `C-b <number>`
- One agent reads the spec files, works through tasks.md
- When context fills up, `/compact` summarizes or you start a new session
- Each new session starts fresh — reloads spec files, no conversational memory
- State lives in spec files on disk
- You are the agent's manager, memory, and merge tool
- All features use the "full" workflow tier (SPEC_STYLE_GUIDE.md)

### Limitations

- Only one thing happens at a time
- Agent finishes a task, waits for you to assign the next
- Big features take many sequential sessions
- No workflow tiers — a one-line fix gets the same process as a new feature

---

## Stage 6 — Parallel Agents, Manual Coordination

Run 3-5 kiro-cli agents simultaneously, each on its own branch and task list.

### What Changes from Stage 5

- Multiple agents run in PARALLEL (not sequential)
- tasks.md splits into per-agent task files
- Each agent on its own branch in its own area
- Introduce workflow tiers (full/standard/quick)
- You manually dispatch, monitor, and merge

### tmux Layout

```
Window 1: nvim         (editor / overview)
Window 2: agent-1      (kiro-cli — feature work)
Window 3: agent-2      (kiro-cli — second feature or bug fixes)
Window 4: agent-3      (kiro-cli — tests / docs / small tasks)
Window 5: lazygit      (merge branches, resolve conflicts)
Window 6: terminal     (builds, tests, misc)
```

### tmux Session Script

```bash
#!/bin/bash
SESSION="chip8"
PROJECT=~/Documents/code/projects/chip8-emulator

tmux has-session -t $SESSION 2>/dev/null

if [ $? != 0 ]; then
  tmux new-session -d -s $SESSION -c $PROJECT

  tmux rename-window -t $SESSION:1 'nvim'
  tmux send-keys -t $SESSION:1 'nvim .' C-m

  tmux new-window -t $SESSION:2 -n 'agent-1' -c $PROJECT
  tmux send-keys -t $SESSION:2 'kiro-cli chat' C-m

  tmux new-window -t $SESSION:3 -n 'agent-2' -c $PROJECT
  tmux send-keys -t $SESSION:3 'kiro-cli chat' C-m

  tmux new-window -t $SESSION:4 -n 'agent-3' -c $PROJECT
  tmux send-keys -t $SESSION:4 'kiro-cli chat' C-m

  tmux new-window -t $SESSION:5 -n 'lazygit' -c $PROJECT
  tmux send-keys -t $SESSION:5 'lazygit' C-m

  tmux new-window -t $SESSION:6 -n 'terminal' -c $PROJECT

  tmux select-window -t $SESSION:1
fi

tmux attach -t $SESSION
```

### Context Files — Split Tasks Per Agent

```
.kiro/specs/
├── requirements.md        <- shared, read-only for agents
├── design.md              <- shared, read-only for agents
└── tasks/                 <- one file PER agent
    ├── agent-1.md
    ├── agent-2.md
    └── agent-3.md
```

Shared spec files guarantee consistent code style across all agents. Per-agent task files prevent agents from stepping on each other.

### Per-Agent Task File Format

```markdown
# Agent 1 — Auth Module

## Context
Read: requirements.md, design.md
Work in: src/auth/, src/middleware/
Branch: feat/a1-auth
Workflow tier: full

## Tasks
- [ ] Implement JWT token generation
- [ ] Add refresh token endpoint
- [ ] Wire up auth middleware

## Rules
- Follow patterns in design.md
- Don't touch files outside your area
- When done, update status to DONE at the top of this file
```

### Workflow Tiers in Practice

At Stage 6, you start matching effort to task size:

```
"Full" tier agent (agent-1):
  Has: requirements.md, design.md, tasks/agent-1.md with PBT checkpoints
  For: the main feature being built

"Standard" tier agent (agent-2):
  Has: brief requirements in task file, tasks with unit tests
  For: medium refactors, secondary features

"Quick" tier agent (agent-3):
  Has: just task descriptions in task file
  For: bug fixes, docs, small tweaks
```

### How It Works

```
  requirements.md ──────┐
  design.md ─────────┐  │
                     │  │  (shared context = consistent code)
          ┌──────────┤  ├──────────┐
          v          v  v          v
    ┌──────────┐ ┌──────────┐ ┌──────────┐
    │ agent-1  │ │ agent-2  │ │ agent-3  │
    │ "full"   │ │"standard"│ │ "quick"  │
    │ tier     │ │ tier     │ │ tier     │
    │ reads:   │ │ reads:   │ │ reads:   │
    │ design + │ │ design + │ │ design + │
    │ its own  │ │ its own  │ │ its own  │
    │ tasks    │ │ tasks    │ │ tasks    │
    └────┬─────┘ └────┬─────┘ └────┬─────┘
         │            │            │
    feat/a1-auth feat/a2-api  feat/a3-fixes
         │            │            │
         └────────────┼────────────┘
                      v
                  you merge
                 (lazygit)
```

### The Crew Cycling Cadence

The most important workflow at Stage 6 is the assign-ignore-harvest loop:

```
1. ASSIGN — Cycle through each agent (C-b <number>),
   give each one a task, then leave it alone.

2. IGNORE — Don't watch them work. Go do something else.
   Read docs, plan the next batch, review design.

3. HARVEST — Cycle back through agents. Some are still
   working (skip them). Some are done. For each finished
   agent:
   - Read what it did (scroll back, understand the output)
   - Accept the work, or send it back with corrections
   - Assign the next task
   - Move to the next agent

4. REPEAT
```

This cadence is more important than the number of agents. Three agents with a good rhythm beats ten agents you're watching type.

Key discipline: when you see an agent is finished, stop everything and read its output before moving on. Act on each agent's response or risk losing work.

### Session Hygiene

Keep sessions short. Hand off (start a new session) after every completed task rather than letting sessions run until context fills up. Long sessions accumulate stale context that leads to mistakes. Short sessions with fresh context and a clear task file are more reliable.

Exception: let sessions run long only when accumulating important context for a big design discussion or complex decision.

### What You Learn at This Stage

- How to prompt agents to stay in their lane
- Branch-per-agent discipline
- Which tasks deserve "full" vs "quick" treatment
- The assign-ignore-harvest rhythm
- When agents finish and sit idle (you won't always notice)
- Merge conflicts multiply with more agents
- You start losing track of who's doing what

These pain points tell you what to automate in Stage 7.

---

## Stage 6.5 — Conventions and Tracking Solidified

Same setup as Stage 6, but with explicit ownership rules and feature-level tracking (convoys).

### Add an Agents Definition File

```markdown
# .kiro/specs/agents.md

## agent-1: Frontend
- Works in: src/components/, src/pages/
- Branch prefix: feat/a1-

## agent-2: Backend API
- Works in: src/api/, src/services/
- Branch prefix: feat/a2-

## agent-3: Tests and Docs
- Works in: tests/, docs/
- Branch prefix: feat/a3-
```

### Add Feature-Level Tracking (Convoys)

Instead of tracking individual tasks, track features as units:

```markdown
# .kiro/specs/work-queue.md

## Active Features (Convoys)
- **User Auth** [agent-1: implementing, agent-3: writing tests]
  - feat/a1-auth: JWT endpoints (in progress)
  - feat/a3-auth-tests: integration tests (in progress)
  - feat/a1-auth-middleware: middleware (queued)

- **API Refactor** [agent-2: implementing]
  - feat/a2-refactor: connection pooling (in progress)

## Done
- **Navbar Fix** — merged 2026-02-25

## Up Next
- WebSocket support
- API docs update
```

### Context File Structure at Stage 6.5

```
.kiro/specs/
├── requirements.md        <- what to build (shared)
├── design.md              <- how to build it (shared)
├── principles.md          <- guiding principles + anti-patterns (shared)
├── agents.md              <- who owns what area
├── work-queue.md          <- feature-level tracking (convoys)
└── tasks/
    ├── agent-1.md         <- current assignment + checklist
    ├── agent-2.md
    └── agent-3.md
```

### Prevent Heresies with a Principles File

When multiple agents work in parallel, wrong assumptions spread. One agent guesses how something works, gets it wrong, and the mistake gets into the code. Other agents see it and copy it. Yegge calls these "heresies."

The fix: a shared principles file that every agent reads on startup.

```markdown
# .kiro/specs/principles.md

## Core Principles
- All API responses use camelCase, never snake_case
- Error handling uses Result types, not try/catch
- State lives in stores, components are stateless
- No direct database access from route handlers

## Anti-Patterns (DO NOT do these)
- Don't create global singletons for services
- Don't put business logic in UI components
- Don't use string concatenation for SQL queries
- Don't import from internal packages across module boundaries
```

This is more targeted than design.md. Design describes *how* to build things. Principles describe *what to never do* — the guardrails that prevent agents from drifting.

### Standing Orders (PR Sheriff Pattern)

Give one agent a permanent task that runs on every session startup. This is a lightweight patrol that doesn't need an orchestrator.

```markdown
# tasks/agent-3.md (permanent standing orders)

## Standing Orders (run on every session startup)
1. Check for any TODO/FIXME comments added in the last 5 commits
2. Run the test suite — flag any failures
3. Check for files that have grown over 300 lines — flag for refactor
4. Report findings at the top of this file under "## Latest Report"

## Latest Report
(agent updates this section each session)

## Regular Tasks
- [ ] Current assigned work goes here
```

This agent becomes your quality sentinel. It runs its standing orders first, then moves on to regular tasks. You get a health check on every startup for free.

### Review Sweep / Fix Sweep Workflow

A two-phase pattern for maintaining code quality across multi-agent work:

```
Phase 1 — REVIEW SWEEP
  Agent reviews code (or a specific area)
  Files issues in work-queue.md for every problem found
  Does NOT fix anything — only documents

Phase 2 — FIX SWEEP
  Agents pick up issues from work-queue.md
  Each agent fixes issues in their area
  Normal per-agent branch + merge workflow
```

This separates finding problems from fixing them. The review agent can be thorough without context-switching into implementation. The fix agents get well-defined, isolated tasks — exactly what they're best at.

---

## Stage 7 — Crysknife Orchestrator

Crysknife automates the pain points from Stage 6 and scales from 6 to ~15 agents with progressive upgrades. Full spec: `crysknife-design.md`.

### What Crysknife Automates

```
STAGE 6 PAIN POINT                    CRYSKNIFE SOLUTION
────────────────────────────────────   ────────────────────────────────
You dispatch work manually             crys sling (mayor runs via execute_bash)
You don't notice idle agents           crys watch (heartbeat + pane diffing)
You merge branches in lazygit          Merger agent with staging branch
Agents share one working directory     Git worktrees per worker
You track status in your head          .crysknife/state.json + crys status
You are mayor + monitor + merger       Three roles: mayor, workers, merger
Agents ignore their assignment         Agent configs with hooks
Agents write outside their area        preToolUse hook blocks writes
Agents run dangerous commands          preToolUse hook guards execute_bash
```

### How It Works (Summary)

Three roles: mayor (plans/dispatches), workers (code on isolated branches), merger (merges to main via staging). Communication flows through shared files and `crys` CLI commands (via `execute_bash`), not direct messaging.

kiro-cli native integration: agent configs define prompt, tools, context, and hooks per role. `agentSpawn` hook loads task on startup. `preToolUse` hooks enforce area boundaries and block dangerous commands. `stop` hook updates heartbeat. Agents run `crys` CLI commands for state interaction.

Adaptive planning: the mayor maintains a living plan.md with tasks grouped by readiness (ready/blocked/discovered/done). No rigid phases.

See `crysknife-design.md` for: agent config examples, hook scripts, CLI command details, templates, state file format, architecture diagrams, and implementation phases.

### Scaling Limits and Breaking Points

Designed for 6 agents, stretches to ~15 with upgrades:

```
AGENTS   WHAT BREAKS                         FIX
──────   ─────────────────────────────────   ──────────────────────────────
6        Nothing. Design target.              --
8-10     You are the bottleneck.              Mayor autonomy, dashboard-first.
10-12    tmux layout.                         Session groups.
12-15    JSON state contention.               SQLite upgrade.
15-20    Area boundaries + single merger.     Soft ownership, second merger.
20-25    Mayor context + manual oversight.    Split mayor, autonomous patrols.
25-30    Shared files across worktrees.       Dolt upgrade.
30+      You've rebuilt Gas Town.             Use Gas Town.
```

Past 15, the fundamental assumptions (single mayor, single merger, hard area boundaries, you at the keyboard) break. That's by design — by the time you hit 15 agents, you'll know exactly *why* Gas Town needs 7 roles and a database.

### What to Skip at This Stage

Gas Town roles not needed with you at the keyboard: Deacon, Dogs, Boot, Wisps, Polecats. Add these only when you need unattended operation (20+ agents).

---

## Gas Town Role Mapping Across Stages

```
Gas Town Role        Stage 5    Stage 6    Stage 6.5  Stage 7         Stage 8
                                                      (Crysknife)
────────────────────────────────────────────────────────────────────────────────
Overseer (you)       you        you        you        you             you
Mayor (planner)      you        you        you        mayor agent     automated
Crew (workers)       1 agent    3-5 agents 3-5 agents workers (4-12)  agents
Polecats (ephemeral) --         --         --         --              automated
Refinery (merger)    you+git    you+git    you+git    merger agent    automated
Witness (monitor)    you        you        you        crys watch      automated
Deacon (heartbeat)   --         --         --         --              automated
Dogs (helpers)       --         --         --         --              automated
Boot (health check)  --         --         --         --              automated
```

---

## Context Preservation — How Each Approach Works

```
KIRO (Stage 5-6):
  .kiro/specs/ files on disk ARE the memory.
  Session dies = start fresh, same files.

CRYSKNIFE (Stage 7):
  Agent configs + task files on disk = continuity.
  Session dies = crys watch detects, restarts with config.
  Hooks rebuild full context automatically.

GAS TOWN (Stage 8):
  Dolt database = persistent agent identity.
  Session dies = agent resumes via GUPP.
  Agent identity survives across sessions.
```

---

## Timeline

```
Now              Stage 5. Single agent, tmux, spec-driven workflow.
Week 1-2         Stage 6: run 3-5 agents manually.
                   Introduce per-agent task files and workflow tiers.
Week 2-3         Stage 6.5: solidify conventions.
                   Add principles.md, agents.md, work-queue.md.
Month 1-2        Stage 7: Build Crysknife (6 agents).
                   Phase 1: Foundation (init, start, stop, status, configs)
                   Phase 2: Work (sling, queue, hooks)
                   Phase 3: Monitoring (watch, nudge, auto-restart)
                   Phase 4: Planning (convoy, dashboard)
Month 2-3        Scale to 8-10 agents.
                   Trust the mayor. Dashboard-driven workflow.
Month 3-4        Scale to 10-12 agents.
                   tmux session groups. Consider SQLite upgrade.
Month 4+         Iterate based on real pain points.
                   Each upgrade unlocks the next scaling tier.
Eventually       Try Gas Town, or keep evolving Crysknife.
```

---

## Key Principles

1. **Externalize all state to files.** Don't keep anything in your head or in a single agent's context. If it's not written down in a shared file, it doesn't exist. This scales from 1 agent to 30.

2. **Keep traceability.** Gas Town optimizes for throughput. We optimize for correctness AND throughput. The acceptance criteria -> design properties -> task subtasks -> tests chain is our advantage. Don't drop it.

3. **Match effort to task size.** Not everything needs a full spec. Use workflow tiers (full/standard/quick) so small fixes don't get buried in process.

4. **Automate what hurts.** Don't build the orchestrator upfront. Use Stage 6 manually until specific pain points demand automation. Build exactly what you need.

5. **Shared specs, split tasks.** Design docs and requirements stay shared — they're the quality guarantee. Only the work assignments (task files) get split per agent.

6. **Short sessions, frequent handoffs.** Don't let sessions run until context fills up. Hand off after each completed task. Fresh context with a clear task file beats a long session with accumulated noise.

7. **Assign, ignore, harvest.** The crew cycling cadence is the core rhythm. Give each agent work, leave it alone, come back to read results. Don't watch agents type.

8. **Guard against heresies.** Multiple agents will propagate wrong assumptions through the codebase. A shared principles file with explicit anti-patterns is the cheapest defense. preToolUse hooks are the hard enforcement. Review sweeps are the third line.

9. **Work with the tool, not against it.** Use kiro-cli's native features (agent configs, hooks, context) instead of fighting it with tmux send-keys. The less you have to work around the tool, the more reliable the system.

10. **Scale by upgrading, not redesigning.** Crysknife's architecture handles 6 agents out of the box. Each scaling tier (8-10, 10-12, 12-15) requires a specific upgrade. Know what breaks next and fix it when it hurts, not before.
