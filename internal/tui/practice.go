package tui

import (
	"fmt"
	"time"

	"pib/internal/agent"
	"pib/internal/model"
	"pib/internal/service"

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

// PracticeWithAIModel is the practice mode with AI comparison
type PracticeWithAIModel struct {
	questionID    string
	question      string
	answer        string
	userAnswer    string
	inputActive   bool
	Completed     bool
	Cancelled     bool
	submitting    bool
	savedAnswer   bool
	analyzing     bool
	typingStats   typingStats
	compareResult *agent.CompareResult
	aiEngine      *agent.CompareEngine
	viewState     string // "input", "analyzing", "result"
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

// submitAnswer saves the user's answer to the database with SM-2 scheduling
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

			// Apply SM-2 algorithm to calculate next review time
			// Default to GradeVague (1) for manual practice without AI
			sm2Calc := service.NewSM2Calculator(2.5, 1.3)
			result := sm2Calc.Calculate(q.EF, q.Interval, model.GradeVague)
			
			// Update question with SM-2 calculated values
			q.EF = result.NewEF
			q.Interval = result.NewInterval
			q.NextReviewAt = &result.NextReviewAt
			
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

// ===========================================
// PracticeWithAI - AI-powered Practice Mode
// ===========================================

var (
	aiResultTitleStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("82")).
				Bold(true).
				Padding(0, 1)

	aiScoreStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86")).
			Bold(true).
			Width(60)

	aiAnalysisStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252")).
			Width(80)
)

// NewPracticeWithAIModel creates a new practice model with AI comparison
func NewPracticeWithAIModel() (*PracticeWithAIModel, error) {
	// Try to load config and create AI engine
	cfg, err := agent.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	aiEngine, err := agent.NewCompareEngine(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize AI engine: %v", err)
	}

	return &PracticeWithAIModel{
		inputActive: true,
		viewState:   "input",
		aiEngine:    aiEngine,
	}, nil
}

// SetQuestion sets the question for AI-powered practice
func (m *PracticeWithAIModel) SetQuestion(id, question, answer string) {
	m.questionID = id
	m.question = question
	m.answer = answer
	m.userAnswer = ""
	m.Completed = false
	m.Cancelled = false
	m.savedAnswer = false
	m.analyzing = false
	m.inputActive = true
	m.viewState = "input"
	m.compareResult = nil
	m.typingStats = typingStats{
		startTime:   time.Now(),
		lastKeyTime: time.Now(),
	}
}

// Init initializes the model
func (m *PracticeWithAIModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for AI-powered practice
func (m *PracticeWithAIModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.viewState == "result" {
				m.Cancelled = true
			} else {
				m.Cancelled = true
			}
		case "b":
			m.Cancelled = true
		case "ctrl+c":
			m.Cancelled = true
			return m, tea.Quit
		case "enter":
			if m.viewState == "input" && m.inputActive && m.userAnswer != "" {
				// Start AI analysis
				m.inputActive = false
				m.analyzing = true
				m.viewState = "analyzing"

				// Run AI comparison in background
				question := m.question
				answer := m.answer
				userAnswer := m.userAnswer
				aiEngine := m.aiEngine
				return m, func() tea.Msg {
					result, err := aiEngine.Compare(question, answer, userAnswer)
					return aiComparisonDoneMsg{Result: result, Err: err}
				}
			}
		case "backspace":
			if m.viewState == "input" && m.inputActive && len(m.userAnswer) > 0 {
				m.userAnswer = m.userAnswer[:len(m.userAnswer)-1]
			}
		case "r":
			// Re-analyze
			if m.viewState == "result" && m.savedAnswer {
				m.viewState = "analyzing"
				m.analyzing = true
				question := m.question
				answer := m.answer
				userAnswer := m.userAnswer
				aiEngine := m.aiEngine
				return m, func() tea.Msg {
					result, err := aiEngine.Compare(question, answer, userAnswer)
					return aiComparisonDoneMsg{Result: result, Err: err}
				}
			}
		default:
			// Handle regular character input
			if m.viewState == "input" && m.inputActive && len(msg.String()) == 1 {
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

	case aiComparisonDoneMsg:
		// AI analysis completed
		m.analyzing = false
		m.viewState = "result"
		if msg.Err != nil {
			// On error, create a basic result
			m.compareResult = &agent.CompareResult{
				SemanticScore:   50,
				ExpressionScore: 50,
				Analysis: agent.Analysis{
					Suggestions: []string{fmt.Sprintf("Analysis failed: %v", msg.Err)},
				},
				FollowUp: []string{},
			}
		} else {
			m.compareResult = msg.Result
		}
		m.savedAnswer = true

		// Apply SM-2 algorithm based on AI scores
		if m.compareResult != nil && db != nil {
			// Calculate overall score
			overallScore := (m.compareResult.SemanticScore + m.compareResult.ExpressionScore) / 2

			// Determine grade based on AI score
			var grade int
			if overallScore >= 70 {
				grade = model.GradeRemembered
			} else if overallScore >= 40 {
				grade = model.GradeVague
			} else {
				grade = model.GradeForgot
			}

			// Load question and update with SM-2 calculated values
			q, err := db.GetQuestionByID(m.questionID)
			if err == nil && q != nil {
				// Append user's answer to the question
				practiceAnswer := fmt.Sprintf("[Practice %s - Score: %.0f]\n%s",
					time.Now().Format("2006-01-02 15:04"), overallScore, m.userAnswer)
				if q.Answer != "" {
					q.Answer = q.Answer + "\n\n" + practiceAnswer
				} else {
					q.Answer = practiceAnswer
				}

				// Apply SM-2 algorithm
				sm2Calc := service.NewSM2Calculator(2.5, 1.3)
				result := sm2Calc.Calculate(q.EF, q.Interval, grade)

				// Update question with SM-2 calculated values
				q.EF = result.NewEF
				q.Interval = result.NewInterval
				q.NextReviewAt = &result.NextReviewAt

				db.UpdateQuestion(q)

				// Create review log with the grade
				reviewLog := &model.ReviewLog{
					QuestionID: m.questionID,
					Grade:      grade,
				}
				db.CreateReviewLog(reviewLog)
			}
		}
	}

	return m, nil
}

// aiComparisonDoneMsg is a custom message for AI comparison completion
type aiComparisonDoneMsg struct {
	Result *agent.CompareResult
	Err    error
}

// View renders the AI-powered practice page
func (m *PracticeWithAIModel) View() string {
	var s string

	header := practiceTitleStyle.Render("🎯 AI-Powered Practice Mode")
	s += header + "\n\n"

	// Question
	s += practiceHintStyle.Render("Question:") + "\n"
	s += practiceQuestionStyle.Render(m.question) + "\n\n"

	switch m.viewState {
	case "input":
		s = m.renderInputView(s)
	case "analyzing":
		s = m.renderAnalyzingView(s)
	case "result":
		s = m.renderResultView(s)
	}

	return s
}

// renderInputView renders the input view
func (m *PracticeWithAIModel) renderInputView(s string) string {
	s += practiceHintStyle.Render("Your Answer:") + "\n"

	if m.inputActive {
		s += practiceInputStyle.Render(m.userAnswer+"▋") + "\n"
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
		s += practiceButtonStyle.Render(" ↵ Submit with AI Analysis ") + " "
		s += practiceHintStyle.Render("[Enter] to submit and get AI feedback")
	}

	s += "\n\n" + practiceHintStyle.Render("Tips: esc/b: Back | q: Quit")

	return s
}

// renderAnalyzingView renders the analyzing view
func (m *PracticeWithAIModel) renderAnalyzingView(s string) string {
	s += practiceHintStyle.Render("Your Answer:") + "\n"
	s += practiceInputStyle.Render(m.userAnswer) + "\n\n"

	s += practiceHintStyle.Render("🤖 AI is analyzing your answer...") + "\n"
	s += "\n"
	s += aiResultTitleStyle.Render("⏳ Processing semantic analysis and generating feedback...") + "\n"

	s += "\n\n" + practiceHintStyle.Render("Please wait...")

	return s
}

// renderResultView renders the result view
func (m *PracticeWithAIModel) renderResultView(s string) string {
	result := m.compareResult

	// Overall Score
	overallScore := (result.SemanticScore + result.ExpressionScore) / 2
	scoreColor := "82" // green
	if overallScore < 60 {
		scoreColor = "227" // yellow
	}
	if overallScore < 40 {
		scoreColor = "196" // red
	}

	s += aiScoreStyle.Render(fmt.Sprintf("📊 Overall Score: %s%d%s / 100",
		lipgloss.NewStyle().Foreground(lipgloss.Color(scoreColor)).Bold(true).Render(fmt.Sprintf("%.0f", overallScore)),
		"", "")) + "\n\n"

	// Detailed Scores
	s += practiceHintStyle.Render("Score Breakdown:") + "\n"
	s += aiAnalysisStyle.Render(fmt.Sprintf("  • Semantic Score: %.0f/100 - Content coverage", result.SemanticScore)) + "\n"
	s += aiAnalysisStyle.Render(fmt.Sprintf("  • Expression Score: %.0f/100 - Structure & clarity", result.ExpressionScore)) + "\n\n"

	// Strengths
	if len(result.Analysis.Strengths) > 0 {
		s += aiResultTitleStyle.Render("✅ Strengths:") + "\n"
		for _, strength := range result.Analysis.Strengths {
			s += "  • " + strength + "\n"
		}
		s += "\n"
	}

	// Areas for Improvement
	if len(result.Analysis.Suggestions) > 0 || len(result.Analysis.LogicGaps) > 0 {
		s += aiResultTitleStyle.Render("📝 Areas for Improvement:") + "\n"
		for _, suggestion := range result.Analysis.Suggestions {
			s += "  • " + suggestion + "\n"
		}
		for _, gap := range result.Analysis.LogicGaps {
			s += "  ⚠ " + gap + "\n"
		}
		s += "\n"
	}

	// Missing Terms
	if len(result.Analysis.MissingTerms) > 0 {
		s += aiResultTitleStyle.Render("🔑 Missing Key Terms:") + "\n"
		for _, term := range result.Analysis.MissingTerms {
			s += "  • " + term + "\n"
		}
		s += "\n"
	}

	// Redundancy
	if result.Analysis.Redundancy != "" {
		s += aiResultTitleStyle.Render("🔄 Expression Analysis:") + "\n"
		s += "  " + result.Analysis.Redundancy + "\n\n"
	}

	// Follow-up Questions
	if len(result.FollowUp) > 0 {
		s += aiResultTitleStyle.Render("❓ Follow-up Questions to Consider:") + "\n"
		for i, q := range result.FollowUp {
			s += fmt.Sprintf("  %d. %s\n", i+1, q)
		}
		s += "\n"
	}

	// Actions
	s += practiceButtonStyle.Render(" [r] Re-analyze ") + " "
	s += practiceHintStyle.Render("analyze again")
	s += " " + practiceButtonStyle.Render(" [b] Back to Question ") + "\n"

	s += "\n" + practiceHintStyle.Render("Tips: r: Re-analyze | b/esc: Back to question")

	return s
}
