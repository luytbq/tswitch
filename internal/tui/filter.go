package tui

import (
	"github.com/sahilm/fuzzy"
	"github.com/luytbq/tswitch/internal/tmux"
)

// FilterSessions returns sessions whose names, or any of their window names,
// fuzzy-match term. windowsBySession may be nil (session-name-only match).
func FilterSessions(sessions []tmux.Session, term string, windowsBySession map[string][]string) []tmux.Session {
	if term == "" {
		return sessions
	}

	// Build search strings: "sessionName windowName1 windowName2 ..."
	// Fuzzy matching against the combined string lets a query like "vim"
	// surface sessions that have a window named "vim".
	searchStrings := make([]string, len(sessions))
	for i, s := range sessions {
		combined := s.Name
		for _, wn := range windowsBySession[s.Name] {
			combined += " " + wn
		}
		searchStrings[i] = combined
	}

	var out []tmux.Session
	for _, m := range fuzzy.Find(term, searchStrings) {
		out = append(out, sessions[m.Index])
	}
	return out
}

// FilterPanes returns panes whose command+dir fuzzy-match term.
func FilterPanes(panes []tmux.Pane, term string) []tmux.Pane {
	if term == "" {
		return panes
	}

	searchStrings := make([]string, len(panes))
	for i, p := range panes {
		searchStrings[i] = p.Command + " " + p.WorkingDir
	}

	var out []tmux.Pane
	for _, m := range fuzzy.Find(term, searchStrings) {
		out = append(out, panes[m.Index])
	}
	return out
}

// FilterWindows returns windows whose names fuzzy-match term.
func FilterWindows(windows []tmux.Window, term string) []tmux.Window {
	if term == "" {
		return windows
	}

	names := make([]string, len(windows))
	for i, w := range windows {
		names[i] = w.Name
	}

	var out []tmux.Window
	for _, m := range fuzzy.Find(term, names) {
		out = append(out, windows[m.Index])
	}
	return out
}

// FilterTree narrows the session list for the tree view. A session is kept when
// its own name matches, or when one of its window names does. In the second
// case the returned map lists which windows hit, so the tree can show only
// those; a session matched by its own name has no entry and keeps all windows.
func FilterTree(sessions []tmux.Session, windows map[string][]tmux.Window, term string) ([]tmux.Session, map[string]map[int]bool) {
	if term == "" {
		return sessions, nil
	}

	names := make([]string, len(sessions))
	for i, s := range sessions {
		names[i] = s.Name
	}
	sessionHit := make(map[string]bool)
	for _, match := range fuzzy.Find(term, names) {
		sessionHit[sessions[match.Index].Name] = true
	}

	windowHit := make(map[string]map[int]bool)
	for _, s := range sessions {
		if sessionHit[s.Name] {
			continue
		}
		sessionWindows := windows[s.Name]
		windowNames := make([]string, len(sessionWindows))
		for i, w := range sessionWindows {
			windowNames[i] = w.Name
		}
		for _, match := range fuzzy.Find(term, windowNames) {
			if windowHit[s.Name] == nil {
				windowHit[s.Name] = make(map[int]bool)
			}
			windowHit[s.Name][sessionWindows[match.Index].Index] = true
		}
	}

	// Keep the display order rather than the fuzzy score, so rows do not jump
	// around under the cursor as the query is typed.
	out := make([]tmux.Session, 0, len(sessions))
	for _, s := range sessions {
		if sessionHit[s.Name] || len(windowHit[s.Name]) > 0 {
			out = append(out, s)
		}
	}
	return out, windowHit
}
