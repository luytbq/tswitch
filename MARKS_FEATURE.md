# Marks Feature

The marks feature allows you to quickly bookmark sessions and windows for fast switching.

## Overview

Press `m` followed by any non-reserved key to mark the current session or window. Later, pressing just that key will instantly switch to the marked target.

## Usage

### Marking a Session

1. Open tswitch
2. Navigate to the session you want to mark
3. Press `m` (you'll see "Press a key to mark this session (ESC to cancel)" in the status bar)
4. Press any key (e.g., `w` for work, `p` for personal, `1`, `2`, etc.)
5. The mark is saved: "Marked 'work' → w"

### Marking a Window

1. Open tswitch
2. Navigate to a session (press Enter to drill in)
3. Navigate to the window you want to mark
4. Press `m` (you'll see "Press a key to mark this window (ESC to cancel)" in the status bar)
5. Press any key (e.g., `e` for editor, `t` for terminal)
6. The mark is saved: "Marked 'work:0' → e"

### Switching to a Mark

While tswitch is open:
- Press the key you marked with (e.g., `w` for work session)
- tswitch instantly switches to that session/window and closes

### Persistence

Marks are saved in `~/.config/tswitch/config.yaml`:

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
  e:
    session: work
    window: 1
    pane: 0
```

## Reserved Keys

These keys cannot be used as marks (they're already reserved for navigation/actions):

- Navigation: `h`, `j`, `k`, `l`, up/down/left/right, space
- Actions: `enter`, `esc`, `m`, `?`, `q`
- Future: `n` (new), `r` (rename), `x` (kill), `t` (tag), `/` (filter)
- Special: `tab`

## Examples

### Workflow 1: Quick Switch Between Work and Personal

```
tmux new-session -d -s work
tmux new-session -d -s personal

# In tswitch:
# 1. Navigate to 'work' session
# 2. Press m, then w
# 3. Navigate to 'personal' session  
# 4. Press m, then p
```

Then later:
```
tswitch      # Open tswitch
w            # Jump to work session
# (tswitch closes, you're in work)

tswitch      # Open again
p            # Jump to personal session
# (tswitch closes, you're in personal)
```

### Workflow 2: Mark Specific Windows

Mark different windows for easy access:

```
# Session: work
# Window 0: editor
# Window 1: terminal
# Window 2: logs

# In tswitch, drill into work:
# Mark window 0 (editor) as 'e'
# Mark window 1 (terminal) as 't'
# Mark window 2 (logs) as 'l'
```

Later:
```
tswitch    # Open
e          # Jump to editor window
# (in work:0)

tswitch    # Open again
t          # Jump to terminal window
# (in work:1)
```

### Workflow 3: Multi-Session Quick Access

Mark your most-used sessions with single keys:

```
# Mark sessions:
# 'a' → api-server
# 'w' → web-frontend
# 'd' → database
# 's' → staging
```

Then rapidly switch between them by repeatedly pressing `tswitch` then the key.

## Configuration

To manually edit marks, edit `~/.config/tswitch/config.yaml`:

```yaml
marks:
  a:
    session: api-server
    window: 0
    pane: 0
  w:
    session: web-frontend
    window: 0
    pane: 0
```

Delete a mark by removing it from the file and saving.

## Removing Marks

Currently, to remove a mark, edit `~/.config/tswitch/config.yaml` and delete the entry. Future versions may add a UI for this.

## Technical Details

- Marks are stored as `key → (session, window, pane)`
- When switching to a mark, tswitch calls `tmux switch-client -t session:window`
- The mark feature works both inside and outside TMUX
- Marks are global and shared across all TMUX invocations

## Future Enhancements

- [ ] List all marks in a submenu
- [ ] Delete marks via UI (press `m`, then `del` to clear mark)
- [ ] Rename marks
- [ ] Share marks across machines
- [ ] Sync marks with dotfiles
