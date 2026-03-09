package model

import "time"

// Question represents a面试 question
type Question struct {
	ID            string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Content       string    `gorm:"type:text;not null" json:"content"`
	Answer        string    `gorm:"type:text" json:"answer"`
	Summary       string    `gorm:"type:text" json:"summary"`
	EF            float64   `gorm:"default:2.5" json:"ef"`            // Ease Factor
	Interval      int       `gorm:"default:0" json:"interval"`          // Days until next review
	NextReviewAt *time.Time `json:"next_review_at" json:"next_review_at"`
	Status       string    `gorm:"default:'draft'" json:"status"`      // draft, active, archived
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Tags         []Tag     `gorm:"many2many:question_tags;" json:"tags,omitempty"`
}

// Tag represents a category tag
type Tag struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

// ReviewLog represents a review history entry
type ReviewLog struct {
	ID          string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	QuestionID  string    `gorm:"type:varchar(36);index;not null" json:"question_id"`
	Grade       int       `gorm:"not null" json:"grade"` // 0: Forgot, 1: Vague, 2: Remembered
	ReviewedAt time.Time `json:"reviewed_at"`
}

// Grade constants
const (
	GradeForgot    = 0
	GradeVague     = 1
	GradeRemembered = 2
)

// QuestionStatus constants
const (
	StatusDraft    = "draft"
	StatusActive   = "active"
	StatusArchived = "archived"
)
