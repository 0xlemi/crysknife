#!/bin/bash
# Tmux session setup for crysknife project

SESSION="crysknife"

tmux has-session -t $SESSION 2>/dev/null
if [ $? != 0 ]; then
  tmux new-session -d -s $SESSION -c /Users/lemi/Documents/code/projects/crysknife

  tmux rename-window -t $SESSION:1 'nvim'
  tmux send-keys -t $SESSION:1 'nvim .' C-m

  tmux new-window -t $SESSION:2 -n 'kiro-cli' -c /Users/lemi/Documents/code/projects/crysknife
  tmux send-keys -t $SESSION:2 'kiro-cli chat --resume-picker' C-m

  tmux new-window -t $SESSION:3 -n 'lazygit' -c /Users/lemi/Documents/code/projects/crysknife
  tmux send-keys -t $SESSION:3 'lazygit' C-m

  tmux new-window -t $SESSION:4 -n 'terminal' -c /Users/lemi/Documents/code/projects/crysknife

  tmux select-window -t $SESSION:1
fi

if [ -z "$DETACHED" ]; then
  tmux attach -t $SESSION
fi
