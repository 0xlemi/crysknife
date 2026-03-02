# Requirements Document

## Introduction

Crysknife is a CLI tool (`crys`) that orchestrates multiple kiro-cli agents running in tmux. It handles project initialization, agent lifecycle (start/stop/restart), work assignment with workflow tiers, status monitoring with auto-nudge and auto-restart, merge queue management, and feature-level tracking. The tool is written in Go and designed for a single developer running 3-6 agents on one machine.

Crysknife manages three agent roles: a mayor (plans and dispatches), workers (code on isolated branches), and a merger (merges completed branches to main). Agents communicate through shared files and run `crys` CLI commands via `execute_bash`. kiro-cli agent configs provide hooks for automatic task loading, area enforcement, command guarding, and heartbeat reporting.

## Glossary

- **Crysknife**: The orchestrator CLI tool. Binary name is `crys`.
- **Agent**: A kiro-cli chat session running in a tmux pane with a specific Agent_Config.
- **Agent_Config**: A JSON file in `.kiro/agents/` that defines an agent's prompt, tools, context, and hooks.
- **State_File**: `.crysknife/state.json` -- the single source of truth for all agent status and orchestrator state.
- **Worktree**: A Git worktree created per worker so each operates on its own branch without conflicts.
- **Role**: One of three agent types: mayor, worker, merger.
- **Heartbeat**: A timestamp in State_File updated by the stop hook after each agent response.
- **Task_File**: A per-agent markdown file in `.kiro/specs/tasks/` containing the agent's current assignment.
- **Symlink**: A filesystem link from a Worktree back to the main project's shared directories.
- **Workflow_Tier**: One of three task complexity levels: full, standard, quick. Each has a corresponding template.
- **Convoy**: A feature-level grouping of related tasks across multiple agents.
- **Merge_Queue**: A list of completed worker branches waiting to be merged to main.
- **Staging_Branch**: `merge/staging` -- an intermediate branch where merges are tested before fast-forwarding to main.
- **Nudge**: A tmux send-keys message that prompts an idle agent to resume work.

## Requirements

### Requirement 1: Project Initialization

**User Story:** As a developer, I want to run `crys init` in my project root so that all Crysknife directories, templates, hook scripts, and base agent configs are created.

#### Acceptance Criteria

1. WHEN `crys init` is run in a directory, THE system SHALL create `.crysknife/` with `state.json`, `hooks/`, and `templates/` subdirectories.
2. WHEN `crys init` is run, THE system SHALL create `.crysknife/state.json` with an empty but valid initial structure (empty agents array, empty queue, empty convoys, empty merge_queue).
3. WHEN `crys init` is run, THE system SHALL copy default template files to `.crysknife/templates/` (mayor.md, worker-prompt.md, merger-prompt.md, full.md, standard.md, quick.md, principles.md).
4. WHEN `crys init` is run, THE system SHALL copy hook scripts to `.crysknife/hooks/` (enforce-area.sh, enforce-commands.sh) and make them executable.
5. WHEN `crys init` is run, THE system SHALL create `.kiro/agents/` with base agent configs (mayor.json, merger.json).
6. WHEN `crys init` is run, THE system SHALL create `.kiro/specs/tasks/` directory, `.kiro/specs/plan.md` (empty template), and `.kiro/specs/principles.md` (from templates/principles.md).
7. IF `.crysknife/` already exists, THEN `crys init` SHALL exit with an error message and not overwrite existing files.

### Requirement 2: Agent Startup

**User Story:** As a developer, I want to run `crys start` so that tmux sessions, Git worktrees, per-worker agent configs, and kiro-cli agents are created and launched automatically.

#### Acceptance Criteria

