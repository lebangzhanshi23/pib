package tui

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"pib/internal/model"

	"github.com/google/uuid"
)

// MarkdownQuestion represents a parsed question from Markdown
type MarkdownQuestion struct {
	Content string
	Answer  string
	Tags    []string
}

// ImportResult represents the result of an import operation
type ImportResult struct {
	TotalQuestions int
	ImportedCount  int
	Errors         []string
}

// ImportFromMarkdown imports questions from Markdown files
func ImportFromMarkdown(filePaths []string) (*ImportResult, error) {
	result := &ImportResult{
		TotalQuestions: 0,
		ImportedCount:  0,
		Errors:         []string{},
	}

	// Initialize database if needed
	if err := initDB(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	for _, filePath := range filePaths {
		questions, err := parseMarkdownFile(filePath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to parse %s: %v", filePath, err))
			continue
		}

		result.TotalQuestions += len(questions)

		for _, q := range questions {
			err := importQuestion(q.Content, q.Answer, q.Tags)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("Failed to import question: %v", err))
				continue
			}
			result.ImportedCount++
		}
	}

	return result, nil
}

// parseMarkdownFile parses a Markdown file and extracts questions
func parseMarkdownFile(filePath string) ([]MarkdownQuestion, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	content := string(data)
	questions := []MarkdownQuestion{}

	// Split by --- separator
	sections := strings.Split(content, "\n---\n")

	for _, section := range sections {
		section = strings.TrimSpace(section)
		if section == "" {
			continue
		}

		q, err := parseMarkdownSection(section)
		if err != nil {
			continue // Skip invalid sections
		}
		if q.Content != "" {
			questions = append(questions, q)
		}
	}

	// If no sections found, try to parse as single question
	if len(questions) == 0 {
		q, err := parseMarkdownSection(content)
		if err == nil && q.Content != "" {
			questions = append(questions, q)
		}
	}

	return questions, nil
}

// parseMarkdownSection parses a single Markdown section into a question
func parseMarkdownSection(section string) (MarkdownQuestion, error) {
	q := MarkdownQuestion{}
	lines := strings.Split(section, "\n")

	state := "question" // question, answer, tags
	var currentContent []string

	for _, line := range lines {
		line = strings.TrimRight(line, "\r") // Handle Windows line endings

		// Check for headers
		if strings.HasPrefix(line, "# ") {
			// Save previous content
			if len(currentContent) > 0 {
				switch state {
				case "question":
					q.Content = strings.TrimSpace(strings.Join(currentContent, "\n"))
				case "answer":
					q.Answer = strings.TrimSpace(strings.Join(currentContent, "\n"))
				}
			}

			// Check if it's the answer section
			if strings.Contains(strings.ToLower(line), "answer") || strings.Contains(strings.ToLower(line), "答案") {
				state = "answer"
				currentContent = []string{}
				continue
			}

			// Check if it's tags
			if strings.Contains(strings.ToLower(line), "tag") || strings.Contains(strings.ToLower(line), "标签") {
				state = "tags"
				currentContent = []string{}
				continue
			}

			// It's a new question
			state = "question"
			currentContent = []string{}
			continue
		}

		// Check for Tags: line (frontmatter style)
		if strings.HasPrefix(line, "Tags:") || strings.HasPrefix(line, "tags:") {
			tagsStr := strings.TrimPrefix(line, "Tags:")
			tagsStr = strings.TrimPrefix(tagsStr, "tags:")
			q.Tags = parseTags(tagsStr)
			continue
		}

		// Check for Tags: line (markdown header style)
		if strings.HasPrefix(line, "## Tags") || strings.HasPrefix(line, "## tags") {
			state = "tags"
			continue
		}

		// Regular content
		currentContent = append(currentContent, line)
	}

	// Save last content
	if len(currentContent) > 0 {
		switch state {
		case "question":
			q.Content = strings.TrimSpace(strings.Join(currentContent, "\n"))
		case "answer":
			q.Answer = strings.TrimSpace(strings.Join(currentContent, "\n"))
		case "tags":
			tagsStr := strings.TrimSpace(strings.Join(currentContent, ","))
			q.Tags = parseTags(tagsStr)
		}
	}

	return q, nil
}

