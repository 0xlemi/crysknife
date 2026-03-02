#!/bin/bash
# enforce-commands.sh — preToolUse hook for execute_bash
# Usage: enforce-commands.sh <agent-id>
# Receives hook event JSON on stdin

AGENT_ID="$1"
EVENT=$(cat)

COMMAND=$(echo "$EVENT" | jq -r '.tool_input.command // empty')

if [ -z "$COMMAND" ]; then
  exit 0
fi

ROLE=$(jq -r ".agents[] | select(.id==\"$AGENT_ID\") | .role" .crysknife/state.json 2>/dev/null)

BLOCKED_PATTERNS=(
  "rm -rf"
  "git push.*--force"
  "git push.*-f"
  "git checkout main"
  "git checkout master"
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

WORKER_ONLY_PATTERNS=(
  "git merge"
  "git rebase"
)

for pattern in "${BLOCKED_PATTERNS[@]}"; do
  if echo "$COMMAND" | grep -qE "$pattern"; then
    echo "BLOCKED: '$COMMAND' matches blocked pattern '$pattern'. If you need this, write it in your Feedback section for the mayor." >&2
    exit 2
  fi
done

if [ "$ROLE" != "merger" ]; then
  for pattern in "${WORKER_ONLY_PATTERNS[@]}"; do
    if echo "$COMMAND" | grep -qE "$pattern"; then
      echo "BLOCKED: '$COMMAND' matches blocked pattern '$pattern'. Only the merger can run git merge/rebase." >&2
      exit 2
    fi
  done
fi

exit 0
