package tui

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/textarea"
)

// ImportPageModel handles importing questions from Markdown files
type ImportPageModel struct {
	pathInput     textinput.Model
	previewText   textarea.Model
	importResult  string
	focus         int // 0: path, 1: import button, 2: done
	Importing    bool
	Completed    bool
	Cancelled    bool
	err           error
}

// NewImportPageModel creates a new import page model
func NewImportPageModel() *ImportPageModel {
	pathInput := textinput.New()
	pathInput.Placeholder = "Enter file path or directory path..."
	pathInput.Focus()
	pathInput.Prompt = "📁 "

	previewText := textarea.New()
	previewText.SetValue("Enter a path above and press Enter to preview, or 'i' to import directly.\n\nSupported formats:\n- Single Markdown file (.md)\n- Directory containing .md files\n\nMarkdown format:\n---\n# Question 1\nAnswer for question 1\n---\n# Question 2\nAnswer for question 2\n---\nTags: Golang, K8s\n")
	previewText.Blur()

	return &ImportPageModel{
		pathInput:   pathInput,
		previewText: previewText,
		focus:       0,
	}
}

// Init initializes the model
func (m *ImportPageModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *ImportPageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.Cancelled = true
			return m, nil

		case "enter":
			if m.focus == 0 {
				// Preview the file
				path := m.pathInput.Value()
				if path == "" {
					m.err = fmt.Errorf("please enter a path")
					return m, nil
				}

				// Check if it's a directory or file
				info, err := os.Stat(path)
				if err != nil {
					m.err = fmt.Errorf("invalid path: %v", err)
					return m, nil
				}

				if info.IsDir() {
					// Directory: show list of files
					files, err := GetMarkdownFilesInDirectory(path)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.previewText.SetValue(fmt.Sprintf("Found %d Markdown files:\n\n", len(files)))
					for i, f := range files {
						if i < 20 {
							m.previewText.SetValue(m.previewText.Value() + fmt.Sprintf("- %s\n", f))
						}
					}
					if len(files) > 20 {
						m.previewText.SetValue(m.previewText.Value() + fmt.Sprintf("\n... and %d more files", len(files)-20))
					}
				} else {
					// File: show preview
					preview, err := ReadMarkdownPreview(path)
					if err != nil {
						m.err = err
						return m, nil
					}
					m.previewText.SetValue(preview)
				}
				m.err = nil

			} else if m.focus == 1 {
				// Import
				m.doImport()
			}

		case "tab":
			// Cycle through inputs
			m.focus = (m.focus + 1) % 3
			switch m.focus {
			case 0:
				m.pathInput.Focus()
			case 1:
				// Import button - no focus needed
			case 2:
				// Done - no focus needed
			}

		case "i":
			// Direct import
			if m.pathInput.Value() != "" {
				m.doImport()
			}

		case "r":
			// Refresh preview
			if m.pathInput.Value() != "" {
				path := m.pathInput.Value()
				info, err := os.Stat(path)
				if err == nil && !info.IsDir() {
					preview, err := ReadMarkdownPreview(path)
					if err == nil {
						m.previewText.SetValue(preview)
					}
				}
			}
		}
	}

	// Update inputs
	m.pathInput, _ = m.pathInput.Update(msg)

	return m, nil
}

// doImport performs the actual import
func (m *ImportPageModel) doImport() {
	path := m.pathInput.Value()
	if path == "" {
		m.err = fmt.Errorf("please enter a path")
		return
	}

	m.Importing = true
	m.err = nil

	info, err := os.Stat(path)
	if err != nil {
		m.err = fmt.Errorf("invalid path: %v", err)
		m.Importing = false
		return
	}

	var result *ImportResult

	if info.IsDir() {
		result, err = ImportFromDirectory(path)
	} else {
		result, err = ImportFromMarkdownFile(path)
	}

	m.Importing = false

	if err != nil {
		m.err = err
		return
	}

	// Show result
	var resultText string
	if len(result.Errors) > 0 {
		resultText = fmt.Sprintf("Import completed with errors:\n\n")
		resultText += fmt.Sprintf("Total: %d, Imported: %d\n\n", result.TotalQuestions, result.ImportedCount)
		resultText += "Errors:\n"
		for _, e := range result.Errors {
			resultText += fmt.Sprintf("- %s\n", e)
		}
	} else {
		resultText = fmt.Sprintf("✅ Import successful!\n\n")
		resultText += fmt.Sprintf("Total questions: %d\n", result.TotalQuestions)
		resultText += fmt.Sprintf("Imported: %d\n", result.ImportedCount)
	}

	m.previewText.SetValue(resultText)
	m.importResult = resultText
	m.Completed = true
}

// View renders the import page
func (m *ImportPageModel) View() string {
	s := ""

	header := titleStyle.Render("📥 Import from Markdown")
	s += header + "\n\n"

	// Path input
	pathLabel := normalStyle.Render("File/Directory path:")
	if m.focus == 0 {
		pathLabel = selectedStyle.Render("File/Directory path:")
	}
	s += pathLabel + "\n"
	s += m.pathInput.View() + "\n\n"

	// Preview area
	s += normalStyle.Render("Preview:") + "\n"
	s += m.previewText.View() + "\n\n"

	// Import button
	importLabel := "  Import  "
	if m.focus == 1 {
		importLabel = buttonActiveStyle.Render("  Import  ")
	} else {
		importLabel = buttonStyle.Render("  Import  ")
	}

	// Status
	if m.Importing {
		s += normalStyle.Render("⏳ Importing...") + "\n\n"
	}

	// Error
	if m.err != nil {
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("Error: %v", m.err)) + "\n\n"
	}

	s += importLabel + "\n\n"

	// Help
	s += helpStyle.Render("Enter: Preview  i: Import  r: Refresh  Tab: Next  esc/q: Cancel")

	return s
}
