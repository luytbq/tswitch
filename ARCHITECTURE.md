# tswitch Architecture

## Overview

`tswitch` is a grid-based TMUX session/window navigator built with Go and Bubbletea. The architecture is organized into logical layers:

```
main.go
  └── internal/
      ├── tmux/          # TMUX system integration
      ├── config/        # Configuration management  
      └── tui/           # User interface components
```

## Module Breakdown

### `main.go`
Entry point. Initializes the Bubbletea TUI program and starts the event loop.

### `internal/tmux/`

**`types.go`**
- Defines core data structures: `Session`, `Window`, `Pane`
- `SessionWithWindows` and `WindowWithPanes` wrapper types for hierarchical data

**`client.go`**
- `Client` struct wraps TMUX command execution
- Methods:
  - List operations: `ListSessions()`, `ListWindows()`, `ListPanes()`
  - Navigation: `SwitchClient()`, `AttachSession()`
  - Management: `NewSession()`, `RenameSession()`, `KillSession()`, `NewWindow()`, etc.
  - Capture: `CapturePane()` for preview panel
- Handles parsing TMUX output format strings into Go structs
- Gracefully handles both in-TMUX and outside-TMUX scenarios

### `internal/config/`

**`config.go`**
- Manages `~/.config/tswitch/config.yaml`
- `Config` struct holds:
  - `Tags` — map of tag names to session lists for grouping
  - `Settings` — app-level preferences (theme, default preview mode, sort order)
- Functions:
  - `LoadConfig()` — read config with sensible defaults
  - `SaveConfig()` — persist changes
  - `GetSessionTags()`, `AddSessionTag()`, `RemoveSessionTag()` — tag management

### `internal/tui/`

**`styles.go`**
- Centralized Lipgloss styling definitions
- `Styles` struct groups related styles:
  - Card styles (focused/unfocused)
  - Preview panel styles
  - Filter, status bar, help overlay styles
- Easy theme customization point

**`grid.go`**
- `Grid` component manages the layout of items in a grid
- Features:
  - Auto-fit columns based on terminal width
  - Focus management (hjkl/arrow navigation)
  - Scrolling support
  - Converts items to cards via rendering
- Methods:
  - `SetItems()` — populate grid and recalculate layout
  - `SetSize()` — respond to terminal resize
  - `MoveFocus()` — navigate with dx/dy deltas
  - `Render()` — generate grid display

**`model.go`**
- `Model` is the main Bubbletea application model
- Implements Bubbletea interface: `Init()`, `Update()`, `View()`
- State management:
  - `sessionGrid` / `windowGrid` — two-level navigation
  - `previewPanel` — side panel
  - `currentMode` — SessionGrid or WindowGrid view
  - `filter` — search term
- Key methods:
  - `loadSessions()` / `loadWindows()` — fetch from TMUX
  - `handleKeyPress()` — input handling
  - `renderSessionView()` / `renderWindowView()` — rendering
  - Session/window management methods (`NewSession()`, `RenameWindow()`, etc.)
- Helper types:
  - `SessionCard` / `WindowCard` — wrapper types for grid items

**`preview.go`**
- `PreviewPanel` displays session/window details
- Supports two modes:
  - `"metadata"` — structured info (names, counts, paths, times)
  - `"capture"` — raw pane output (future: will use `CapturePane()`)
- Methods:
  - `SetSessionMetadata()` / `SetWindowMetadata()` — update content
  - `ToggleMode()` — switch between modes
  - `Render()` — generate bordered preview display

**`filter.go`**
- Fuzzy search/filter functions
- `FilterSessions()` / `FilterWindows()` — use `github.com/sahilm/fuzzy` for matching
- Returns filtered slice maintaining TMUX order

**`dialog.go`**
- `Dialog` struct for confirmation/input dialogs
- Factory functions: `NewConfirmDialog()`, `NewInputDialog()`
- Simple rendering with bordered display

## Data Flow

### Session Browse → Window Browse → Switch

```
Model.handleKeyPress("enter")
  └─> SessionCard.GetFocused() 
      └─> Model.loadWindows(sessionName)
          └─> Client.ListWindows()
              └─> Model.currentMode = ModeWindowGrid
                  └─> View() renders WindowGrid
                      
User presses enter on window
  └─> Model.handleKeyPress("enter")
      └─> WindowCard.GetFocused()
          └─> Client.SwitchClient(sessionName, windowIndex)
              └─> return tea.Quit  (exit app, switched)
```

### Navigation & Preview Update

```
Model.handleKeyPress("j")  // down
  └─> sessionGrid.MoveFocus(0, 1)
      └─> sessionGrid.GetFocused() → SessionCard
          └─> previewPanel.SetSessionMetadata(card.session)
              └─> View() re-renders with updated preview
```

## Rendering Layout (60/40 Split)

```
┌─────────────────────────────────────────────────────────────────┐
│ TMUX Sessions                                                   │ Header
├─────────────────────────────────────────┬───────────────────────┤
│                                         │                       │
│  [Grid of Session Cards]                │  [Preview Panel]      │ Main Area
│  (60% width)                            │  (40% width)          │
│                                         │                       │
├─────────────────────────────────────────┴───────────────────────┤
│ j/k:nav  enter:select  space:quick  tab:toggle  ?:help  q:quit │ Status Bar
└─────────────────────────────────────────────────────────────────┘
```

Implemented using Lipgloss `JoinHorizontal()` and `JoinVertical()`.

## State Machine

```
ModeSessionGrid
  ├─> Enter
  │   └─> Load windows → ModeWindowGrid
  ├─> Space
  │   └─> Quick switch → Quit
  └─> Esc → Quit

ModeWindowGrid
  ├─> Enter
  │   └─> Switch client → Quit
  ├─> Esc
  │   └─> ModeSessionGrid
  └─> Navigation
      └─> Update preview
```

## Future Expansion Points

1. **Pane-level detail** — Extend to `ModePaneGrid` for viewing panes within a window
2. **Dialogs** — Integrate `Dialog` component for create/rename/kill operations
3. **Tagging** — Use `config.Tags` to filter grid display
4. **Capture preview** — Call `Client.CapturePane()` in preview when mode="capture"
5. **Search/filter** — Integrate `filter.go` functions to apply to grid items
6. **Custom keybindings** — Load from config file
7. **Themes** — Multiple color schemes via `styles.go`

## Dependencies

- **bubbletea** — TUI framework and event loop
- **bubbles** — Pre-built components (not yet used, but available)
- **lipgloss** — Terminal styling and layout
- **fuzzy** — Fuzzy matching for search
- **yaml.v3** — Config file parsing

## Testing Strategy

1. **TMUX integration tests** — Mock TMUX commands
2. **Grid layout tests** — Verify column/row calculations
3. **Filter tests** — Ensure fuzzy matching works correctly
4. **End-to-end** — Full flow: list → navigate → switch

Currently manual testing only. Future: add test suite.

## Performance Considerations

- TMUX command execution is sequential (could parallelize)
- Grid rendering is O(n) where n = visible items
- Fuzzy matching is O(n*m) for n sessions and m search term length
- Memory-bounded by session/window count (typically < 1000)

## Error Handling

Currently minimal — errors are set in `Model.lastErr` and displayed (future work).

Proper error handling needed for:
- TMUX unavailable
- Session/window not found
- Invalid operations
- Terminal size edge cases
