package tmux

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

var remoteCommands = map[string]bool{
	"ssh": true, "mosh": true, "mosh-client": true,
	"ftp": true, "sftp": true,
}

// shellCommands lists processes that may wrap a remote command (e.g. zsh -c ... ssh ...).
var shellCommands = map[string]bool{
	"zsh": true, "bash": true, "sh": true, "fish": true, "tcsh": true, "csh": true,
}

// sshFlagsWithValue lists ssh flags that consume the next token as their value.
var sshFlagsWithValue = map[string]bool{
	"-b": true, "-c": true, "-D": true, "-e": true, "-F": true,
	"-i": true, "-I": true, "-J": true, "-l": true, "-L": true,
	"-m": true, "-o": true, "-O": true, "-p": true, "-Q": true,
	"-R": true, "-S": true, "-w": true, "-W": true,
}

// RemoteInfo holds parsed connection details for a remote pane.
type RemoteInfo struct {
	Cmd  string // actual remote command name, e.g. "ssh", "mosh", "ftp"
	User string
	Host string
	Port string
}

func (r *RemoteInfo) Display() string {
	base := r.Host
	if r.User != "" {
		base = r.User + "@" + r.Host
	}
	if r.Port != "" {
		base += ":" + r.Port
	}
	return base
}

// DetectRemoteConnection returns (info, true) if the pane is running a known
// remote command (ssh, ftp, …) — either directly or wrapped inside a shell
// (e.g. "zsh -c … ssh user@host"). It searches the process subtree rooted at
// pid, then falls back to parsing pane_title.
func DetectRemoteConnection(command, title string, pid int) (*RemoteInfo, bool) {
	cmd := strings.ToLower(strings.TrimSpace(command))

	// Search the process subtree when the pane runs a remote command directly
	// OR when it runs a shell that may be wrapping one (zsh -c ... ssh ...).
	if pid > 0 && (remoteCommands[cmd] || shellCommands[cmd]) {
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

// readProcessArgs finds a remote command (ssh/ftp/…) in the process subtree
// rooted at pid. pane_pid is the shell PID; the actual remote command runs as
// a child or grandchild of that shell, so we search up to 4 levels deep.
func readProcessArgs(pid int) string {
	return findRemoteChildArgs(pid, 4)
}

// findRemoteChildArgs recursively searches children of pid for a process whose
// argv[0] is a known remote command, returning its full args string.
func findRemoteChildArgs(pid, depth int) string {
	if depth == 0 {
		return ""
	}
	out, err := exec.Command("pgrep", "-P", strconv.Itoa(pid)).Output()
	if err != nil || len(out) == 0 {
		return ""
	}
	for _, pidStr := range strings.Fields(string(out)) {
		childPID, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}
		argsOut, err := exec.Command("ps", "-p", pidStr, "-o", "args=").Output()
		if err != nil {
			continue
		}
		args := strings.TrimSpace(string(argsOut))
		if args == "" {
			continue
		}
		fields := strings.Fields(args)
		base := strings.ToLower(filepath.Base(fields[0]))
		if remoteCommands[base] {
			return args
		}
		// Not a remote command; search its children.
		if grandchild := findRemoteChildArgs(childPID, depth-1); grandchild != "" {
			return grandchild
		}
	}
	return ""
}

// parseSSHArgs parses an ssh command line and extracts user, host, port.
// Handles: ssh [opts] [user@]host [command]
// Handles combined flag syntax like -p22 as well as separate -p 22.
func parseSSHArgs(args string) (*RemoteInfo, bool) {
	tokens := strings.Fields(args)
	if len(tokens) == 0 {
		return nil, false
	}
	// Accept both "ssh ..." and full paths like "/usr/bin/ssh ...".
	base := strings.ToLower(filepath.Base(tokens[0]))
	if base != "ssh" && base != "mosh" && base != "sftp" && base != "ftp" {
		return nil, false
	}

	info := &RemoteInfo{Cmd: base}
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
func parseSSHTitle(title string) (*RemoteInfo, bool) {
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
	info := &RemoteInfo{User: user}
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
