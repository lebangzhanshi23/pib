package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/bubbles/textinput"
)

var (
	inputStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Padding(0, 1)

	inputFocusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("211")).
			Padding(0, 1)
)

// AddQuestionModel handles adding new questions
type AddQuestionModel struct {
	contentInput textinput.Model
	answerInput  textinput.Model
	tagsInput    textinput.Model
	focus        int // 0: content, 1: answer, 2: tags, 3: confirm
	Completed    bool
	Cancelled   bool
	err          error
}

// NewAddQuestionModel creates a new add question model
func NewAddQuestionModel() *AddQuestionModel {
	contentInput := textinput.New()
	contentInput.Placeholder = "Enter your question..."
	contentInput.Focus()
	contentInput.Prompt = "Q: "

	answerInput := textinput.New()
	answerInput.Placeholder = "Enter the answer (optional)..."
	answerInput.Prompt = "A: "

	tagsInput := textinput.New()
	tagsInput.Placeholder = "e.g. Golang, K8s, System-Design"
	tagsInput.Prompt = "Tags: "

	return &AddQuestionModel{
		contentInput: contentInput,
		answerInput:  answerInput,
		tagsInput:    tagsInput,
		focus:        0,
	}
}

// Init initializes the model
func (m *AddQuestionModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *AddQuestionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Cycle through inputs
			m.focus = (m.focus + 1) % 4
			switch m.focus {
			case 0:
				m.contentInput.Focus()
			case 1:
				m.answerInput.Focus()
			case 2:
				m.tagsInput.Focus()
			case 3:
				// Confirm button - no focus needed
			}
		case "enter":
			if m.focus < 3 {
				m.focus++
				switch m.focus {
				case 0:
					m.contentInput.Focus()
				case 1:
					m.answerInput.Focus()
				case 2:
					m.tagsInput.Focus()
				}
			} else {
				// Save the question
				if m.contentInput.Value() != "" {
					err := saveQuestion(m.contentInput.Value(), m.answerInput.Value(), m.tagsInput.Value())
					if err != nil {
						m.err = err
					} else {
						m.Completed = true
					}
				}
			}
		case "esc":
			m.Cancelled = true
		}
	}

	// Update inputs
	m.contentInput, _ = m.contentInput.Update(msg)
	m.answerInput, _ = m.answerInput.Update(msg)
	m.tagsInput, _ = m.tagsInput.Update(msg)

	return m, nil
}

// View renders the add question form
func (m *AddQuestionModel) View() string {
	s := ""

	header := titleStyle.Render("➕ Add New Question")
	s += header + "\n\n"

	// Content input
	contentLabel := normalStyle.Render("Question:")
	if m.focus == 0 {
		contentLabel = selectedStyle.Render("Question:")
	}
	s += contentLabel + "\n"
	s += m.contentInput.View() + "\n\n"

	// Answer input
	answerLabel := normalStyle.Render("Answer:")
	if m.focus == 1 {
		answerLabel = selectedStyle.Render("Answer:")
	}
	s += answerLabel + "\n"
	s += m.answerInput.View() + "\n\n"

	// Tags input
	tagsLabel := normalStyle.Render("Tags (comma separated):")
	if m.focus == 2 {
		tagsLabel = selectedStyle.Render("Tags (comma separated):")
	}
	s += tagsLabel + "\n"
	s += m.tagsInput.View() + "\n\n"

	// Save button
	saveLabel := "  Save  "
	if m.focus == 3 {
		saveLabel = buttonActiveStyle.Render("  Save  ")
	} else {
		saveLabel = buttonStyle.Render("  Save  ")
	}
	s += saveLabel + "\n\n"

	// Help
	s += helpStyle.Render("Tab: Next field  Enter: Confirm/Save  esc: Cancel  q: Quit")

	// Error
	if m.err != nil {
		s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("Error: %v", m.err))
	}

	return s
}
