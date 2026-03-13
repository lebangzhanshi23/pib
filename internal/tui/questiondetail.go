package tui

import (
	"fmt"

	"pib/internal/agent"

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
	questionID     string
	content        string
	answer         string
	tags           []string
	BackToList     bool
	StartPractice  bool
	StartPracticeAI bool
	// AI Scout fields
	scoutResult  *agent.ScoutResult
	scoutLoading bool
	scoutError   string
	showScout    int  // 0: detail, 1: beginner, 2: expert, 3: bigtech
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
			if m.showScout > 0 {
				m.showScout = 0
			} else {
				m.BackToList = true
			}
		case "b":
			if m.showScout > 0 {
				m.showScout = 0
			} else {
				m.BackToList = true
			}
		case "a":
			// Trigger AI Scout
			if !m.scoutLoading && m.showScout == 0 {
				m.scoutLoading = true
				m.scoutError = ""
				return m, m.runAIScout()
			}
		case "p":
			// Start immersive practice mode
			if m.showScout == 0 && !m.scoutLoading {
				m.StartPractice = true
			}
		case "i":
			// Start AI-powered practice mode
			if m.showScout == 0 && !m.scoutLoading {
				m.StartPracticeAI = true
			}
		case "1":
			// Show beginner answer
			if m.scoutResult != nil && m.showScout == 0 {
				m.showScout = 1
			}
		case "2":
			// Show expert answer
			if m.scoutResult != nil && m.showScout == 0 {
				m.showScout = 2
			}
		case "3":
			// Show big tech answer
			if m.scoutResult != nil && m.showScout == 0 {
				m.showScout = 3
			}
		}

	case ScoutResultMsg:
		m.scoutLoading = false
		m.scoutResult = msg.Result

	case ScoutErrorMsg:
		m.scoutLoading = false
		m.scoutError = msg.Error
	}

	return m, nil
}

// runAIScout runs the AI Scout in background
func (m *QuestionDetailModel) runAIScout() tea.Cmd {
	return func() tea.Msg {
		// Load config
		cfg, err := agent.LoadConfig()
		if err != nil {
			return ScoutErrorMsg{Error: fmt.Sprintf("配置加载失败: %v", err)}
		}

		if cfg.LLM.APIKey == "" {
			return ScoutErrorMsg{Error: "请先配置 LLM API Key (按 c 进入配置)"}
		}

		// Create LLM client
		client := agent.NewLLMClient(cfg)

		// Generate answers
		result, err := client.GenerateAnswers(m.content, m.tags)
		if err != nil {
			return ScoutErrorMsg{Error: fmt.Sprintf("生成失败: %v", err)}
		}

		return ScoutResultMsg{Result: result}
	}
}

// ScoutResultMsg represents AI Scout result message
type ScoutResultMsg struct {
	Result *agent.ScoutResult
}

// ScoutErrorMsg represents AI Scout error message
type ScoutErrorMsg struct {
	Error string
}

// View renders the detail page
func (m *QuestionDetailModel) View() string {
	s := ""

	header := detailTitleStyle.Render("📝 Question Detail")
	s += header + "\n\n"

	// Content
	s += normalStyle.Render("Question:") + "\n"
	s += detailContentStyle.Render(m.content) + "\n\n"

	// Answer (only show manual answer in detail view, not scout)
	if m.answer != "" && m.showScout == 0 {
		s += normalStyle.Render("Answer:") + "\n"
		s += detailAnswerStyle.Render(m.answer) + "\n\n"
	}

	// AI Scout Results
	if m.showScout > 0 && m.scoutResult != nil {
		var answerTitle, answerContent string
		switch m.showScout {
		case 1:
			answerTitle = "🌱 入门级答案"
			answerContent = m.scoutResult.Beginner
		case 2:
			answerTitle = "🚀 专家级答案"
			answerContent = m.scoutResult.Expert
		case 3:
			answerTitle = "🏢 大厂面试官版"
			answerContent = m.scoutResult.BigTech
		}

		s += lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Bold(true).Render(answerTitle) + "\n"
		s += detailAnswerStyle.Render(answerContent) + "\n\n"

		// Back to detail
		s += helpStyle.Render("esc/b: 返回题目详情") + "\n"
	} else {
		// AI Scout Button
		if m.scoutLoading {
			s += normalStyle.Render("🤖 AI Scout 生成中...") + "\n\n"
		} else if m.scoutError != "" {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("⚠️ "+m.scoutError) + "\n\n"
			s += buttonStyle.Render("  a: 重试 ") + "\n\n"
		} else if m.scoutResult != nil {
			s += normalStyle.Render("🤖 AI 答案已生成:") + "\n"
			s += buttonStyle.Render(" 1: 入门级 ") + " "
			s += buttonStyle.Render(" 2: 专家级 ") + " "
			s += buttonStyle.Render(" 3: 大厂版 ") + "\n\n"
		} else {
			s += buttonStyle.Render("  a: AI Scout 生成参考答案 ") + "\n\n"
		}

		// Tags
		if len(m.tags) > 0 {
			s += normalStyle.Render("Tags:") + " "
			for _, tag := range m.tags {
				s += tagStyle.Render(tag)
			}
			s += "\n\n"
		}

		// Practice buttons
		s += buttonStyle.Render(" p: 基础练习 ") + " "
		s += buttonActiveStyle.Render(" i: AI 对练 ") + "\n\n"

		// Navigation help
		s += "\n" + helpStyle.Render("esc/b: Back to list  a: AI Scout  p: Practice  q: Quit")
	}

	return s
}
