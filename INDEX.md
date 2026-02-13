# tswitch Documentation Index

## 🚀 Getting Started

Start here if you're new to tswitch:

1. **[SETUP.md](SETUP.md)** — Step-by-step installation and TMUX integration
2. **[QUICKREF.md](QUICKREF.md)** — Quick reference of keybindings and usage
3. **[README.md](README.md)** — Full feature overview and usage guide

## 📖 Documentation

### For Users

- **[SETUP.md](SETUP.md)** — Installation and initial configuration
- **[QUICKREF.md](QUICKREF.md)** — Keybindings and quick reference
- **[TMUX_INTEGRATION.md](TMUX_INTEGRATION.md)** — Detailed TMUX setup guide
  - How to add `bind s display-popup -E -w 80% -h 80% "tswitch"`
  - Multiple setup methods (popup, new-window, split-pane)
  - Troubleshooting and customization
- **[MARKS_FEATURE.md](MARKS_FEATURE.md)** — Bookmarking sessions/windows
  - How to mark with `m` + key
  - Quick switching with marked keys
  - Workflow examples

### For Developers

- **[ARCHITECTURE.md](ARCHITECTURE.md)** — Code structure and design
- **[DEVELOPMENT.md](DEVELOPMENT.md)** — Development guide
- **[MARKS_IMPLEMENTATION.md](MARKS_IMPLEMENTATION.md)** — Technical details of marks feature

### Summaries & Overview

- **[PROJECT_SUMMARY.md](PROJECT_SUMMARY.md)** — What was built and why
- **[MARKS_SUMMARY.md](MARKS_SUMMARY.md)** — Marks feature overview
- **[INDEX.md](INDEX.md)** — This file

## 📝 Example Files

- **[example.tmux.conf](example.tmux.conf)** — Complete TMUX config example
- **[test_marks.sh](test_marks.sh)** — Test script for marks feature

## 🎯 Quick Answers

### How do I install tswitch?

See **[SETUP.md](SETUP.md)** — Step 1: Build from source

### How do I add it to TMUX?

See **[SETUP.md](SETUP.md)** — Step 3: Add to .tmux.conf

The simplest line to add to `~/.tmux.conf`:
```tmux
bind s display-popup -E -w 80% -h 80% "tswitch"
```

Then reload: `tmux source-file ~/.tmux.conf`

### What keybindings does tswitch have?

See **[QUICKREF.md](QUICKREF.md)** — Keybindings section

Quick summary:
- `j/k` — Navigate up/down
- `h/l` — Navigate left/right
- `Enter` — Drill in / Switch
- `m + key` — Mark session/window (new!)
- `key` — Jump to marked session/window (new!)

### How do I mark sessions?

See **[MARKS_FEATURE.md](MARKS_FEATURE.md)**

Quick summary:
1. Open tswitch: `prefix + s` (or just `./tswitch`)
2. Navigate to session
3. Press `m` to enter marking mode
4. Press a key (e.g., `w` for work)
5. Later: press `w` to jump to that session

### What are all the TMUX setup options?

See **[TMUX_INTEGRATION.md](TMUX_INTEGRATION.md)**

Methods:
1. **Popup** (recommended) — `display-popup` (TMUX 3.2+)
2. **New window** — `new-window` (older TMUX)
3. **Split pane** — `split-window`
4. **Full screen** — Fullscreen pane

### I'm getting an error. What do I do?

1. Check **[SETUP.md](SETUP.md)** — Troubleshooting section
2. Check **[TMUX_INTEGRATION.md](TMUX_INTEGRATION.md)** — Troubleshooting section
3. Run: `tmux list-keys | grep tswitch` to verify binding exists
4. Run: `which tswitch` to verify it's in PATH

### How does the marks system work?

See **[MARKS_FEATURE.md](MARKS_FEATURE.md)** for user guide
See **[MARKS_IMPLEMENTATION.md](MARKS_IMPLEMENTATION.md)** for technical details

Quick: You mark sessions/windows with `m + key`, then press that key to switch.

## 📊 File Organization

```
tswitch/
├── SOURCE CODE
│   ├── main.go
│   ├── go.mod / go.sum
│   └── internal/
│       ├── tmux/
│       ├── config/
│       └── tui/
│
├── DOCUMENTATION (YOU ARE HERE)
│   ├── INDEX.md (this file)
│   ├── README.md (overview)
│   ├── SETUP.md (installation)
│   ├── QUICKREF.md (quick reference)
│   ├── TMUX_INTEGRATION.md (detailed setup)
│   ├── MARKS_FEATURE.md (user guide for marks)
│   ├── MARKS_IMPLEMENTATION.md (technical)
│   ├── MARKS_SUMMARY.md (overview)
│   ├── ARCHITECTURE.md (code structure)
│   ├── DEVELOPMENT.md (dev guide)
│   ├── PROJECT_SUMMARY.md (project overview)
│   └── INDEX.md (this file)
│
├── EXAMPLES
│   ├── example.tmux.conf
│   └── test_marks.sh
│
└── BINARY
    └── tswitch (compiled executable)
```

