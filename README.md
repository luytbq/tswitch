# tswitch

A grid-based TUI for browsing and switching tmux sessions and windows.

## Features

- **Grid layout** — responsive card grid that auto-fits columns to terminal width
- **Two-level navigation** — browse sessions, drill into windows
- **Fuzzy search** — filter sessions and windows by name
- **Marks** — bookmark sessions/windows with single-key hotkeys for instant switching
- **Preview panel** — toggle between pane capture and session/window metadata
- **Reorder** — rearrange sessions and windows with Shift+H/J/K/L, persisted across runs
- **Session management** — create, rename, and kill sessions and windows
- **Custom key bindings** — override default keys via JSON config
- **`tswitch last`** — switch to the previous tmux session from the command line

## Installation

### Prerequisites

- Go 1.24+
- tmux 3.0+ (3.2+ for popup support)

### Install

```bash
go install github.com/luytbq/tswitch@latest
```

Or build from source:

```bash
git clone https://github.com/luytbq/tswitch
cd tswitch
go build -o tswitch
```

## tmux Integration

Add to your `~/.tmux.conf` to launch tswitch as a popup overlay:

```tmux
bind-key s display-popup -E -w 80% -h 80% "tswitch"
```

Reload with `tmux source-file ~/.tmux.conf`, then press `prefix + s` to open.

## Subcommands

| Command | Description |
|---------|-------------|
| `tswitch` | Open the TUI |
| `tswitch last` | Switch to the previous tmux session |

## Key Bindings

| Key | Action |
|-----|--------|
| `h/j/k/l` or arrows | Navigate the grid |
| `Enter` | Drill into session / switch to window |
| `Space` | Quick-switch to session's active window |
| `Esc` | Back to sessions / quit |
| `H/J/K/L` | Reorder focused item (Shift + direction) |
| `m` + key | Mark current item with a hotkey |
| _mark key_ | Jump to marked session/window |
| `/` | Fuzzy search filter |
| `Tab` | Toggle preview panel |
| `n` | New session or window |
| `r` | Rename focused item |
| `x` | Kill focused item (with confirmation) |
| `t` | Tag focused session |
| `?` | Help overlay |
| `q` | Quit |

## Configuration

### Custom key bindings — `tswitch-config.json`

Override default keys by placing a `tswitch-config.json` file next to the binary or in `~/.tswitch/`. The file is checked in that order; the first one found wins.

```json
{
  "keys": {
    "quit": "Q",
    "filter": "f"
  }
}
```

Action names match the defaults: `move_up`, `move_down`, `move_left`, `move_right`, `confirm`, `quick_swap`, `back`, `start_mark`, `new`, `rename`, `kill`, `tag`, `reorder_up`, `reorder_down`, `reorder_left`, `reorder_right`, `toggle_preview`, `toggle_help`, `filter`, `quit`.

### Runtime state — `~/.tswitch/state.yaml`

Auto-managed by tswitch. Stores marks, session/window ordering, and tags. You normally don't need to edit this by hand.

## Troubleshooting

**No sessions found** — make sure tmux is running (`tmux list-sessions`).

**Popup doesn't work** — tmux 3.2+ is required for `display-popup`. Check with `tmux -V`.

## License

MIT
