package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// Compile-time check: Client must satisfy Service.
var _ Service = (*Client)(nil)

// Executor abstracts command execution so the client can be tested
// without shelling out to a real tmux binary.
type Executor interface {
	Run(args ...string) (string, error)
}

// shellExecutor runs real tmux commands.
type shellExecutor struct{}

func (e *shellExecutor) Run(args ...string) (string, error) {
	cmd := exec.Command("tmux", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tmux command failed: %w, output: %s", err, string(output))
	}
	return string(output), nil
}

// Client wraps tmux commands and implements Service.
type Client struct {
	exec           Executor
	inTmux         bool
	currentSession string // session tswitch is running in (empty if not in tmux)
}

// NewClient creates a Client that shells out to the real tmux binary.
func NewClient() *Client {
	c := &Client{
		exec:   &shellExecutor{},
		inTmux: os.Getenv("TMUX") != "",
	}
	if c.inTmux {
		if out, err := c.exec.Run("display-message", "-p", "#{session_name}"); err == nil {
			c.currentSession = strings.TrimSpace(out)
		}
	}
	return c
}

// NewClientWith creates a Client using the given Executor (useful for tests).
func NewClientWith(exec Executor) *Client {
	return &Client{
		exec:   exec,
		inTmux: os.Getenv("TMUX") != "",
	}
}

// IsInTmux returns true if running inside a tmux session.
func (c *Client) IsInTmux() bool {
	return c.inTmux
}

// ---------------------------------------------------------------------------
// Queries
// ---------------------------------------------------------------------------

func (c *Client) ListSessions() ([]Session, error) {
	output, err := c.exec.Run("list-sessions", "-F",
		"#{session_name}|#{session_windows}|#{session_attached}|#{session_created}|#{session_last_attached}|#{session_width}|#{session_height}")
	if err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}

	var sessions []Session
	for _, line := range splitLines(output) {
		s, err := parseSessionLine(line)
		if err != nil {
			continue
		}
		sessions = append(sessions, s)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastActive.After(sessions[j].LastActive)
	})
	return sessions, nil
}

func (c *Client) ListWindows(sessionName string) ([]Window, error) {
	output, err := c.exec.Run("list-windows", "-t", sessionName, "-F",
		"#{window_index}|#{window_name}|#{window_panes}|#{window_active}|#{window_layout}|#{window_path}")
	if err != nil {
		return nil, fmt.Errorf("failed to list windows in session %s: %w", sessionName, err)
	}

	var windows []Window
	for _, line := range splitLines(output) {
		w, err := parseWindowLine(line)
		if err != nil {
			continue
		}
		windows = append(windows, w)
	}
	return windows, nil
}

// ListAllWindowNames returns a map of session name -> window names by querying
// all sessions at once with list-windows -a.
func (c *Client) ListAllWindowNames() (map[string][]string, error) {
	output, err := c.exec.Run("list-windows", "-a", "-F", "#{session_name}|#{window_name}")
	if err != nil {
		return nil, fmt.Errorf("failed to list all windows: %w", err)
	}

	result := make(map[string][]string)
	for _, line := range splitLines(output) {
		parts := strings.Split(line, "|")
		if len(parts) < 2 {
			continue
		}
		sessName, winName := parts[0], parts[1]
		result[sessName] = append(result[sessName], winName)
	}
	return result, nil
}

func (c *Client) ListPanes(sessionName string, windowIndex int) ([]Pane, error) {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	output, err := c.exec.Run("list-panes", "-t", target, "-F",
		"#{pane_index}|#{pane_active}|#{pane_width}|#{pane_height}|#{pane_current_command}|#{pane_current_path}")
	if err != nil {
		return nil, fmt.Errorf("failed to list panes: %w", err)
	}

	var panes []Pane
	for _, line := range splitLines(output) {
		p, err := parsePaneLine(line)
		if err != nil {
			continue
		}
		panes = append(panes, p)
	}
	return panes, nil
}

func (c *Client) CapturePane(sessionName string, windowIndex int, paneIndex int) (string, error) {
	var target string
	if windowIndex < 0 {
		// Capture the active window/pane of the session.
		target = sessionName
	} else {
		target = fmt.Sprintf("%s:%d.%d", sessionName, windowIndex, paneIndex)
	}
	return c.exec.Run("capture-pane", "-t", target, "-p")
}

// ---------------------------------------------------------------------------
// Navigation
// ---------------------------------------------------------------------------

// SwitchToSession switches to a session without specifying a window,
// letting tmux choose the current/last-active window automatically.
func (c *Client) SwitchToSession(sessionName string) error {
	_, err := c.exec.Run("switch-client", "-t", sessionName)
	return err
}

