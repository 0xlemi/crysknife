# Crysknife

A lightweight CLI orchestrator for managing multiple [kiro-cli](https://github.com/aws/kiro-cli) agents in tmux. Named after the Fremen crysknife from Dune — a personal weapon, forged from experience.

```
$ crys status

  AGENT      ROLE     STATUS    TASK              BRANCH            TIER
  mayor      mayor    working   Planning auth     —                 —
  worker-1   worker   working   Auth backend      feat/auth-api     full
  worker-2   worker   working   Auth frontend     feat/auth-ui      full
  worker-3   worker   idle      —                 —                 —
  merger     merger   waiting   Standing orders   —                 —

  QUEUE (1 task):
  task-001  Add WebSocket support  full

  CONVOYS:
  User Auth [in-progress] worker-1, worker-2
```

## What It Does

Crysknife manages 3-6 kiro-cli agents running in parallel on a single machine. It handles:

- **Agent lifecycle** — spin up/down agents in tmux with isolated Git worktrees
- **Work assignment** — assign tasks with workflow tiers (full/standard/quick) via templates
- **Monitoring** — detect idle/dead agents, auto-nudge and auto-restart
- **Merge flow** — dedicated merger agent merges branches through a staging branch
- **Safety** — preToolUse hooks block file writes outside assigned areas and dangerous commands

## Three Roles

```
┌─────────────────────────────────────────────┐
│  MAYOR (1 agent)                            │
│  Plans work, assigns tasks, adapts the plan │
│  You talk to this one the most              │
├─────────────────────────────────────────────┤
│  WORKERS (3-4 agents)                       │
│  Each gets a task + branch + area           │
│  Code, test, commit — don't touch main      │
├─────────────────────────────────────────────┤
│  MERGER (1 agent)                           │
│  Merges completed branches to main          │
│  One at a time, through staging             │
└─────────────────────────────────────────────┘
```

Agents communicate through shared files (task files, plan.md, state.json), not direct messaging. Each worker operates in its own Git worktree on its own branch.

## Quick Start

### Prerequisites

- Go 1.21+
- [tmux](https://github.com/tmux/tmux)
- [kiro-cli](https://github.com/aws/kiro-cli)
- Git

### Install

```bash
go install github.com/0xlemi/crysknife@latest
```

Or build from source:

```bash
git clone https://github.com/0xlemi/crysknife.git
cd crysknife
go build -o crys .
```

### Initialize a Project

```bash
cd your-project
crys init
```

This creates:
- `.crysknife/` — state file, templates, hook scripts
- `.kiro/agents/` — mayor and merger agent configs
- `.kiro/specs/tasks/` — per-agent task files
- `.kiro/specs/plan.md` — living plan (mayor owns this)
- `.kiro/specs/principles.md` — shared guardrails

### Start Agents

```bash
crys start                  # 4 workers (default) + mayor + merger
crys start --workers 2      # 2 workers + mayor + merger
```

This generates per-worker agent configs, creates Git worktrees, sets up the tmux session, and launches all agents.

### Assign Work

```bash
crys sling worker-1 --task "Add auth endpoint" --tier full --area "src/auth/" --branch feat/auth
```

Or let the mayor do it — the mayor runs `crys sling` via `execute_bash` as part of its planning loop.

### Monitor

```bash
crys watch
```

Runs in the dashboard tmux pane. Detects idle agents (auto-nudge after 2 min) and dead agents (auto-restart after 5 min).

### Harvest Results

```bash
crys status                 # see who's done
crys merge-queue            # see what's ready to merge
```

When a worker finishes, it runs `crys done worker-1`, which updates state, adds the branch to the merge queue, and nudges the mayor and merger.

### Stop

```bash
crys stop                   # kill all agent panes
crys stop worker-2          # kill one agent
```

Worktrees persist for inspection. The tmux session stays alive (lazygit, nvim, terminal windows survive).

## Commands

| Command | Description |
|---|---|
| `crys init` | Scaffold project with configs, templates, hooks |
| `crys start` | Generate configs, worktrees, tmux, launch agents |
| `crys stop` | Kill agent panes, update state |
| `crys status` | Display agent table, queue, convoys |
| `crys sling` | Assign work with template + config regen + nudge |
| `crys queue` | Manage work backlog (add/list/remove) |
| `crys watch` | Monitor loop with auto-nudge and auto-restart |
| `crys nudge` | Manually poke an idle agent |
| `crys done` | Mark worker done, add to merge queue |
| `crys merge-queue` | Show branches ready to merge |
| `crys merge-done` | Mark merge complete or failed |
| `crys convoy` | Feature-level tracking across agents |
| `crys heartbeat` | Update activity timestamp (called by hooks) |
| `crys my-task` | Print task summary (called by hooks) |

## Workflow Tiers

Not all work needs the same process:

| Tier | Steps | Use for |
|---|---|---|
| **full** | Review → design → implement → self-review → test → summary | New features, complex changes |
| **standard** | Implement → test → summary | Medium features, refactors |
| **quick** | Implement → verify → done | Bug fixes, small tweaks |

## How It Works

```
crys start
  → generates .kiro/agents/worker-N.json (per worker)
  → creates Git worktrees with symlinked shared dirs
  → creates tmux session with all windows/panes
  → launches kiro-cli chat --agent <name> in each pane
  → agentSpawn hooks inject task context automatically

crys sling worker-1 --task "Add auth" --tier full --area "src/auth/"
  → renders full.md template → .kiro/specs/tasks/worker-1.md
  → regenerates worker-1.json with area restrictions
  → updates state.json
  → nudges worker via tmux send-keys

crys watch (polling loop)
  → reads heartbeat timestamps from state.json
  → idle > 2 min → auto-nudge
  → dead > 5 min → auto-restart (kill pane, relaunch, hooks rebuild context)

crys done worker-1 (called by worker via execute_bash)
  → status → "done", branch → merge queue
  → nudges mayor + merger
```

## Safety

Two preToolUse hooks protect against agent mistakes:

- **enforce-area.sh** — blocks `fs_write` outside the worker's assigned area. Workers can only modify files in their area and their own task file.
- **enforce-commands.sh** — blocks dangerous `execute_bash` commands (rm -rf, git push --force, sudo, etc.). Role-aware: the merger is allowed git merge/rebase, workers are not.

Both hooks exit with code 2 to block the operation and return an error message to the agent.

## Project Structure

```
your-project/
├── .crysknife/
│   ├── state.json              # single source of truth
│   ├── templates/              # customizable workflow templates
│   │   ├── full.md, standard.md, quick.md
│   │   ├── mayor.md, worker-prompt.md, merger-prompt.md
│   │   └── principles.md
│   └── hooks/
│       ├── enforce-area.sh     # area boundary enforcement
│       └── enforce-commands.sh # command guard
├── .kiro/
│   ├── agents/                 # kiro-cli agent configs
│   │   ├── mayor.json
│   │   ├── worker-1.json       # generated by crys start
│   │   ├── worker-2.json
│   │   └── merger.json
│   └── specs/
│       ├── plan.md             # living plan (mayor updates)
│       ├── principles.md       # shared guardrails
│       └── tasks/
│           ├── mayor.md
│           ├── worker-1.md     # generated by crys sling
│           └── merger.md       # standing orders
```

Worker worktrees are created as siblings: `../<project>-worktrees/worker-1/` etc., with `.crysknife/` and `.kiro/specs/` symlinked back to the main project.

## Inspired By

Crysknife is inspired by Steve Yegge's [Gas Town](https://github.com/steveyegge/gas-town) — a multi-agent orchestrator for large-scale agentic development. Crysknife implements a subset of Gas Town's ideas adapted for a single developer running 3-6 agents on one machine. It's a learning tool and stepping stone — simple enough to understand completely, useful enough to run daily.

## License

MIT
