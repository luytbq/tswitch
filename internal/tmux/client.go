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
	exec   Executor
	inTmux bool
}

// NewClient creates a Client that shells out to the real tmux binary.
func NewClient() *Client {
	return &Client{
		exec:   &shellExecutor{},
		inTmux: os.Getenv("TMUX") != "",
	}
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
	target := fmt.Sprintf("%s:%d.%d", sessionName, windowIndex, paneIndex)
	return c.exec.Run("capture-pane", "-t", target, "-p")
}

// ---------------------------------------------------------------------------
// Navigation
// ---------------------------------------------------------------------------

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
	_, err := c.exec.Run("kill-session", "-t", sessionName)
	return err
}

// ---------------------------------------------------------------------------
// Window management
// ---------------------------------------------------------------------------

func (c *Client) NewWindow(sessionName string, windowName string) error {
	_, err := c.exec.Run("new-window", "-t", sessionName, "-n", windowName)
	return err
}

func (c *Client) RenameWindow(sessionName string, windowIndex int, newName string) error {
	target := fmt.Sprintf("%s:%d", sessionName, windowIndex)
	_, err := c.exec.Run("rename-window", "-t", target, newName)
	return err
}

func (c *Client) KillWindow(sessionName string, windowIndex int) error {
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
	fmt.Sscanf(parts[1], "%d", &windowCount)

	var width, height int
	fmt.Sscanf(parts[5], "%d", &width)
	fmt.Sscanf(parts[6], "%d", &height)

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
	fmt.Sscanf(parts[0], "%d", &index)
	fmt.Sscanf(parts[2], "%d", &paneCount)

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
	fmt.Sscanf(parts[0], "%d", &index)
	fmt.Sscanf(parts[2], "%d", &width)
	fmt.Sscanf(parts[3], "%d", &height)

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
	fmt.Sscanf(s, "%d", &unix)
	if unix > 0 {
		return time.Unix(unix, 0)
	}
	return time.Now()
}
