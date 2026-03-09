package main

import (
	"fmt"
	"log"

	"pib/config"
	"pib/internal/model"
	"pib/internal/repository"
	"pib/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize JSON database
	db, err := repository.NewJSONDB("data/pib.json")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize SM-2 calculator
	sm2 := service.NewSM2Calculator(cfg.SRS.InitialEF, cfg.SRS.MinEF)

	// Setup Gin router
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Serve static frontend at /app path
	r.Static("/app", "frontend")

	// Redirect root to /app
	r.GET("/", func(c *gin.Context) {
		c.Redirect(302, "/app")
	})

	// API routes
	v1 := r.Group("/api/v1")
	{
		// Questions
		v1.POST("/questions", func(c *gin.Context) {
			var req struct {
				Content string   `json:"content" binding:"required"`
				Answer  string   `json:"answer"`
				Tags    []string `json:"tags"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			q := &model.Question{
				Content: req.Content,
				Answer:  req.Answer,
				Status:  "draft",
				EF:      2.5,
				Interval: 0,
			}

			if err := db.CreateQuestion(q); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			// Add tags
			for _, tagName := range req.Tags {
				tag, _ := db.GetOrCreateTag(tagName)
				db.AddTagToQuestion(q.ID, tag.ID)
			}

			// Load tags
			q.Tags = db.GetTagsForQuestion(q.ID)

			c.JSON(200, q)
		})

		v1.GET("/questions", func(c *gin.Context) {
			status := c.DefaultQuery("status", "draft")
			questions := db.ListQuestionsByStatus(status)
			
			// Attach tags to each question
			for i := range questions {
				questions[i].Tags = db.GetTagsForQuestion(questions[i].ID)
			}
			
			c.JSON(200, questions)
		})

		v1.GET("/questions/review", func(c *gin.Context) {
			limit := 20
			questions := db.GetQuestionsForReview(limit)
			
			// Attach tags to each question
			for i := range questions {
				questions[i].Tags = db.GetTagsForQuestion(questions[i].ID)
			}
			
			c.JSON(200, questions)
		})

		v1.POST("/questions/:id/review", func(c *gin.Context) {
			id := c.Param("id")
			var req struct {
				Grade int `json:"grade"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}

			// Validate grade
			if req.Grade < 0 || req.Grade > 2 {
				c.JSON(400, gin.H{"error": "grade must be 0, 1, or 2"})
				return
			}

			// Get question
			q, err := db.GetQuestionByID(id)
			if err != nil || q == nil {
				c.JSON(404, gin.H{"error": "question not found"})
				return
			}

			// Calculate next review using SM-2
			result := sm2.Calculate(q.EF, q.Interval, req.Grade)

			// Update question
			q.EF = result.NewEF
			q.Interval = result.NewInterval
			q.NextReviewAt = &result.NextReviewAt
			if q.Status == "draft" {
				q.Status = "active"
			}

			if err := db.UpdateQuestion(q); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}

			// Create review log
			reviewLog := &model.ReviewLog{
				QuestionID: id,
				Grade:      req.Grade,
			}
			db.CreateReviewLog(reviewLog)

			c.JSON(200, gin.H{
				"question":     q,
				"next_review":  result,
			})
		})

		v1.DELETE("/questions/:id", func(c *gin.Context) {
			id := c.Param("id")
			if err := db.DeleteQuestion(id); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"status": "deleted"})
		})
	}

	// Start server
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Printf("Starting PIB server on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
