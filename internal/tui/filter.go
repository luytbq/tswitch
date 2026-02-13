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
	byName := make(map[string]tmux.Session, len(sessions))
	for i, s := range sessions {
		names[i] = s.Name
		byName[s.Name] = s
	}

	var out []tmux.Session
	for _, m := range fuzzy.Find(term, names) {
		if s, ok := byName[m.Str]; ok {
			out = append(out, s)
		}
	}
	return out
}

// FilterWindows returns windows whose names fuzzy-match term.
func FilterWindows(windows []tmux.Window, term string) []tmux.Window {
	if term == "" {
		return windows
	}

	names := make([]string, len(windows))
	byName := make(map[string]tmux.Window, len(windows))
	for i, w := range windows {
		names[i] = w.Name
		byName[w.Name] = w
	}

	var out []tmux.Window
	for _, m := range fuzzy.Find(term, names) {
		if w, ok := byName[m.Str]; ok {
			out = append(out, w)
		}
	}
	return out
}
