# Implementation Plan: Crysknife

## Overview

Build the `crys` CLI tool in Go using Cobra. Implementation follows four phases: Foundation (project scaffold, state, tmux, agent lifecycle), Work (assignment, queue, merge flow, hooks), Monitoring (watch loop, nudge, auto-restart), and Planning (convoys, dashboard). Each phase builds on the previous.

**Testing Strategy:**
- Unit tests for every internal package (`internal/state`, `internal/tmux`, `internal/agentcfg`, `internal/template`, `internal/monitor`)
- Property-based tests (PBT) using `rapid` (Go PBT library) for correctness properties from design.md
- PBT checkpoints after every 3-4 task groups
- Integration tests for tmux operations gated behind `testing.Short()` skip
- Hook scripts tested with shell-based tests (bats or plain bash)
- Run all tests: `go test ./...`

## Tasks

### Phase 1: Foundation

- [ ] 1. Project Scaffold and CLI Skeleton
  - [ ] 1.1 Initialize Go module and install Cobra
    - `go mod init github.com/0xlemi/crysknife`
    - `go get github.com/spf13/cobra`
    - Create `main.go` with root command (binary name: `crys`)
    - _Requirements: none (scaffold)_

  - [ ] 1.2 Create empty command files with Cobra registration
    - New files: `cmd/root.go`, `cmd/init.go`, `cmd/start.go`, `cmd/stop.go`, `cmd/status.go`, `cmd/heartbeat.go`, `cmd/mytask.go`, `cmd/sling.go`, `cmd/queue.go`, `cmd/done.go`, `cmd/mergequeue.go`, `cmd/mergedone.go`, `cmd/watch.go`, `cmd/nudge.go`, `cmd/convoy.go`
    - Each registers with root command, has placeholder RunE
    - _Requirements: none (scaffold)_

  - [ ] 1.3 Embed default templates and hook scripts
    - New dir: `templates/` with embedded files: mayor.md, worker-prompt.md, merger-prompt.md, full.md, standard.md, quick.md, principles.md
    - New dir: `hooks/` with embedded files: enforce-area.sh, enforce-commands.sh
    - Use `//go:embed` directive in a `embed.go` file
    - _Requirements: 1.3, 1.4_

- [ ] 2. State Package
  - [ ] 2.1 Implement state types and Load/Save
    - New file: `internal/state/state.go`
    - Structs: `State`, `Agent`, `QueueItem`, `Convoy`, `MergeEntry` (per design.md)
    - `Load(root)`: read `.crysknife/state.json`, return empty default if missing, error on invalid JSON
    - `Save(root)`: atomic write (temp file + rename)
    - _Requirements: 15.1, 15.2, 15.3_

  - [ ] 2.2 Implement state helper methods
    - `FindAgent(id)`, `UpdateHeartbeat(agentID)`, `SetAgentStatus(agentID, status)`
    - `AddToMergeQueue(branch, worker)`, `RemoveFromMergeQueue(branch)`
    - `AddQueueItem(title, tier, area)`, `RemoveQueueItem(id)`, `PopQueueItem()`
    - _Requirements: 15.1_

  - [ ] 2.3 Unit tests for state package
    - Test Load with missing file returns empty default state
    - Test Load with invalid JSON returns error
    - Test Save + Load roundtrip preserves all fields
    - Test atomic write (file is valid JSON even if process interrupted — verify temp file pattern)
    - Test FindAgent returns nil for unknown ID
    - Test UpdateHeartbeat sets timestamp within 1s of now
    - Test AddToMergeQueue / RemoveFromMergeQueue
    - Test AddQueueItem generates unique IDs
    - Test RemoveQueueItem with unknown ID returns error
    - Test PopQueueItem from empty queue returns error
    - Run: `go test ./internal/state/ -v`
    - _Requirements: 15.1, 15.2, 15.3, 12.1_

