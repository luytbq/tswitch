# tswitch Quick Reference

## Installation

```bash
cd /path/to/tswitch
go build -o tswitch
```

## Basic Usage

```bash
# Run tswitch
./tswitch

# Or from anywhere after installing
tswitch
```

## Keybindings

### Navigation
| Key | Action |
|-----|--------|
| `j` / `k` / `↓` / `↑` | Move down/up |
| `h` / `l` / `←` / `→` | Move left/right |
| `Enter` | Drill in (sessions) / Switch (windows) |
| `Space` | Quick switch to session |
| `Esc` | Back / Quit |

### Marks ⭐ (NEW!)
| Key | Action |
|-----|--------|
| `m` + key | Mark current session/window |
| key | Jump to marked session/window |

**Examples:**
```
m w    # Mark as 'w'
m p    # Mark as 'p'
w      # Later: jump to marked session
p      # Later: jump to marked session
```

### Other
| Key | Action |
|-----|--------|
| `Tab` | Toggle preview mode |
| `?` | Show help |
| `q` | Quit |

## TMUX Integration

Add to `~/.tmux.conf`:

```tmux
bind-key s display-popup -E -w 80% -h 80% "tswitch"
```

Then press `prefix+s` to open tswitch as a popup.

## Configuration

File: `~/.config/tswitch/config.yaml`

### Example

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
  work: [project-a, project-b]
  personal: [dotfiles]

settings:
  default_preview: metadata
  theme: default
  sort_by: activity
```

## Common Workflows

### Quick Switch Between Sessions

1. Mark sessions: `m` + `w` (work), `p` (personal), etc.
2. Later: Press key to switch instantly

### Navigate to Specific Window

1. Open tswitch
2. Press `Enter` to drill into session
3. Press `m` + key to mark window
4. Later: Press key to jump there

### TMUX Keybinding

```tmux
# In ~/.tmux.conf
bind-key s display-popup -E -w 80% -h 80% "tswitch"
bind-key p send-keys -X copy-pipe "tswitch"  # Alternative
```

## Troubleshooting

### No sessions appear
```bash
tmux list-sessions    # Check if sessions exist
```

### Marks not saving
- Check `~/.config/tswitch/config.yaml` exists
- Ensure directory is writable: `chmod 755 ~/.config/tswitch`

### App crashes
- Check terminal size is at least 80x24
- Try in different terminal emulator

## Tips & Tricks

1. **Use number keys for marks**: `m1`, `m2`, etc. for numeric sessions
2. **Use letters for common sessions**: `mw` (work), `mp` (personal), `md` (dev)
3. **Mark first window of session**: `m` when on session card for quick access
4. **Check marks in config**: `cat ~/.config/tswitch/config.yaml | grep -A 20 marks:`

## Files

- `README.md` — Full documentation
- `MARKS_FEATURE.md` — Marks detailed guide
- `MARKS_IMPLEMENTATION.md` — Technical details
- `ARCHITECTURE.md` — Code structure
- `DEVELOPMENT.md` — Developer guide
- `PROJECT_SUMMARY.md` — Feature overview

## Build & Install

```bash
# Build
cd /home/luytbq/projects/l/test-ai/tswitch
go build -o tswitch

# Install (optional)
sudo mv tswitch /usr/local/bin/

# Verify
which tswitch
tswitch --help  # (Currently doesn't show help, just runs app)
```

## Environment

- **Go**: 1.24+
- **TMUX**: 3.0+ (3.2+ for popup)
- **Terminal**: 256-color support recommended
- **OS**: Linux/macOS (tested on Linux)

## Project Status

✅ Fully functional  
✅ Grid layout with marking  
✅ Session/window management ready (keybindings defined)  
⏳ Dialog UI for management operations (coming)  
⏳ Capture preview rendering (coming)  

## Support

Run: `tswitch` → Press `?` for help overlay

Check documentation files for detailed info.

---

**Version**: 1.0 with Marks  
**Last Updated**: Feb 2026  
**Enjoy!** 🚀
