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

## Subcommands

| Command | Description |
|---------|-------------|
| `tswitch` | Open the TUI |
| `tswitch last` | Switch to the previous tmux session (non-interactive) |
| `tswitch browse` | fzf directory browser; creates/attaches a session in the chosen dir. Requires `browse_dirs` in config and `fzf` in PATH |

## Architecture

The codebase has three main layers:

### 1. TMUX Layer (`internal/tmux/`)
- `types.go` — Core data structures: `Session`, `Window`, `Pane`, `Mark`
- `client.go` — Wraps tmux CLI commands via the `Service` interface; parses output using format strings like `#{session_name}|#{session_windows}|...`. `ListAllWindows` / `ListAllPanes` fetch every window and pane on the server in one call each: the tree view shows children of sessions it never drilled into, so fetching per session would mean one tmux call per visible row. `CurrentSession` reports the session of the calling client — the `session_attached` flag cannot stand in for it, because several sessions can be attached at once by different clients.
- `Service` interface enables dependency injection and future testability via mock implementations
- `ssh.go` — `DetectRemoteConnection()` inspects a pane's process tree (via `ps`) to identify SSH/mosh/ftp connections, including cases where the remote command is wrapped inside a shell (`zsh -c … ssh …`). Falls back to parsing `pane_title`. Used by the preview panel to show connection info.

### 2. Config Layer (`internal/config/`)
Two separate config files with distinct roles:
- **`~/.tswitch/state.yaml`** (`config.go`) — runtime state: marks, session/window ordering, tags, tree fold state (`expanded`), and the last-used view, sort and preview modes. Auto-managed; users rarely edit by hand.
- **`~/.tswitch/tswitch-config.json`** (`appconfig.go`) — read-only user preferences: key overrides (`keys`), browse directories (`browse_dirs`/`browse_exclude`), and UI settings (`ui.card_min_width`). Also searched next to the binary. Loaded once at startup via `LoadAppConfig()`; never written back.

`ui.preview_refresh_ms` (default 1000, `0` disables, floor 100) sets the capture preview's live refresh period.

`ui.card_min_width` (default 16) controls the minimum card content width. Increase it to fit longer session names — wider cards mean fewer columns.

`settings.pin_current_first` (default true) is the only setting with no key bound to it; it lives in `state.yaml` rather than the JSON config because it sits beside the sort mode it modifies.

### 3. Keys Layer (`internal/keys/`)
- `keys.go` — `Action` enum decouples key presses from business logic; supports vim (hjkl) and arrow keys. `ApplyOverrides` merges user config on top of defaults without wiping untouched bindings.

### 4. TUI Layer (`internal/tui/`)
- Follows the Bubbletea MVU pattern: `Model` → `Update` → `View`
- `model.go` — Central state: current mode, sessions/windows/panes lists, focus index, search query, preview visibility, active dialog, pending cut clipboard
- `handler.go` — Maps key input (`keys.Action`) to state mutations and TMUX commands. Also owns cross-session move via `handleCut`/`handlePaste` (clipboard persists across mode changes).
- `grid.go` — Responsive grid layout; calculates columns dynamically from terminal width. Minimum card content width is controlled by `minCardContentW` const (16) or `appConfig.UI.CardMinWidth` if set. Manages focus navigation (flat index ↔ row/col conversion).
- `view.go` — Top-level rendering, assembles grid + preview panel + dialog overlay + clipboard banner. `renderHelp` lays the key reference out in two columns and clamps it to the terminal height: Bubbletea's inline renderer discards whatever overflows the last row instead of letting the terminal scroll it, so a taller help screen loses its top silently. `statusLine` cuts the key hints from the right, which is why `?:help` sits at the front of every hint string.
- `cards.go` — `SessionCard`, `WindowCard`, and `PaneCard` implement the `GridItem` interface for polymorphic display
- `tree.go` — The list view widget: a flat slice of `TreeRow` rendered as an indented, scrollable list. Each row carries its own identity (`TreeRow.Key()`), used for mark lookup and for holding focus across rebuilds.
- `treebuild.go` — Builds the visible rows from the session list, the active filter and the saved fold state. Rows are rebuilt from scratch on every change rather than mutated.
- `preview.go` — Side panel with two modes, cycled by Tab: `PreviewCapture` (live pane content via `tmux capture-pane`) and `PreviewMetadata` (session/window info). Mode persists in `state.yaml`. In capture mode a `previewTickMsg` timer refetches the focused pane every `ui.preview_refresh_ms`; exactly one tick is ever in flight because each one reschedules itself on arrival. Captures run off the update loop, so every `captureResultMsg` carries the `previewGen` it was issued under and a result from a focus the cursor has already left is dropped instead of painted over the new one. The timer path does not clear the panel first (that would blink it empty once a second) and swallows capture errors, so a pane that dies between ticks leaves the last good content on screen. Layout split: session view is 60% grid / 40% preview; window/pane view is 50/50.
- `filter.go` — Fuzzy search filtering via `github.com/sahilm/fuzzy`
- `dialog.go` — Modal confirm/input overlays (e.g. rename, kill-session confirmation)
- `browse.go` — `tea.ExecProcess`-based fzf browser over `appConfig.BrowseDirs` for creating/attaching sessions rooted in a chosen directory. Also exposed as `tswitch browse` non-TUI subcommand via `main.go`.
- `styles.go` — All Lipgloss styling constants

### View Modes

`v` toggles `Settings.ViewMode` between `grid` and `list`, persisted in `state.yaml`.

The grid walks the three-level state machine below. The tree list spans all three levels at once, so while it is active `currentMode` stays pinned at `ModeSessionGrid` and depth is carried by each row's fold state instead. `handleTreeAction` intercepts the actions whose meaning changes in the tree (navigation, folding, reordering) and lets everything else fall through to the handlers shared with the grid.

### Session Ordering

`loadSessions` runs the list through `orderSessions`: `recent` (by `session_last_attached`, which tmux already sorts for us), `name` (`naturalLess`, so `2` sorts before `19`), or `manual` (`SessionOrder` from `state.yaml`). The session of the calling client is then pinned to the front unless `settings.pin_current_first` is false, and `initialFocusIndex` starts the cursor on the row *after* it — the one row the user will never pick is the one they are already in. Reorder keys refuse to act outside `manual` mode.

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
- **`focusTarget`**: identifies the item under the cursor (session, window or pane, with its coordinates) independently of the view mode. The grid derives it from the drill-down level, the tree from the depth of the focused row, so `handleRename`, `handleKill`, `handleCut`, `handlePaste`, `syncPreview` and friends have a single implementation each. Dialogs capture it as `pendingTarget` when they open, rather than re-reading focus on submit.
- **`switchTo`**: the one place that moves the client. Selecting a pane only changes which pane is active inside its own window, so a pane target needs `select-pane` *and* `switch-client` to actually take the user there.
- **Interface-based**: `tmux.Service` and `Executor` interfaces allow command-execution to be swapped (e.g., for testing)
- **`GridItem` interface**: `SessionCard`, `WindowCard`, and `PaneCard` all implement it, keeping grid rendering generic
- **Flat focus index**: The grid stores a single `focusIdx int` and converts to row/col as needed; navigation wraps at column boundaries
- **Layered interception in `handleKey`**: dialog → filter → marking → action dispatch. New overlays (like cut/paste) slot in as short-circuit checks before normal action handling.
