package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"pib/internal/model"

	"github.com/google/uuid"
)

// JSONDB is a simple JSON file-based database
type JSONDB struct {
	filePath string
	Data     *Database `json:"data"`
}

// Database holds all data
type Database struct {
	Questions   []model.Question   `json:"questions"`
	Tags        []model.Tag        `json:"tags"`
	ReviewLogs  []model.ReviewLog  `json:"review_logs"`
	QuestionTags []QuestionTag     `json:"question_tags"`
}

// QuestionTag represents a question-tag relationship
type QuestionTag struct {
	QuestionID string `json:"question_id"`
	TagID      string `json:"tag_id"`
}

// NewJSONDB creates a new JSON database
func NewJSONDB(filePath string) (*JSONDB, error) {
	jdb := &JSONDB{
		filePath: filePath,
		Data: &Database{},
	}

	// Ensure directory exists
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Load existing data if file exists
	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(data, jdb.Data); err != nil {
			return nil, err
		}
	}

	return jdb, nil
}

// Save saves the database to file
func (j *JSONDB) Save() error {
	data, err := json.MarshalIndent(j.Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(j.filePath, data, 0644)
}

// CreateQuestion creates a new question
func (j *JSONDB) CreateQuestion(q *model.Question) error {
	if q.ID == "" {
		q.ID = uuid.New().String()
	}
	q.CreatedAt = time.Now()
	q.UpdatedAt = time.Now()
	j.Data.Questions = append(j.Data.Questions, *q)
	return j.Save()
}

// GetQuestionByID retrieves a question by ID
func (j *JSONDB) GetQuestionByID(id string) (*model.Question, error) {
	for i := range j.Data.Questions {
		if j.Data.Questions[i].ID == id {
			return &j.Data.Questions[i], nil
		}
	}
	return nil, nil
}

// UpdateQuestion updates a question
func (j *JSONDB) UpdateQuestion(q *model.Question) error {
	q.UpdatedAt = time.Now()
	for i := range j.Data.Questions {
		if j.Data.Questions[i].ID == q.ID {
			j.Data.Questions[i] = *q
			return j.Save()
		}
	}
	return nil
}

// DeleteQuestion deletes a question
func (j *JSONDB) DeleteQuestion(id string) error {
	questions := make([]model.Question, 0)
	for _, q := range j.Data.Questions {
		if q.ID != id {
			questions = append(questions, q)
		}
	}
	j.Data.Questions = questions
	
	// Also remove related tags and logs
	j.Data.QuestionTags = removeQuestionTags(j.Data.QuestionTags, id)
	j.Data.ReviewLogs = removeReviewLogs(j.Data.ReviewLogs, id)
	
	return j.Save()
}

func removeQuestionTags(tags []QuestionTag, qid string) []QuestionTag {
	result := make([]QuestionTag, 0)
	for _, t := range tags {
		if t.QuestionID != qid {
			result = append(result, t)
		}
	}
	return result
}

func removeReviewLogs(logs []model.ReviewLog, qid string) []model.ReviewLog {
	result := make([]model.ReviewLog, 0)
	for _, l := range logs {
		if l.QuestionID != qid {
			result = append(result, l)
		}
	}
	return result
}

// ListQuestionsByStatus lists questions by status
func (j *JSONDB) ListQuestionsByStatus(status string) []model.Question {
	var result []model.Question
	for _, q := range j.Data.Questions {
		if q.Status == status {
			result = append(result, q)
		}
	}
	return result
}

// GetQuestionsForReview returns questions due for review
func (j *JSONDB) GetQuestionsForReview(limit int) []model.Question {
	var result []model.Question
	now := time.Now()
	for _, q := range j.Data.Questions {
		if q.Status == model.StatusActive {
			if q.NextReviewAt == nil || q.NextReviewAt.Before(now) || q.NextReviewAt.Equal(now) {
				result = append(result, q)
				if limit > 0 && len(result) >= limit {
					break
				}
			}
		}
	}
	return result
}

// CreateTag creates a new tag
func (j *JSONDB) CreateTag(t *model.Tag) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.CreatedAt = time.Now()
	
	// Check if tag exists
	for _, existing := range j.Data.Tags {
		if existing.Name == t.Name {
			*t = existing
			return nil
		}
	}
	
	j.Data.Tags = append(j.Data.Tags, *t)
	return j.Save()
}

// GetOrCreateTag gets or creates a tag
func (j *JSONDB) GetOrCreateTag(name string) (*model.Tag, error) {
	for _, t := range j.Data.Tags {
		if t.Name == name {
			return &t, nil
		}
	}
	
	tag := &model.Tag{
		ID:   uuid.New().String(),
		Name: name,
	}
	tag.CreatedAt = time.Now()
	j.Data.Tags = append(j.Data.Tags, *tag)
	return tag, j.Save()
}

// AddTagToQuestion adds a tag to a question
func (j *JSONDB) AddTagToQuestion(qid, tid string) error {
	// Check if already exists
	for _, qt := range j.Data.QuestionTags {
		if qt.QuestionID == qid && qt.TagID == tid {
			return nil
		}
	}
	j.Data.QuestionTags = append(j.Data.QuestionTags, QuestionTag{QuestionID: qid, TagID: tid})
	return j.Save()
}

// CreateReviewLog creates a review log entry
func (j *JSONDB) CreateReviewLog(r *model.ReviewLog) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	r.ReviewedAt = time.Now()
	j.Data.ReviewLogs = append(j.Data.ReviewLogs, *r)
	return j.Save()
}

// GetTagsForQuestion gets tags for a question
func (j *JSONDB) GetTagsForQuestion(qid string) []model.Tag {
	var result []model.Tag
	for _, qt := range j.Data.QuestionTags {
		if qt.QuestionID == qid {
			for _, t := range j.Data.Tags {
				if t.ID == qt.TagID {
					result = append(result, t)
					break
				}
			}
		}
	}
	return result
}