- [ ] 3. tmux Package
  - [ ] 3.1 Implement tmux wrapper functions
    - New file: `internal/tmux/tmux.go`
    - Functions: `CreateSession`, `CreateWindow`, `SplitPane`, `SendKeys`, `KillPane`, `KillSession`, `SessionExists`, `CapturePaneContent`
    - All functions shell out via `os/exec` to `tmux` binary
    - _Requirements: 2.4, 2.5, 3.1_

  - [ ] 3.2 Unit tests for tmux package
    - Test command construction (verify args passed to tmux)
    - Test SessionExists returns false for nonexistent session
    - Integration tests (skip with `testing.Short()`): create session, create window, split pane, send keys, capture content, kill session
    - Run: `go test ./internal/tmux/ -v`
    - _Requirements: 2.4, 3.1_

- [ ] 4. Agent Config Package
  - [ ] 4.1 Implement agent config generation
    - New file: `internal/agentcfg/agentcfg.go`
    - `AgentConfig` struct (per design.md)
    - `GenerateWorkerConfig(workerID, taskFile, area)`: builds worker JSON with correct hooks (agentSpawn: crys my-task, preToolUse: enforce-area.sh + enforce-commands.sh, stop: crys heartbeat), toolsSettings with allowedPaths/deniedPaths, resources
    - `GenerateMayorConfig()`: builds mayor JSON with hooks (agentSpawn: crys status --json, preToolUse: enforce-commands.sh, stop: crys heartbeat)
    - `GenerateMergerConfig()`: builds merger JSON with hooks (agentSpawn: crys merge-queue --json, preToolUse: enforce-commands.sh with merger role, stop: crys heartbeat)
    - `WriteConfig(root, config)`: marshal to JSON, write to `.kiro/agents/<name>.json`
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5_

  - [ ] 4.2 Unit tests for agent config package
    - Test GenerateWorkerConfig output has correct name, hooks, toolsSettings, resources
    - Test GenerateWorkerConfig deniedPaths includes `.crysknife/**` and `.kiro/specs/design.md`
    - Test GenerateMayorConfig output has correct hooks and resources
    - Test GenerateMergerConfig output has correct hooks
    - Test WriteConfig creates valid JSON file at correct path
    - Run: `go test ./internal/agentcfg/ -v`
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5_

- [ ] 5. Property Test Checkpoint 1
  - [ ] 5.1 Write PBT for Properties 1, 2, 3, 4
    - **Property 1: Atomic state writes** — generate random State structs, Save concurrently from goroutines, verify file always contains valid JSON after each write
    - **Property 2: Init idempotency guard** — generate random project dirs, run init twice, verify second fails and files unchanged
    - **Property 3: Agent count consistency** — generate random N (1-10), verify state has N+2 agents after start logic
    - **Property 4: Heartbeat timestamp freshness** — generate random agent IDs, call UpdateHeartbeat, verify last_activity within 1s of now
    - Library: `pgregory.net/rapid`
    - Run: `go test ./internal/state/ -v -run TestProperty`
    - _Validates: 15.1, 1.7, 2.6, 2.7, 12.1_

  - [ ] 5.2 Run all tests and verify no regressions
    - Run: `go test ./... -v`
    - Fix any failures from tasks 1-4

- [ ] 6. `crys init` Command
  - [ ] 6.1 Implement init command
    - File: `cmd/init.go`
    - Check if `.crysknife/` exists → error if so
    - Create dirs: `.crysknife/`, `.crysknife/hooks/`, `.crysknife/templates/`
    - Write `state.json` with empty default state (project name from directory)
    - Copy embedded templates to `.crysknife/templates/`
    - Copy embedded hooks to `.crysknife/hooks/`, chmod +x
    - Create `.kiro/agents/` with mayor.json and merger.json (via agentcfg package)
    - Create `.kiro/specs/tasks/`, `.kiro/specs/plan.md`, `.kiro/specs/principles.md`
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_

  - [ ] 6.2 Unit tests for init command
    - Test init creates all expected directories and files
    - Test init writes valid state.json with correct structure
    - Test init copies all 7 templates
    - Test init copies 2 hook scripts and they are executable
    - Test init creates mayor.json and merger.json
    - Test init fails with error when `.crysknife/` already exists
    - Run: `go test ./cmd/ -v -run TestInit`
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7_

