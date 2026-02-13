#!/bin/bash

# Test script for marks feature
# Creates test TMUX sessions and demonstrates marks

set -e

echo "🧪 Setting up test TMUX sessions..."

# Kill existing test sessions
tmux kill-session -t test-work 2>/dev/null || true
tmux kill-session -t test-personal 2>/dev/null || true
tmux kill-session -t test-api 2>/dev/null || true

# Create test sessions
tmux new-session -d -s test-work -x 80 -y 24
tmux new-session -d -s test-personal -x 80 -y 24
tmux new-session -d -s test-api -x 80 -y 24

# Add windows to test-work
tmux new-window -t test-work -n editor
tmux new-window -t test-work -n terminal
tmux new-window -t test-work -n logs

echo "✓ Created test sessions:"
echo "  - test-work (windows: 0:default, 1:editor, 2:terminal, 3:logs)"
echo "  - test-personal"
echo "  - test-api"
echo ""

# Create a test config
mkdir -p ~/.config/tswitch
cat > ~/.config/tswitch/test_marks.yaml << 'YAML'
marks:
  w:
    session: test-work
    window: 0
    pane: 0
  e:
    session: test-work
    window: 1
    pane: 0
  t:
    session: test-work
    window: 2
    pane: 0
  p:
    session: test-personal
    window: 0
    pane: 0
  a:
    session: test-api
    window: 0
    pane: 0
tags: {}
settings:
  default_preview: metadata
  theme: default
  sort_by: activity
YAML

echo "📝 Test marks configured:"
echo "  w → test-work (window 0)"
echo "  e → test-work (window 1 - editor)"
echo "  t → test-work (window 2 - terminal)"
echo "  p → test-personal (window 0)"
echo "  a → test-api (window 0)"
echo ""

echo "🎮 How to test marks:"
echo "  1. Run: tswitch"
echo "  2. Press 'w' to jump to test-work:0"
echo "  3. Reopen: tswitch"
echo "  4. Press 'e' to jump to test-work:1 (editor)"
echo "  5. Reopen: tswitch"
echo "  6. Press 'p' to jump to test-personal"
echo ""

echo "📋 Manual testing steps:"
echo "  1. Mark test-work with 'w': nav → m → w"
echo "  2. Mark test-work:1 with 'e': enter → nav → m → e"
echo "  3. Press 'w' to jump to test-work"
echo "  4. Press 'e' to jump to test-work:1"
echo ""

echo "✨ Setup complete! Run './tswitch' to test."
