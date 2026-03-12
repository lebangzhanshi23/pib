package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"pib/config"
)

// LLMClient handles LLM API calls
type LLMClient struct {
	cfg     *config.Config
	client  *http.Client
}

// NewLLMClient creates a new LLM client
func NewLLMClient(cfg *config.Config) *LLMClient {
	return &LLMClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

// Chat calls the LLM API
func (c *LLMClient) Chat(messages []Message) (string, error) {
	if c.cfg.LLM.Provider == "" {
		return "", fmt.Errorf("LLM provider not configured")
	}

	req := ChatRequest{
		Model:    c.cfg.LLM.Model,
		Messages: messages,
	}

	var url string
	var headers map[string]string

	switch c.cfg.LLM.Provider {
	case "deepseek":
		url = "https://api.deepseek.com/v1/chat/completions"
		headers = map[string]string{
			"Authorization": "Bearer " + c.cfg.LLM.APIKey,
			"Content-Type":  "application/json",
		}
	case "openai":
		url = "https://api.openai.com/v1/chat/completions"
		headers = map[string]string{
			"Authorization": "Bearer " + c.cfg.LLM.APIKey,
			"Content-Type":  "application/json",
		}
	default:
		return "", fmt.Errorf("unsupported provider: %s", c.cfg.LLM.Provider)
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error: %s", string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", err
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return chatResp.Choices[0].Message.Content, nil
}

// GetLLMConfigPath returns the path to config.yaml
func GetLLMConfigPath() string {
	// Try multiple paths
	paths := []string{
		"config/config.yaml",
		"../config/config.yaml",
		"../../config/config.yaml",
		"../../../config/config.yaml",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return "config/config.yaml"
}

// LoadConfig loads LLM configuration
func LoadConfig() (*config.Config, error) {
	path := GetLLMConfigPath()
	return config.Load(path)
}