1. WHEN `crys start` is run, THE system SHALL generate per-worker agent configs (`.kiro/agents/worker-N.json`) from the worker template for each worker.
2. WHEN `crys start` is run, THE system SHALL create a Git worktree for each worker and the mayor via `git worktree add`, each on its own branch (workers on task branches, mayor on main). The merger operates in the main project directory (no worktree).
3. WHEN `crys start` is run, THE system SHALL create symlinks in each worktree (workers and mayor) pointing `.crysknife/` and `.kiro/specs/` back to the main project.
4. WHEN `crys start` is run, THE system SHALL create a tmux session with windows for: mayor, dashboard, worker pairs (2 workers per window as panes), merger, lazygit, nvim, and terminal.
5. WHEN `crys start` is run, THE system SHALL launch each agent with `kiro-cli chat --agent <name>` in its respective tmux pane, inside its Git worktree.
6. WHEN `crys start` is run, THE system SHALL update State_File with all agent entries (id, role, status, tmux_pane, task_file).
7. WHEN `crys start --workers N` is provided, THE system SHALL create exactly N workers (plus 1 mayor and 1 merger).
8. IF a tmux session with the project name already exists, THEN `crys start` SHALL exit with an error message.
9. WHEN `crys start --agent <agent-id>` is provided, THE system SHALL start only that specific agent.

### Requirement 3: Agent Shutdown

**User Story:** As a developer, I want to run `crys stop` so that all agents are shut down and tmux panes are cleaned up.

#### Acceptance Criteria

1. WHEN `crys stop` is run, THE system SHALL kill all agent tmux panes.
2. WHEN `crys stop` is run, THE system SHALL update State_File setting all agent statuses to "stopped".
3. WHEN `crys stop <agent-id>` is run with a specific agent, THE system SHALL kill only that agent's tmux pane and update only that agent's status.
4. WHEN `crys stop` is run, THE system SHALL NOT remove Git worktrees (they persist for inspection).

### Requirement 4: Status Display

**User Story:** As a developer, I want to run `crys status` so that I can see the current state of all agents, the work queue, and convoys at a glance.

#### Acceptance Criteria

1. WHEN `crys status` is run, THE system SHALL read State_File and display a table with columns: AGENT, ROLE, STATUS, TASK, BRANCH, TIER.
2. WHEN `crys status` is run, THE system SHALL display the work queue (pending tasks) below the agent table.
3. WHEN `crys status` is run, THE system SHALL display active convoys below the queue.
4. WHEN `crys status --json` is provided, THE system SHALL output the raw State_File contents as JSON to stdout.

### Requirement 5: Work Assignment

**User Story:** As a developer or mayor agent, I want to run `crys sling` so that a worker gets a task file generated from a workflow tier template, an updated agent config with area restrictions, and a nudge to start working.

#### Acceptance Criteria

1. WHEN `crys sling <worker-id> --task <name> --tier <tier>` is run, THE system SHALL read the template for the specified Workflow_Tier from `.crysknife/templates/<tier>.md`.
2. WHEN `crys sling` is run, THE system SHALL fill template variables (AGENT_ID, TASK_NAME, AREA, BRANCH) and write the result to `.kiro/specs/tasks/<worker-id>.md`.
3. WHEN `crys sling` is run with `--area <path>`, THE system SHALL regenerate the worker's Agent_Config with updated `toolsSettings.fs_write.allowedPaths` for the new area.
4. WHEN `crys sling` is run with `--branch <name>`, THE system SHALL create a new Git branch in the worker's Worktree.
5. WHEN `crys sling` is run, THE system SHALL update State_File (worker status to "working", task, branch, area, tier fields).
6. WHEN `crys sling` is run, THE system SHALL nudge the worker via tmux send-keys to pick up the new task.
7. WHEN `crys sling <worker-id> --from-queue` is provided, THE system SHALL pick the next task from the work queue and remove it from the queue.
8. IF the specified worker does not exist in State_File, THEN `crys sling` SHALL exit with an error.

### Requirement 6: Work Queue Management

**User Story:** As a developer or mayor agent, I want to manage a backlog of tasks so that work can be queued and assigned later.

#### Acceptance Criteria

1. WHEN `crys queue add <title> --tier <tier>` is run, THE system SHALL add a new entry to the queue array in State_File with a generated ID, title, tier, and creation timestamp.
2. WHEN `crys queue add` is run with `--area <path>`, THE system SHALL include the area in the queue entry.
3. WHEN `crys queue list` is run, THE system SHALL display all queued tasks with their ID, title, tier, and area.
4. WHEN `crys queue remove <task-id>` is run, THE system SHALL remove the task from the queue array in State_File.
5. IF the task-id does not exist in the queue, THEN `crys queue remove` SHALL exit with an error.

### Requirement 7: Monitoring Loop

