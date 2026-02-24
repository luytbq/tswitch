package main

import (
	"fmt"
	"log"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/tswitch/internal/tmux"
	"github.com/user/tswitch/internal/tui"
)

func main() {
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