- [ ] 7. `crys start` Command
  - [ ] 7.1 Implement start command
    - File: `cmd/start.go`
    - Flags: `--workers N` (default 4), `--agent <id>` (single agent)
    - Check `.crysknife/` exists → error if not
    - Check tmux session doesn't exist → error if so
    - Check `tmux` and `kiro-cli` in PATH → error if not
    - Generate per-worker agent configs via agentcfg package
    - Create Git worktrees: `git worktree add ../<project>-worktrees/<id>` for each worker and mayor
    - Create symlinks in each worktree: `.crysknife/` → main, `.kiro/specs/` → main
    - Create tmux session + windows: mayor, dashboard, worker pairs, merger, lazygit, nvim, terminal
    - Launch agents: `kiro-cli chat --agent <name>` in each pane
    - Update state.json with all agent entries
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.9, 2.10, 2.11, 2.12_

  - [ ] 7.2 Unit tests for start command
    - Test worker config generation produces N configs
    - Test state.json has N+2 agents after start
    - Test default --workers is 4 when flag omitted
    - Test initial statuses: workers "idle", mayor "working", merger "waiting"
    - Test error when `.crysknife/` missing
    - Test error when tmux session already exists
    - Test error when tmux not in PATH
    - Test error when kiro-cli not in PATH
    - Test `--agent` flag starts only specified agent
    - Run: `go test ./cmd/ -v -run TestStart`
    - _Requirements: 2.1, 2.6, 2.7, 2.8, 2.9, 2.10, 2.11, 2.12_

- [ ] 8. `crys stop`, `crys status`, `crys heartbeat`, `crys my-task` Commands
  - [ ] 8.1 Implement stop command
    - File: `cmd/stop.go`
    - Args: optional `<agent-id>` for single agent stop
    - Kill tmux panes (all or single), update state statuses to "stopped"
    - Do NOT remove worktrees
    - _Requirements: 3.1, 3.2, 3.3, 3.4_

  - [ ] 8.2 Implement status command
    - File: `cmd/status.go`
    - Flag: `--json` for raw JSON output
    - Read state, print table (AGENT, ROLE, STATUS, TASK, BRANCH, TIER), queue, convoys
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [ ] 8.3 Implement heartbeat command
    - File: `cmd/heartbeat.go`
    - Args: `<agent-id>` (required)
    - Call `state.UpdateHeartbeat(agentID)`, error if agent not found
    - _Requirements: 12.1, 12.2_

  - [ ] 8.4 Implement my-task command
    - File: `cmd/mytask.go`
    - Args: `<agent-id>` (required)
    - Read agent from state, print summary (id, role, status, task, branch, area, task_file)
    - Print "no task assigned" if task is nil
    - Error if agent not found
    - _Requirements: 13.1, 13.2, 13.3_

  - [ ] 8.5 Unit tests for stop, status, heartbeat, my-task
    - Test stop updates all agent statuses to "stopped"
    - Test stop kills agent panes but tmux session survives
    - Test stop with agent-id only updates that agent
    - Test status outputs table format with correct columns
    - Test status --json outputs valid JSON matching state
    - Test heartbeat updates timestamp
    - Test heartbeat errors on unknown agent
    - Test my-task prints all non-null fields
    - Test my-task prints "no task" when task is nil
    - Test my-task errors on unknown agent
    - Run: `go test ./cmd/ -v -run "TestStop|TestStatus|TestHeartbeat|TestMyTask"`
    - _Requirements: 3.1, 3.2, 3.3, 4.1, 4.4, 12.1, 12.2, 13.1, 13.2, 13.3_

- [ ] 9. Property Test Checkpoint 2
  - [ ] 9.1 Write PBT for Properties 5, 6, 7, 8
    - **Property 5: Worktree isolation** — generate N workers, verify each has unique branch and directory
    - **Property 6: Symlink validity** — generate worktrees, verify symlinks resolve to main project dirs
    - **Property 7: Stop updates state** — generate agents with various statuses, run stop, verify all "stopped"
    - **Property 8: Task context completeness** — generate agents with random tasks, verify my-task outputs all non-null fields
    - Run: `go test ./... -v -run TestProperty`
    - _Validates: 2.2, 2.3, 3.2, 3.3, 13.1_

  - [ ] 9.2 Run all tests and verify no regressions
    - Run: `go test ./... -v`
    - Fix any failures from tasks 6-8

### Phase 2: Work Assignment

