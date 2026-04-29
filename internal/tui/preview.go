package tui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/luytbq/tswitch/internal/tmux"
)

// PreviewMode is a typed enum for the preview panel display mode.
type PreviewMode int

const (
	PreviewMetadata PreviewMode = iota
	PreviewCapture
)

// PreviewPanel displays metadata or captured content for the focused item.
type PreviewPanel struct {
	width   int
	height  int
	styles  Styles
	mode    PreviewMode
	content string
	title   string
}

// NewPreviewPanel creates a new preview panel.
func NewPreviewPanel(width, height int, styles Styles) *PreviewPanel {
	return &PreviewPanel{
		width:  width,
		height: height,
		styles: styles,
		mode:   PreviewCapture,
		title:  "Preview",
	}
}

// SetSize updates the panel dimensions.
func (pp *PreviewPanel) SetSize(width, height int) {
	pp.width = width
	pp.height = height
}

// IsCapture reports whether the panel is in capture mode.
func (pp *PreviewPanel) IsCapture() bool { return pp.mode == PreviewCapture }

// ToggleMode switches between metadata and capture modes.
func (pp *PreviewPanel) ToggleMode() {
	if pp.mode == PreviewCapture {
		pp.mode = PreviewMetadata
	} else {
		pp.mode = PreviewCapture
	}
}

// SetSessionMetadata populates the panel for a session.
func (pp *PreviewPanel) SetSessionMetadata(session tmux.Session) {
	pp.title = "Session"

	var lines []string
	lines = append(lines, pp.styles.CardTitle.Render(session.Name))
	lines = append(lines, "")

	lines = append(lines, fmt.Sprintf("Windows:     %d", session.WindowCount))
	lines = append(lines, fmt.Sprintf("Created:     %s", formatTime(session.Created)))
	lines = append(lines, fmt.Sprintf("Last Active: %s", formatTime(session.LastActive)))

	if session.ActivePaneCmd != "" || session.ActivePaneDir != "" {
		lines = append(lines, "")
		lines = append(lines, pp.styles.CardSubtle.Render("Active pane:"))
		if session.ActivePaneDir != "" {
			lines = append(lines, fmt.Sprintf("  Dir:       %s", session.ActivePaneDir))
		}
		if session.ActivePaneCmd != "" {
			lines = append(lines, fmt.Sprintf("  Command:   %s", session.ActivePaneCmd))
		}
		if ssh, ok := detectRemoteConnection(session.ActivePaneCmd, session.ActivePaneTitle, session.ActivePanePID); ok {
			lines = append(lines, fmt.Sprintf("  Remote:    %s", ssh.Display()))
		}
	}

	pp.content = strings.Join(lines, "\n")
}

// SetWindowMetadata populates the panel for a window.
func (pp *PreviewPanel) SetWindowMetadata(window tmux.Window) {
	pp.title = "Window"

	var lines []string
	lines = append(lines, pp.styles.CardTitle.Render(fmt.Sprintf("%d: %s", window.Index, window.Name)))
	lines = append(lines, "")

	lines = append(lines, fmt.Sprintf("Panes:       %d", window.PaneCount))
	lines = append(lines, fmt.Sprintf("Layout:      %s", window.Layout))

	if window.WorkingDir != "" {
		lines = append(lines, fmt.Sprintf("Dir:         %s", window.WorkingDir))
	}
	if window.ActivePaneCmd != "" {
		lines = append(lines, fmt.Sprintf("Command:     %s", window.ActivePaneCmd))
		if ssh, ok := detectRemoteConnection(window.ActivePaneCmd, window.ActivePaneTitle, window.ActivePanePID); ok {
			lines = append(lines, fmt.Sprintf("Remote:      %s", ssh.Display()))
		}
	}

	pp.content = strings.Join(lines, "\n")
}

// SetPaneMetadata populates the panel for a pane.
func (pp *PreviewPanel) SetPaneMetadata(pane tmux.Pane) {
	pp.title = "Pane"

	var lines []string
	lines = append(lines, pp.styles.CardTitle.Render(fmt.Sprintf("Pane %d", pane.Index)))
	lines = append(lines, "")

	if pane.WorkingDir != "" {
		lines = append(lines, fmt.Sprintf("Dir:         %s", pane.WorkingDir))
	}
	lines = append(lines, fmt.Sprintf("Command:     %s", pane.Command))
	if ssh, ok := detectRemoteConnection(pane.Command, pane.Title, pane.PID); ok {
		lines = append(lines, fmt.Sprintf("Remote:      %s", ssh.Display()))
	}
	lines = append(lines, fmt.Sprintf("Size:        %dx%d", pane.Width, pane.Height))

	pp.content = strings.Join(lines, "\n")
}

// SetCaptureContent sets raw capture-pane output.
func (pp *PreviewPanel) SetCaptureContent(content string) {
	pp.title = "Preview"
	pp.content = content
}

