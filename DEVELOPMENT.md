# Development Guide

## Building from Source

```bash
git clone https://github.com/user/tswitch
cd tswitch
go build -o tswitch
```

## Running the App

### In TMUX

```bash
# Inside a TMUX session
./tswitch
```

### Outside TMUX

```bash
# Will show sessions and allow attaching
./tswitch
```

### As TMUX Popup

```bash
# From ~/.tmux.conf
bind-key s display-popup -E -w 80% -h 80% "tswitch"

# Then press prefix+s
```

## Code Organization

- `main.go` — Entry point
- `internal/tmux/` — TMUX integration
  - `types.go` — Data structures
  - `client.go` — Command wrapper
- `internal/config/` — Configuration
  - `config.go` — YAML config management
- `internal/tui/` — User interface
  - `model.go` — Main Bubbletea model
  - `grid.go` — Grid layout component
  - `preview.go` — Preview panel
  - `styles.go` — Styling
  - `filter.go` — Search/filter
  - `dialog.go` — Modal dialogs

## Key Concepts

### Grid Layout

The grid component auto-fits columns based on terminal width. Each "card" represents a session or window.

```go
// Calculate columns
cardWidthWithGap := 18 + 1  // 18 chars + 1 gap
columns := (width / cardWidthWithGap)
```

### Navigation

Vim-style navigation (hjkl) or arrow keys. Focus is tracked as a single index into a flat grid.

```go
// Convert flat index to row/col
row := focusIndex / columns
col := focusIndex % columns
```

### Two-Level Views

- **SessionGrid**: Top level, shows all sessions
- **WindowGrid**: Drilled into one session, shows windows
- Use `currentMode` to track which view is active

### Preview Panel

Side panel (40% width) shows metadata about the focused item. Can toggle between "metadata" and "capture" modes.

## Common Tasks

### Add a New Keybinding

1. Add to `handleKeyPress()` in `model.go`:
```go
case "x":
    // Handle 'x' key
```

### Add a New Status Line Message

1. Modify `renderStatusBar()` in `model.go`

### Change Styles/Colors

1. Edit `NewStyles()` in `styles.go`
2. Use Lipgloss color codes (e.g., `Color("39")` for blue)

### Add Configuration Option

1. Add to `Config` struct in `internal/config/config.go`
2. Update `config.yaml` example in README
3. Access via `m.config.<option>`

### Filter Results

1. Use `FilterSessions()` or `FilterWindows()` from `filter.go`
2. Apply to grid via `SetItems()`

## Testing

Currently manual testing. To test features:

```bash
# Create test TMUX sessions
tmux new-session -d -s work
tmux new-session -d -s personal
tmux new-window -t work -n editor
tmux new-window -t work -n terminal

# Run app
./tswitch

# Navigate with j/k, press enter to switch
```

## Debugging

### Print statements

```go
fmt.Printf("Debug: %#v\n", variable)
```

Bubbletea redirects stderr, so use a log file:

```go
f, _ := os.OpenFile("/tmp/tswitch.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
fmt.Fprintf(f, "Debug: %#v\n", variable)
```

### Common Issues

**No sessions appear**
- Ensure TMUX is running: `tmux list-sessions`
- Check `Client.IsInTmux()` — may need to handle both cases

**Grid doesn't display**
- Check terminal size: `echo $COLUMNS $LINES`
- Verify `width` and `height` are set correctly in Update()

**TMUX commands fail**
- Run manually: `tmux list-sessions -F '#{session_name}'`
- Check error in `Client.tmuxCommand()`

## Next Steps / TODOs

- [ ] Implement dialog integration (create/rename/kill)
- [ ] Add `n/r/x` keybindings with dialogs
- [ ] Integrate tag filtering (`/tag:work`)
- [ ] Capture-pane preview rendering
- [ ] Error handling and user feedback
- [ ] Unit tests
- [ ] Performance optimizations (parallel TMUX calls)
- [ ] Custom theme support from config
- [ ] Keybinding customization

## Build & Release

```bash
# Clean build
rm tswitch && go build -o tswitch

# Cross-compile for Linux/macOS
GOOS=linux GOARCH=amd64 go build -o tswitch-linux
GOOS=darwin GOARCH=amd64 go build -o tswitch-macos

# Strip binary for smaller size
strip tswitch
```

## Questions?

Refer to:
- `README.md` — User documentation
- `ARCHITECTURE.md` — Technical overview
- `go doc` — Package documentation
- Bubbletea examples: https://github.com/charmbracelet/bubbletea/tree/main/examples