- [ ] 10. Template Package
  - [ ] 10.1 Implement template rendering
    - New file: `internal/template/template.go`
    - `TemplateVars` struct: AgentID, TaskName, Area, Branch
    - `Render(tier, vars, templatesDir)`: read `<tier>.md` from templatesDir, replace `{{AGENT_ID}}`, `{{TASK_NAME}}`, `{{AREA}}`, `{{BRANCH}}`
    - Error if template file missing
    - _Requirements: 5.1, 5.2_

  - [ ] 10.2 Unit tests for template package
    - Test Render substitutes all 4 variables correctly
    - Test Render with each tier (full, standard, quick)
    - Test Render errors on missing template file
    - Test Render with empty/nil variables produces output with empty strings
    - Run: `go test ./internal/template/ -v`
    - _Requirements: 5.1, 5.2_

- [ ] 11. `crys sling` Command
  - [ ] 11.1 Implement sling command
    - File: `cmd/sling.go`
    - Args: `<worker-id>` (required)
    - Flags: `--task`, `--tier`, `--area`, `--branch`, `--from-queue`
    - Render template via template package, write to `.kiro/specs/tasks/<worker-id>.md`
    - Regenerate worker agent config with new area (via agentcfg package)
    - Create git branch in worker's worktree if `--branch` provided
    - Update state (status "working", task, branch, area, tier)
    - Nudge worker via tmux send-keys
    - `--from-queue`: pop from queue, use its title/tier/area
    - Error if worker not found, error if invalid tier, error if queue empty (for --from-queue)
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7, 5.8, 5.9_

  - [ ] 11.2 Unit tests for sling command
    - Test sling creates task file with correct content
    - Test sling regenerates agent config with updated area
    - Test sling updates state with all fields
    - Test sling --from-queue pops from queue
    - Test sling errors on unknown worker
    - Test sling errors on invalid tier
    - Test sling --from-queue errors on empty queue
    - Run: `go test ./cmd/ -v -run TestSling`
    - _Requirements: 5.1, 5.2, 5.3, 5.5, 5.7, 5.8, 5.9_

- [ ] 12. `crys queue` Command
  - [ ] 12.1 Implement queue command with subcommands
    - File: `cmd/queue.go`
    - Subcommands: `add`, `list`, `remove`
    - `add <title> --tier <tier> [--area <path>]`: add to state queue with generated ID
    - `list`: print all queued tasks
    - `remove <task-id>`: remove from queue, error if not found
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

  - [ ] 12.2 Unit tests for queue command
    - Test add creates entry with ID, title, tier, timestamp
    - Test add with --area includes area
    - Test list displays all entries
    - Test remove deletes entry
    - Test remove errors on unknown ID
    - Run: `go test ./cmd/ -v -run TestQueue`
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ] 13. `crys done`, `crys merge-queue`, `crys merge-done` Commands
  - [ ] 13.1 Implement done command
    - File: `cmd/done.go`
    - Args: `<worker-id>` (required)
    - Update state: status → "done", add branch to merge queue with status "ready"
    - Nudge mayor and merger via tmux send-keys
    - Error if worker not found, error if no branch assigned
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

  - [ ] 13.2 Implement merge-queue command
    - File: `cmd/mergequeue.go`
    - Flag: `--json`
    - Display merge queue entries (branch, worker, status, completed_at)
    - _Requirements: 10.1, 10.2_

  - [ ] 13.3 Implement merge-done command
    - File: `cmd/mergedone.go`
    - Args: `<branch>` (required)
    - Flag: `--failed <reason>`
    - Remove from merge queue, update worker status to "merged" (or "merge-failed" with reason)
    - Nudge mayor
    - Error if branch not in queue
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_

  - [ ] 13.4 Unit tests for done, merge-queue, merge-done
    - Test done updates status to "done" and adds to merge queue
    - Test done nudges mayor and merger
    - Test done errors on unknown worker
    - Test done errors when worker has no branch
    - Test merge-queue displays entries
    - Test merge-queue --json outputs valid JSON
    - Test merge-done removes branch from queue
    - Test merge-done updates worker status to "merged"
    - Test merge-done --failed sets "merge-failed" and stores reason
    - Test merge-done errors on unknown branch
    - Run: `go test ./cmd/ -v -run "TestDone|TestMergeQueue|TestMergeDone"`
    - _Requirements: 9.1, 9.2, 9.5, 10.1, 10.2, 11.1, 11.2, 11.3, 11.5_

