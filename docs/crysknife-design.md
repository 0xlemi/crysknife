# Crysknife — Design Document

A lightweight agent orchestrator for managing 3-6 kiro-cli agents in tmux. Named after the Fremen crysknife from Dune — a personal weapon, forged from experience. CLI shortcut: `crys`.

---

## Overview

Crysknife is a CLI tool written in Go that manages multiple kiro-cli agents running in tmux sessions. It handles agent lifecycle (start, stop, restart), work assignment, status monitoring, and nudging of idle/stuck agents. It is designed as a learning tool and stepping stone — simple enough to understand completely, useful enough to run daily.

Crysknife does not aim to replace Gas Town. It implements a subset of Gas Town's ideas adapted for a single developer running 3-6 agents on one machine.

### Key Design Decisions

1. **CLI-first, no daemon.** All commands are run by the user. A `crys watch` command runs a polling loop in the foreground for monitoring, but there is no background process. This keeps the system transparent and debuggable. The binary is named `crys` for brevity (short for Crysknife).

2. **tmux is the runtime.** Crysknife does not manage processes directly. It creates tmux sessions/windows/panes and sends commands to them. tmux is the process manager, Crysknife is the orchestrator on top.

3. **JSON file is the state store.** All state lives in a single `.crysknife/state.json` file in the project root, committed to Git. No database. Any agent can read it. The user can read it. It's the single source of truth.

4. **Agents are kiro-cli agent configs.** Each agent role (mayor, worker, merger) is defined as a kiro-cli agent configuration (`.kiro/agents/<role>.json`). Agent configs specify the system prompt, available tools, auto-loaded context files, hooks, and MCP servers. `crys start` generates per-worker configs from templates and launches each agent with `kiro-cli chat --agent <name>`. This replaces tmux send-keys prompting with native kiro-cli integration.

5. **Shared specs, split tasks.** Design docs and requirements stay shared across all agents. Only task files are per-agent. This preserves code quality and consistency.

6. **Workflow tiers.** Not all work needs the same process. Crysknife supports three tiers (full, standard, quick) with templates for each.

7. **Three agent roles.** Not all agents are equal. Crysknife defines three roles: Mayor (plans and dispatches), Workers (code), and Merger (merges branches to main and resolves conflicts). Agents communicate through shared files, not direct messaging.

8. **Adaptive planning, not rigid phases.** The plan is a living document that the mayor updates continuously. No pre-planned phase graph. Tasks are grouped by readiness (ready / blocked / discovered). The mayor adapts as workers finish, get stuck, or discover new requirements. You make product and architecture decisions when the mayor asks.

9. **Git worktrees for isolation.** Each worker and the merger operate in their own Git worktree. Workers never share a working directory. The merger works in the main worktree. `crys start` creates worktrees automatically via `git worktree add`. Shared directories (`.crysknife/`, `.kiro/specs/`) are symlinked back to the main worktree so all agents read/write the same state and task files.

10. **Only `crys` CLI writes state.json.** Agents never modify state.json directly — this avoids concurrent write corruption. Agents run `crys` CLI commands via `execute_bash` (e.g. `crys done worker-1`), which routes through the `crys` binary. A `preToolUse` hook on `execute_bash` blocks dangerous commands.

11. **Staging branch for safe merges.** The merger never merges directly to main. It merges to `merge/staging` first, runs tests there, and only fast-forwards main if tests pass. If tests fail, main stays clean and the merger flags the issue.

12. **`crys watch` nudges the mayor.** When `crys watch` detects a worker has gone idle or completed a task, it nudges the mayor agent (not just the worker) so the mayor can reassign work without you having to notice and intervene manually.

13. **kiro-cli native integration.** Crysknife uses kiro-cli's agent configs, hooks, and context system rather than fighting the tool with tmux send-keys. Hooks provide GUPP (agents auto-load their assignment on startup). Agents run `crys` CLI commands via `execute_bash` for state interaction. Context resources auto-load spec files. preToolUse hooks enforce area boundaries and block dangerous commands.

---

## Agent Roles

Crysknife defines three roles. A typical 6-agent setup: 1 mayor, 3-4 workers, 1 merger.

```
┌──────────────────────────────────────────────────────┐
│                                                      │
│  MAYOR (1 agent — your command center)               │
│  - Updates plan.md, design.md, task files            │
│  - Reads feedback from workers and merger            │
│  - Decides what to do next, assigns work             │
│  - You talk to this one the most                     │
│                                                      │
├──────────────────────────────────────────────────────┤
│                                                      │
│  WORKERS (3-4 agents — do the actual coding)         │
│  - Each gets a task file + branch + area             │
│  - Code, test, commit to their branch                │
│  - When done: write summary in their task file       │
│  - Don't touch main, don't merge                     │
│                                                      │
├──────────────────────────────────────────────────────┤
│                                                      │
│  MERGER (1 agent — the Refinery)                     │
│  - Watches for completed worker branches             │
│  - Merges them to main one at a time                 │
│  - Resolves conflicts using design.md as reference   │
│  - Writes merge reports: what conflicted, what       │
│    changed, any concerns                             │
│  - Flags issues back to mayor                        │
│                                                      │
└──────────────────────────────────────────────────────┘
```

### Feedback Loop

Agents don't message each other directly. All communication flows through shared files:

```
                    ┌─────────┐
          ┌────────│  MAYOR  │◄────────────────────┐
          │        │ (plans) │                      │
          │        └────┬────┘                      │
          │             │                           │
          │  updates    │ assigns tasks             │ merge reports
          │  plan.md    │                           │ + feedback
          │             v                           │
          │   ┌──────────────────┐           ┌─────┴──────┐
          │   │     WORKERS      │           │   MERGER   │
          │   │                  │  done     │            │
          │   │  worker-1 ─────────────────►│  merges    │
          │   │  worker-2 ─────────────────►│  one at    │
          │   │  worker-3 ─────────────────►│  a time    │
          │   │  worker-4 ─────────────────►│  to main   │
          │   │                  │           │            │
          │   └──────────────────┘           └────────────┘
          │         │                              │
          │         │ worker feedback               │
          │         │ "design.md is wrong about Y"  │
          └─────────┘                              │
                                                   │
          merger feedback:                         │
          "auth-backend conflicted with models"    │
          "suggest updating design.md section 3" ──┘
```

Workers write feedback in their task files. The merger writes merge reports in its task file. You read both through the mayor and update the plan accordingly.

### Merger Standing Orders

The merger gets a permanent task file that persists across sessions:

```markdown
# .kiro/specs/tasks/merger.md

## Standing Orders

You are the Merger. Your job is to merge completed worker branches
to main, one at a time, through the staging branch.

## Merge Runbook

For each branch in the merge queue:

### 1. Check what's ready
  crys status
Look for workers with status "done". Pick the branch with the smallest diff:
  git diff main...<branch> --stat

### 2. Prepare staging
  git checkout merge/staging
  git reset --hard main

### 3. Merge the branch
  git merge <branch> --no-edit

If conflicts appear, go to "Conflict Resolution" below.

### 4. Run tests
  go test ./...
(or whatever the project's test command is)

### 5a. Tests pass — land it
  git checkout main
  git merge merge/staging --ff-only
  crys merge-done <branch>
Write a merge report below, then go back to step 1.

### 5b. Tests fail — diagnose
Read the test output carefully.
- If the failure is caused by the merge (missing import, wrong function
  signature after conflict resolution), fix it on staging and re-run tests.
- If the failure is a pre-existing issue or something you can't fix,
  do NOT touch main. Instead:
  crys merge-done <branch> --failed "tests failed: <one-line summary>"
Write what happened in the merge report and move to the next branch.

## Conflict Resolution

When `git merge` reports conflicts:
1. Run `git diff` to see all conflicting files
2. For each conflict:
   - Read both sides of the conflict
   - Check design.md for the intended architecture/patterns
   - Check principles.md for guardrails
   - Pick the approach that matches the design, or combine both if they
     touch different parts of the same file
3. If a conflict involves two workers changing the same logic differently
   and you can't tell which is correct from the docs:
   - Pick the simpler approach
   - Write in Feedback: "conflict in <file>, resolved by keeping <worker>'s
     approach. Mayor please verify."
4. After resolving all conflicts:
   git add .
   git commit --no-edit
   Continue to step 4 (run tests).

## Rules
- NEVER merge two branches at the same time
- ALWAYS run tests after each merge
- NEVER push directly to main — always go through merge/staging
- If a conflict is too complex or ambiguous, flag it — don't guess
- Read design.md and principles.md to resolve conflicts correctly
- If design.md contradicts the code, flag it in Feedback for the mayor

## Merge Reports
(write one entry per merge, most recent first)

### <branch> — <date>
- Files changed: <count>
- Conflicts: none / resolved <list>
- Tests: pass / fail (<summary>)
- Notes: <anything the mayor should know>
```

---

## Parallel Planning

### Philosophy

Plans are living documents, not rigid scripts. You can't predict which worker finishes first, which task turns out harder than expected, or what new work gets discovered mid-flight. The mayor adapts the plan on the fly based on what's actually happening.

The mayor's job is to keep workers busy with the right work at the right time. Your job is to answer the mayor's product and architecture questions. Workers just code.

```
YOU (product decisions, architecture, priorities)
  ↕  conversation
MAYOR (generates/modifies plan, assigns work, adapts)
  ↓  task files
WORKERS (implement, give feedback)
  ↓  completed branches
MERGER (merges, reports conflicts)
  ↓  merge reports + feedback
MAYOR (reads feedback, adapts plan, reassigns)
  ↕  asks you when unsure
YOU
```

### How It Works in Practice

There is no pre-planned phase graph. Instead:

1. You tell the mayor what you want built: "We need user authentication"
2. The mayor asks you product questions: "OAuth or email/password? Which pages need auth? What about password reset?"
3. You answer. The mayor writes an initial plan.md with tasks it can identify
4. The mayor assigns the first batch of parallel tasks to available workers
5. As workers finish or get stuck, the mayor adapts:
   - Worker-1 finished fast → mayor assigns it the next task
   - Worker-2 is stuck → mayor reads its feedback, adjusts the task or reassigns
   - Worker-3 discovered a new requirement → mayor asks you about it, updates plan
   - Merger reports a conflict → mayor updates design.md to prevent future conflicts
6. The plan grows and changes throughout execution

### Plan File

The plan is a snapshot of current understanding, not a contract. The mayor updates it continuously.

```markdown
# .kiro/specs/plan.md

## Goal: User Authentication

## Current Status
Phase: implementation
Workers active: 3
Merge queue: 1 branch ready

## Tasks

### Ready (can be assigned now)
- [ ] Auth backend — JWT endpoints, refresh tokens
      Area: src/auth/
      Tier: full

- [ ] Auth frontend — login form, token storage
      Area: src/components/auth/
      Tier: full

- [ ] Database models — user table, migrations
      Area: src/models/
      Tier: standard

### Blocked (waiting on something)
- [ ] Integration tests — needs backend + frontend merged first
- [ ] Password reset flow — waiting on product decision (asked overseer)

### Discovered During Execution
- [ ] Rate limiting on auth endpoints — worker-1 flagged this as needed
      Area: src/middleware/
      Tier: standard

### Done
- [x] Design auth architecture
- [x] API documentation
```

The key difference from the previous rigid phase approach: there are no numbered phases with explicit dependencies. Instead, tasks are grouped by readiness — what can run now, what's blocked, and what was discovered along the way. The mayor moves tasks between groups as the situation changes.

### The Mayor's Decision Loop

The mayor continuously runs this loop (driven by you or by reading worker/merger feedback):

```
1. Any workers idle?
   → Check "Ready" tasks, pick one, assign it

2. Any workers stuck?
   → Read their feedback
   → Can I solve it? → Update their task file with guidance
   → Need product input? → Ask the overseer
   → Task is wrong? → Reassign to different work

3. Any branches merged?
   → Does this unblock something? → Move from "Blocked" to "Ready"
   → Did the merger flag issues? → Update design.md or principles.md

4. Any new work discovered?
   → Add to plan.md under "Discovered During Execution"
   → Prioritize with overseer if unclear

5. All tasks done?
   → Report to overseer, propose next steps
```

The mayor doesn't execute this as code — it's a kiro-cli agent that you talk to. You say "worker-2 just finished, what's next?" and the mayor reads the plan, checks state, and either assigns work or asks you a question.

### Branch Strategy

Branches are created on-the-fly as the mayor assigns tasks, not pre-planned:

```
main ──────────────────────────────────────────→
  │                                          ↑
  ├──→ feat/auth-models ──→ merger merges ───┤
  │    (worker-3 finished first)             │
  │                                          │
  ├──→ feat/auth-backend ──→ merger merges ──┤
  │    (worker-1 finished second)            │
  │                                          │
  ├──→ feat/auth-frontend ──→ merger merges ─┤
  │    (worker-2 finished third)             │
  │                                          │
  ├──→ feat/rate-limiting ──→ merger merges ─┘
  │    (discovered mid-flight, assigned to
  │     worker-3 after it finished models)
  │
  └──→ feat/auth-docs ──→ merged whenever
       (independent, low priority)
```

The order is determined by who finishes when, not by a pre-planned sequence. The merger handles whatever comes in.

### Your Role as Overseer

You are the product owner. The mayor handles implementation details but comes to you for:

- **Product decisions:** "Should we support OAuth or just email/password?"
- **Priority calls:** "Worker-1 found we need rate limiting. Should I prioritize that or keep going with the current plan?"
- **Architecture questions:** "Worker-2 says the token storage approach in design.md won't work on mobile. What do you want to do?"
- **Scope decisions:** "This is getting bigger than expected. Should we cut password reset from this round?"

You don't need to know which worker is doing what. You just answer the mayor's questions and review the plan when you want to steer direction.

## kiro-cli Integration

