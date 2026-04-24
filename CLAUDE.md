# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**tswitch** is a TUI application for browsing and switching TMUX sessions and windows. It replaces the default TMUX session/window chooser with a grid-based layout featuring fuzzy search, marks/bookmarks, and session management. Built in Go with the Bubbletea/Lipgloss (charmbracelet) ecosystem.

## Commands

```bash
# Build (injects version from git tag)
make build

# Run
./tswitch

# Run with go
go run main.go

# Tidy dependencies
go mod tidy
```

There is no automated test suite — testing is done manually by running the app against live TMUX sessions.

## Architecture

The codebase has three main layers:

### 1. TMUX Layer (`internal/tmux/`)
- `types.go` — Core data structures: `Session`, `Window`, `Pane`, `Mark`
- `client.go` — Wraps tmux CLI commands via the `Service` interface; parses output using format strings like `#{session_name}|#{session_windows}|...`
- `Service` interface enables dependency injection and future testability via mock implementations

### 2. Config Layer (`internal/config/`)
- Manages persistent YAML state at `~/.tswitch/state.yaml`
- Handles the marks system (single-key bookmarks for sessions/windows), session tags, and user settings
- Falls back to sensible defaults when config is absent

### 3. Keys Layer (`internal/keys/`)
- `keys.go` — `Action` enum decouples key presses from business logic; supports vim (hjkl) and arrow keys. `ApplyOverrides` merges user config on top of defaults without wiping untouched bindings.

### 4. TUI Layer (`internal/tui/`)
- Follows the Bubbletea MVU pattern: `Model` → `Update` → `View`
- `model.go` — Central state: current mode, sessions/windows/panes lists, focus index, search query, preview visibility, active dialog, pending cut clipboard
- `handler.go` — Maps key input (`keys.Action`) to state mutations and TMUX commands. Also owns cross-session move via `handleCut`/`handlePaste` (clipboard persists across mode changes).
- `grid.go` — Responsive grid layout; calculates columns dynamically from terminal width using fixed card dimensions (22-char content width); manages focus navigation (flat index ↔ row/col conversion)
- `view.go` — Top-level rendering, assembles grid + preview panel + dialog overlay + clipboard banner
- `cards.go` — `SessionCard`, `WindowCard`, and `PaneCard` implement the `GridItem` interface for polymorphic display
- `preview.go` — Toggleable side panel (Tab key) showing session metadata or pane capture
- `filter.go` — Fuzzy search filtering via `github.com/sahilm/fuzzy`
- `dialog.go` — Modal confirm/input overlays (e.g. rename, kill-session confirmation)
- `browse.go` — `tea.ExecProcess`-based fzf browser over `appConfig.BrowseDirs` for creating/attaching sessions rooted in a chosen directory
- `styles.go` — All Lipgloss styling constants

### Navigation State Machine

```
ModeSessionGrid ──'o'──▶ ModeWindowGrid ──'o'──▶ ModePaneGrid
       ▲                       │                       │
       └──────── Esc ──────────┴──────── Esc ──────────┘
```

Three modes in a linear drill-down: `ModeSessionGrid` → `ModeWindowGrid` → `ModePaneGrid`. `o` (ActionConfirm) drills into the focused item. `Enter` (ActionDirectSwitch) switches immediately to the focused item without drilling. `Esc` (ActionBack) pops one level up. `l` is grid-right movement, not drill-down.

### Cut/Paste Clipboard
`handleCut` captures the focused window or pane into `m.clipboard`; a second press clears it. `handlePaste` commits the move via `MoveWindow` (tmux `move-window`) or `JoinPane` (tmux `join-pane`) once the destination mode matches — windows paste onto sessions, panes paste onto windows. The clipboard banner rides across mode changes so the user can freely navigate the grids to pick a destination.

### Key Design Patterns
- **Interface-based**: `tmux.Service` and `Executor` interfaces allow command-execution to be swapped (e.g., for testing)
- **`GridItem` interface**: `SessionCard`, `WindowCard`, and `PaneCard` all implement it, keeping grid rendering generic
- **Flat focus index**: The grid stores a single `focusIdx int` and converts to row/col as needed; navigation wraps at column boundaries
- **Layered interception in `handleKey`**: dialog → filter → marking → action dispatch. New overlays (like cut/paste) slot in as short-circuit checks before normal action handling.
