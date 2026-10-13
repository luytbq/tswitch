package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/luytbq/tswitch/internal/tmux"
)

// rebuildTree regenerates the visible rows from the session list, the current
// filter and the saved expansion state, keeping focus on the same node.
func (m *Model) rebuildTree() {
	sessions := m.sessions
	var windowMatches map[string]map[int]bool
	if m.filterQuery != "" {
		sessions, windowMatches = FilterTree(m.sessions, m.allWindows, m.filterQuery)
	}

	rows := make([]TreeRow, 0, len(sessions))
	for _, s := range sessions {
		matched := windowMatches[s.Name]
		windows := m.windowsOf(s.Name)
		// A session pulled into the results only by a window name is opened
		// so the matching window is actually visible.
		expanded := len(windows) > 0 &&
			(m.config.IsSessionExpanded(s.Name) || len(matched) > 0)

		rows = append(rows, TreeRow{
			Kind:       NodeSession,
			Depth:      0,
			Session:    s.Name,
			Window:     -1,
			Pane:       -1,
			Label:      s.Name,
			Meta:       sessionMeta(s),
			Indicator:  indicator(s.Attached),
			Expandable: len(windows) > 0,
			Expanded:   expanded,
		})
		if !expanded {
			continue
		}

		for _, w := range windows {
			// A session matched by its own name shows all its windows; one
			// pulled in only by a window name shows just the windows that hit.
			if matched != nil && !matched[w.Index] {
				continue
			}
			panes := m.panesOf(s.Name, w.Index)
			// A single-pane window has nothing to reveal, so it stays a leaf.
			expandable := len(panes) > 1
			winExpanded := expandable && m.config.IsWindowExpanded(s.Name, w.Index)

			rows = append(rows, TreeRow{
				Kind:       NodeWindow,
				Depth:      1,
				Session:    s.Name,
				Window:     w.Index,
				Pane:       -1,
				Label:      fmt.Sprintf("%d: %s", w.Index, w.Name),
				Meta:       windowMeta(w),
				Indicator:  indicator(w.Active),
				Expandable: expandable,
				Expanded:   winExpanded,
			})
			if !winExpanded {
				continue
			}

			for _, p := range panes {
				rows = append(rows, TreeRow{
					Kind:      NodePane,
					Depth:     2,
					Session:   s.Name,
					Window:    w.Index,
					Pane:      p.Index,
					Label:     fmt.Sprintf("%d: %s", p.Index, p.Command),
					Meta:      paneMeta(p),
					Indicator: indicator(p.Active),
				})
			}
		}
	}

	var focusKey, focusSession string
	if row, ok := m.tree.Focused(); ok {
		focusKey, focusSession = row.Key(), row.Session
	}
	m.tree.SetRows(rows)
	if focusKey == "" {
		return
	}
	if m.tree.FocusRowWhere(func(r TreeRow) bool { return r.Key() == focusKey }) {
		return
	}
	// The node is gone: filtered out, folded away or killed. Fall back to its
	// session, then to the top, rather than leaving the cursor on whatever
	// row happens to have inherited the old position.
	if m.tree.FocusRowWhere(func(r TreeRow) bool {
		return r.Kind == NodeSession && r.Session == focusSession
	}) {
		return
	}
	m.tree.SetFocusIndex(0)
}

// windowsOf returns a session's windows in display order: the saved manual
// order first, then anything tmux reports that the saved order does not cover.
func (m *Model) windowsOf(session string) []tmux.Window {
	windows := m.allWindows[session]
	if session == m.currentSess && len(m.windows) > 0 {
		windows = m.windows
	}
	order, ok := m.config.WindowOrder[session]
	if !ok || len(order) == 0 {
		return windows
	}

	byIndex := make(map[int]tmux.Window, len(windows))
	for _, w := range windows {
		byIndex[w.Index] = w
	}
	result := make([]tmux.Window, 0, len(windows))
	seen := make(map[int]bool, len(windows))
	for _, idx := range order {
		if w, ok := byIndex[idx]; ok {
			result = append(result, w)
			seen[idx] = true
		}
	}
	for _, w := range windows {
		if !seen[w.Index] {
			result = append(result, w)
		}
	}
	return result
}

// panesOf returns a window's panes ordered by pane index.
func (m *Model) panesOf(session string, windowIndex int) []tmux.Pane {
	panes := m.allPanes[session][windowIndex]
	if len(panes) < 2 {
		return panes
	}
	out := make([]tmux.Pane, len(panes))
	copy(out, panes)
	sort.Slice(out, func(i, j int) bool { return out[i].Index < out[j].Index })
	return out
}

// ---------------------------------------------------------------------------
// Row text
// ---------------------------------------------------------------------------

func sessionMeta(s tmux.Session) string {
	return fmt.Sprintf("%s · %s · %s",
		count(s.WindowCount, "win", "wins"),
		count(s.PaneCount, "pane", "panes"),
		formatTimeSince(s.LastActive))
}

func windowMeta(w tmux.Window) string {
	meta := count(w.PaneCount, "pane", "panes")
	if dir := baseName(w.WorkingDir); dir != "" {
		meta += " · " + dir
	}
	return meta
}

func paneMeta(p tmux.Pane) string {
	meta := fmt.Sprintf("%dx%d", p.Width, p.Height)
	if dir := baseName(p.WorkingDir); dir != "" {
		meta = dir + " · " + meta
	}
	return meta
}

func count(n int, one, many string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, one)
	}
	return fmt.Sprintf("%d %s", n, many)
}

func indicator(on bool) string {
	if on {
		return "●"
	}
	return ""
}

func baseName(dir string) string {
	if idx := strings.LastIndexByte(dir, '/'); idx >= 0 && idx < len(dir)-1 {
		return dir[idx+1:]
	}
	return dir
}