Crysknife leverages kiro-cli's native features instead of fighting the tool with tmux send-keys. This section covers how agent configs, hooks, MCP server, and context work together.

### Agent Configs

Each role gets a kiro-cli agent config in `.kiro/agents/`. `crys init` generates the base configs. `crys start` generates per-worker configs dynamically (worker-1.json through worker-N.json) from the worker template, each pointing to the correct task file and area.

Generated per-worker configs are the primary approach because each worker needs different context files (its own task file), different area restrictions, and different branch assignments. A single shared worker.json can't express this — the task file path and area boundaries change per worker.

Alternative: a single worker.json with an `agentSpawn` hook that dynamically loads the right task file based on the worker's identity (detected from tmux pane or environment variable). This is simpler but less reliable — the hook has to figure out which worker it is at runtime, and area boundary enforcement via preToolUse hooks would need the same runtime detection.

#### `.kiro/agents/mayor.json`

```json
{
  "name": "mayor",
  "description": "Crysknife Mayor -- plans, dispatches, adapts",
  "prompt": "file://.crysknife/templates/mayor.md",
  "tools": [
    "fs_read", "fs_write", "execute_bash",
    "grep", "glob"
  ],
  "allowedTools": ["fs_read", "grep", "glob"],
  "resources": [
    "file://.kiro/specs/plan.md",
    "file://.kiro/specs/design.md",
    "file://.kiro/specs/principles.md",
    "file://.crysknife/state.json",
    "file://.kiro/specs/tasks/mayor.md"
  ],
  "hooks": {
    "agentSpawn": [{
      "command": "crys status --json",
      "description": "Load current agent status into context on startup"
    }],
    "preToolUse": [{
      "matcher": "execute_bash",
      "command": ".crysknife/hooks/enforce-commands.sh mayor",
      "description": "Block dangerous shell commands"
    }],
    "stop": [{
      "command": "crys heartbeat mayor",
      "description": "Update last_activity timestamp after each response"
    }]
  },
  "keyboardShortcut": "ctrl+shift+m",
  "welcomeMessage": "Mayor online. What are we building?"
}
```

#### `.kiro/agents/worker-1.json` (generated by `crys start`)

```json
{
  "name": "worker-1",
  "description": "Crysknife Worker 1 -- implements tasks in assigned area",
  "prompt": "file://.crysknife/templates/worker-prompt.md",
  "tools": [
    "fs_read", "fs_write", "execute_bash",
    "grep", "glob", "code"
  ],
  "allowedTools": ["fs_read", "grep", "glob", "code"],
  "toolsSettings": {
    "fs_write": {
      "allowedPaths": ["src/auth/**"],
      "deniedPaths": [".crysknife/**", ".kiro/specs/design.md"]
    }
  },
  "resources": [
    "file://.kiro/specs/design.md",
    "file://.kiro/specs/principles.md",
    "file://.kiro/specs/tasks/worker-1.md"
  ],
  "hooks": {
    "agentSpawn": [{
      "command": "crys my-task worker-1",
      "description": "Load current task assignment into context"
    }],
    "preToolUse": [{
      "matcher": "fs_write",
      "command": ".crysknife/hooks/enforce-area.sh worker-1",
      "description": "Block writes outside assigned area"
    }, {
      "matcher": "execute_bash",
      "command": ".crysknife/hooks/enforce-commands.sh worker-1",
      "description": "Block dangerous shell commands"
    }],
    "stop": [{
      "command": "crys heartbeat worker-1",
      "description": "Update last_activity timestamp"
    }]
  }
}
```

#### `.kiro/agents/merger.json`

```json
{
  "name": "merger",
  "description": "Crysknife Merger -- merges completed branches to main",
  "prompt": "file://.crysknife/templates/merger-prompt.md",
  "tools": [
    "fs_read", "fs_write", "execute_bash",
    "grep", "glob"
  ],
  "allowedTools": ["fs_read", "grep", "glob"],
  "resources": [
    "file://.kiro/specs/design.md",
    "file://.kiro/specs/principles.md",
    "file://.kiro/specs/tasks/merger.md"
  ],
  "hooks": {
    "agentSpawn": [{
      "command": "crys merge-queue --json",
      "description": "Load pending merge queue into context"
    }],
    "preToolUse": [{
      "matcher": "execute_bash",
      "command": ".crysknife/hooks/enforce-commands.sh merger",
      "description": "Block dangerous shell commands (allows git merge/rebase)"
    }],
    "stop": [{
      "command": "crys heartbeat merger",
      "description": "Update last_activity timestamp"
    }]
  }
}
```

### Hooks — GUPP for Crysknife

kiro-cli hooks replace most of what Gas Town's GUPP does. Instead of nudging agents to read their task files, hooks inject the right context automatically.

```
HOOK TRIGGER         WHEN IT FIRES              CRYSKNIFE USE
─────────────────    ────────────────────────   ──────────────────────────────
agentSpawn           Session starts             Load state.json + task file
                                                into context. Agent knows its
                                                assignment immediately. This
                                                IS our GUPP.

userPromptSubmit     Every user message         (not used currently — could
                                                inject fresh state if needed)

preToolUse           Before any tool runs       TWO GUARDS:
                                                1. fs_write → enforce-area.sh
                                                   Block writes outside area.
                                                2. execute_bash → enforce-commands.sh
                                                   Block dangerous commands
                                                   (rm -rf, git push --force, etc.)
                                                Exit code 2 = blocked,
                                                STDERR returned to agent.

postToolUse          After any tool runs        (not used currently — could
                                                log file modifications)

stop                 Agent finishes responding  Update last_activity in
                                                state.json via crys heartbeat.
                                                Replaces pane diffing for
                                                activity detection.
```

#### Area Boundary Enforcement Script

`.crysknife/hooks/enforce-area.sh` — blocks writes outside the worker's assigned area:

```bash
#!/bin/bash
# Usage: enforce-area.sh <worker-id>
# Receives hook event JSON on stdin, checks if fs_write target is in allowed area

WORKER_ID="$1"
EVENT=$(cat)

# Extract the file path from the hook event
FILE_PATH=$(echo "$EVENT" | jq -r '.tool_input.path // .tool_input.operations[0].path // empty')

if [ -z "$FILE_PATH" ]; then
  exit 0  # no path to check, allow
fi

# Read this worker's allowed area from state.json
AREA=$(jq -r ".agents[] | select(.id==\"$WORKER_ID\") | .area" .crysknife/state.json)

if [ -z "$AREA" ] || [ "$AREA" = "null" ]; then
  exit 0  # no area restriction, allow
fi

# Check if file is within allowed area
case "$FILE_PATH" in
  $AREA*|.kiro/specs/tasks/$WORKER_ID.md)
    exit 0  # allowed
    ;;
  *)
    echo "BLOCKED: $FILE_PATH is outside your assigned area ($AREA). Write your request in the Feedback section of your task file and the mayor will assign it to the correct worker." >&2
    exit 2  # block tool execution, return message to agent
    ;;
esac
```

