package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	detailTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true).
				Padding(0, 1)

	detailContentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Padding(1, 2)

	detailAnswerStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("75")).
				Background(lipgloss.Color("236")).
				Padding(1, 2)

	buttonStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Background(lipgloss.Color("235")).
			Padding(0, 2).
			Margin(1)

	buttonActiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235")).
				Background(lipgloss.Color("86")).
				Bold(true).
				Padding(0, 2).
				Margin(1)
)

// QuestionDetailModel displays question details
type QuestionDetailModel struct {
	questionID   string
	content      string
	answer       string
	tags         []string
	BackToList   bool
}

// NewQuestionDetailModel creates a new detail model
func NewQuestionDetailModel() *QuestionDetailModel {
	return &QuestionDetailModel{}
}

// SetQuestion sets the question to display
func (m *QuestionDetailModel) SetQuestion(id string, content string) {
	m.questionID = id
	m.content = content
	// Load full question details from file
	q := loadQuestionByID(id)
	if q != nil {
		m.answer = q.Answer
		m.tags = make([]string, len(q.Tags))
		for i, tag := range q.Tags {
			m.tags[i] = tag.Name
		}
	}
}

// Init initializes the model
func (m *QuestionDetailModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *QuestionDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.BackToList = true
		case "b":
			m.BackToList = true
		}
	}

	return m, nil
}

// View renders the detail page
func (m *QuestionDetailModel) View() string {
	s := ""

	header := detailTitleStyle.Render("📝 Question Detail")
	s += header + "\n\n"

	// Content
	s += normalStyle.Render("Question:") + "\n"
	s += detailContentStyle.Render(m.content) + "\n\n"

	// Answer
	if m.answer != "" {
		s += normalStyle.Render("Answer:") + "\n"
		s += detailAnswerStyle.Render(m.answer) + "\n\n"
	}

	// Tags
	if len(m.tags) > 0 {
		s += normalStyle.Render("Tags:") + " "
		for _, tag := range m.tags {
			s += tagStyle.Render(tag)
		}
		s += "\n\n"
	}

	// Navigation help
	s += "\n" + helpStyle.Render("esc/b: Back to list  q: Quit")

	return s
}
