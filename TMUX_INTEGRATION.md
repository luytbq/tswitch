# TMUX Integration Guide

## Quick Setup

Add this line to your `~/.tmux.conf`:

```tmux
bind-key s display-popup -E -w 80% -h 80% "tswitch"
```

Then reload TMUX:

```bash
tmux source-file ~/.tmux.conf
```

Now press `prefix+s` to open tswitch.

## Setup Methods

### Method 1: Popup (Recommended for TMUX 3.2+)

```tmux
# ~/.tmux.conf
bind-key s display-popup -E -w 80% -h 80% "tswitch"
```

**Pros:**
- Appears as overlay, doesn't replace pane
- Clean, non-intrusive
- Automatically closes after switching

**Cons:**
- Requires TMUX 3.2+
- Needs terminal with popup support

### Method 2: New Window

```tmux
# ~/.tmux.conf
bind-key s new-window -n tswitch "tswitch"
```

**Pros:**
- Works with older TMUX versions
- Can keep window open for multiple switches

**Cons:**
- Creates a window (clutter)
- Need to close window manually

### Method 3: Split Pane

```tmux
# ~/.tmux.conf
bind-key s split-window -h "tswitch"
```

**Pros:**
- Easy horizontal split
- Side-by-side view

**Cons:**
- Splits current pane
- Less clean UX

### Method 4: Full Screen Pane

```tmux
# ~/.tmux.conf
bind-key s split-window -h -f -c "#{pane_current_path}" "tswitch"
```

**Pros:**
- Opens in new pane in current directory
- Maintains context

**Cons:**
- Requires manual navigation

## Full Configuration Example

Here's a complete `.tmux.conf` setup with tswitch:

```tmux
# ~/.tmux.conf

# Set prefix to Ctrl+a (or keep Ctrl+b)
set -g prefix C-a
unbind C-b
bind C-a send-prefix

# Enable 256 colors
set -g default-terminal "screen-256color"

# Windows and panes keybindings
bind -n M-h select-window -t :-
bind -n M-l select-window -t :+
bind -n C-Left select-window -t :-
bind -n C-Right select-window -t :+

# === tswitch Integration ===

# Method 1: Popup overlay (TMUX 3.2+)
bind s display-popup -E -w 80% -h 80% "tswitch"

# OR Method 2: New window
# bind s new-window -n tswitch "tswitch"

# OR Method 3: Split pane
# bind s split-window -h "tswitch"

# Quick mark switching (assuming marks are set up)
# Jump to marked sessions directly
bind w send-keys -X copy-mode "tswitch w"
bind p send-keys -X copy-mode "tswitch p"
```

## Choosing Your Prefix Key

Common options:

```tmux
# Default
set -g prefix C-b

# Ctrl+a (GNU Screen style)
set -g prefix C-a
unbind C-b
bind C-a send-prefix

# Ctrl+Space
set -g prefix C-Space
unbind C-b

# Ctrl+x
set -g prefix C-x
unbind C-b
```

## Full vs Relative Keys

```tmux
# Full specification
bind-key -n prefix+s display-popup -E -w 80% -h 80% "tswitch"

# Short form (bind-key is optional)
bind s display-popup -E -w 80% -h 80% "tswitch"

# With -n flag (no prefix needed, global hotkey)
bind -n M-s display-popup -E -w 80% -h 80% "tswitch"

# With -c flag (run in specific directory)
bind s display-popup -E -w 80% -h 80% -c "#{pane_current_path}" "tswitch"
```

## Popup Customization

```tmux
# Size options
display-popup -w 80% -h 80% "tswitch"    # 80% of terminal
display-popup -w 120 -h 40 "tswitch"     # 120 cols × 40 lines
display-popup -w 0 -h 0 "tswitch"        # Full screen

# Position options
display-popup -x 10 -y 5 "tswitch"       # 10 cols from left, 5 from top
display-popup -C "tswitch"               # Centered (default)

# Flags
-E                                        # Close popup on exit
-B                                        # Draw border
```

### Full Popup Example

```tmux
# Fancy popup with border, 85% size, centered
bind s display-popup -B -w 85% -h 85% -E "tswitch"

# Smaller popup, positioned at top-right
bind s display-popup -w 60% -h 60% -x 40% -y 5 -E "tswitch"

# Full screen without border
bind s display-popup -w 100% -h 100% -E "tswitch"
```

