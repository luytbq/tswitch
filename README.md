# tswitch

A grid-based TUI for browsing and switching tmux sessions and windows.

## Features

- **Grid layout** — responsive card grid that auto-fits columns to terminal width
- **Three-level navigation** — browse sessions, drill into windows, drill into panes
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
| `o` | Drill into focused item (session → windows → panes) |
| `Enter` | Switch directly to focused item |
| `Space` | Quick-switch to session's active window |
| `Esc` | Back one level / quit |
| `H/J/K/L` | Reorder focused item (Shift + direction) |
| `m` + key | Mark current item with a hotkey |
| _mark key_ | Jump to marked session/window |
| `/` | Fuzzy search filter |
| `Tab` | Toggle preview panel |
| `n` | New session or window |
| `r` | Rename focused item |
| `d` | Kill focused item (with confirmation) |
| `x` | Cut focused window/pane to clipboard |
| `p` | Paste clipboard onto focused destination |
| `t` | Tag focused session |
| `?` | Help overlay |
| `q` | Quit |

## Configuration

### `tswitch-config.json`

tswitch looks for `tswitch-config.json` in two locations, in this order:

1. **Next to the binary** — the same directory as the `tswitch` executable (e.g. `~/go/bin/tswitch-config.json` if you installed via `go install`).
2. **User home** — `~/.tswitch/tswitch-config.json`.

The first file found wins; if neither exists, tswitch runs with defaults (no custom keys, no browse directories).

A complete reference config listing every supported key binding, `browse_dirs`, and `browse_exclude` is checked into the repo at [`tswitch-config.json`](./tswitch-config.json) — use it as a starting template. Save it to `~/.tswitch/tswitch-config.json` and it will be picked up by any `tswitch` binary on your system.

Example config:

```json
{
  "keys": {
    "quit": "Q",
    "filter": "f"
  },
  "browse_dirs": [
    {"path": "~/projects", "depth": 4},
    {"path": "~/.config", "depth": 3}
  ],
  "browse_exclude": [
    ".git",
    "node_modules",
    "vendor"
  ]
}
```

**`keys`** — override default key bindings. Action names: `move_up`, `move_down`, `move_left`, `move_right`, `confirm`, `quick_swap`, `back`, `start_mark`, `new`, `rename`, `kill`, `cut`, `paste`, `tag`, `reorder_up`, `reorder_down`, `reorder_left`, `reorder_right`, `toggle_preview`, `toggle_help`, `filter`, `quit`.

**`browse_dirs`** — directories that `tswitch browse` scans for subdirectories to open as new tmux sessions. Each entry is a `{path, depth}` pair; `depth` is how many levels to descend. Requires [`fzf`](https://github.com/junegunn/fzf) to be available in `PATH`.

Install `fzf`:

| OS | Command |
|----|---------|
| macOS | `brew install fzf` |
| Ubuntu / Debian | `sudo apt install fzf` |
| Fedora | `sudo dnf install fzf` |
| Arch Linux | `sudo pacman -S fzf` |
| Any (Go) | `go install github.com/junegunn/fzf@latest` |

**`browse_exclude`** — directory names to skip while scanning `browse_dirs` (matched by basename).

### Runtime state — `~/.tswitch/state.yaml`

Auto-managed by tswitch. Stores marks, session/window ordering, and tags. You normally don't need to edit this by hand.

## Troubleshooting

**No sessions found** — make sure tmux is running (`tmux list-sessions`).

**Popup doesn't work** — tmux 3.2+ is required for `display-popup`. Check with `tmux -V`.

**`tswitch browse` prints `Error: no browse directories configured`** — either no `tswitch-config.json` was found in the two lookup locations (see [Configuration](#tswitch-configjson)), or the config file exists but has no `browse_dirs` entries. Add a `browse_dirs` array and restart.

## License

MIT
