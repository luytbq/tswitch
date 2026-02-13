# Quick Setup Guide for tswitch

## Step 1: Build tswitch

```bash
cd /path/to/tswitch
go build -o tswitch
```

## Step 2: Install (Optional)

```bash
# Copy to system path
sudo cp tswitch /usr/local/bin/

# Or add to PATH
export PATH="/path/to/tswitch:$PATH"
```

## Step 3: Add to .tmux.conf

### Quick Version (Copy-Paste)

Open your `~/.tmux.conf` and add this line:

```tmux
bind s display-popup -E -w 80% -h 80% "tswitch"
```

### Step-by-Step

1. **Open your tmux config**:
   ```bash
   vim ~/.tmux.conf
   # or
   nano ~/.tmux.conf
   ```

2. **Add the tswitch binding**:
   ```tmux
   # Navigate sessions with tswitch (prefix + s)
   bind s display-popup -E -w 80% -h 80% "tswitch"
   ```

3. **Save and exit** (vim: `:wq`, nano: Ctrl+X → Y → Enter)

## Step 4: Reload TMUX

Option A - Reload without restarting:
```bash
tmux source-file ~/.tmux.conf
```

Option B - Restart TMUX server:
```bash
tmux kill-server
# Start new session
tmux new-session -s mysession
```

## Step 5: Test It!

1. **Create test sessions**:
   ```bash
   tmux new-session -d -s work
   tmux new-session -d -s personal
   ```

2. **Open TMUX**:
   ```bash
   tmux attach
   ```

3. **Press: `prefix + s`** (e.g., Ctrl+B + S if using default prefix)
   - tswitch should open as a popup overlay
   - Navigate with `j/k` or arrows
   - Press `Enter` to switch

## Step 6: Set Up Marks (Optional)

Once tswitch is working, you can add bookmarks:

1. **Open tswitch**: Press `prefix + s`
2. **Mark a session**: 
   - Navigate to "work" session
   - Press `m` (enter marking mode)
   - Press `w` (assign mark)
   - See: "Marked 'work' → w"
3. **Later - quickly switch**:
   - Open tswitch: `prefix + s`
   - Press `w` to jump directly to work

## Troubleshooting

### "Command not found: tswitch"

**Solution**: Add full path to .tmux.conf:
```tmux
bind s display-popup -E -w 80% -h 80% "/full/path/to/tswitch"
```

Or ensure tswitch is in your PATH:
```bash
export PATH="$PATH:/path/to/tswitch/directory"
```

### Popup doesn't appear

**Check TMUX version**:
```bash
tmux -V
```

**If < 3.2**, use new-window instead:
```tmux
bind s new-window -n tswitch "tswitch"
```

### Key binding not working

**Verify it was added**:
```bash
tmux list-keys | grep tswitch
```

**Reload config**:
```bash
tmux source-file ~/.tmux.conf
```

### tswitch shows "No sessions"

**Check if TMUX sessions exist**:
```bash
tmux list-sessions
```

**Create a test session**:
```bash
tmux new-session -d -s test
```

## Common TMUX Prefix Keys

```tmux
# Default (Ctrl+B)
set -g prefix C-b

# GNU Screen style (Ctrl+A) - recommended
set -g prefix C-a
unbind C-b
bind C-a send-prefix

# Other popular options
set -g prefix C-Space
set -g prefix C-x
set -g prefix M-a  # Alt+A
```

## Full .tmux.conf Example

See `example.tmux.conf` in this directory for a complete configuration.

Copy relevant sections or use as template:
```bash
cat example.tmux.conf >> ~/.tmux.conf
```

## Customize Popup Size

```tmux
# Smaller popup (60%)
bind s display-popup -E -w 60% -h 60% "tswitch"

# Larger popup (90%)
bind s display-popup -E -w 90% -h 90% "tswitch"

# Full screen
bind s display-popup -E -w 100% -h 100% "tswitch"

# Fixed size (120 cols × 40 lines)
bind s display-popup -E -w 120 -h 40 "tswitch"

# With border
bind s display-popup -B -E -w 80% -h 80% "tswitch"
```

## Advanced: Multiple Session Groups

Set up shortcuts to specific sessions:

```tmux
# Quick switch to 'work' session (no menu)
bind W send-keys "tswitch" Enter && sleep 0.2 && send-keys "w"

# Quick switch to 'personal' session
bind P send-keys "tswitch" Enter && sleep 0.2 && send-keys "p"

# Note: This is a workaround; marks feature is better (see MARKS_FEATURE.md)
```

## Verify Installation

```bash
# Check tswitch is in PATH
which tswitch

# Check TMUX config is correct
grep -n "tswitch" ~/.tmux.conf

# Test tswitch directly
tswitch

# Test from TMUX
# (Already in tmux session) Press: prefix + s
```

## Next Steps

1. ✅ Build tswitch
2. ✅ Add to .tmux.conf
3. ✅ Reload TMUX
4. ✅ Test with `prefix + s`
5. 📖 Read MARKS_FEATURE.md for bookmarking
6. 📖 Read QUICKREF.md for keybindings

## Getting Help

Run `tswitch` and press `?` for help overlay.

Or check documentation:
- `README.md` — Full guide
- `QUICKREF.md` — Quick reference
- `TMUX_INTEGRATION.md` — Detailed TMUX setup
- `MARKS_FEATURE.md` — Session bookmarking

---

**You're all set!** 🚀

Press `prefix + s` in TMUX to open tswitch.

