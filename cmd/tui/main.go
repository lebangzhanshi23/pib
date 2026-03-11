package main

import (
	"os"

	"pib/internal/tui"

	"github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(
		tui.NewMainModel(),
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	)

	if _, err := p.Run(); err != nil {
		os.Stderr.WriteString("Error running TUI: " + err.Error() + "\n")
		os.Exit(1)
	}
}