### Command Guard — Safe Bash Execution

Workers and the merger have access to `execute_bash` for running tests, builds, and `crys` commands. But they could also run destructive commands (`rm -rf`, `git push --force`, `git checkout main`). The command guard hook blocks dangerous commands before they execute.

`.crysknife/hooks/enforce-commands.sh` — whitelist-based command filtering:

```bash
#!/bin/bash
# Usage: enforce-commands.sh <agent-id>
# Receives hook event JSON on stdin, blocks dangerous bash commands

AGENT_ID="$1"
EVENT=$(cat)

COMMAND=$(echo "$EVENT" | jq -r '.tool_input.command // empty')

if [ -z "$COMMAND" ]; then
  exit 0
fi

# Block list — destructive or out-of-scope commands
BLOCKED_PATTERNS=(
  "rm -rf"
  "git push.*--force"
  "git push.*-f"
  "git checkout main"
  "git checkout master"
  "git merge"
  "git rebase"
  "git reset --hard"
  "git branch -[dD]"
  "chmod"
  "chown"
  "sudo"
  "curl.*|.*sh"
  "wget.*|.*sh"
  "npm publish"
  "go install"
)

for pattern in "${BLOCKED_PATTERNS[@]}"; do
  if echo "$COMMAND" | grep -qE "$pattern"; then
    echo "BLOCKED: '$COMMAND' matches blocked pattern '$pattern'. If you need this, write it in your Feedback section for the mayor." >&2
    exit 2
  fi
done

exit 0
```

The mayor gets a shorter block list (no git restrictions — it may need to inspect branches). The merger gets git merge/rebase allowed but push --force blocked. Each role gets its own variant, or the script reads the role from state.json and adjusts.

### MCP Server — Deferred

The design originally included a `crys mcp-serve` MCP server so agents would call `@crysknife/done` as native tool calls with role-based filtering. This is deferred.

Agents instead use `execute_bash` to run `crys` CLI commands directly:
- Workers run: `crys done worker-1`, `crys status`
- Mayor runs: `crys sling worker-1 ...`, `crys status`, `crys queue ...`
- Merger runs: `crys status`, `crys merge-done ...`

The `execute_bash` approach is simpler (no protocol handler, no per-agent processes), more reliable (execute_bash is battle-tested), and sufficient at 6 agents. Role enforcement comes from prompts + the command guard hook. The preToolUse hook on execute_bash blocks destructive commands regardless of role.

If agents routinely call commands outside their role and it causes problems, add the MCP server then. In practice, well-prompted agents stay in their lane.

### Context Resources — Auto-Loaded Specs

Agent configs use the `resources` field to auto-load files into context on startup:

```
ROLE       AUTO-LOADED CONTEXT
─────────  ──────────────────────────────────────────────
Mayor      plan.md, design.md, principles.md, state.json,
           tasks/mayor.md

Worker-N   design.md, principles.md, tasks/worker-N.md

Merger     design.md, principles.md, tasks/merger.md
```

Workers don't load plan.md or state.json — they don't need the big picture. They get their task file and the shared guardrails. The mayor gets everything.

Combined with the `agentSpawn` hook (which runs `crys status --json` or `crys my-task`), agents have full context the moment they start. No nudge needed.

### How It All Fits Together

```
SESSION STARTUP (what happens when crys start launches a worker):

  1. crys start generates .kiro/agents/worker-1.json
     (from template, with correct task file, area, branch)

  2. crys start runs: kiro-cli chat --agent worker-1
     (in the worker's tmux pane, inside its Git worktree)

  3. kiro-cli loads worker-1.json:
     - System prompt from .crysknife/templates/worker-prompt.md
     - Context: design.md, principles.md, tasks/worker-1.md
     - Hooks registered

  4. agentSpawn hook fires:
     - Runs: crys my-task worker-1
     - Output (task summary) injected into context
     - Agent now knows exactly what to do

  5. Agent starts working. No nudge needed.

DURING WORK:

  6. Agent calls fs_write to edit a file
     - preToolUse hook fires: enforce-area.sh worker-1
     - If file is in area → allowed
     - If file is outside area → BLOCKED, agent gets error message

  7. Agent finishes a response
     - stop hook fires: crys heartbeat worker-1
     - state.json last_activity updated
     - crys watch sees fresh timestamp (no pane diffing needed)

TASK COMPLETE:

  8. Agent runs: crys done worker-1 (via execute_bash)
     - state.json updated (status → "done")
     - Branch added to merge queue
     - Mayor nudged: "worker-1 finished, check plan.md"
     - Merger nudged: "new branch in merge queue"
```

---

## Architecture

### Git Worktree Layout

Each agent operates in its own Git worktree to avoid branch conflicts. `crys start` creates these automatically and symlinks shared directories.

```
project/                          <- main worktree (merger works here)
├── .crysknife/                   <- shared state (real location)
├── .kiro/specs/                  <- shared specs (real location)
├── src/
└── ...

project-worktrees/                <- created by crys start
├── mayor/                        <- mayor's worktree (main branch)
│   ├── .crysknife/ -> ../../project/.crysknife/   (symlink)
│   ├── .kiro/specs/ -> ../../project/.kiro/specs/  (symlink)
│   └── src/
├── worker-1/                     <- worker-1's worktree (own branch)
│   ├── .crysknife/ -> ../../project/.crysknife/   (symlink)
│   ├── .kiro/specs/ -> ../../project/.kiro/specs/  (symlink)
│   └── src/
├── worker-2/
│   └── (same symlink pattern)
├── worker-3/
│   └── (same symlink pattern)
└── worker-4/
    └── (same symlink pattern)
```

Symlinks ensure all agents see the same state.json, plan.md, task files, and spec docs regardless of which worktree they're in. Workers edit code in their isolated worktree but read/write shared files through the symlinks.

Created via:
```bash
git worktree add ../project-worktrees/worker-1 -b feat/task-name
ln -s ../../project/.crysknife ../project-worktrees/worker-1/.crysknife
ln -s ../../project/.kiro/specs ../project-worktrees/worker-1/.kiro/specs
```

When a worker finishes and its branch is merged, `crys` removes the worktree and creates a fresh one for the next task.

### Merge Flow

The merger never touches main directly. All merges go through a staging branch:

```
worker branch ──→ merge/staging ──→ tests ──→ main
                                      │
                                 tests fail?
                                      │
                                 revert staging,
                                 flag for mayor,
                                 main stays clean
```

```bash
# Merger's process:
git checkout merge/staging
git reset --hard main              # staging = clean copy of main
git merge feat/auth-backend        # merge worker's branch
# run tests
# if pass: git checkout main && git merge merge/staging --ff-only
# if fail: flag issue, don't touch main
```

### State Ownership

```
WHO WRITES WHAT:

  state.json        ← ONLY crys CLI and crys watch
                       agents NEVER write this directly

  tasks/worker-N.md ← the worker writes its own task file
                       (progress, feedback, summaries)

  tasks/merger.md   ← the merger writes merge reports here

  tasks/mayor.md    ← the mayor writes planning notes here

  plan.md           ← the mayor updates this

  design.md         ← the mayor updates this (with your approval)

  principles.md     ← you update this (through the mayor)
```

