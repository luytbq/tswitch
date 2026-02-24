package tui

import (
	"github.com/sahilm/fuzzy"
	"github.com/user/tswitch/internal/tmux"
)

// FilterSessions returns sessions whose names fuzzy-match term.
func FilterSessions(sessions []tmux.Session, term string) []tmux.Session {
	if term == "" {
		return sessions
	}

	names := make([]string, len(sessions))
	for i, s := range sessions {
		names[i] = s.Name
	}

	var out []tmux.Session
	for _, m := range fuzzy.Find(term, names) {
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
