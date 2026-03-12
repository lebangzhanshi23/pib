package tui

import (
	"fmt"
	"time"

	"pib/internal/model"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	practiceTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("86")).
				Bold(true).
				Padding(0, 1)

	practiceQuestionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Background(lipgloss.Color("235")).
				Padding(2, 2).
				Width(80)

	practiceInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("252")).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("86")).
				Padding(1, 2).
				Width(80)

	practiceButtonStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("235")).
				Background(lipgloss.Color("86")).
				Padding(0, 3).
				Margin(1)

	practiceHintStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241"))
)

// PracticeModel is the immersive practice mode
type PracticeModel struct {
	questionID   string
	question     string
	answer       string
	userAnswer   string
	inputActive  bool
	Completed    bool
	Cancelled    bool
	submitting   bool
	savedAnswer  bool
	typingStats  typingStats
}

// typingStats tracks user typing patterns
type typingStats struct {
	startTime     time.Time
	typingEvents  []time.Time
	pauseCount    int
	lastKeyTime   time.Time
	charsPerMinute float64
}

// NewPracticeModel creates a new practice model
func NewPracticeModel() *PracticeModel {
	return &PracticeModel{
		inputActive: true,
	}
}

// SetQuestion sets the question for practice
func (m *PracticeModel) SetQuestion(id, question, answer string) {
	m.questionID = id
	m.question = question
	m.answer = answer
	m.userAnswer = ""
	m.Completed = false
	m.Cancelled = false
	m.savedAnswer = false
	m.inputActive = true
	m.typingStats = typingStats{
		startTime:   time.Now(),
		lastKeyTime: time.Now(),
	}
}

// Init initializes the model
func (m *PracticeModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *PracticeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Cancelled = true
		case "b":
			// Also go back
			m.Cancelled = true
		case "ctrl+c":
			m.Cancelled = true
			return m, tea.Quit
		case "enter":
			// If input is active and user pressed Enter in the input area
			if m.inputActive && m.userAnswer != "" {
				// Submit the answer
				m.submitAnswer()
				m.inputActive = false
			}
		case "backspace":
			// Handle backspace
			if m.inputActive && len(m.userAnswer) > 0 {
				m.userAnswer = m.userAnswer[:len(m.userAnswer)-1]
			}
		default:
			// Handle regular character input
			if m.inputActive && len(msg.String()) == 1 {
				m.userAnswer += msg.String()
				
				// Track typing
				now := time.Now()
				// Detect pause (more than 2 seconds between keys)
				if now.Sub(m.typingStats.lastKeyTime) > 2*time.Second {
					m.typingStats.pauseCount++
				}
				m.typingStats.lastKeyTime = now
				
				// Update typing speed
				elapsed := now.Sub(m.typingStats.startTime)
				if elapsed > 0 {
					m.typingStats.charsPerMinute = float64(len(m.userAnswer)) / elapsed.Minutes()
				}
			}
		}

	case tea.WindowSizeMsg:
		// Handle window resize if needed
	}

	return m, nil
}

// submitAnswer saves the user's answer to the database
func (m *PracticeModel) submitAnswer() {
	if m.questionID == "" || m.userAnswer == "" {
		return
	}

	m.submitting = true

	// Try to save to SQLite first
	if db != nil {
		// Load the question and update with user's answer
		q, err := db.GetQuestionByID(m.questionID)
		if err == nil && q != nil {
			// Update the answer field with user's practice answer
			// We'll store the practice answer in a separate field or append
			practiceAnswer := fmt.Sprintf("[Practice %s]\n%s", 
				time.Now().Format("2006-01-02 15:04"), m.userAnswer)
			
			if q.Answer != "" {
				q.Answer = q.Answer + "\n\n" + practiceAnswer
			} else {
				q.Answer = practiceAnswer
			}
			
			db.UpdateQuestion(q)
		}
		
		// Create a review log entry
		reviewLog := &model.ReviewLog{
			QuestionID: m.questionID,
			Grade:      model.GradeVague, // Default grade for practice
		}
		db.CreateReviewLog(reviewLog)
	}

	m.submitting = false
	m.savedAnswer = true
}

// View renders the practice page
func (m *PracticeModel) View() string {
	var s string

	header := practiceTitleStyle.Render("🎯 Immersive Practice Mode")
	s += header + "\n\n"

	// Question
	s += practiceHintStyle.Render("Question:") + "\n"
	s += practiceQuestionStyle.Render(m.question) + "\n\n"

	// User's answer input
	s += practiceHintStyle.Render("Your Answer:") + "\n"
	
	if m.inputActive {
		s += practiceInputStyle.Render(m.userAnswer + "▋") + "\n"
	} else {
		s += practiceInputStyle.Render(m.userAnswer) + "\n"
	}

	// Show typing stats
	if len(m.userAnswer) > 0 {
		s += "\n"
		s += practiceHintStyle.Render(fmt.Sprintf("⌨️ Typing: %.0f chars/min | Pauses: %d | Total: %d chars",
			m.typingStats.charsPerMinute, m.typingStats.pauseCount, len(m.userAnswer))) + "\n"
	}

	s += "\n"

	// Buttons
	if !m.savedAnswer {
		// Submit button
		s += practiceButtonStyle.Render(" ↵ Submit ") + " "
		s += practiceHintStyle.Render("[Enter] to submit your answer")
	} else {
		// Show saved confirmation
		s += lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true).Render("✅ Answer saved!") + "\n"
		s += practiceHintStyle.Render("Press [b] or [esc] to return to question details")
	}

	// Help text
	s += "\n\n" + practiceHintStyle.Render("Tips: Type your answer, press Enter to submit. | esc/b: Back | q: Quit")

	return s
}