### Done Detection

When a worker finishes a task, it runs `crys done <worker-id>` via `execute_bash`, which:

1. Updates state.json (worker status -> "done", branch -> ready for merge)
2. Adds the branch to the merge queue
3. Nudges the mayor: "worker-1 finished task X. Check plan.md for next assignment."
4. Nudges the merger: "New branch in merge queue."

Workers are prompted in their task file template to run this command when done:

```markdown
## When Done
Run: crys done {{AGENT_ID}}
```

If a worker forgets, `crys watch` detects idle state (no heartbeat updates) and runs the done flow automatically.

### System Overview

```
┌──────────────────────────────────────────────────────────┐
│  YOU (the Overseer)                                      │
│                                                          │
│  $ crys start          spin up agents in tmux            │
│  $ crys status         show agent status dashboard       │
│  $ crys sling <agent>  assign work to an agent           │
│  $ crys done <worker>  mark worker done, notify mayor    │
│  $ crys watch          monitor loop (foreground)         │
│  $ crys nudge <agent>  poke an idle/stuck agent          │
│  $ crys stop           shut down all agents              │
└──────────────┬───────────────────────────────────────────┘
               │
               │  reads/writes
               v
┌──────────────────────────────────────────────────────────┐
│  .crysknife/state.json  (single source of truth)         │
│                                                          │
│  - agent definitions (id, role, status, task, branch)    │
│  - work queue (pending tasks)                            │
│  - convoy tracking (feature-level grouping)              │
│  - phase tracking (current phase, dependencies)          │
└──────────────────────────────────────────────────────────┘
               │
               │  crys manages
               v
┌──────────────────────────────────────────────────────────┐
│  tmux                                                    │
│                                                          │
│  session: <project>                                      │
│  ┌────────────────────┬─────────────────────┐            │
│  │ window: mayor      │ window: dashboard   │            │
│  │ (kiro-cli)         │ (crys watch output) │            │
│  ├────────────────────┼─────────────────────┤            │
│  │ window: workers-1  │ window: workers-2   │            │
│  │ ┌────────┬───────┐ │ ┌────────┬────────┐ │            │
│  │ │worker-1│work-2 │ │ │worker-3│work-4  │ │            │
│  │ │kiro-cli│kiro-cl│ │ │kiro-cli│kiro-cli│ │            │
│  │ └────────┴───────┘ │ └────────┴────────┘ │            │
│  ├────────────────────┼─────────────────────┤            │
│  │ window: merger     │ window: lazygit     │            │
│  │ (kiro-cli)         │ (manual fallback)   │            │
│  ├────────────────────┼─────────────────────┤            │
│  │ window: nvim       │ window: terminal    │            │
│  └────────────────────┴─────────────────────┘            │
└──────────────────────────────────────────────────────────┘
               │
               │  agents read/write
               v
┌──────────────────────────────────────────────────────────┐
│  Project Files                                           │
│                                                          │
│  .kiro/specs/                                            │
│  ├── requirements.md      (shared, read-only for agents) │
│  ├── design.md            (shared, read-only for agents) │
│  ├── principles.md        (shared guardrails)            │
│  ├── plan.md              (phases + parallel tasks)      │
│  └── tasks/                                              │
│      ├── mayor.md         (mayor notes + decisions)      │
│      ├── worker-1.md      (per-worker task file)         │
│      ├── worker-2.md                                     │
│      ├── worker-3.md                                     │
│      ├── worker-4.md                                     │
│      └── merger.md        (standing orders + reports)    │
│                                                          │
│  .crysknife/                                             │
│  ├── state.json           (orchestrator state)           │
│  ├── hooks/               (preToolUse scripts)           │
│  │   ├── enforce-area.sh                                 │
│  │   └── enforce-commands.sh                             │
│  └── templates/           (workflow tier templates)      │
│      ├── full.md                                         │
│      ├── standard.md                                     │
│      └── quick.md                                        │
│                                                          │
│  .kiro/agents/            (kiro-cli agent configs)       │
│  ├── mayor.json                                          │
│  ├── worker-1.json        (generated by crys start)      │
│  ├── worker-2.json                                       │
│  ├── merger.json                                         │
│  └── ...                                                 │
└──────────────────────────────────────────────────────────┘
```

### Data Flow

```
1. ASSIGN WORK

   you ──> crys sling worker-1 --task "Add auth" --tier full --area "src/auth/"
              │
              ├── reads .crysknife/templates/full.md
              ├── generates .kiro/specs/tasks/worker-1.md (from template)
              ├── regenerates .kiro/agents/worker-1.json (area restrictions)
              ├── updates .crysknife/state.json (worker-1 status = "working")
              └── nudges worker via tmux send-keys
                  (agentSpawn hook loads task on next session/restart)

2. MONITOR

   crys watch (polling loop, every 30s)
              │
              ├── for each agent in state.json:
              │     read last_activity timestamp (from stop hook heartbeat)
              │     calculate time since last heartbeat
              │     determine: WORKING / IDLE / DEAD
              │     (fallback: tmux pane content diffing if no heartbeat)
              │
              ├── if IDLE > 2 min:
              │     tmux send-keys nudge message
              │     update state.json (status = "nudged")
              │
              ├── if DEAD > 5 min:
              │     kill pane, restart: kiro-cli chat --agent <id>
              │     agent config auto-loads context + hooks
              │     update state.json (status = "restarted")
              │
              └── display dashboard to stdout

3. HARVEST RESULTS

   you ──> crys status
              │
              └── reads state.json, shows:
                  worker-1: WORKING  "Add auth"         feat/a1-auth
                  worker-2: IDLE     "Fix cache bug"    feat/a2-cache
                  worker-3: DONE     "Write tests"      feat/a3-tests

   you ──> cycle to worker-3 window, read output, then:
   you ──> crys sling worker-3 --task "API docs" --tier quick
```

---

## CLI Commands

### `crys init`

Initialize Crysknife in the current project.

```
$ crys init

Creates:
  .crysknife/
  ├── state.json
  ├── hooks/
  │   ├── enforce-area.sh      (preToolUse area boundary script)
  │   └── enforce-commands.sh  (preToolUse command guard script)
  └── templates/
      ├── mayor.md
      ├── worker-prompt.md
      ├── merger-prompt.md
      ├── full.md
      ├── standard.md
      ├── quick.md
      └── principles.md

  .kiro/agents/
  ├── mayor.json               (kiro-cli agent config)
  └── merger.json              (kiro-cli agent config)

  .kiro/specs/
  ├── plan.md                  (empty plan template)
  ├── principles.md            (from templates/principles.md)
  └── tasks/                   (empty directory)
```

Worker agent configs (worker-1.json, etc.) are generated dynamically by `crys start`.

### `crys start`

Spin up the tmux layout with all configured agents.

