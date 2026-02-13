# tswitch — A Grid-Based TMUX Navigator

`tswitch` is a terminal user interface (TUI) application for browsing, managing, and switching between TMUX sessions and windows. It replaces the default TMUX session/window chooser with a modern, grid-based layout featuring fuzzy search, tagging, and session management.

## Features

- **Grid layout** — Auto-fit columns based on terminal width
- **Two-level navigation** — Browse sessions, then drill into windows
- **Quick navigation** — Keyboard-driven with vim keybindings (h/j/k/l)
- **Marks** — Bookmark sessions/windows with single keys for instant switching (with visual hotkey display)
- **Preview panel** — See pane content or session metadata (toggle with Tab)
- **Session management** — Create, rename, and kill sessions/windows
- **Fuzzy search** — Filter sessions and windows by name
- **Session tagging** — Organize sessions with custom tags/groups
- **TMUX integration** — Launch as a popup overlay in TMUX

## Installation

### Build from source

```bash
git clone https://github.com/user/tswitch
cd tswitch
go build -o tswitch
sudo mv tswitch /usr/local/bin/
```

### Prerequisites

- Go 1.24+
- TMUX 3.0+ (3.2+ for popup support)

## Quick Start

### Run directly

```bash
tswitch
```

### Integrate with TMUX

Add this to your `~/.tmux.conf`:

```tmux
# Replace the default session chooser
bind-key s display-popup -E -w 80% -h 80% "tswitch"
```

Then reload TMUX config:

```bash
tmux source-file ~/.tmux.conf
```

Now press `prefix+s` to open `tswitch` as a popup.

## Usage

### Navigation

| Key | Action |
|-----|--------|
| `h/j/k/l` or `↑↓←→` | Move between cards in the grid |
| `Enter` | Zoom into session / Switch to window |
| `Space` | Quick switch to session's active window |
| `Esc` | Back to sessions / Quit |

### Marks

| Key | Action |
|-----|--------|
| `m` + key | Mark current session/window with a key |
| key | Switch to marked session/window (if marked) |

### Management

| Key | Action |
|-----|--------|
| `n` | New session (sessions view) or new window (windows view) |
| `r` | Rename focused item |
| `x` | Kill focused item (with confirmation) |
| `t` | Tag focused session |

### Other

| Key | Action |
|-----|--------|
| `/` | Filter sessions/windows by name (fuzzy search) |
| `Tab` | Toggle preview mode (capture-pane ↔ metadata) |
| `?` | Show help overlay |
| `q` | Quit |

## Configuration

Configuration is stored in `~/.config/tswitch/config.yaml`.

### Example config

```yaml
marks:
  w:
    session: work
    window: 0
    pane: 0
  p:
    session: personal
    window: 0
    pane: 0

tags:
  work:
    - project-a
    - project-b
  personal:
    - dotfiles
    - notes

settings:
  default_preview: capture    # "capture" or "metadata"
  theme: default
  sort_by: activity           # "activity", "name", or "tag"
```

### Marks Feature

Mark sessions and windows for quick access:

1. Press `m` to enter marking mode
2. Press any key (e.g., `w` for work, `p` for personal)
3. The session/window is bookmarked
4. Later, press that key in tswitch to instantly switch

See [MARKS_FEATURE.md](MARKS_FEATURE.md) for detailed documentation and examples.

Marked items display their hotkey in brackets: `work [w]` means press `w` to jump to the work session.

## Architecture

```
tswitch/
├── main.go                    # Entry point
├── internal/
│   ├── tmux/
│   │   ├── client.go          # TMUX command wrapper
│   │   ├── types.go           # Session, Window, Pane structs
│   │   └── capture.go         # Pane capture utilities
│   ├── config/
│   │   └── config.go          # Config file management
│   └── tui/
│       ├── model.go           # Main Bubbletea model
│       ├── grid.go            # Grid component
│       ├── styles.go          # Lipgloss theme
│       └── ...other components
└── scripts/
    └── tswitch.tmux           # TMUX integration example
```

## Roadmap

- [x] Grid layout with auto-fit columns
- [x] Session and window browsing
- [x] Quick switch functionality
- [x] Marks system (bookmark sessions/windows)
- [ ] Full preview panel (capture-pane + metadata)
- [ ] Fuzzy search/filter
- [ ] Session/window management (create, rename, kill)
- [ ] Session tagging/groups
- [ ] Status bar with keybinding hints
- [ ] Help overlay
- [ ] Error handling and edge cases
- [ ] TMUX 3.2+ popup support
- [ ] Configuration file support

## Troubleshooting

### No sessions found

Ensure TMUX is running and you have active sessions:

```bash
tmux list-sessions
```

### Popup doesn't work

TMUX 3.2+ is required for `display-popup` support. Check your version:

```bash
tmux -V
```

### Config not loading

Ensure `~/.config/tswitch/config.yaml` is valid YAML. Example:

```bash
cat ~/.config/tswitch/config.yaml
```

## Development

### Running with TMUX debug output

```bash
tswitch 2>&1 | tee /tmp/tswitch.log
```

### Testing without TMUX

The app gracefully handles being run outside TMUX and will attempt to attach to sessions instead of switching.

## License

MIT

## Contributing

Contributions welcome! Please open an issue or PR on GitHub.
