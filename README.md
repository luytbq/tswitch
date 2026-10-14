# tswitch

A TUI for browsing and switching tmux sessions, windows and panes. It replaces the default tmux chooser with a view that stays usable when you have forty sessions open.

## Features

- Two views, switchable with one key: a card grid, or a folding tree that puts sessions, windows and panes on one screen
- Live preview of the focused pane, refreshed once a second
- Fuzzy search across session and window names
- Marks: pin anything to a single key and jump straight to it
- Sort by most recent, by name, or in an order you arrange by hand
- Create, rename and kill sessions and windows without leaving the switcher
- Move a window to another session, or a pane into another window
- Open any directory as a session through an fzf browser

## Requirements

- tmux 3.0 or newer, 3.2 or newer for popup support
- Go 1.24 or newer, to install
- fzf, only for the directory browser

## Install

```bash
go install github.com/luytbq/tswitch@latest
```

The binary lands in `~/go/bin`, which has to be on your `PATH`. Check with `tswitch version`.

To build from a clone instead:

```bash
git clone https://github.com/luytbq/tswitch
cd tswitch
make build
```

Installing fzf, if you want the directory browser:

| OS | Command |
|----|---------|
| macOS | `brew install fzf` |
| Ubuntu / Debian | `sudo apt install fzf` |
| Fedora | `sudo dnf install fzf` |
| Arch Linux | `sudo pacman -S fzf` |
| Any (Go) | `go install github.com/junegunn/fzf@latest` |

## tmux setup

Add to your `~/.tmux.conf`:

```tmux
# Open the switcher as a popup
bind-key S display-popup -E -w 80% -h 80% "tswitch"

# Switch to the previous session
bind-key l run-shell "tswitch last"

# Open the directory browser as a popup (needs fzf and browse_dirs)
bind-key f display-popup -E -w 80% -h 80% "tswitch browse"
```

Reload with `tmux source-file ~/.tmux.conf`. `display-popup` needs tmux 3.2 or newer; without it, run `tswitch` in a window.

## Commands

| Command | What it does |
|---------|--------------|
| `tswitch` | Open the switcher |
| `tswitch last` | Switch to the previous session and exit, no UI |
| `tswitch browse` | Pick a directory with fzf and open it as a session |
| `tswitch version` | Print the installed version |

## Keys

Press `?` inside tswitch for the same list.

| Key | Action |
|-----|--------|
| `h/j/k/l` or arrows | Move the cursor |
| `o` | Grid: drill in one level. List: fold or unfold |
| `Enter` or `Space` | Switch to whatever the cursor is on, and quit |
| `Esc` | Back one level, or fold, or quit |
| `v` | Switch between card grid and tree list |
| `s` | Cycle the sort order |
| `H/L` | List: fold or unfold every session |
| `J/K` | Move the focused item up or down (sort `manual` only) |
| `m` then a key | Mark the focused item with that key |
| a mark key | Jump to whatever that key is marked to |
| `/` | Fuzzy search |
| `Tab` | Switch the preview between pane contents and metadata |
| `n` | New session or window |
| `r` | Rename |
| `d` | Kill, with confirmation |
| `x` | Cut the focused window or pane, press again to cancel |
| `p` | Paste what you cut onto the focused destination |
| `f` | Open the directory browser (needs fzf and `browse_dirs`) |
| `?` | Show this list |
| `q` | Quit |

In the grid, `H/J/K/L` all reorder, since the cursor moves with `h/j/k/l`.

## Views

`v` switches between two ways of showing the same sessions, and tswitch remembers which one you used last.

**Grid** lays sessions out as cards, left to right then top to bottom. `o` drills down one level at a time: sessions, then windows, then panes, and `Esc` comes back up.

**List** is an indented tree where sessions, windows and panes share one screen. Every node starts folded, so forty sessions still fit without scrolling. `o` or `l` unfolds the node under the cursor, `h` folds it again, and `H`/`L` do it to all of them at once. What you left unfolded is still unfolded next time. `Enter` switches to whatever the cursor is on, at any depth, so you can jump straight to a pane three levels down.

