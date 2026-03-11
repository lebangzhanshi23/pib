package main

import (
	"fmt"
	"os"

	"pib/internal/tui"

	"github.com/charmbracelet/bubbletea"
)

func main() {
	// Test mode for debugging
	if os.Getenv("TEST_MODE") == "1" {
		fmt.Println("Running in test mode...")
		// This will trigger loading questions
		model := tui.NewMainModel()
		// We can't call Init directly as it returns a tea.Cmd
		// Instead, let's just run the program briefly
		p := tea.NewProgram(model, tea.WithoutSignalHandler())
		if err := p.Start(); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		return
	}

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
