package tui

import (
	"github.com/charmbracelet/bubbletea"
)

// PageType represents the type of page in the TUI
type PageType int

const (
	PageList PageType = iota
	PageDetail
	PageAdd
	PageConfig
	PageImport
	PagePractice
)

// MainModel is the root model that manages page navigation
type MainModel struct {
	currentPage   PageType
	listModel     *QuestionListModel
	detailModel   *QuestionDetailModel
	addModel      *AddQuestionModel
	configModel   *ConfigModel
	importModel   *ImportPageModel
	practiceModel *PracticeModel
}

// NewMainModel creates a new main model
func NewMainModel() *MainModel {
	return &MainModel{
		currentPage:   PageList,
		listModel:     NewQuestionListModel(),
		detailModel:   NewQuestionDetailModel(),
		addModel:      NewAddQuestionModel(),
		configModel:   NewConfigModel(),
		importModel:   NewImportPageModel(),
		practiceModel: NewPracticeModel(),
	}
}

// Init initializes the model
func (m *MainModel) Init() tea.Cmd {
	return m.listModel.LoadQuestions()
}

// Update handles messages
func (m *MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			// Go back to list from detail or add
			if m.currentPage != PageList {
				m.currentPage = PageList
				return m, m.listModel.LoadQuestions()
			}
		}
	}

	switch m.currentPage {
	case PageList:
		list, cmd := m.listModel.Update(msg)
		m.listModel = list.(*QuestionListModel)
		// Check if a question was selected
		if m.listModel.SelectedID != "" {
			// Find the selected question content
			for _, q := range m.listModel.questions {
				if q.ID == m.listModel.SelectedID {
					m.detailModel.SetQuestion(m.listModel.SelectedID, q.Content)
					break
				}
			}
			m.currentPage = PageDetail
			m.listModel.SelectedID = ""
		}
		// Check if add was triggered
		if m.listModel.AddingNew {
			m.currentPage = PageAdd
			m.listModel.AddingNew = false
		}
		// Check if import was triggered
		if m.listModel.OpenImport {
			m.currentPage = PageImport
			m.listModel.OpenImport = false
		}
		// Check if config was triggered
		if m.listModel.OpenConfig {
			m.currentPage = PageConfig
			m.listModel.OpenConfig = false
		}
		return m, cmd

	case PageDetail:
		detail, cmd := m.detailModel.Update(msg)
		m.detailModel = detail.(*QuestionDetailModel)
		// Check if we should go back to list
		if m.detailModel.BackToList {
			m.currentPage = PageList
			m.detailModel.BackToList = false
			return m, m.listModel.LoadQuestions()
		}
		// Check if we should start practice
		if m.detailModel.StartPractice {
			m.detailModel.StartPractice = false
			// Set up practice model with current question
			m.practiceModel.SetQuestion(m.detailModel.questionID, m.detailModel.content, m.detailModel.answer)
			m.currentPage = PagePractice
		}
		return m, cmd

	case PageAdd:
		add, cmd := m.addModel.Update(msg)
		m.addModel = add.(*AddQuestionModel)
		// Check if add was completed or cancelled
		if m.addModel.Completed || m.addModel.Cancelled {
			m.currentPage = PageList
			m.addModel = NewAddQuestionModel()
			return m, m.listModel.LoadQuestions()
		}
		return m, cmd

	case PageConfig:
		cfg, cmd := m.configModel.Update(msg)
		m.configModel = cfg.(*ConfigModel)
		// Check if config was completed or cancelled
		if m.configModel.Completed || m.configModel.Cancelled {
			m.currentPage = PageList
			m.configModel = NewConfigModel()
		}
		return m, cmd

	case PageImport:
		imp, cmd := m.importModel.Update(msg)
		m.importModel = imp.(*ImportPageModel)
		// Check if import was completed or cancelled
		if m.importModel.Completed || m.importModel.Cancelled {
			m.currentPage = PageList
			m.importModel = NewImportPageModel()
			return m, m.listModel.LoadQuestions()
		}
		return m, cmd

	case PagePractice:
		practice, cmd := m.practiceModel.Update(msg)
		m.practiceModel = practice.(*PracticeModel)
		// Check if practice was completed or cancelled
		if m.practiceModel.Completed || m.practiceModel.Cancelled {
			m.currentPage = PageDetail
			m.practiceModel = NewPracticeModel()
		}
		return m, cmd
	}

	return m, nil
}

// View renders the UI
func (m *MainModel) View() string {
	switch m.currentPage {
	case PageList:
		return m.listModel.View()
	case PageDetail:
		return m.detailModel.View()
	case PageAdd:
		return m.addModel.View()
	case PageConfig:
		return m.configModel.View()
	case PageImport:
		return m.importModel.View()
	case PagePractice:
		return m.practiceModel.View()
	default:
		return "Unknown page"
	}
}