// Render returns the rendered panel string.
//
// pp.width  = content width passed to lipgloss Width(). Rendered = pp.width + 2 (border).
// pp.height = total body height. We set lipgloss Height(pp.height - 2) so
//
//	rendered height = (pp.height - 2) + 2 = pp.height.
func (pp *PreviewPanel) Render() string {
	if pp.width < 10 {
		return ""
	}

	titleLine := pp.styles.PreviewTitle.Render(pp.title)

	body := pp.content
	if body == "" {
		body = pp.styles.CardSubtle.Render("(no content)")
	}

	// Usable lines inside the box: total height - border(2) - padding(2) - title(1) - blank after title(1).
	maxLines := pp.height - 6
	if maxLines < 1 {
		maxLines = 1
	}
	lines := strings.Split(body, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}

	// Truncate lines that exceed the panel width to prevent layout overflow.
	// Use lipgloss.Width for accurate visual width (handles wide/multi-byte chars).
	for i, line := range lines {
		if lipgloss.Width(line) > pp.width {
			runes := []rune(line)
			cut := len(runes)
			for cut > 0 && lipgloss.Width(string(runes[:cut])) > pp.width {
				cut--
			}
			lines[i] = string(runes[:cut])
		}
	}

	inner := titleLine + "\n" + strings.Join(lines, "\n")

	// lipgloss: Width/Height set content area; border adds +2 each axis.
	contentH := pp.height - 2
	if contentH < 1 {
		contentH = 1
	}

	// lipgloss Width includes padding; pp.width is the text content width,
	// so add 2 for horizontal padding (Padding(1) = 1 left + 1 right).
	return pp.styles.PreviewBorder.
		Copy().
		Padding(1).
		Width(pp.width + 2).
		Height(contentH).
		Render(inner)
}

// formatTime formats a time for display, handling zero times.
func formatTime(t interface{ Format(string) string }) string {
	type zeroChecker interface {
		IsZero() bool
	}
	if z, ok := t.(zeroChecker); ok && z.IsZero() {
		return "?"
	}
	return t.Format("2006-01-02 15:04")
}

// ---------------------------------------------------------------------------
// SSH / remote connection detection
// ---------------------------------------------------------------------------

var remoteCommands = map[string]bool{
	"ssh": true, "mosh": true, "mosh-client": true,
	"ftp": true, "sftp": true,
}

// sshFlagsWithValue lists ssh flags that consume the next token as their value.
var sshFlagsWithValue = map[string]bool{
	"-b": true, "-c": true, "-D": true, "-e": true, "-F": true,
	"-i": true, "-I": true, "-J": true, "-l": true, "-L": true,
	"-m": true, "-o": true, "-O": true, "-p": true, "-Q": true,
	"-R": true, "-S": true, "-w": true, "-W": true,
}

type sshInfo struct {
	User string
	Host string
	Port string
}

func (s *sshInfo) Display() string {
	base := s.Host
	if s.User != "" {
		base = s.User + "@" + s.Host
	}
	if s.Port != "" {
		base += ":" + s.Port
	}
	return base
}

// detectRemoteConnection returns (info, true) if command is a known remote
// process. It tries to read process args via ps first (reliable), then falls
// back to parsing pane_title (best-effort, depends on remote shell config).
func detectRemoteConnection(command, title string, pid int) (*sshInfo, bool) {
	if !remoteCommands[strings.ToLower(strings.TrimSpace(command))] {
		return nil, false
	}
	// Primary: read full command line from ps.
	if pid > 0 {
		if args := readProcessArgs(pid); args != "" {
			if info, ok := parseSSHArgs(args); ok {
				return info, true
			}
		}
	}
	// Fallback: parse pane_title set by remote shell.
	if t := strings.TrimSpace(title); t != "" {
		return parseSSHTitle(t)
	}
	return nil, false
}

// readProcessArgs returns the full command line of a process via ps.
func readProcessArgs(pid int) string {
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "args=").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// parseSSHArgs parses an ssh command line and extracts user, host, port.
// Handles: ssh [opts] [user@]host [command]
// Handles combined flag syntax like -p22 as well as separate -p 22.
func parseSSHArgs(args string) (*sshInfo, bool) {
	tokens := strings.Fields(args)
	if len(tokens) == 0 {
		return nil, false
	}
	// Accept both "ssh ..." and full paths like "/usr/bin/ssh ...".
	base := strings.ToLower(filepath.Base(tokens[0]))
	if base != "ssh" && base != "mosh" && base != "sftp" && base != "ftp" {
		return nil, false
	}

	info := &sshInfo{}
	i := 1
	for i < len(tokens) {
		t := tokens[i]
		if t == "--" {
			i++
			break
		}
		// Combined -p22 form.
		if strings.HasPrefix(t, "-p") && len(t) > 2 {
			info.Port = t[2:]
			i++
			continue
		}
		if sshFlagsWithValue[t] && i+1 < len(tokens) {
			if t == "-p" {
				info.Port = tokens[i+1]
			} else if t == "-l" {
				info.User = tokens[i+1]
			}
			i += 2
			continue
		}
		if strings.HasPrefix(t, "-") {
			i++
			continue
		}
		// First non-flag argument is [user@]host.
		if at := strings.Index(t, "@"); at > 0 {
			if info.User == "" {
				info.User = t[:at]
			}
			info.Host = t[at+1:]
		} else {
			info.Host = t
		}
		break
	}
	if info.Host == "" {
		return nil, false
	}
	return info, true
}

// parseSSHTitle attempts to extract user@host[:port] from a terminal title.
// Common formats set by shells on the remote end:
//
//	"user@host: ~/path"   (bash/zsh with PROMPT_COMMAND)
//	"user@host"           (minimal)
func parseSSHTitle(title string) (*sshInfo, bool) {
	if idx := strings.Index(title, ": "); idx != -1 {
		title = title[:idx]
	}
	atIdx := strings.Index(title, "@")
	if atIdx < 1 {
		return nil, false
	}
	user := title[:atIdx]
	hostPart := title[atIdx+1:]
	if hostPart == "" {
		return nil, false
	}
	info := &sshInfo{User: user}
	if colonIdx := strings.LastIndex(hostPart, ":"); colonIdx != -1 {
		info.Host = hostPart[:colonIdx]
		info.Port = hostPart[colonIdx+1:]
	} else {
		info.Host = hostPart
	}
	if info.Host == "" {
		return nil, false
	}
	return info, true
}