- [ ] 14. Hook Scripts
  - [ ] 14.1 Write enforce-area.sh
    - File: `hooks/enforce-area.sh` (embedded, copied by init)
    - Read worker-id from $1, hook event JSON from stdin
    - Extract file path from tool_input
    - Read worker's area from state.json via jq
    - Allow if path matches area or is worker's own task file
    - Block (exit 2 + stderr message) if outside area
    - _Requirements: 17.1, 17.2_

  - [ ] 14.2 Write enforce-commands.sh
    - File: `hooks/enforce-commands.sh` (embedded, copied by init)
    - Read agent-id from $1, hook event JSON from stdin
    - Extract command from tool_input
    - Check against blocked patterns (rm -rf, git push --force, sudo, etc.)
    - Allow git merge/rebase for merger role (read role from state.json)
    - Block (exit 2 + stderr message) if matches blocked pattern
    - _Requirements: 17.3, 17.4, 17.5_

  - [ ] 14.3 Tests for hook scripts
    - Test enforce-area.sh allows writes inside area
    - Test enforce-area.sh allows writes to own task file
    - Test enforce-area.sh blocks writes outside area (exit code 2)
    - Test enforce-area.sh allows when no area set
    - Test enforce-commands.sh blocks rm -rf, git push --force, sudo
    - Test enforce-commands.sh allows safe commands (go test, ls, cat)
    - Test enforce-commands.sh allows git merge for merger role
    - Test enforce-commands.sh blocks git merge for worker role
    - Run: `bash hooks/test_hooks.sh` or `go test ./hooks/ -v`
    - _Requirements: 17.1, 17.2, 17.3, 17.4, 17.5_

- [ ] 15. Property Test Checkpoint 3
  - [ ] 15.1 Write PBT for Properties 9, 10, 11, 12, 13
    - **Property 9: Sling state consistency** — generate random sling inputs, verify state entry + task file + agent config all match
    - **Property 10: Done triggers merge queue** — generate workers with branches, call done, verify branch in merge queue with "ready"
    - **Property 11: Merge-done cleans queue** — generate merge queue entries, call merge-done, verify branch removed
    - **Property 12: Area enforcement** — generate random file paths and areas, verify enforce-area.sh exit codes
    - **Property 13: Command guard** — generate random commands, verify enforce-commands.sh exit codes for blocked/allowed
    - Run: `go test ./... -v -run TestProperty`
    - _Validates: 5.2, 5.3, 5.5, 16.6, 9.1, 9.2, 11.1, 17.1, 17.2, 17.3, 17.4_

  - [ ] 15.2 Run all tests and verify no regressions
    - Run: `go test ./... -v`
    - Fix any failures from tasks 10-14

### Phase 3: Monitoring

- [ ] 16. Monitor Package
  - [ ] 16.1 Implement monitor detection logic
    - New file: `internal/monitor/monitor.go`
    - `AgentState` enum: Working, Idle, Dead, Done
    - `MonitorConfig` struct: Interval, NudgeAfter, RestartAfter, NoNudge, NoRestart
    - `CheckAgent(agent, prevHash)`: determine state from heartbeat timestamp, fall back to pane hash diff
    - `HashPaneContent(target)`: capture pane via tmux, return hash
    - _Requirements: 7.2, 7.3, 7.4, 7.5_

  - [ ] 16.2 Unit tests for monitor package
    - Test CheckAgent returns Working when heartbeat is fresh
    - Test CheckAgent returns Idle when heartbeat exceeds nudge threshold
    - Test CheckAgent returns Dead when heartbeat exceeds restart threshold
    - Test CheckAgent falls back to hash diff when no heartbeat
    - Test HashPaneContent returns consistent hash for same content
    - Run: `go test ./internal/monitor/ -v`
    - _Requirements: 7.2, 7.3, 7.4, 7.5_

