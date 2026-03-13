package tui

import (
	"fmt"
	"os"
	"path/filepath"

	"pib/internal/model"
	"pib/internal/repository"

	"github.com/google/uuid"
)

// QuestionsData represents the JSON structure (kept for backward compatibility)
type QuestionsData struct {
	Questions    []model.Question       `json:"questions"`
	Tags         []model.Tag            `json:"tags"`
	QuestionTags []QuestionTagRelation  `json:"question_tags"`
}

// QuestionTagRelation represents a question-tag relationship in JSON
type QuestionTagRelation struct {
	QuestionID string `json:"question_id"`
	TagID      string `json:"tag_id"`
}

const dataFilePath = "data/pib.json"

// db is the global database instance
var db *repository.SQLiteDB

// initDB initializes the database connection
func initDB() error {
	if db != nil {
		return nil
	}

	// Get current working directory
	wd, err := getWorkingDir()
	if err != nil {
		return err
	}

	// Try multiple paths to find the database
	dbPaths := []string{
		filepath.Join(wd, "data", "pib.db"),
		filepath.Join(wd, "..", "data", "pib.db"),
		filepath.Join(wd, "..", "..", "data", "pib.db"),
		filepath.Join(wd, "..", "..", "..", "data", "pib.db"),
	}

	for _, dbPath := range dbPaths {
		db, err = repository.NewSQLiteDB(dbPath)
		if err == nil {
			return nil
		}
	}

	// If all paths fail, return the last error
	return err
}

// loadQuestionsFromFile loads questions from SQLite database
func loadQuestionsFromFile() ([]model.Question, error) {
	// Initialize database if needed
	if err := initDB(); err != nil {
		// Fall back to JSON file if SQLite fails
		fmt.Fprintf(os.Stderr, "[DEBUG] initDB failed, falling back to JSON: %v\n", err)
		return loadQuestionsFromJSON()
	}

	// Load questions from SQLite
	questions, err := db.ListQuestionsByStatus("")
	if err != nil {
		// Fall back to JSON on error
		fmt.Fprintf(os.Stderr, "[DEBUG] ListQuestionsByStatus error, falling back to JSON: %v\n", err)
		return loadQuestionsFromJSON()
	}

	fmt.Fprintf(os.Stderr, "[DEBUG] Loaded %d questions from SQLite\n", len(questions))

	// If no questions in DB, try JSON (for backward compatibility)
	if len(questions) == 0 {
		fmt.Fprintf(os.Stderr, "[DEBUG] SQLite empty, falling back to JSON\n")
		return loadQuestionsFromJSON()
	}

	return questions, nil
}

// loadQuestionsFromJSON loads questions from the JSON file (fallback)
func loadQuestionsFromJSON() ([]model.Question, error) {
	// Get current working directory
	wd, err := getWorkingDir()
	if err != nil {
		return nil, err
	}

	// Try multiple paths to find the JSON file
	paths := []string{
		filepath.Join(wd, "data", "pib.json"),
		filepath.Join(wd, "..", "data", "pib.json"),
		filepath.Join(wd, "..", "..", "data", "pib.json"),
		filepath.Join(wd, "..", "..", "..", "data", "pib.json"),
	}

	var data []byte
	for _, p := range paths {
		data, err = readFile(p)
		if err == nil {
			break
		}
	}

	if data == nil {
		// Return empty list if file doesn't exist
		return []model.Question{}, nil
	}

	var questionsData QuestionsData
	if err := unmarshalJSON(data, &questionsData); err != nil {
		return nil, err
	}

	// Build tag lookup map
	tagMap := make(map[string]model.Tag)
	for _, tag := range questionsData.Tags {
		tagMap[tag.ID] = tag
	}

	// Build question-tag relationships
	qtMap := make(map[string][]model.Tag)
	for _, qt := range questionsData.QuestionTags {
		if tag, ok := tagMap[qt.TagID]; ok {
			qtMap[qt.QuestionID] = append(qtMap[qt.QuestionID], tag)
		}
	}

	// Attach tags to questions
	for i := range questionsData.Questions {
		qid := questionsData.Questions[i].ID
		if tags, ok := qtMap[qid]; ok {
			questionsData.Questions[i].Tags = tags
		}
	}

	return questionsData.Questions, nil
}

// loadQuestionByID loads a single question by ID from SQLite
func loadQuestionByID(id string) *model.Question {
	// Try SQLite first
	if db != nil {
		q, err := db.GetQuestionByID(id)
		if err == nil && q != nil {
			return q
		}
	}

	// Fall back to JSON
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

// saveQuestion saves a new question to SQLite
func saveQuestion(content, answer, tagsStr string) error {
	// Initialize database if needed
	if err := initDB(); err != nil {
		// Fall back to JSON file if SQLite fails
		return saveQuestionToJSON(content, answer, tagsStr)
	}

	// Create new question
	q := &model.Question{
		ID:      uuid.New().String(),
		Content: content,
		Answer:  answer,
		Status:  model.StatusDraft,
		EF:      2.5,
	}

	if err := db.CreateQuestion(q); err != nil {
		return err
	}

	// Handle tags
	if tagsStr != "" {
		tagNames := parseTags(tagsStr)
		for _, name := range tagNames {
			tag, err := db.GetOrCreateTag(name)
			if err == nil && tag != nil {
				db.AddTagToQuestion(q.ID, tag.ID)
			}
		}
	}

	return nil
}

// saveQuestionToJSON saves a new question to the JSON file (fallback)
func saveQuestionToJSON(content, answer, tagsStr string) error {
	// Get current working directory
	wd, err := getWorkingDir()
	if err != nil {
		return err
	}

	// Try different paths (from cmd/tui)
	paths := []string{
		filepath.Join(wd, "..", "..", dataFilePath),
		filepath.Join(wd, "..", dataFilePath),
		filepath.Join(wd, dataFilePath),
	}

	var filePath string
	for _, p := range paths {
		_, err := osStat(p)
		if err == nil {
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
	if data, err := readFile(filePath); err == nil {
		unmarshalJSON(data, &questionsData)
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
	data, err := marshalJSON(questionsData)
	if err != nil {
		return err
	}

	return writeFile(filePath, data)
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

// GetDB returns the database instance
func GetDB() *repository.SQLiteDB {
	if db == nil {
		initDB()
	}
	return db
}
