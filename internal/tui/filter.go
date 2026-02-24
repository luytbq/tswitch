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
