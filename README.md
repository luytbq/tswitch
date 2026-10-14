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

### Install dependency `fzf`:

| OS | Command |
|----|---------|
| macOS | `brew install fzf` |
| Ubuntu / Debian | `sudo apt install fzf` |
| Fedora | `sudo dnf install fzf` |
| Arch Linux | `sudo pacman -S fzf` |
| Any (Go) | `go install github.com/junegunn/fzf@latest` |


## tmux Integration

Add to your `~/.tmux.conf`:

```tmux
# Open session/window switcher as a popup
bind-key S display-popup -E -w 80% -h 80% "tswitch"

# Switch to the previous session
bind-key l run-shell "tswitch last"

# Open directory browser as a popup (requires fzf + browse_dirs config)
bind-key f display-popup -E -w 80% -h 80% "tswitch browse"
```

Reload with `tmux source-file ~/.tmux.conf`. Requires tmux 3.2+ for `display-popup`.

## Subcommands

| Command | Description |
|---------|-------------|
| `tswitch` | Open the TUI |
| `tswitch last` | Switch to the previous tmux session |

## Key Bindings

| Key | Action |
|-----|--------|
| `h/j/k/l` or arrows | Navigate |
| `v` | Switch between card grid and tree list |
| `s` | Cycle session sort: `recent`, `name`, `manual` |
| `o` | Grid: drill into focused item. List: fold/unfold |
| `Enter` | Switch directly to focused item |
| `Space` | Quick-switch to session's active window |
| `Esc` | Back one level / fold / quit |
| `H/J/K/L` | Grid: reorder focused item. List: `J/K` reorder, `H/L` fold/unfold all |
| `m` + key | Mark current item with a hotkey |
| _mark key_ | Jump to marked session/window |
| `/` | Fuzzy search filter |
| `Tab` | Switch the preview between pane capture and metadata |
| `n` | New session or window |
| `r` | Rename focused item |
| `d` | Kill focused item (with confirmation) |
| `x` | Cut focused window/pane to clipboard |
| `p` | Paste clipboard onto focused destination |
| `t` | Reserved for tagging; no action bound yet |
| `?` | Key reference overlay (`?` or `Esc` closes it) |
| `q` | Quit |

## Views

`v` switches between two ways of showing the same sessions, and the choice is remembered.

**Grid** — cards laid out left to right, then top to bottom. `o` drills down one level at a time: sessions, then windows, then panes.

**List** — an indented tree where sessions, windows and panes share one screen. Every node starts folded, so a long session list fits without scrolling; `o` or `l` unfolds the node under the cursor and `h` folds it again. Fold state is remembered between runs. `Enter` switches to whatever the cursor is on, at any depth.

## Sorting

`s` cycles the session order and remembers the choice:

| Mode | Order |
|------|-------|
| `recent` | most recently attached first (the default) |
| `name` | by name, with digit runs compared as numbers so `2` sorts before `19` |
| `manual` | the order you built with `H/J/K/L`, stored in `session_order` |

The session you are currently in is pinned to the top in every mode, and the cursor starts on the row below it, so opening the switcher and pressing `Enter` takes you back where you came from. Reordering keys only apply in `manual` mode.

## Configuration

### `tswitch-config.json`

tswitch looks for `tswitch-config.json` in two locations, in this order:

1. **Next to the binary** — the same directory as the `tswitch` executable (e.g. `~/go/bin/tswitch-config.json` if you installed via `go install`).
2. **User home** — `~/.tswitch/tswitch-config.json`.

The first file found wins; if neither exists, tswitch runs with defaults (no custom keys, no browse directories).

A complete reference config listing every supported key binding, `browse_dirs`, and `browse_exclude` is checked into the repo at [`tswitch-config.json`](./tswitch-config.json) — use it as a starting template. Save it to `~/.tswitch/tswitch-config.json` and it will be picked up by any `tswitch` binary on your system.

**`keys`** — override default key bindings. Action names: `move_up`, `move_down`, `move_left`, `move_right`, `confirm`, `quick_swap`, `direct_switch`, `back`, `start_mark`, `new`, `rename`, `kill`, `cut`, `paste`, `tag`, `reorder_up`, `reorder_down`, `reorder_left`, `reorder_right`, `toggle_view`, `cycle_sort`, `toggle_preview`, `toggle_help`, `filter`, `quit`, `browse_dirs`.

**`ui.card_min_width`** — minimum card content width in characters (default: `16`). Increase this to fit longer session/window names without truncation; for example, `20` is a good value if your names regularly exceed 11–12 characters. Wider cards mean fewer columns on the same terminal width.

**`browse_dirs`** — directories that `tswitch browse` scans for subdirectories to open as new tmux sessions. Each entry is a `{path, depth}` pair; `depth` is how many levels to descend. Requires [`fzf`](https://github.com/junegunn/fzf) to be available in `PATH`.


**`browse_exclude`** — directory names to skip while scanning `browse_dirs` (matched by basename).

### Runtime state — `~/.tswitch/state.yaml`

Auto-managed by tswitch. Stores marks, session/window ordering, tags, the fold state of the tree list, and the last-used view, sort and preview modes. You normally don't need to edit this by hand.

`settings.pin_current_first` is the one knob with no key bound to it: set it to `false` to stop pinning the current session to the top of the list.

## Troubleshooting

**No sessions found** — make sure tmux is running (`tmux list-sessions`).

**Popup doesn't work** — tmux 3.2+ is required for `display-popup`. Check with `tmux -V`.

**`tswitch browse` prints `Error: no browse directories configured`** — either no `tswitch-config.json` was found in the two lookup locations (see [Configuration](#tswitch-configjson)), or the config file exists but has no `browse_dirs` entries. Add a `browse_dirs` array and restart.

## License

MIT