```
$ crys start                    # start all agents
$ crys start --workers 3        # start with 3 workers (+ mayor + merger)
$ crys start --agent worker-1   # start specific agent
```

Steps:
1. Generate per-worker agent configs (`.kiro/agents/worker-N.json`) from template
2. Create Git worktrees for each worker, symlink shared directories
3. Create tmux session with windows: mayor, dashboard, worker pairs, merger, lazygit, terminal
4. Launch each agent with `kiro-cli chat --agent <name>` in its tmux pane
5. Agent configs handle the rest: prompt, context, hooks, MCP server all load automatically

### `crys stop`

Shut down all agents gracefully.

```
$ crys stop                     # stop all agents
$ crys stop agent-2             # stop specific agent
```

Updates state.json, kills tmux panes.

### `crys status`

Show current state of all agents.

```
$ crys status

  AGENT     STATUS    TASK              BRANCH          TIER
  agent-1   WORKING   Add auth          feat/a1-auth    full
  agent-2   IDLE      —                 —               —
  agent-3   DONE      Write tests       feat/a3-tests   standard

  QUEUE (2 tasks):
  - Add WebSocket support                               full
  - Fix typo in README                                  quick

  CONVOYS:
  - User Auth [in-progress] agent-1, agent-3
```

### `crys sling <agent>`

Assign work to an agent. Can be called by you from the terminal or by the mayor via `execute_bash`.

```
$ crys sling worker-1 --task "Add auth endpoint" --tier full --branch feat/a1-auth --area "src/auth/"
$ crys sling worker-2 --task "Fix navbar bug" --tier quick
$ crys sling worker-3 --from-queue    # pick next task from queue
```