// importQuestion imports a single question into the database
func importQuestion(content, answer string, tags []string) error {
	q := &model.Question{
		ID:      uuid.New().String(),
		Content: content,
		Answer:  answer,
		Status:  model.StatusDraft,
		EF:      2.5,
	}

	if err := db.CreateQuestion(q); err != nil {
		return fmt.Errorf("failed to create question: %v", err)
	}

	// Handle tags
	for _, tagName := range tags {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}
		tag, err := db.GetOrCreateTag(tagName)
		if err == nil && tag != nil {
			db.AddTagToQuestion(q.ID, tag.ID)
		}
	}

	return nil
}

// ImportFromMarkdownFile imports questions from a single Markdown file (for TUI)
func ImportFromMarkdownFile(filePath string) (*ImportResult, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %v", err)
	}

	if _, err := os.Stat(absPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file does not exist: %s", absPath)
	}

	return ImportFromMarkdown([]string{absPath})
}

// ImportFromDirectory imports all Markdown files from a directory
func ImportFromDirectory(dirPath string) (*ImportResult, error) {
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, fmt.Errorf("invalid directory path: %v", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("directory does not exist: %v", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", absPath)
	}

	// Find all .md files
	var mdFiles []string
	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan directory: %v", err)
	}

	if len(mdFiles) == 0 {
		return &ImportResult{
			TotalQuestions: 0,
			ImportedCount:  0,
			Errors:         []string{"No Markdown files found in directory"},
		}, nil
	}

	return ImportFromMarkdown(mdFiles)
}

// ReadMarkdownPreview reads a Markdown file and returns a preview
func ReadMarkdownPreview(filePath string) (string, error) {
	_, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Count questions
	questions, err := parseMarkdownFile(filePath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("File: %s\n", filepath.Base(filePath)))
	buf.WriteString(fmt.Sprintf("Questions found: %d\n\n", len(questions)))

	for i, q := range questions {
		if i >= 5 { // Show max 5 previews
			buf.WriteString(fmt.Sprintf("... and %d more questions", len(questions)-5))
			break
		}
		buf.WriteString(fmt.Sprintf("%d. %s\n", i+1, truncateString(q.Content, 50)))
		if len(q.Tags) > 0 {
			buf.WriteString(fmt.Sprintf("   Tags: %s\n", strings.Join(q.Tags, ", ")))
		}
		buf.WriteString("\n")
	}

	return buf.String(), nil
}

func truncateString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GetMarkdownFilesInDirectory returns a list of Markdown files in a directory
func GetMarkdownFilesInDirectory(dirPath string) ([]string, error) {
	absPath, err := filepath.Abs(dirPath)
	if err != nil {
		return nil, fmt.Errorf("invalid directory path: %v", err)
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("directory does not exist: %v", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", absPath)
	}

	var mdFiles []string
	err = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".md") {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})

	return mdFiles, err
}

// AskUserForImportPath asks the user to input a file or directory path (for non-interactive use)
func AskUserForImportPath() (string, error) {
	// This is a placeholder - in the TUI we'll handle this interactively
	return "", nil
}

// ImportFromStdin reads Markdown content from stdin
func ImportFromStdin() (*ImportResult, error) {
	scanner := bufio.NewScanner(os.Stdin)
	var content strings.Builder
	for scanner.Scan() {
		content.WriteString(scanner.Text())
		content.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read stdin: %v", err)
	}

	// Parse as a single question
	q, err := parseMarkdownSection(content.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse content: %v", err)
	}

	if q.Content == "" {
		return &ImportResult{
			TotalQuestions: 0,
			ImportedCount:  0,
			Errors:         []string{"No content found in stdin"},
		}, nil
	}

	// Initialize database
	if err := initDB(); err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	result := &ImportResult{
		TotalQuestions: 1,
		ImportedCount:  0,
		Errors:         []string{},
	}

	err = importQuestion(q.Content, q.Answer, q.Tags)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}

	result.ImportedCount = 1
	return result, nil
}
