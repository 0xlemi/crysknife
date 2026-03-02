#!/bin/bash
# enforce-area.sh — preToolUse hook for fs_write
# Usage: enforce-area.sh <worker-id>
# Receives hook event JSON on stdin

WORKER_ID="$1"
EVENT=$(cat)

FILE_PATH=$(echo "$EVENT" | jq -r '.tool_input.path // .tool_input.operations[0].path // empty')

if [ -z "$FILE_PATH" ]; then
  exit 0
fi

AREA=$(jq -r ".agents[] | select(.id==\"$WORKER_ID\") | .area" .crysknife/state.json)

if [ -z "$AREA" ] || [ "$AREA" = "null" ]; then
  exit 0
fi

case "$FILE_PATH" in
  $AREA*|.kiro/specs/tasks/$WORKER_ID.md)
    exit 0
    ;;
  *)
    echo "BLOCKED: $FILE_PATH is outside your assigned area ($AREA). Write your request in the Feedback section of your task file and the mayor will assign it to the correct worker." >&2
    exit 2
    ;;
esac