Steps:
1. Read the template for the specified tier
2. Fill in variables (task name, branch, area)
3. Write to `.kiro/specs/tasks/<agent>.md`
4. Regenerate the worker's agent config (`.kiro/agents/<agent>.json`) with updated area restrictions, task file path, and toolsSettings
5. Update state.json
6. Nudge the worker via tmux send-keys to pick up the new task (agent's agentSpawn hook will load the updated context on next session or restart)

### `crys queue`

Manage the work queue.

```
$ crys queue add "Add WebSocket support" --tier full
$ crys queue add "Fix typo" --tier quick
$ crys queue list
$ crys queue remove <task-id>
```

### `crys watch`

Run the monitoring loop in the foreground. Designed to run in the dashboard tmux pane.

```
$ crys watch

  === CRYS MONITOR 2026-02-28 19:45:00 ===

  agent-1   WORKING   (active 2m ago)    Add auth
  agent-2   IDLE      (idle 3m)          Fix cache → NUDGING
  agent-3   WORKING   (active 30s ago)   Write tests

  [auto-nudged agent-2 at 19:44:30]
  [next check in 30s]
```

Options:
- `--interval 30` — check every N seconds (default: 30)
- `--nudge-after 120` — nudge idle agents after N seconds (default: 120)
- `--restart-after 300` — restart dead agents after N seconds (default: 300)
- `--no-nudge` — detection only, no auto-nudge
- `--no-restart` — no auto-restart

### `crys nudge <agent>`

Manually nudge a specific agent.

```
$ crys nudge agent-1
$ crys nudge --all               # nudge all idle agents
```

Sends a tmux `send-keys` message to the agent's pane telling it to check its task file.

### `crys done <worker>`

Mark a worker as done. Called by the worker via `execute_bash`, or manually by you from the terminal.

```
$ crys done worker-1
```

Steps:
1. Updates state.json (worker status -> "done")
2. Adds the worker's branch to the merge queue
3. Nudges the mayor: "worker-1 finished. Check plan.md for next assignment."
4. Nudges the merger: "New branch ready in merge queue."

Workers run this as part of their normal workflow. If a worker forgets, `crys watch` detects idle state and triggers the done flow automatically.

### `crys convoy`

Track feature-level work.

```
$ crys convoy create "User Auth" --tasks task-1,task-2,task-3
$ crys convoy list
$ crys convoy status "User Auth"
```

### `crys heartbeat <agent>`

Update an agent's last_activity timestamp in state.json. Called by the `stop` hook after each agent response — not intended for manual use.

```
$ crys heartbeat worker-1
```

### `crys my-task <agent>`

Output the agent's current task summary to stdout. Called by the `agentSpawn` hook to inject task context on session startup.

```
$ crys my-task worker-1

# Output:
# Worker-1 — Auth Backend
# Status: working
# Branch: feat/auth-backend
# Area: src/auth/
# Task file: .kiro/specs/tasks/worker-1.md
```

---

## State File

### `.crysknife/state.json`

```json
{
  "project": "chip8-emulator",
  "created": "2026-02-28T19:00:00Z",
  "agents": [
    {
      "id": "mayor",
      "role": "mayor",
      "status": "working",
      "task": "Planning auth feature",
      "task_file": "tasks/mayor.md",
      "tmux_pane": "chip8:mayor.0",
      "last_activity": "2026-02-28T19:43:00Z"
    },
    {
      "id": "worker-1",
      "role": "worker",
      "status": "working",
      "task": "Auth backend",
      "branch": "feat/auth-backend",
      "area": "src/auth/",
      "tier": "full",
      "task_file": "tasks/worker-1.md",
      "tmux_pane": "chip8:workers-1.0",
      "last_activity": "2026-02-28T19:43:00Z",
      "nudge_count": 0
    },
    {
      "id": "worker-2",
      "role": "worker",
      "status": "working",
      "task": "Auth frontend",
      "branch": "feat/auth-frontend",
      "area": "src/components/auth/",
      "tier": "full",
      "task_file": "tasks/worker-2.md",
      "tmux_pane": "chip8:workers-1.1",
      "last_activity": "2026-02-28T19:40:00Z",
      "nudge_count": 0
    },
    {
      "id": "worker-3",
      "role": "worker",
      "status": "idle",
      "task": null,
      "branch": null,
      "area": null,
      "tier": null,
      "task_file": "tasks/worker-3.md",
      "tmux_pane": "chip8:workers-2.0",
      "last_activity": "2026-02-28T19:30:00Z",
      "nudge_count": 0
    },
    {
      "id": "merger",
      "role": "merger",
      "status": "waiting",
      "task": "Standing orders — merge completed branches",
      "task_file": "tasks/merger.md",
      "tmux_pane": "chip8:merger.0",
      "last_activity": "2026-02-28T19:35:00Z"
    }
  ],
  "queue": [
    {
      "id": "task-001",
      "title": "Add WebSocket support",
      "tier": "full",
      "area": "src/ws/",
      "created": "2026-02-28T18:00:00Z"
    }
  ],
  "convoys": [
    {
      "id": "convoy-001",
      "name": "User Auth",
      "status": "in-progress",
      "phase": 2,
      "tasks": ["task-auth-backend", "task-auth-frontend", "task-auth-models"],
      "agents": ["worker-1", "worker-2"],
      "created": "2026-02-28T17:00:00Z"
    }
  ],
  "merge_queue": [
    {
      "branch": "feat/auth-models",
      "worker": "worker-3",
      "status": "ready",
      "completed_at": "2026-02-28T19:30:00Z"
    }
  ]
}
```

---

## Workflow Templates

### `.crysknife/templates/mayor.md`

Sent to the mayor agent on session startup by `crys start`.

```markdown
# Mayor — Crysknife Orchestrator

You are the Mayor. You plan work, assign tasks to workers, and adapt the plan based on feedback. You do NOT write code yourself.

## Your Files
- `.kiro/specs/plan.md` — the living plan. You own this file. Update it constantly.
- `.kiro/specs/design.md` — architecture and patterns. Update with overseer approval.
- `.kiro/specs/principles.md` — guardrails and anti-patterns. Suggest changes to overseer.
- `.kiro/specs/tasks/` — per-agent task files. You generate these when assigning work.
- `.crysknife/state.json` — read-only for you. Use `crys status` to check agent state.

## Your Tools
- `crys status` — see which workers are idle, working, or done
- `crys sling <worker> --task "..." --tier ...` — assign work to a worker
- `crys nudge <agent>` — poke a stuck agent
- `crys queue add "..."` — add/remove tasks from the work queue

## Your Decision Loop
1. Any workers idle? → Check plan.md "Ready" tasks, pick one, run `crys sling`
2. Any workers stuck? → Read their task file Feedback section, help or reassign
3. Merger flagged issues? → Read tasks/merger.md reports, update design.md if needed
4. New work discovered? → Add to plan.md "Discovered During Execution"
5. Need product/architecture input? → Ask the overseer (me)
6. All tasks done? → Report to overseer, propose next steps

## Rules
- NEVER write code. You plan and dispatch.
- ALWAYS ask the overseer for product decisions, priorities, and scope changes.
- Keep plan.md updated — it's the source of truth for what needs doing.
- When assigning work, make sure areas don't overlap between workers.
- Prefer smaller, well-defined tasks over large ambiguous ones.
```

### `.crysknife/templates/full.md`

```markdown
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
```

### `.crysknife/templates/standard.md`

```markdown
# {{AGENT_ID}} — {{TASK_NAME}}

## Context
Read: .kiro/specs/design.md, .kiro/specs/principles.md
Work in: {{AREA}}
Branch: {{BRANCH}}
Workflow tier: standard

## Tasks
- [ ] Implement {{TASK_NAME}}
- [ ] Write unit tests
- [ ] Run test suite, fix regressions
- [ ] Write summary of what you did at the top of this file
- [ ] Run: crys done {{AGENT_ID}}

## Rules
- Follow patterns in design.md
- Don't touch files outside your area ({{AREA}})
- Commit to branch {{BRANCH}}
- When fully done, run `crys done {{AGENT_ID}}`

## Feedback
(write any issues, questions, or discoveries here for the mayor)
```

### `.crysknife/templates/quick.md`

```markdown
# {{AGENT_ID}} — {{TASK_NAME}}

## Context
Read: .kiro/specs/principles.md
Work in: {{AREA}}
Branch: {{BRANCH}}
Workflow tier: quick

## Tasks
- [ ] Implement {{TASK_NAME}}
- [ ] Verify it works
- [ ] Run: crys done {{AGENT_ID}}

## Rules
- Don't touch files outside your area ({{AREA}})
- Commit to branch {{BRANCH}}
- When done, run `crys done {{AGENT_ID}}`
```

### `.crysknife/templates/principles.md`

Default principles file generated by `crys init`. Project-specific rules should be added by the user.

```markdown
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
```

---

## Agent Detection

How `crys watch` determines agent state:

### Primary: Heartbeat via stop Hook

Every agent has a `stop` hook that runs `crys heartbeat <agent-id>` after each response. This updates `last_activity` in state.json automatically. `crys watch` reads these timestamps to determine state:

```
Detection Method: heartbeat timestamps (primary)

Every check interval (default 30s):
  1. Read last_activity for each agent from state.json
  2. Calculate time since last heartbeat

  < nudge threshold (2 min)    → WORKING or WAITING
  > nudge threshold (2 min)    → IDLE (needs nudge)
  > restart threshold (5 min)  → DEAD (needs restart)

Additional signals:
  - Pane process exited (tmux pane_dead flag)  → DEAD
  - Agent ran crys done                        → DONE
```

### Fallback: tmux Pane Content Diffing

If heartbeat hooks fail (agent config issue, hook timeout), `crys watch` falls back to pane content diffing:

```
Fallback Method: tmux pane content diffing

  1. Capture pane content: tmux capture-pane -t <pane> -p
  2. Hash the content (md5/sha256)
  3. Compare to previous hash

  Hash changed    → WORKING
  Hash unchanged  → apply same thresholds as heartbeat
```

### Nudge Message

When nudging, Crysknife sends via `tmux send-keys`:

```
Check your task file and continue working on your current task.
```

Because agents have their task file auto-loaded via the `resources` field in their agent config, the nudge just needs to prompt action — the agent already has context.

### Restart Sequence

When restarting a dead agent:

```
1. Kill the tmux pane
2. Create new pane in same position
3. Start: kiro-cli chat --agent <agent-id>
4. Agent config auto-loads: prompt, context files, hooks, MCP server
5. agentSpawn hook fires: loads task assignment into context
6. Agent resumes working automatically (no manual nudge needed)
7. Update state.json (status = "working", nudge_count = 0)
```

The agent config + task file on disk is the continuity mechanism. Restarts are seamless because kiro-cli reconstructs the full agent context from the config file.

---

## Project Structure

```
crysknife/
├── main.go                 # entry point, CLI routing (binary: crys)
├── cmd/                    # CLI command implementations
│   ├── init.go             # crys init
│   ├── start.go            # crys start (generates agent configs, worktrees)
│   ├── stop.go             # crys stop
│   ├── status.go           # crys status
│   ├── sling.go            # crys sling
│   ├── queue.go            # crys queue
│   ├── watch.go            # crys watch (monitor loop)
│   ├── nudge.go            # crys nudge
│   ├── done.go             # crys done
│   ├── convoy.go           # crys convoy
│   ├── heartbeat.go        # crys heartbeat (called by stop hooks)
│   ├── mytask.go           # crys my-task (called by agentSpawn hooks)
│   └── mergequeue.go       # crys merge-queue
├── internal/
│   ├── state/
│   │   └── state.go        # state.json read/write
│   ├── tmux/
│   │   └── tmux.go         # tmux session/pane management
│   ├── monitor/
│   │   └── monitor.go      # agent detection (heartbeat + pane diffing)
│   ├── template/
│   │   └── template.go     # workflow template rendering
│   └── agentcfg/
│       └── agentcfg.go     # kiro-cli agent config generation
├── templates/               # default workflow templates
│   ├── mayor.md
│   ├── worker-prompt.md     # worker system prompt (shared by all workers)
│   ├── merger-prompt.md
│   ├── full.md
│   ├── standard.md
│   ├── quick.md
│   └── principles.md
├── hooks/                   # hook scripts (copied to .crysknife/hooks/)
│   ├── enforce-area.sh     # preToolUse: block writes outside area
│   └── enforce-commands.sh # preToolUse: block dangerous bash commands
├── go.mod
├── go.sum
└── README.md
```

### Key Dependencies

- `cobra` — CLI framework
- `encoding/json` — state file handling, agent config generation, MCP protocol
- `os/exec` — tmux command execution
- `crypto/md5` — pane content hashing (fallback detection)
- `text/template` — task file and agent config generation
- `bufio` + `os.Stdin/Stdout` — interactive output

No external dependencies beyond cobra. tmux interaction is all through `os/exec` calling `tmux` commands.

---

## Implementation Phases

### Phase 1: Foundation

- `crys init` — create `.crysknife/` directory with state.json, default templates, hook scripts, and base agent configs (mayor.json, merger.json)
- `crys start` — generate per-worker agent configs, create Git worktrees with symlinks, create tmux layout, launch agents with `kiro-cli chat --agent <name>`
- `crys stop` — kill agent panes, clean up generated worker agent configs
- `crys status` — read and display state.json with role grouping
- `crys heartbeat` — update last_activity in state.json (called by stop hooks)
- `crys my-task` — output current task summary for agentSpawn hooks
- State file read/write (with role and merge_queue support)
- tmux session/pane management
- Agent config generation (internal/agentcfg)

### Phase 2: Work Assignment

- `crys sling` — template rendering, task file generation, regenerate worker agent config with new task/area, nudge worker
- `crys queue` — add/remove/list tasks in state.json
- Workflow tier templates (full/standard/quick)
- Merger standing orders template (generated on init)
- Area boundary enforcement hook script (enforce-area.sh)
- Command guard hook script (enforce-commands.sh)

### Phase 3: Monitoring

- `crys watch` — polling loop using heartbeat timestamps (primary) + pane content diffing (fallback)
- `crys nudge` — manual nudge via tmux send-keys
- Auto-nudge for idle workers (based on last_activity from heartbeat)
- Auto-restart for dead agents (relaunch with `kiro-cli chat --agent <name>`, agentSpawn hook handles context)
- Merge queue display in dashboard

### Phase 4: Planning and Tracking

- `crys convoy` — feature-level grouping and status tracking
- Plan file support — readiness-based task grouping (ready/blocked/discovered)
- `crys sling --next` — mayor assigns next ready task to an idle worker
- Dashboard improvements in `crys watch` output (merge queue, worker feedback, blocked tasks)
- (Optional) MCP server if role enforcement via prompts proves insufficient

---

## Gas Town Concepts Mapping

```
Gas Town                Crysknife Equivalent
────────────────────    ────────────────────────────────
gt (CLI binary)         crys (CLI binary)
Beads (Dolt database)   .crysknife/state.json
gt sling                crys sling (CLI command, agents call via execute_bash)
Molecules               .kiro/specs/tasks/<agent>.md
Protomolecules          .crysknife/templates/<tier>.md
Hooks                   kiro-cli agent configs + agentSpawn hooks
GUPP                    agentSpawn hook + stop hook heartbeat
                          (agent auto-loads task on startup,
                           auto-reports activity on each response)
gt nudge                crys nudge (tmux send-keys, fallback)
Mayor                   mayor agent (.kiro/agents/mayor.json)
Crew                    workers (.kiro/agents/worker-N.json)
Polecats                —  (workers are persistent, not ephemeral)
Witness                 crys watch (monitor loop)
Refinery                merger agent (merges + conflict resolution)
Convoys                 crys convoy
Patrols                 crys watch polling loop
                        merger standing orders
Deacon/Dogs/Boot        not implemented (you are these)
```

---

## What Crysknife Does NOT Do

- **No daemon.** You run `crys watch` in a tmux pane. Close it and monitoring stops.
- **No agent-to-agent messaging.** Agents communicate through shared files (task files, plan.md, state.json) only.
- **No multi-project support.** One `.crysknife/` per project. Run separate instances for separate repos.
- **No remote agents.** All agents run locally in tmux on your machine.
- **No persistent agent identity.** Sessions are disposable. The task file is the only continuity.
- **No automatic merge ordering.** The merger agent decides order on-the-fly. No DAG solver.

These are intentional constraints for simplicity. Each one can be added later if needed.

---

## Future: State Storage Upgrade Path

Crysknife currently uses JSON files with symlinks for all agent communication and state. This works at 6 agents with the constraint that only `crys` CLI writes state.json. If we outgrow this, here are the upgrade options:

### Current: Shared JSON Files

```
.crysknife/state.json          <- orchestrator state
.kiro/specs/tasks/worker-1.md  <- agent task files (markdown)
.kiro/specs/plan.md            <- living plan
```

- Agents write their own task files (markdown)
- Only `crys` CLI writes state.json (no concurrent write issues)
- Symlinked across worktrees so all agents see the same files
- Human readable, Git-trackable, zero dependencies

### Upgrade 1: SQLite

Keep agents writing markdown task files, but move orchestrator state to SQLite for structured queries and concurrent safety.

```
.crysknife/crysknife.db        <- SQLite database
.kiro/specs/tasks/worker-1.md  <- still markdown (agents don't change)
```

- `crys status` becomes a SQL query instead of JSON parsing
- Concurrent-safe reads/writes
- History and audit log for free
- Agents still interact through markdown files (no change for them)
- Single file, portable, no server process

### Upgrade 2: Dolt (What Beads Uses)

Dolt is a SQL database with Git-like version control built in — branching, merging, diffing, all at the database level.

```
.crysknife/dolt-db/            <- Dolt database
```

- Each worker can have its own Dolt branch for state changes
- Cell-level merge: two workers creating different beads merge cleanly (no conflicts like with JSON/JSONL files)
- Hash-based IDs prevent merge collisions in multi-agent workflows
- Full version history of all state changes
- Native SQL querying
- Heavier dependency (Dolt binary required)

This is what Beads evolved to. Early Beads used JSONL files in Git (like our state.json). Current Beads uses Dolt for concurrent-safe, branch-aware state management.

```
How Dolt branching maps to our worktrees:

  Git worktree (code)          Dolt branch (data)
  ─────────────────────        ─────────────────────────
  worker-1 worktree            worker-1 Dolt branch
    edits code on               creates/updates state on
    git branch feat/auth        dolt branch worker-1

  merger worktree              main Dolt branch
    merges git branches         merges dolt branches
    to main                     (cell-level, clean merge
                                 if different rows)
```

Git manages code, Dolt manages work tracking data. They branch and merge independently.

### When to Upgrade

```
JSON files (now)     → works at 3-6 agents, crys CLI is sole writer
SQLite (Stage 7)     → when you need querying, history, or audit logs
Dolt (Stage 8+)      → when multiple processes need to write state
                        concurrently, or you want branch-per-agent
                        state isolation
```

No need to decide now. Start with JSON, feel the pain, upgrade when it hurts.