## Sorting

`s` cycles the session order and remembers the choice:

| Mode | Order |
|------|-------|
| `recent` | most recently used first, the default |
| `name` | by name, with digits compared as numbers, so `2` comes before `19` |
| `manual` | the order you built with `J/K` |

The session you are in sits at the top in every mode, and the cursor starts one row below it. Open the switcher, press `Enter`, and you are back where you came from. `J/K` only reorder in `manual` mode; in the other two tswitch tells you to press `s` first.

## Marks

Press `m` and then any unused key to mark whatever the cursor is on. A session, a window or a pane all work. After that, pressing that key anywhere in the switcher takes you straight there, and the mark shows up beside the item as `[key]`.

Marks are saved, so they survive restarts. Keys already bound to a command are refused, and tswitch says so.

## Moving windows and panes

`x` cuts the focused window or pane and shows a banner at the top. Navigate wherever you want, then press `p`: a window pastes onto a session, a pane pastes onto a window. The banner follows you across views, so you can take your time finding the destination. Pressing `x` again cancels.

## Configuration

### Preferences: `tswitch-config.json`

tswitch looks for this file next to the binary first (for example `~/go/bin/tswitch-config.json`), then at `~/.tswitch/tswitch-config.json`. The first one found wins, and if there is none, tswitch runs on defaults.

A complete example listing every option ships with the source: [`tswitch-config.json`](./tswitch-config.json). Save it as `~/.tswitch/tswitch-config.json` and any tswitch on your machine picks it up.

**`keys`** rebinds any command. Names: `move_up`, `move_down`, `move_left`, `move_right`, `confirm`, `quick_swap`, `direct_switch`, `back`, `start_mark`, `new`, `rename`, `kill`, `cut`, `paste`, `tag`, `reorder_up`, `reorder_down`, `reorder_left`, `reorder_right`, `toggle_view`, `cycle_sort`, `toggle_preview`, `toggle_help`, `filter`, `quit`, `browse_dirs`.

**`ui.preview_refresh_ms`** is how often the preview refetches the focused pane, in milliseconds, default `1000`. Set it to `0` to refresh only when the cursor moves. Anything below `100` is raised to `100`, because every refresh runs a `tmux capture-pane`. Only the pane-contents preview refreshes; metadata holds still, and nothing is fetched while a dialog or the key list covers the panel.

**`ui.card_min_width`** is the narrowest a grid card may get, in characters, default `16`. Raise it if your session names are being truncated; `20` suits names longer than about a dozen characters. Wider cards mean fewer columns.

**`browse_dirs`** lists the directories `tswitch browse` offers, as `{path, depth}` pairs, where `depth` is how many levels down to look. Needs [fzf](https://github.com/junegunn/fzf) on your `PATH`.

**`browse_exclude`** lists directory names to skip while scanning, matched on the name alone, so `node_modules` skips every one of them.

### State: `~/.tswitch/state.yaml`

tswitch writes this itself. It holds your marks, the manual order, which nodes are unfolded, and the view, sort and preview modes you last used. You do not normally edit it.

One setting there has no key bound to it: set `settings.pin_current_first` to `false` to stop pinning the session you are in to the top.

## Troubleshooting

**No sessions found.** Check that tmux is actually running, with `tmux list-sessions`.

**The popup does nothing.** `display-popup` needs tmux 3.2 or newer. Check with `tmux -V`, and run `tswitch` in an ordinary window otherwise.

**`tswitch` is not found after installing.** `go install` puts it in `~/go/bin`. Add that to your `PATH`.

**`J/K` do not move anything.** Reordering only works in `manual` sort. Press `s` until the header says `sort: manual`.

**A key will not take a mark.** Keys bound to commands cannot be marks. Pick a letter or digit that does not appear in the key list above.

**`tswitch browse` says no browse directories configured.** Either no `tswitch-config.json` was found in the two places tswitch looks, or it has no `browse_dirs`. See [Configuration](#preferences-tswitch-configjson).

## License

MIT
