# Mark Display Feature

## Overview

Marked sessions and windows now display their hotkey in the **top-right corner** of their grid cards for easy visual identification.

## Visual Example

### Before (No Mark Display)
```
┌──────────┐
│ work     │
│ 3 wins   │
│ 2m ago   │
└──────────┘
```

### After (With Mark Display)
```
┌──────────────┐
│ work    [w]  │
│ 3 wins       │
│ 2m ago       │
└──────────────┘
```

The `[w]` indicates this session is marked with the `w` key.

## How It Works

1. When tswitch renders the grid, it checks the config for marks
2. For each marked item, it displays the mark key in brackets: `[key]`
3. Both sessions and windows can be marked and displayed
4. Unmarked items show no indicator

## Examples

### Marked Sessions View
```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ work    [w]  │  │ personal [p]  │  │ dev      [d] │
│ 3 wins       │  │ 2 wins        │  │ 1 win        │
│ 2m ago       │  │ 15m ago       │  │ 30m ago      │
└──────────────┘  └──────────────┘  └──────────────┘

┌──────────────┐  ┌──────────────┐
│ staging      │  │ testing  [t]  │
│ 1 win        │  │ 2 wins        │
│ 1h ago       │  │ 45m ago       │
└──────────────┘  └──────────────┘
```

### Marked Windows View
```
┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│0: editor [e] │  │1: terminal[t] │  │2: logs   [l] │
│ 2 panes      │  │ 1 pane        │  │ 1 pane       │
└──────────────┘  └──────────────┘  └──────────────┘
```

## Implementation Details

### Code Changes

**Location:** `internal/tui/grid.go`

- Added `markMap` field to Grid struct: `map[string]string`
- New method: `SetMarks(markMap)` — Update marks for display
- Updated `renderCard()` — Include mark indicator in output

**Location:** `internal/tui/model.go`

- New method: `buildMarkMap(isSession)` — Build mapping of items to mark keys
- New method: `getWindowName()` — Helper to get window name by index
- Updated `renderSessionView()` — Call `SetMarks()` before rendering
- Updated `renderWindowView()` — Call `SetMarks()` before rendering

### Mark Display Format

- **Session format:** `session_name [key]`
- **Window format:** `window_index: window_name [key]`
- **No mark:** Item name without brackets

### Performance

- Mark lookup is O(1) HashMap access
- Rebuilt only during render (not on every frame)
- No impact on navigation or other operations

## Workflow Example

```
1. Open tswitch
   └─ Session grid shows all sessions

2. Mark "work" with 'w'
   └─ Now displays as: "work [w]"
   └─ Config saved to ~/.config/tswitch/config.yaml

3. Later, open tswitch again
   └─ Previously marked sessions still show: "work [w]"

4. Quick reference
   └─ Immediately see which sessions/windows are marked
   └─ Quickly press the key to jump there

5. Mark another session with 'p'
   └─ Display updates: "personal [p]"
```

## Visual Identification Benefits

1. **Quick Recognition** — Instantly see marked items
2. **Mnemonic Aids** — Remember which key maps to which item
3. **Organization** — Visually group related sessions
4. **Discovery** — Find marked items at a glance

## Technical Details

### Mark Storage

Marks are stored in `~/.config/tswitch/config.yaml`:

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
```

### Display Logic

```go
// In renderCard():
mark, hasMark := g.markMap[name]
if hasMark {
    content = fmt.Sprintf("%s [%s]\n%s", name, mark, metadata)
} else {
    content = fmt.Sprintf("%s\n%s", name, metadata)
}
```

## Customization

The mark display is automatically handled. To customize:

1. **Bracket style** — Edit `renderCard()` in `grid.go`
   - Change `[%s]` to `{%s}`, `<%s>`, or custom format

2. **Position** — Move mark to different location in card
   - Modify format string in `renderCard()`

3. **Color** — Highlight mark with different color
   - Add Lipgloss styling in `renderCard()` or `styles.go`

### Example Customization

```go
// Change bracket style
markIndicator := fmt.Sprintf("→%s←", mark)  // Use arrows instead

// Right-align the mark
content = fmt.Sprintf("%-8s [%s]\n%s", name, mark, metadata)

// Color the mark (requires Lipgloss)
styledMark := lipgloss.NewStyle().Foreground(Color("11")).Render(mark)
content = fmt.Sprintf("%s [%s]\n%s", name, styledMark, metadata)
```

## Future Enhancements

- [ ] Colored marks (e.g., work marks in blue, personal in green)
- [ ] Custom mark symbols (e.g., ⭐ instead of brackets)
- [ ] Mark indicators on session metadata line
- [ ] Mark count display (e.g., "3 marks")
- [ ] Show all marks submenu
- [ ] Keyboard shortcut to jump to next marked item

## Testing

To test the mark display feature:

```bash
# Build
cd /home/luytbq/projects/l/test-ai/tswitch && go build -o tswitch

# Create test sessions
tmux new-session -d -s work
tmux new-session -d -s personal
tmux new-window -t work -n editor

# Run tswitch
./tswitch

# Mark sessions
# Navigate to "work" → press m → press w
# Navigate to "personal" → press m → press p

# View mark display
# Session grid now shows:
#   [work [w]]  [personal [p]]
```

## Notes

- Mark display is real-time and updates when marks are created/deleted
- Works with both session and window marking
- Display persists across app restarts (marks are in config)
- No performance impact on app performance

## Questions?

- Check MARKS_FEATURE.md for marking usage
- Check ARCHITECTURE.md for code structure
- Review grid.go and model.go for implementation