## 🔍 Documentation Map

| Document | Audience | Purpose |
|----------|----------|---------|
| SETUP.md | End users | Installation & setup |
| QUICKREF.md | End users | Fast lookup of keybindings |
| README.md | Everyone | Feature overview |
| TMUX_INTEGRATION.md | TMUX users | Detailed integration guide |
| MARKS_FEATURE.md | End users | How to use marks |
| MARKS_IMPLEMENTATION.md | Developers | Marks technical details |
| ARCHITECTURE.md | Developers | Code structure |
| DEVELOPMENT.md | Developers | Contributing guide |
| PROJECT_SUMMARY.md | Everyone | What was built |
| MARKS_SUMMARY.md | Everyone | Marks feature summary |
| example.tmux.conf | TMUX users | Complete config example |

## 🎓 Learning Path

### New User
1. Read: SETUP.md
2. Build: `go build -o tswitch`
3. Try: `./tswitch`
4. Test: Add to TMUX and press `prefix + s`
5. Explore: Try keybindings from QUICKREF.md

### Want to Use Marks?
1. Read: MARKS_FEATURE.md
2. Try: Open tswitch and press `m` + `w`
3. Use: Press `w` to jump to marked session

### Power User
1. Study: TMUX_INTEGRATION.md for advanced setup
2. Configure: Customize popup size and position
3. Optimize: Set up marks for your workflow

### Developer
1. Read: ARCHITECTURE.md
2. Study: DEVELOPMENT.md
3. Explore: Review code in `internal/`
4. Extend: Check roadmap for next features

## 🆘 Getting Help

### Common Questions

**Q: How do I add tswitch to TMUX?**
A: Edit `~/.tmux.conf` and add:
```tmux
bind s display-popup -E -w 80% -h 80% "tswitch"
```

**Q: What keys can I use for marks?**
A: Any key except: h, j, k, l, m, enter, space, esc, tab, ?, q

**Q: Can I customize the popup size?**
A: Yes! Change `-w 80% -h 80%` to `-w 60% -h 60%` or fixed size like `-w 120 -h 40`

**Q: Does tswitch work outside TMUX?**
A: Yes! It will show sessions and allow attaching instead of switching.

### Still Stuck?

1. Check the appropriate documentation file from the table above
2. Search for your issue in the docs (Ctrl+F)
3. Try the troubleshooting sections:
   - SETUP.md — Common setup issues
   - TMUX_INTEGRATION.md — TMUX-specific issues
4. Review example.tmux.conf for working configuration

## 📋 Checklists

### Installation Checklist
- [ ] Go 1.24+ installed (`go version`)
- [ ] TMUX 3.0+ installed (`tmux -V`)
- [ ] tswitch built (`go build -o tswitch`)
- [ ] Added to .tmux.conf
- [ ] TMUX reloaded (`tmux source-file ~/.tmux.conf`)
- [ ] Tested with `prefix + s`

### Marks Setup Checklist
- [ ] tswitch is working
- [ ] Created test sessions (`tmux new-session -d -s work`)
- [ ] Opened tswitch (`./tswitch` or `prefix + s`)
- [ ] Marked a session (`m` + `w`)
- [ ] Jumped to mark (press `w`)
- [ ] Checked config file (`cat ~/.config/tswitch/config.yaml`)

### TMUX Integration Checklist
- [ ] Updated `~/.tmux.conf`
- [ ] Verified syntax (`tmux source-file ~/.tmux.conf`)
- [ ] Tested binding (`prefix + s`)
- [ ] Customized if needed (size, position, etc.)
- [ ] Set up additional marks (optional)

## 🎁 Bonus Content

### Example Workflows

**Fast Development Setup:**
```
1. Mark sessions: tswitch → m w (work), m a (api), m d (db)
2. In TMUX: just press w, a, or d to switch
```

**Multi-Window Navigation:**
```
1. Mark windows in session: m e (editor), m t (terminal)
2. Jump between windows instantly
```

**Window Management:**
```
1. Session view: navigate with hjkl
2. Drill in: enter
3. Window view: navigate and mark
4. Switch: enter or marked key
```

## 📞 Support Resources

- **In-app help**: Open tswitch and press `?`
- **Command help**: `tmux list-keys | grep tswitch`
- **Config validation**: `tmux source-file ~/.tmux.conf`
- **Marks debug**: `cat ~/.config/tswitch/config.yaml | grep marks:`

## ✨ What's Next?

After setup, explore:

- [ ] Set up marks for your favorite sessions
- [ ] Customize popup size in .tmux.conf
- [ ] Create TMUX keybinds for marked sessions
- [ ] Share your config with teammates
- [ ] Check out the code (it's approachable!)

---

**Last Updated**: Feb 2026
**Version**: 1.0 with Marks Feature
**Status**: ✅ Fully Functional

**Enjoy using tswitch!** 🚀

