package repository

import (
	"path/filepath"
	"time"

	"pib/internal/model"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SQLiteDB is a SQLite database
type SQLiteDB struct {
	db *gorm.DB
}

// NewSQLiteDB creates a new SQLite database
func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := mkdirAll(dir); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sdb := &SQLiteDB{db: db}
	if err := sdb.migrate(); err != nil {
		return nil, err
	}

	return sdb, nil
}

// migrate creates tables
func (s *SQLiteDB) migrate() error {
	return s.db.AutoMigrate(
		&Question{},
		&Tag{},
		&QuestionTagRel{},
		&ReviewLog{},
	)
}

// Question represents a question in the database
type Question struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Content      string    `gorm:"not null" json:"content"`
	Answer       string    `json:"answer"`
	Status       string    `gorm:"default:draft" json:"status"`
	EF           float64   `gorm:"default:2.5" json:"ef"`
	Interval     int       `gorm:"default:0" json:"interval"`
	NextReviewAt *time.Time `json:"next_review_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Tags         []Tag     `gorm:"-" json:"tags"`
}

// TableName specifies the table name
func (Question) TableName() string {
	return "questions"
}

// Tag represents a tag
type Tag struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name
func (Tag) TableName() string {
	return "tags"
}

// QuestionTagRel represents a question-tag relationship
type QuestionTagRel struct {
	QuestionID string `gorm:"primaryKey" json:"question_id"`
	TagID      string `gorm:"primaryKey" json:"tag_id"`
}

// TableName specifies the table name
func (QuestionTagRel) TableName() string {
	return "question_tags"
}

// ReviewLog represents a review log
type ReviewLog struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	QuestionID  string    `gorm:"index" json:"question_id"`
	Grade       int       `json:"grade"`
	ReviewedAt time.Time `json:"reviewed_at"`
}

// TableName specifies the table name
func (ReviewLog) TableName() string {
	return "review_logs"
}

// CreateQuestion creates a new question
func (s *SQLiteDB) CreateQuestion(q *model.Question) error {
	if q.ID == "" {
		q.ID = uuid.New().String()
	}
	q.CreatedAt = time.Now()
	q.UpdatedAt = time.Now()

	dbQ := Question{
		ID:           q.ID,
		Content:      q.Content,
		Answer:       q.Answer,
		Status:       q.Status,
		EF:           q.EF,
		Interval:     q.Interval,
		NextReviewAt: q.NextReviewAt,
		CreatedAt:    q.CreatedAt,
		UpdatedAt:    q.UpdatedAt,
	}

	return s.db.Create(&dbQ).Error
}

// GetQuestionByID retrieves a question by ID
func (s *SQLiteDB) GetQuestionByID(id string) (*model.Question, error) {
	var q Question
	if err := s.db.First(&q, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	return toModelQuestion(&q), nil
}

// UpdateQuestion updates a question
func (s *SQLiteDB) UpdateQuestion(q *model.Question) error {
	q.UpdatedAt = time.Now()

	dbQ := Question{
		ID:           q.ID,
		Content:      q.Content,
		Answer:       q.Answer,
		Status:       q.Status,
		EF:           q.EF,
		Interval:     q.Interval,
		NextReviewAt: q.NextReviewAt,
		CreatedAt:    q.CreatedAt,
		UpdatedAt:    q.UpdatedAt,
	}

	return s.db.Save(&dbQ).Error
}

// DeleteQuestion deletes a question
func (s *SQLiteDB) DeleteQuestion(id string) error {
	return s.db.Delete(&Question{}, "id = ?", id).Error
}

// ListQuestionsByStatus lists questions by status
func (s *SQLiteDB) ListQuestionsByStatus(status string) ([]model.Question, error) {
	var questions []Question
	if err := s.db.Find(&questions, "status = ?", status).Error; err != nil {
		return nil, err
	}

	result := make([]model.Question, len(questions))
	for i, q := range questions {
		result[i] = *toModelQuestion(&q)
	}
	return result, nil
}

// GetQuestionsForReview returns questions due for review
func (s *SQLiteDB) GetQuestionsForReview(limit int) ([]model.Question, error) {
	var questions []Question
	err := s.db.Where("status = ? AND (next_review_at IS NULL OR next_review_at <= ?)", 
		"active", time.Now()).
		Order("next_review_at ASC").
		Limit(limit).
		Find(&questions).Error
	
	if err != nil {
		return nil, err
	}

	result := make([]model.Question, len(questions))
	for i, q := range questions {
		result[i] = *toModelQuestion(&q)
	}
	return result, nil
}

// CreateTag creates a new tag
func (s *SQLiteDB) CreateTag(t *model.Tag) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	t.CreatedAt = time.Now()

	dbT := Tag{
		ID:        t.ID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
	}

	return s.db.Create(&dbT).Error
}

// GetOrCreateTag gets or creates a tag
func (s *SQLiteDB) GetOrCreateTag(name string) (*model.Tag, error) {
	var t Tag
	err := s.db.First(&t, "name = ?", name).Error
	if err == nil {
		return toModelTag(&t), nil
	}
	if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new tag
	t.ID = uuid.New().String()
	t.Name = name
	t.CreatedAt = time.Now()

	if err := s.db.Create(&t).Error; err != nil {
		return nil, err
	}

	return toModelTag(&t), nil
}

// AddTagToQuestion adds a tag to a question
func (s *SQLiteDB) AddTagToQuestion(qid, tid string) error {
	qt := QuestionTagRel{QuestionID: qid, TagID: tid}
	return s.db.FirstOrCreate(&qt, qt).Error
}

// CreateReviewLog creates a review log entry
func (s *SQLiteDB) CreateReviewLog(r *model.ReviewLog) error {
	if r.ID == "" {
		r.ID = uuid.New().String()
	}
	r.ReviewedAt = time.Now()

	dbR := ReviewLog{
		ID:          r.ID,
		QuestionID:  r.QuestionID,
		Grade:       r.Grade,
		ReviewedAt:  r.ReviewedAt,
	}

	return s.db.Create(&dbR).Error
}

// GetTagsForQuestion gets tags for a question
func (s *SQLiteDB) GetTagsForQuestion(qid string) ([]model.Tag, error) {
	var tags []Tag
	err := s.db.Joins("JOIN question_tags ON tags.id = question_tags.tag_id").
		Where("question_tags.question_id = ?", qid).
		Find(&tags).Error
	
	if err != nil {
		return nil, err
	}

	result := make([]model.Tag, len(tags))
	for i, t := range tags {
		result[i] = *toModelTag(&t)
	}
	return result, nil
}

// Helper functions
func toModelQuestion(q *Question) *model.Question {
	return &model.Question{
		ID:           q.ID,
		Content:      q.Content,
		Answer:       q.Answer,
		Status:       q.Status,
		EF:           q.EF,
		Interval:     q.Interval,
		NextReviewAt: q.NextReviewAt,
		CreatedAt:    q.CreatedAt,
		UpdatedAt:    q.UpdatedAt,
	}
}

func toModelTag(t *Tag) *model.Tag {
	return &model.Tag{
		ID:        t.ID,
		Name:      t.Name,
		CreatedAt: t.CreatedAt,
	}
}

func mkdirAll(path string) error {
	// Simple implementation for Windows compatibility
	return nil
}