**User Story:** As a developer, I want to run `crys watch` in a tmux pane so that idle agents are auto-nudged and dead agents are auto-restarted.

#### Acceptance Criteria

1. WHEN `crys watch` is run, THE system SHALL enter a polling loop that checks agent state every `--interval` seconds (default: 30).
2. WHEN checking agent state, THE system SHALL read `last_activity` timestamps from State_File (primary detection via Heartbeat).
3. IF an agent's last_activity is older than `--nudge-after` seconds (default: 120), THEN THE system SHALL send a Nudge to that agent and update its status to "nudged".
4. IF an agent's last_activity is older than `--restart-after` seconds (default: 300), THEN THE system SHALL kill the agent's pane, create a new pane, relaunch with `kiro-cli chat --agent <id>`, and update State_File.
5. WHEN Heartbeat detection is unavailable (no timestamp), THE system SHALL fall back to tmux pane content diffing (capture pane, hash content, compare to previous).
6. WHEN `crys watch` detects a worker has completed (status "done"), THE system SHALL nudge the mayor.
7. WHEN `crys watch` is running, THE system SHALL display a dashboard to stdout showing each agent's status and time since last activity.
8. WHEN `crys watch --no-nudge` is provided, THE system SHALL detect state but not auto-nudge.
9. WHEN `crys watch --no-restart` is provided, THE system SHALL detect state but not auto-restart.

### Requirement 8: Manual Nudge

**User Story:** As a developer, I want to manually nudge an idle agent so that it resumes working on its task.

#### Acceptance Criteria

1. WHEN `crys nudge <agent-id>` is run, THE system SHALL send a tmux send-keys message to the agent's pane telling it to check its task file and continue working.
2. WHEN `crys nudge --all` is run, THE system SHALL nudge all agents with status "idle" or "nudged".
3. WHEN an agent is nudged, THE system SHALL increment the agent's `nudge_count` in State_File.

### Requirement 9: Done Notification

**User Story:** As a worker agent, I want to run `crys done` so that my status is updated, my branch is added to the merge queue, and the mayor and merger are notified.

#### Acceptance Criteria

1. WHEN `crys done <worker-id>` is run, THE system SHALL update State_File (worker status to "done").
2. WHEN `crys done` is run, THE system SHALL add the worker's branch to the Merge_Queue in State_File with status "ready".
3. WHEN `crys done` is run, THE system SHALL nudge the mayor via tmux send-keys with a message about the completed worker.
4. WHEN `crys done` is run, THE system SHALL nudge the merger via tmux send-keys with a message about the new branch in the merge queue.
5. IF the worker has no branch in State_File, THEN `crys done` SHALL exit with an error.

### Requirement 10: Merge Queue Display

**User Story:** As the merger agent, I want to run `crys merge-queue` so that I can see which branches are ready to merge.

#### Acceptance Criteria

1. WHEN `crys merge-queue` is run, THE system SHALL display all entries in the Merge_Queue from State_File with branch name, worker, status, and completion time.
2. WHEN `crys merge-queue --json` is provided, THE system SHALL output the merge queue as JSON to stdout.

### Requirement 11: Merge Completion

**User Story:** As the merger agent, I want to run `crys merge-done` so that a merged branch is removed from the queue and the mayor is notified.

#### Acceptance Criteria

1. WHEN `crys merge-done <branch>` is run, THE system SHALL remove the branch from the Merge_Queue in State_File.
2. WHEN `crys merge-done` is run successfully, THE system SHALL update the corresponding worker's status to "merged" in State_File.
3. WHEN `crys merge-done <branch> --failed <reason>` is provided, THE system SHALL update the worker's status to "merge-failed" and store the failure reason.
4. WHEN `crys merge-done` is run, THE system SHALL nudge the mayor with a message about the merge result.
5. IF the branch does not exist in the Merge_Queue, THEN `crys merge-done` SHALL exit with an error.

### Requirement 12: Heartbeat

**User Story:** As an agent's stop hook, I want to run `crys heartbeat <agent-id>` so that the agent's last_activity timestamp is updated in State_File.

#### Acceptance Criteria

