package tui

import (
	"fmt"
	"time"

	"pib/internal/model"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("211")).
			Background(lipgloss.Color("57")).
			Bold(true)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("244"))

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("76")).
			Background(lipgloss.Color("235")).
			Padding(0, 1).
			MarginRight(1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))
)

// QuestionListModel displays a list of questions
type QuestionListModel struct {
	questions       []QuestionItem
	selected        int
	SelectedID      string
	SelectedContent string
	AddingNew       bool
	loading         bool
	err             error
}

// QuestionItem represents a question in the list
type QuestionItem struct {
	ID         string
	Content    string
	Status     string
	Tags       []string
	NextReview time.Time
}

// NewQuestionListModel creates a new list model
func NewQuestionListModel() *QuestionListModel {
	return &QuestionListModel{
		questions: []QuestionItem{},
		selected:  0,
	}
}

// LoadQuestions loads questions from the data file
func (m *QuestionListModel) LoadQuestions() tea.Cmd {
	return func() tea.Msg {
		// Load questions from JSON file
		questions, err := loadQuestionsFromFile()
		if err != nil {
			return errorMsg{err}
		}

		items := make([]QuestionItem, len(questions))
		for i, q := range questions {
			tags := make([]string, len(q.Tags))
			for j, tag := range q.Tags {
				tags[j] = tag.Name
			}
			var nextReview time.Time
			if q.NextReviewAt != nil {
				nextReview = *q.NextReviewAt
			}
			items[i] = QuestionItem{
				ID:         q.ID,
				Content:    q.Content,
				Status:     q.Status,
				Tags:       tags,
				NextReview: nextReview,
			}
		}

		return questionsLoaded{items}
	}
}

type questionsLoaded struct {
	items []QuestionItem
}

type errorMsg struct {
	err error
}

// Init initializes the model
func (m *QuestionListModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *QuestionListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case questionsLoaded:
		m.questions = msg.items
		m.loading = false

	case errorMsg:
		m.err = msg.err
		m.loading = false

	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.questions)-1 {
				m.selected++
			}
		case "enter":
			if len(m.questions) > 0 && m.selected < len(m.questions) {
				m.SelectedID = m.questions[m.selected].ID
				m.AddingNew = false
			}
		case "n":
			m.AddingNew = true
		case "r":
			m.loading = true
			return m, m.LoadQuestions()
		}
	}

	return m, nil
}

// View renders the list
func (m *QuestionListModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\nPress 'r' to retry", m.err)
	}

	if m.loading {
		return "Loading questions...\nPress 'q' to quit"
	}

	if len(m.questions) == 0 {
		header := titleStyle.Render("📚 PIB - Question Library")
		help := helpStyle.Render("\n\n↑/↓: Navigate  Enter: View  n: Add new  r: Refresh  q: Quit")
		return header + "\n\nNo questions yet. Press 'n' to add your first question." + help
	}

	var s string
	header := titleStyle.Render("📚 PIB - Question Library")
	s += header + "\n\n"

	for i, q := range m.questions {
		cursor := "  "
		if i == m.selected {
			cursor = "▶ "
		}

		status := ""
		switch q.Status {
		case model.StatusDraft:
			status = "📝 Draft"
		case model.StatusActive:
			status = "✅ Active"
		case model.StatusArchived:
			status = "📦 Archived"
		}

		// Truncate content if too long
		content := q.Content
		if len(content) > 60 {
			content = content[:57] + "..."
		}

		line := cursor + content
		if i == m.selected {
			line = selectedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		s += line + "\n"
		s += normalStyle.Render("   "+status) + "  "

		// Render tags
		for _, tag := range q.Tags {
			s += tagStyle.Render(tag)
		}

		// Show next review time
		if !q.NextReview.IsZero() {
			days := int(time.Since(q.NextReview).Hours() / 24)
			if days < 0 {
				s += normalStyle.Render(fmt.Sprintf("  Due in %d days", -days))
			} else if days == 0 {
				s += normalStyle.Render("  Due today")
			} else {
				s += normalStyle.Render(fmt.Sprintf("  %d days ago", days))
			}
		}

		s += "\n\n"
	}

	help := helpStyle.Render("↑/↓: Navigate  Enter: View  n: Add new  r: Refresh  q: Quit")
	s += help

	return s
}
