package main

import (
	"fmt"
	"log"
	"os"

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
		if err := runSubcommand(os.Args[1]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	model, err := tui.NewModel()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func runSubcommand(cmd string) error {
	client := tmux.NewClient()
	if !client.IsInTmux() {
		return fmt.Errorf("not inside a tmux session")
	}

	switch cmd {
	case "last":
		return client.SwitchToLast()
	default:
		return fmt.Errorf("unknown command: %s\nUsage: tswitch [last]", cmd)
	}
}
