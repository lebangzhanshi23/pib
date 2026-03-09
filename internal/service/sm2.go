package service

import (
	"math"
	"time"

	"pib/internal/model"
)

// SM2Calculator calculates next review time based on SM-2 algorithm
type SM2Calculator struct {
	InitialEF float64
	MinEF     float64
}

// NewSM2Calculator creates a new SM-2 calculator
func NewSM2Calculator(initialEF, minEF float64) *SM2Calculator {
	return &SM2Calculator{
		InitialEF: initialEF,
		MinEF:     minEF,
	}
}

// ReviewResult contains the result of a review calculation
type ReviewResult struct {
	NewEF        float64   `json:"ef"`
	NewInterval  int       `json:"interval"`
	NextReviewAt time.Time `json:"next_review_at"`
}

// Calculate calculates the next review based on grade
//
// Grade levels:
//   - 0 (Forgot): Reset interval to 1 day, decrease EF
//   - 1 (Vague): Multiply interval by 1.2, keep EF
//   - 2 (Remembered): Multiply interval by EF, increase EF
func (s *SM2Calculator) Calculate(currentEF float64, currentInterval int, grade int) ReviewResult {
	var newEF float64
	var newInterval int

	switch grade {
	case model.GradeForgot:
		// Forgot: reset to 1 day, decrease EF
		newEF = currentEF + (0.1 - (2 - float64(model.GradeForgot)) * 0.08)
		newInterval = 1

	case model.GradeVague:
		// Vague: multiply by 1.2, keep EF
		newEF = currentEF
		newInterval = int(math.Ceil(float64(currentInterval) * 1.2))

	case model.GradeRemembered:
		// Remembered: multiply by EF, increase EF
		newEF = currentEF + (0.1 - (2 - float64(model.GradeRemembered)) * 0.08)
		if currentInterval == 0 {
			newInterval = 1
		} else {
			newInterval = int(math.Ceil(float64(currentInterval) * newEF))
		}
	}

	// Enforce minimum EF
	if newEF < s.MinEF {
		newEF = s.MinEF
	}

	// Ensure minimum interval of 1 day
	if newInterval < 1 {
		newInterval = 1
	}

	nextReviewAt := time.Now().AddDate(0, 0, newInterval)

	return ReviewResult{
		NewEF:        newEF,
		NewInterval:  newInterval,
		NextReviewAt: nextReviewAt,
	}
}

// GetInitialReviewResult returns the result for a new question's first review
func (s *SM2Calculator) GetInitialReviewResult(grade int) ReviewResult {
	return s.Calculate(s.InitialEF, 0, grade)
}