- [ ] 17. `crys watch` and `crys nudge` Commands
  - [ ] 17.1 Implement watch command
    - File: `cmd/watch.go`
    - Flags: `--interval` (30), `--nudge-after` (120), `--restart-after` (300), `--no-nudge`, `--no-restart`
    - Polling loop: read state, check each agent via monitor package
    - Auto-nudge idle agents (unless --no-nudge): send-keys + update status to "nudged"
    - Auto-restart dead agents (unless --no-restart): kill pane, relaunch, update status to "working"
    - Nudge mayor when worker status is "done"
    - Print dashboard to stdout each interval
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7, 7.8, 7.9_

  - [ ] 17.2 Implement nudge command
    - File: `cmd/nudge.go`
    - Args: `<agent-id>` or `--all`
    - Send tmux send-keys message to agent pane
    - `--all`: nudge agents with status "idle" or "nudged"
    - Increment nudge_count in state
    - _Requirements: 8.1, 8.2, 8.3_

  - [ ] 17.3 Unit tests for watch and nudge
    - Test watch loop detects idle agent and nudges
    - Test watch loop detects dead agent and restarts
    - Test watch --no-nudge skips nudging
    - Test watch --no-restart skips restarting
    - Test watch nudges mayor when worker is done
    - Test nudge sends keys to correct pane
    - Test nudge --all nudges only idle/nudged agents
    - Test nudge increments nudge_count
    - Run: `go test ./cmd/ -v -run "TestWatch|TestNudge"`
    - _Requirements: 7.1, 7.3, 7.4, 7.8, 7.9, 8.1, 8.2, 8.3_

- [ ] 18. Property Test Checkpoint 4
  - [ ] 18.1 Write PBT for Property 14
    - **Property 14: Watch detection accuracy** — generate agents with various last_activity timestamps relative to thresholds, verify correct classification (Working/Idle/Dead)
    - Run: `go test ./internal/monitor/ -v -run TestProperty`
    - _Validates: 7.3_

  - [ ] 18.2 Run all tests and verify no regressions
    - Run: `go test ./... -v`
    - Fix any failures from tasks 16-17

### Phase 4: Planning and Tracking

- [ ] 19. `crys convoy` Command
  - [ ] 19.1 Implement convoy command with subcommands
    - File: `cmd/convoy.go`
    - Subcommands: `create`, `list`, `status`
    - `create <name> --tasks <ids>`: add convoy to state with generated ID, status "in-progress"
    - `list`: display all convoys with name, status, agents
    - `status <name>`: display convoy tasks with individual statuses
    - _Requirements: 14.1, 14.2, 14.3_

  - [ ] 19.2 Unit tests for convoy command
    - Test create adds convoy to state with correct fields
    - Test list displays all convoys
    - Test status shows tasks with statuses
    - Run: `go test ./cmd/ -v -run TestConvoy`
    - _Requirements: 14.1, 14.2, 14.3_

- [ ] 20. Final Property Test Checkpoint
  - [ ] 20.1 Run full test suite
    - Run: `go test ./... -v`
    - Verify all 14 properties pass
    - Verify all unit tests pass
    - Fix any regressions

  - [ ] 20.2 Integration smoke test
    - Run `crys init` in a temp directory
    - Verify all files created
    - Run `crys status` — verify empty table
    - Run `crys queue add "test task" --tier quick`
    - Run `crys queue list` — verify task appears
    - Run `crys queue remove <id>` — verify removed
    - Run: manual or scripted bash test

## Notes

- `rapid` (pgregory.net/rapid) is the Go PBT library. Install with `go get pgregory.net/rapid`.
- tmux integration tests require tmux installed and are skipped with `testing.Short()`. CI should run with `-short` flag unless tmux is available.
- Hook script tests need `jq` installed (used by enforce-area.sh and enforce-commands.sh to parse JSON).
- Phase 4 is intentionally light — convoy is the only new command. `crys sling --next` (from design doc Phase 4) is deferred until the mayor workflow proves it's needed.
- The embedded templates in `templates/` and `hooks/` are the source of truth. `.crysknife/templates/` and `.crysknife/hooks/` are copies that users can customize after init.
- Worker agent configs are regenerated on every `crys sling` call, so customizations to `.kiro/agents/worker-N.json` will be overwritten. This is by design — the config is derived from state.