// SwitchToLast switches to the last (previously active) session.
func (c *Client) SwitchToLast() error {
	_, err := c.exec.Run("switch-client", "-l")
	return err
}

func (c *Client) SwitchClient(sessionName string, windowIndex int) error {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	_, err := c.exec.Run("switch-client", "-t", target)
	return err
}

func (c *Client) AttachSession(sessionName string) error {
	_, err := c.exec.Run("attach-session", "-t", sessionName)
	return err
}

// ---------------------------------------------------------------------------
// Session management
// ---------------------------------------------------------------------------

func (c *Client) NewSession(sessionName string) error {
	_, err := c.exec.Run("new-session", "-d", "-s", sessionName)
	return err
}

func (c *Client) RenameSession(oldName, newName string) error {
	_, err := c.exec.Run("rename-session", "-t", oldName, newName)
	return err
}

func (c *Client) KillSession(sessionName string) error {
	// If we're running inside the session being killed, switch to another
	// session first so tswitch isn't terminated mid-execution.
	if c.currentSession == sessionName {
		c.exec.Run("switch-client", "-n") // ignore error — best effort
	}
	_, err := c.exec.Run("kill-session", "-t", sessionName)
	return err
}

// ---------------------------------------------------------------------------
// Window management
// ---------------------------------------------------------------------------

func (c *Client) NewWindow(sessionName string, windowName string) error {
	home, _ := os.UserHomeDir()
	_, err := c.exec.Run("new-window", "-t", sessionName, "-n", windowName, "-c", home)
	return err
}

func (c *Client) RenameWindow(sessionName string, windowIndex int, newName string) error {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	_, err := c.exec.Run("rename-window", "-t", target, newName)
	return err
}

func (c *Client) KillWindow(sessionName string, windowIndex int) error {
	// If we're running in the same session, switch focus away first to avoid
	// tswitch being terminated if it happens to be in this window.
	if c.currentSession == sessionName {
		c.exec.Run("switch-client", "-n") // ignore error — best effort
	}
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	_, err := c.exec.Run("kill-window", "-t", target)
	return err
}

// ---------------------------------------------------------------------------
// Parsing helpers
// ---------------------------------------------------------------------------

func splitLines(output string) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

func parseSessionLine(line string) (Session, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 7 {
		return Session{}, fmt.Errorf("invalid session line: need 7 fields, got %d", len(parts))
	}

	var windowCount int
	if _, err := fmt.Sscanf(parts[1], "%d", &windowCount); err != nil {
		return Session{}, fmt.Errorf("invalid window count %q: %w", parts[1], err)
	}

	var width, height int
	fmt.Sscanf(parts[5], "%d", &width)  // non-critical, zero is acceptable
	fmt.Sscanf(parts[6], "%d", &height) // non-critical, zero is acceptable

	return Session{
		Name:        parts[0],
		WindowCount: windowCount,
		Attached:    parts[2] == "1",
		Created:     parseUnixTime(parts[3]),
		LastActive:  parseUnixTime(parts[4]),
		Width:       width,
		Height:      height,
	}, nil
}

func parseWindowLine(line string) (Window, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 6 {
		return Window{}, fmt.Errorf("invalid window line: need 6 fields, got %d", len(parts))
	}

	var index, paneCount int
	if _, err := fmt.Sscanf(parts[0], "%d", &index); err != nil {
		return Window{}, fmt.Errorf("invalid window index %q: %w", parts[0], err)
	}
	fmt.Sscanf(parts[2], "%d", &paneCount) // zero is acceptable fallback

	return Window{
		Index:      index,
		Name:       parts[1],
		PaneCount:  paneCount,
		Active:     parts[3] == "1",
		Layout:     parts[4],
		WorkingDir: parts[5],
	}, nil
}

func parsePaneLine(line string) (Pane, error) {
	parts := strings.Split(line, "|")
	if len(parts) < 6 {
		return Pane{}, fmt.Errorf("invalid pane line: need 6 fields, got %d", len(parts))
	}

	var index, width, height int
	if _, err := fmt.Sscanf(parts[0], "%d", &index); err != nil {
		return Pane{}, fmt.Errorf("invalid pane index %q: %w", parts[0], err)
	}
	fmt.Sscanf(parts[2], "%d", &width)  // zero is acceptable fallback
	fmt.Sscanf(parts[3], "%d", &height) // zero is acceptable fallback

	return Pane{
		Index:      index,
		Active:     parts[1] == "1",
		Width:      width,
		Height:     height,
		Command:    parts[4],
		WorkingDir: parts[5],
	}, nil
}

func parseUnixTime(s string) time.Time {
	var unix int64
	if _, err := fmt.Sscanf(s, "%d", &unix); err == nil && unix > 0 {
		return time.Unix(unix, 0)
	}
	return time.Time{}
}
