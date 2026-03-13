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
	OpenConfig      bool
	OpenImport      bool
	OpenAnalytics   bool
	loading         bool
	err             error
	filterTag       string   // Current filter tag
	allTags         []string // All available tags
	showDueOnly     bool     // Show only questions due for review
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
		// Initialize database if needed
		if err := initDB(); err == nil {
			// Try to load from SQLite with tag filter or due for review
			var questions []model.Question
			var err error
			
			if m.showDueOnly {
				// Load questions due for review
				questions, err = db.GetQuestionsForReview(50)
			} else if m.filterTag != "" {
				questions, err = db.GetQuestionsByTag(m.filterTag)
			} else {
				questions, err = db.ListQuestionsByStatus("")
			}
			
			if err == nil && len(questions) > 0 {
				// Load tags for each question
				for i := range questions {
					tags, err := db.GetTagsForQuestion(questions[i].ID)
					if err == nil {
						questions[i].Tags = tags
					}
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

				// Load all tags for filter
				var allTags []string
				tagCounts, err := db.GetQuestionCountByTag()
				if err == nil {
					for name := range tagCounts {
						allTags = append(allTags, name)
					}
				}

				return questionsLoaded{items, allTags}
			}
		}

		// Fall back to JSON file
		questions, err := loadQuestionsFromFile()
		if err != nil {
			return errorMsg{err}
		}

		items := make([]QuestionItem, len(questions))
		tagSet := make(map[string]bool)
		for i, q := range questions {
			tags := make([]string, len(q.Tags))
			for j, tag := range q.Tags {
				tags[j] = tag.Name
				tagSet[tag.Name] = true
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

		// Collect all unique tags
		allTags := make([]string, 0, len(tagSet))
		for tag := range tagSet {
			allTags = append(allTags, tag)
		}

		return questionsLoaded{items, allTags}
	}
}

type questionsLoaded struct {
	items   []QuestionItem
	allTags []string
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
		m.allTags = msg.allTags
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
		case "i":
			m.OpenImport = true
		case "c":
			m.OpenConfig = true
		case "a":
			m.OpenAnalytics = true
		case "r":
			m.loading = true
			m.filterTag = "" // Clear filter on refresh
			return m, m.LoadQuestions()
		case "t":
			// Cycle through tags for filtering
			m.loading = true
			m.showDueOnly = false // Clear due filter when cycling tags
			if len(m.allTags) == 0 {
				m.loading = false
				m.filterTag = ""
			} else {
				// Find current tag index
				currentIdx := -1
				for i, tag := range m.allTags {
					if tag == m.filterTag {
						currentIdx = i
						break
					}
				}
				// Move to next tag
				nextIdx := (currentIdx + 1) % (len(m.allTags) + 1)
				if nextIdx == len(m.allTags) {
					m.filterTag = ""
				} else {
					m.filterTag = m.allTags[nextIdx]
				}
				return m, m.LoadQuestions()
			}
		case "d":
			// Toggle showing only due questions
			m.loading = true
			m.showDueOnly = !m.showDueOnly
			m.filterTag = "" // Clear tag filter when toggling due
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
	
	// Show filter status
	if m.showDueOnly {
		header += " " + lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("📅 Due for Review")
	} else if m.filterTag != "" {
		header += " " + tagStyle.Render("Filter: "+m.filterTag)
	}
	
	s += header + "\n\n"

	// Show available tags if any
	if len(m.allTags) > 0 && m.filterTag == "" {
		s += normalStyle.Render("Tags: ")
		for _, tag := range m.allTags {
			s += tagStyle.Render(tag)
		}
		s += "\n\n"
	}

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

	help := helpStyle.Render("↑/↓: Navigate  Enter: View  n: Add new  i: Import  c: Config  a: Analytics  t: Filter by tag  d: Due for review  r: Refresh  q: Quit")
	s += help

	return s
}
