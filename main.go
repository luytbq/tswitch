package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/luytbq/tswitch/internal/config"
	"github.com/luytbq/tswitch/internal/keys"
	"github.com/luytbq/tswitch/internal/tmux"
	"github.com/luytbq/tswitch/internal/tui"
)

func main() {
	// Load JSON app config and apply key overrides before anything else.
	appCfg, err := config.LoadAppConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to load app config: %v\n", err)
	}
	if appCfg != nil && len(appCfg.Keys) > 0 {
		keys.ApplyOverrides(appCfg.Keys)
	}

	if len(os.Args) > 1 {
		if err := runSubcommand(os.Args[1], appCfg); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	model, err := tui.NewModel(appCfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func runSubcommand(cmd string, appCfg *config.AppConfig) error {
	client := tmux.NewClient()
	if !client.IsInTmux() {
		return fmt.Errorf("not inside a tmux session")
	}

	switch cmd {
	case "last":
		return client.SwitchToLast()
	case "browse":
		return runBrowse(client, appCfg)
	default:
		return fmt.Errorf("unknown command: %s\nUsage: tswitch [last|browse]", cmd)
	}
}

func runBrowse(client *tmux.Client, appCfg *config.AppConfig) error {
	if appCfg == nil || len(appCfg.BrowseDirs) == 0 {
		return fmt.Errorf("no browse directories configured")
	}

	shellCmd, tmpPath, err := tui.BuildBrowseCommand(appCfg)
	if err != nil {
		return err
	}

	// Run fzf interactively. fzf renders its UI on stderr and writes the
	// selection to stdout (redirected to tmpfile by the shell command).
	c := exec.Command("sh", "-c", shellCmd)
	c.Stdin = os.Stdin
	c.Stderr = os.Stderr
	fzfErr := c.Run()

	defer os.Remove(tmpPath)

	if fzfErr != nil {
		// User cancelled or fzf error — exit silently.
		return nil
	}

	data, err := os.ReadFile(tmpPath)
	if err != nil {
		return err
	}
	selected := strings.TrimSpace(string(data))
	if selected == "" {
		return nil
	}

	return tui.SwitchOrCreateSession(client, selected)
}
