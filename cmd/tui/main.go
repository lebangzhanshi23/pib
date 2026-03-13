package main

import (
	"flag"
	"fmt"
	"os"

	"pib/internal/tui"

	"github.com/charmbracelet/bubbletea"
)

func main() {
	// Command line flags
	importFile := flag.String("import", "", "Import questions from Markdown file or directory")
	flag.Parse()

	// Handle import from command line
	if *importFile != "" {
		fmt.Printf("Importing from: %s\n", *importFile)
		result, err := tui.ImportFromMarkdownFile(*importFile)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Import completed!\n")
		fmt.Printf("  Total: %d\n", result.TotalQuestions)
		fmt.Printf("  Imported: %d\n", result.ImportedCount)
		if len(result.Errors) > 0 {
			fmt.Println("  Errors:")
			for _, e := range result.Errors {
				fmt.Printf("    - %s\n", e)
			}
		}
		return
	}

	// Test mode for debugging
	if os.Getenv("TEST_MODE") == "1" {
		fmt.Println("Running in test mode...")
		// This will trigger loading questions
		model := tui.NewMainModel(tui.GetDB())
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
		tui.NewMainModel(tui.GetDB()),
		tea.WithAltScreen(),
		tea.WithMouseAllMotion(),
	)

	if _, err := p.Run(); err != nil {
		os.Stderr.WriteString("Error running TUI: " + err.Error() + "\n")
		os.Exit(1)
	}
}
