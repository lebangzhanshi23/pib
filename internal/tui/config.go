package tui

import (
	"fmt"
	"path/filepath"

	"pib/config"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// ConfigModel handles LLM configuration
type ConfigModel struct {
	providerInput textinput.Model
	apiKeyInput   textinput.Model
	modelInput    textinput.Model
	focus         int // 0: provider, 1: apiKey, 2: model, 3: save
	Completed     bool
	Cancelled     bool
	err           error
	cfg           *config.Config
}

// NewConfigModel creates a new config model
func NewConfigModel() *ConfigModel {
	// Load existing config
	cfg, _ := config.Load(getConfigPath())

	providerInput := textinput.New()
	providerInput.Placeholder = "deepseek or openai"
	providerInput.Prompt = "Provider: "
	if cfg != nil && cfg.LLM.Provider != "" {
		providerInput.SetValue(cfg.LLM.Provider)
	} else {
		providerInput.Focus()
	}

	apiKeyInput := textinput.New()
	apiKeyInput.Placeholder = "Your API key (will be masked)"
	apiKeyInput.EchoMode = textinput.EchoPassword
	apiKeyInput.Prompt = "API Key: "
	if cfg != nil && cfg.LLM.APIKey != "" {
		apiKeyInput.SetValue(cfg.LLM.APIKey)
	}

	modelInput := textinput.New()
	modelInput.Placeholder = "deepseek-chat or gpt-4o-mini"
	modelInput.Prompt = "Model: "
	if cfg != nil && cfg.LLM.Model != "" {
		modelInput.SetValue(cfg.LLM.Model)
	} else {
		modelInput.SetValue("deepseek-chat")
	}

	return &ConfigModel{
		providerInput: providerInput,
		apiKeyInput:   apiKeyInput,
		modelInput:    modelInput,
		focus:        0,
		cfg:           cfg,
	}
}

// getConfigPath returns the path to config.yaml
func getConfigPath() string {
	wd, _ := getWorkingDir()
	paths := []string{
		filepath.Join(wd, "config", "config.yaml"),
		filepath.Join(wd, "..", "config", "config.yaml"),
		filepath.Join(wd, "..", "..", "config", "config.yaml"),
	}
	for _, p := range paths {
		if _, err := osStat(p); err == nil {
			return p
		}
	}
	return paths[0]
}

// Init initializes the model
func (m *ConfigModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *ConfigModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			m.focus = (m.focus + 1) % 4
			switch m.focus {
			case 0:
				m.providerInput.Focus()
			case 1:
				m.apiKeyInput.Focus()
			case 2:
				m.modelInput.Focus()
			case 3:
				// Save button
			}
		case "enter":
			if m.focus < 3 {
				m.focus++
				switch m.focus {
				case 0:
					m.providerInput.Focus()
				case 1:
					m.apiKeyInput.Focus()
				case 2:
					m.modelInput.Focus()
				}
			} else {
				// Save config
				m.saveConfig()
			}
		case "esc":
			m.Cancelled = true
		}
	}

	m.providerInput, _ = m.providerInput.Update(msg)
	m.apiKeyInput, _ = m.apiKeyInput.Update(msg)
	m.modelInput, _ = m.modelInput.Update(msg)

	return m, nil
}

// saveConfig saves the configuration to file
func (m *ConfigModel) saveConfig() error {
	provider := m.providerInput.Value()
	apiKey := m.apiKeyInput.Value()
	model := m.modelInput.Value()

	if provider == "" {
		m.err = fmt.Errorf("provider is required")
		return m.err
	}

	// Load existing config or create new
	cfg := &config.Config{
		App: config.AppConfig{
			Name: "PIB",
			Port: 8081,
		},
		Database: config.DatabaseConfig{
			Path: "./data/pib.db",
		},
		LLM: config.LLMConfig{
			Provider: provider,
			APIKey:   apiKey,
			Model:    model,
		},
		SRS: config.SRSConfig{
			InitialEF: 2.5,
			MinEF:     1.3,
		},
	}

	// Write config
	cfgPath := getConfigPath()
	data, err := config.MarshalYAML(cfg)
	if err != nil {
		m.err = err
		return err
	}

	err = writeFile(cfgPath, data)
	if err != nil {
		m.err = err
		return err
	}

	m.Completed = true
	return nil
}

// View renders the config form
func (m *ConfigModel) View() string {
	s := ""

	header := titleStyle.Render("⚙️ LLM Configuration")
	s += header + "\n\n"

	// Provider input
	providerLabel := normalStyle.Render("Provider (deepseek/openai):")
	if m.focus == 0 {
		providerLabel = selectedStyle.Render("Provider (deepseek/openai):")
	}
	s += providerLabel + "\n"
	s += m.providerInput.View() + "\n\n"

	// API Key input
	apiKeyLabel := normalStyle.Render("API Key:")
	if m.focus == 1 {
		apiKeyLabel = selectedStyle.Render("API Key:")
	}
	s += apiKeyLabel + "\n"
	s += m.apiKeyInput.View() + "\n\n"

	// Model input
	modelLabel := normalStyle.Render("Model:")
	if m.focus == 2 {
		modelLabel = selectedStyle.Render("Model:")
	}
	s += modelLabel + "\n"
	s += m.modelInput.View() + "\n\n"

	// Save button
	saveLabel := "  Save  "
	if m.focus == 3 {
		saveLabel = buttonActiveStyle.Render("  Save  ")
	} else {
		saveLabel = buttonStyle.Render("  Save  ")
	}
	s += saveLabel + "\n\n"

	// Help
	s += helpStyle.Render("Tab: Next field  Enter: Confirm/Save  esc: Cancel")

	// Error
	if m.err != nil {
		s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Success
	if m.Completed {
		s += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render("✓ Configuration saved!")
	}

	return s
}
