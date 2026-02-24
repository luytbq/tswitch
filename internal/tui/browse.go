package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/luytbq/tswitch/internal/config"
	"github.com/luytbq/tswitch/internal/tmux"
)

// handleBrowseDirs launches fzf over configured browse directories.
func (m *Model) handleBrowseDirs() (tea.Model, tea.Cmd) {
	dirs := m.appConfig.BrowseDirs
	if len(dirs) == 0 {
		m.setStatusError("no browse directories configured")
		return m, nil
	}

	shellCmd, tmpPath, err := BuildBrowseCommand(m.appConfig)
	if err != nil {
		m.setStatusError(err.Error())
		return m, nil
	}

	c := exec.Command("sh", "-c", shellCmd)
	return m, tea.ExecProcess(c, func(err error) tea.Msg {
		defer os.Remove(tmpPath)

		// fzf exit code 130 = user cancelled (Ctrl-C/Esc), 1 = no match.
		if err != nil {
			return fzfResultMsg{}
		}

		data, readErr := os.ReadFile(tmpPath)
		if readErr != nil {
			return fzfResultMsg{err: readErr}
		}

		selected := strings.TrimSpace(string(data))
		if selected == "" {
			return fzfResultMsg{}
		}
		return fzfResultMsg{path: selected}
	})
}

// handleFzfResult processes the result from the fzf directory browser.
func (m *Model) handleFzfResult(msg fzfResultMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.setStatusError(msg.err.Error())
		return m, nil
	}
	if msg.path == "" {
		// User cancelled or no selection — return to TUI.
		return m, nil
	}

	if err := SwitchOrCreateSession(m.tmux, msg.path); err != nil {
		m.setStatusError(err.Error())
		return m, nil
	}
	return m, tea.Quit
}

// BuildBrowseCommand constructs the shell command that pipes find into fzf,
// writing the selection to a temp file. Returns the command string, temp file
// path, or an error.
func BuildBrowseCommand(appCfg *config.AppConfig) (shellCmd string, tmpPath string, err error) {
	home, _ := os.UserHomeDir()

	// Build -name exclusion flags from browse_exclude patterns.
	var excludeArgs string
	if len(appCfg.BrowseExclude) > 0 {
		var parts []string
		for _, pattern := range appCfg.BrowseExclude {
			parts = append(parts, fmt.Sprintf("-name %q", pattern))
		}
		excludeArgs = fmt.Sprintf("\\( %s \\) -prune -o ", strings.Join(parts, " -o "))
	}

	var findParts []string
	for _, d := range appCfg.BrowseDirs {
		p := d.Path
		if strings.HasPrefix(p, "~/") {
			p = filepath.Join(home, p[2:])
		}
		depth := d.Depth
		if depth < 1 {
			depth = 1
		}
		findParts = append(findParts, fmt.Sprintf("find %q -mindepth 1 -maxdepth %d %s-type d -print", p, depth, excludeArgs))
	}

	tmpFile, tmpErr := os.CreateTemp("", "tswitch-fzf-*")
	if tmpErr != nil {
		return "", "", fmt.Errorf("failed to create temp file: %w", tmpErr)
	}
	tmpPath = tmpFile.Name()
	tmpFile.Close()

	findCmd := strings.Join(findParts, " && ")
	shellCmd = fmt.Sprintf("{ %s; } | fzf --prompt='directory> ' > %q", findCmd, tmpPath)
	return shellCmd, tmpPath, nil
}

// SwitchOrCreateSession creates a new tmux session in the given directory
// (if one doesn't already exist) and switches to it.
func SwitchOrCreateSession(svc tmux.Service, dir string) error {
	name := NormalizeSessionName(dir)
	if name == "" {
		return fmt.Errorf("could not derive session name from path")
	}

	if !svc.HasSession(name) {
		if err := svc.NewSessionInDir(name, dir); err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
	}

	if err := svc.SwitchToSession(name); err != nil {
		return fmt.Errorf("failed to switch: %w", err)
	}
	return nil
}

// NormalizeSessionName derives a valid tmux session name from a directory path.
func NormalizeSessionName(path string) string {
	name := filepath.Base(path)
	name = strings.ReplaceAll(name, ".", "-")
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.TrimLeft(name, "-")
	if name == "" {
		// Fallback: use parent-child.
		parent := filepath.Base(filepath.Dir(path))
		child := filepath.Base(path)
		name = parent + "-" + child
		name = strings.ReplaceAll(name, ".", "-")
		name = strings.ReplaceAll(name, " ", "-")
		name = strings.TrimLeft(name, "-")
	}
	return name
}
