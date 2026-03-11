package tui

import (
	"encoding/json"
	"os"
	"path/filepath"

	"pib/internal/model"

	"github.com/google/uuid"
)

// QuestionsData represents the JSON structure
type QuestionsData struct {
	Questions []model.Question `json:"questions"`
	Tags      []model.Tag       `json:"tags"`
}

const dataFilePath = "data/pib.json"

// loadQuestionsFromFile loads questions from the JSON file
func loadQuestionsFromFile() ([]model.Question, error) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Try different paths
	paths := []string{
		filepath.Join(wd, dataFilePath),
		filepath.Join(wd, "..", dataFilePath),
		filepath.Join(wd, "..", "..", dataFilePath),
	}

	var data []byte
	for _, p := range paths {
		data, err = os.ReadFile(p)
		if err == nil {
			break
		}
	}

	if data == nil {
		// Return empty list if file doesn't exist
		return []model.Question{}, nil
	}

	var questionsData QuestionsData
	if err := json.Unmarshal(data, &questionsData); err != nil {
		return nil, err
	}

	return questionsData.Questions, nil
}

// loadQuestionByID loads a single question by ID
func loadQuestionByID(id string) *model.Question {
	questions, err := loadQuestionsFromFile()
	if err != nil {
		return nil
	}

	for i := range questions {
		if questions[i].ID == id {
			return &questions[i]
		}
	}

	return nil
}

// saveQuestion saves a new question to the JSON file
func saveQuestion(content, answer, tagsStr string) error {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Try different paths
	paths := []string{
		filepath.Join(wd, dataFilePath),
		filepath.Join(wd, "..", dataFilePath),
		filepath.Join(wd, "..", "..", dataFilePath),
	}

	var filePath string
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			filePath = p
			break
		}
	}

	if filePath == "" {
		// Use default path
		filePath = paths[0]
	}

	// Read existing data
	var questionsData QuestionsData
	if data, err := os.ReadFile(filePath); err == nil {
		json.Unmarshal(data, &questionsData)
	}

	// Create new question
	q := model.Question{
		ID:      uuid.New().String(),
		Content: content,
		Answer:  answer,
		Status:  "draft",
		EF:      2.5,
	}

	// Parse tags
	if tagsStr != "" {
		// Simple tag parsing - in real app, would handle this better
		// Just add to question's Tags field
		tagNames := parseTags(tagsStr)
		for _, name := range tagNames {
			// Check if tag exists
			tagFound := false
			for _, tag := range questionsData.Tags {
				if tag.Name == name {
					q.Tags = append(q.Tags, tag)
					tagFound = true
					break
				}
			}
			if !tagFound {
				newTag := model.Tag{
					ID:   uuid.New().String(),
					Name: name,
				}
				questionsData.Tags = append(questionsData.Tags, newTag)
				q.Tags = append(q.Tags, newTag)
			}
		}
	}

	questionsData.Questions = append(questionsData.Questions, q)

	// Write back to file
	data, err := json.MarshalIndent(questionsData, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// parseTags parses comma-separated tags
func parseTags(tagsStr string) []string {
	var tags []string
	var current string
	for _, c := range tagsStr {
		if c == ',' {
			if current != "" {
				tags = append(tags, trimSpace(current))
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		tags = append(tags, trimSpace(current))
	}
	return tags
}

func trimSpace(s string) string {
	// Simple trim
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}