## Advanced: Conditional Integration

### Use Different Commands Based on TMUX Version

```tmux
# Check TMUX version and use appropriate command
if-shell "tmux -V | grep -q 3.2" \
  "bind s display-popup -E -w 80% -h 80% 'tswitch'" \
  "bind s new-window -n tswitch 'tswitch'"
```

### Change Keybinding Per OS

```tmux
# Linux: use Ctrl+s
if-shell "uname | grep -q Linux" \
  "bind C-s display-popup -E -w 80% -h 80% 'tswitch'"

# macOS: use Ctrl+t (avoid terminal suspend)
if-shell "uname | grep -q Darwin" \
  "bind C-t display-popup -E -w 80% -h 80% 'tswitch'"
```

## Troubleshooting

### Keybinding not working

1. **Check syntax**: `tmux list-keys | grep tswitch`
2. **Reload config**: `tmux source-file ~/.tmux.conf`
3. **Check conflicts**: See if key is already bound

### Popup doesn't appear

1. **Check TMUX version**: `tmux -V` (need 3.2+)
2. **Try new-window instead**: `bind s new-window 'tswitch'`
3. **Check terminal support**: Some terminals don't support popups

### tswitch doesn't find sessions

1. **Ensure tswitch is in PATH**: `which tswitch`
2. **Use full path**: `bind s display-popup -E 'tswitch'`
3. **Check TMUX sessions**: `tmux list-sessions`

### tswitch closes immediately

1. **Add -E flag**: `display-popup -E ... "tswitch"` (closes popup on exit)
2. **Check for errors**: `tswitch 2>&1 | tee /tmp/tswitch.log`

## Tips & Best Practices

### 1. Use Consistent Keybinding

```tmux
# Good: mnemonic key
bind s display-popup -E -w 80% -h 80% "tswitch"   # 's' = switch

# Also good: function key
bind F12 display-popup -E -w 80% -h 80% "tswitch"
```

### 2. Avoid Conflicts

```tmux
# Check existing bindings
tmux list-keys | grep -E "C-s|M-s|F12"

# Don't use:
bind C-b          # Default prefix
bind c            # New window
bind ,            # Rename window
bind x            # Kill pane
```

### 3. Add Comment Documenting Hotkey

```tmux
# Navigate TMUX sessions with tswitch (prefix + s)
bind s display-popup -E -w 80% -h 80% "tswitch"
```

### 4. Use Marks for Quick Navigation

```tmux
# After setting up marks in tswitch config:
# Mark sessions: tswitch → m + key

# Quick access to marked sessions
bind W send-keys "tswitch && w" Enter    # Open tswitch, then jump to 'w' mark
bind P send-keys "tswitch && p" Enter    # Open tswitch, then jump to 'p' mark
```

## Complete Minimal `.tmux.conf`

```tmux
# Minimal TMUX config with tswitch

# Set prefix
set -g prefix C-a
unbind C-b
bind C-a send-prefix

# Enable 256 colors
set -g default-terminal "screen-256color"

# tswitch integration
bind s display-popup -E -w 80% -h 80% "tswitch"

# Basic keybinds
bind r source-file ~/.tmux.conf
bind c new-window
bind | split-window -h
bind - split-window -v
```

## Using with Marks

Once marks are configured in tswitch, you can create quick switches:

```tmux
# ~/.tmux.conf

# Open tswitch
bind s display-popup -E -w 80% -h 80% "tswitch"

# Direct mark switches (if using marks w, p, d)
bind W run-shell "tmux send-keys -X copy-mode 'tswitch w'"
bind P run-shell "tmux send-keys -X copy-mode 'tswitch p'"
bind D run-shell "tmux send-keys -X copy-mode 'tswitch d'"
```

## Reload Without Killing Sessions

```bash
# Reload tmux config without restarting sessions
tmux source-file ~/.tmux.conf

# Or use alias
alias tmux-reload='tmux source-file ~/.tmux.conf'
```

## References

- TMUX man page: `man tmux`
- Display-popup docs: Look for `display-popup` in TMUX 3.2+ docs
- tswitch docs: See README.md and MARKS_FEATURE.md

---

**Note**: TMUX 3.2+ is recommended for popup support. For older versions, use `new-window` or `split-window` instead.