1. WHEN `crys heartbeat <agent-id>` is run, THE system SHALL update the `last_activity` field for that agent in State_File to the current UTC timestamp.
2. IF the agent-id does not exist in State_File, THEN `crys heartbeat` SHALL exit with a non-zero exit code and print an error to stderr.

### Requirement 13: Task Context Injection

**User Story:** As an agent's agentSpawn hook, I want to run `crys my-task <agent-id>` so that the agent's current task summary is printed to stdout and injected into context on startup.

#### Acceptance Criteria

1. WHEN `crys my-task <agent-id>` is run, THE system SHALL read the agent's entry from State_File and print a summary to stdout containing: agent id, role, status, task name, branch, area, and task file path.
2. IF the agent has no current task, THEN `crys my-task` SHALL print a message indicating no task is assigned.
3. IF the agent-id does not exist in State_File, THEN `crys my-task` SHALL exit with a non-zero exit code and print an error to stderr.

### Requirement 14: Feature-Level Tracking

**User Story:** As a developer, I want to group related tasks into convoys so that I can track feature completion across multiple agents.

#### Acceptance Criteria

1. WHEN `crys convoy create <name> --tasks <task-ids>` is run, THE system SHALL create a convoy entry in State_File with a generated ID, name, task list, and status "in-progress".
2. WHEN `crys convoy list` is run, THE system SHALL display all convoys with their name, status, and assigned agents.
3. WHEN `crys convoy status <name>` is run, THE system SHALL display the convoy's tasks with their individual statuses.

### Requirement 15: State File Integrity

**User Story:** As the orchestrator, I want state.json to be the single source of truth with consistent read/write behavior so that concurrent access from hooks and CLI commands does not corrupt data.

#### Acceptance Criteria

1. WHEN any `crys` command writes to State_File, THE system SHALL use atomic file writes (write to temp file, then rename) to prevent partial writes.
2. WHEN any `crys` command reads State_File, THE system SHALL handle the case where the file does not exist by returning an empty default state.
3. THE State_File SHALL conform to a consistent JSON schema with fields: project, created, agents (array), queue (array), convoys (array), merge_queue (array).

### Requirement 16: Agent Config Generation

**User Story:** As the orchestrator, I want to generate per-worker kiro-cli agent configs dynamically so that each worker gets the correct task file, area restrictions, and hooks.

#### Acceptance Criteria

1. WHEN `crys start` generates a worker config, THE config SHALL include: name, description, prompt (pointing to worker-prompt.md), tools, allowedTools, resources (design.md, principles.md, worker's task file), and hooks (agentSpawn, preToolUse for fs_write and execute_bash, stop).
2. WHEN `crys start` generates a worker config, THE config SHALL set `toolsSettings.fs_write.deniedPaths` to include `.crysknife/**` and `.kiro/specs/design.md`.
3. WHEN a worker config is generated, THE agentSpawn hook SHALL run `crys my-task <worker-id>`.
4. WHEN a worker config is generated, THE stop hook SHALL run `crys heartbeat <worker-id>`.
5. WHEN a worker config is generated, THE preToolUse hooks SHALL include matchers for both `fs_write` (enforce-area.sh) and `execute_bash` (enforce-commands.sh).
6. WHEN `crys sling` regenerates a worker config, THE config SHALL update `toolsSettings.fs_write.allowedPaths` to match the new area assignment.

### Requirement 17: Hook Scripts

**User Story:** As the orchestrator, I want preToolUse hook scripts that block dangerous file writes and shell commands so that agents stay within their assigned boundaries.

#### Acceptance Criteria

1. WHEN a worker calls `fs_write` on a file outside its assigned area, THE enforce-area.sh hook SHALL exit with code 2 and return an error message to the agent via stderr.
2. WHEN a worker calls `fs_write` on a file inside its assigned area or its own Task_File, THE enforce-area.sh hook SHALL exit with code 0 (allow).
3. WHEN any agent calls `execute_bash` with a command matching a blocked pattern (rm -rf, git push --force, sudo, etc.), THE enforce-commands.sh hook SHALL exit with code 2 and return an error message to the agent via stderr.
4. WHEN any agent calls `execute_bash` with a safe command, THE enforce-commands.sh hook SHALL exit with code 0 (allow).
5. THE enforce-commands.sh hook SHALL allow the merger role to run git merge and git rebase commands that are blocked for workers.
