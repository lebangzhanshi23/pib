package agent

import (
	"fmt"

	"pib/config"
)

// CompareResult contains the result of AI comparison
type CompareResult struct {
	SemanticScore    float64  `json:"semantic_score"`    // 0-100
	ExpressionScore float64  `json:"expression_score"`   // 0-100
	Analysis        Analysis `json:"analysis"`
	FollowUp        []string `json:"follow_up"` // Follow-up questions
}

// Analysis contains detailed expression analysis
type Analysis struct {
	LogicGaps       []string `json:"logic_gaps"`       // Logical gaps detected
	Redundancy     string   `json:"redundancy"`       // Redundancy feedback
	MissingTerms   []string `json:"missing_terms"`    // Missing professional terms
	Strengths      []string `json:"strengths"`       // Strengths of the answer
	Suggestions    []string `json:"suggestions"`      // Improvement suggestions
}

// CompareEngine handles AI-powered answer comparison
type CompareEngine struct {
	client *LLMClient
}

// NewCompareEngine creates a new compare engine
func NewCompareEngine(cfg *config.Config) (*CompareEngine, error) {
	client := NewLLMClient(cfg)
	
	// Test if LLM is configured
	testMsg := []Message{{Role: "user", Content: "Hi"}}
	_, err := client.Chat(testMsg)
	if err != nil {
		return nil, fmt.Errorf("LLM not configured: %v", err)
	}
	
	return &CompareEngine{
		client: client,
	}, nil
}

// Compare performs AI comparison between user answer and standard answer
func (e *CompareEngine) Compare(question, standardAnswer, userAnswer string) (*CompareResult, error) {
	// Build prompt for comprehensive analysis
	prompt := buildComparePrompt(question, standardAnswer, userAnswer)
	
	messages := []Message{
		{Role: "system", Content: "You are an expert interview coach. Analyze the user's answer to the given question and provide detailed feedback in JSON format."},
		{Role: "user", Content: prompt},
	}
	
	response, err := e.client.Chat(messages)
	if err != nil {
		return nil, fmt.Errorf("failed to get AI comparison: %v", err)
	}
	
	// Parse the response
	result, err := parseCompareResponse(response)
	if err != nil {
		// If parsing fails, create a fallback result
		result = &CompareResult{
			SemanticScore:    50,
			ExpressionScore:  50,
			Analysis:         Analysis{
				Suggestions: []string{"Unable to analyze. Please try again."},
			},
			FollowUp: []string{},
		}
	}
	
	return result, nil
}

// buildComparePrompt creates a prompt for the LLM to analyze the answer
func buildComparePrompt(question, standardAnswer, userAnswer string) string {
	return fmt.Sprintf(`Please analyze the following interview answer and provide feedback in JSON format.

Question: %s

Standard/Expected Answer: %s

User's Answer: %s

Please analyze and return a JSON object with the following structure:
{
  "semantic_score": <number 0-100 - how well the answer covers the key concepts>,
  "expression_score": <number 0-100 - how well the answer is structured and expressed>,
  "analysis": {
    "logic_gaps": [<array of logical gaps or missing points>],
    "redundancy": "<feedback on any redundant or filler content>",
    "missing_terms": [<array of important professional terms that are missing>],
    "strengths": [<array of what the user did well>],
    "suggestions": [<array of specific suggestions for improvement>]
  },
  "follow_up": [<array of 1-2 follow-up questions based on the weaknesses in the answer>]
}

Respond ONLY with valid JSON, no additional text.`, question, standardAnswer, userAnswer)
}

// parseCompareResponse parses the AI response into CompareResult
func parseCompareResponse(response string) (*CompareResult, error) {
	// Try to extract JSON from the response
	result := &CompareResult{
		Analysis: Analysis{},
		FollowUp: []string{},
	}
	
	// Simple parsing - look for numeric scores and arrays
	// In a production version, you'd use proper JSON parsing
	
	// Try to extract semantic score
	fmt.Sscanf(response, `"semantic_score": %f`, &result.SemanticScore)
	fmt.Sscanf(response, `"expression_score": %f`, &result.ExpressionScore)
	
	// If scores are 0, try alternative format
	if result.SemanticScore == 0 {
		fmt.Sscanf(response, `"semantic_score":%f`, &result.SemanticScore)
	}
	if result.ExpressionScore == 0 {
		fmt.Sscanf(response, `"expression_score":%f`, &result.ExpressionScore)
	}
	
	// Default scores if not found
	if result.SemanticScore == 0 {
		result.SemanticScore = 50
	}
	if result.ExpressionScore == 0 {
		result.ExpressionScore = 50
	}
	
	// Extract logic gaps
	// This is a simplified extraction - in production use proper JSON parsing
	result.Analysis.LogicGaps = extractArrayField(response, "logic_gaps")
	result.Analysis.MissingTerms = extractArrayField(response, "missing_terms")
	result.Analysis.Strengths = extractArrayField(response, "strengths")
	result.Analysis.Suggestions = extractArrayField(response, "suggestions")
	result.FollowUp = extractArrayField(response, "follow_up")
	
	// Extract redundancy
	result.Analysis.Redundancy = extractStringField(response, "redundancy")
	
	return result, nil
}

// extractArrayField extracts an array field from JSON-like string
func extractArrayField(jsonStr, fieldName string) []string {
	// Simple extraction - find the field and extract items between brackets
	start := findFieldStart(jsonStr, fieldName)
	if start == -1 {
		return []string{}
	}
	
	// Find the opening bracket
	bracketStart := -1
	for i := start; i < len(jsonStr) && i < start+50; i++ {
		if jsonStr[i] == '[' {
			bracketStart = i
			break
		}
	}
	if bracketStart == -1 {
		return []string{}
	}
	
	// Find matching closing bracket
	bracketCount := 0
	bracketEnd := -1
	for i := bracketStart; i < len(jsonStr); i++ {
		if jsonStr[i] == '[' {
			bracketCount++
		} else if jsonStr[i] == ']' {
			bracketCount--
			if bracketCount == 0 {
				bracketEnd = i
				break
			}
		}
	}
	
	if bracketEnd == -1 {
		return []string{}
	}
	
	// Extract and parse items
	bracketContent := jsonStr[bracketStart+1 : bracketEnd]
	
	// Split by comma and clean up
	var items []string
	var current string
	inQuote := false
	for _, c := range bracketContent {
		if c == '"' {
			inQuote = !inQuote
		} else if c == ',' && !inQuote {
			if current != "" {
				items = append(items, cleanString(current))
			}
			current = ""
		} else if !inQuote && (c == ' ' || c == '\n' || c == '\t') {
			// Skip whitespace outside quotes
		} else {
			current += string(c)
		}
	}
	if current != "" {
		items = append(items, cleanString(current))
	}
	
	return items
}

// extractStringField extracts a string field from JSON-like string
func extractStringField(jsonStr, fieldName string) string {
	items := extractArrayField(jsonStr, fieldName)
	if len(items) > 0 {
		return items[0]
	}
	return ""
}

// findFieldStart finds the start position of a field name
func findFieldStart(jsonStr, fieldName string) int {
	target := fmt.Sprintf(`"%s"`, fieldName)
	return findSubstring(jsonStr, target)
}

// findSubstring finds a substring (simple implementation)
func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// cleanString removes quotes and whitespace from a string
func cleanString(s string) string {
	result := ""
	inQuote := false
	for _, c := range s {
		if c == '"' {
			inQuote = !inQuote
			continue
		}
		if !inQuote && (c == ' ' || c == '\n' || c == '\t' || c == '\r') {
			continue
		}
		result += string(c)
	}
	return result
}

// QuickCompare performs a quick semantic comparison without detailed analysis
func (e *CompareEngine) QuickCompare(standardAnswer, userAnswer string) (float64, error) {
	prompt := fmt.Sprintf(`Compare these two answers and give a semantic similarity score from 0 to 100.

Standard Answer: %s

User Answer: %s

Respond with ONLY a number (0-100). No other text.`, standardAnswer, userAnswer)
	
	messages := []Message{
		{Role: "system", Content: "You are a scoring assistant. Respond with only a number."},
		{Role: "user", Content: prompt},
	}
	
	response, err := e.client.Chat(messages)
	if err != nil {
		return 0, err
	}
	
	// Extract number from response
	var score float64
	fmt.Sscanf(response, "%f", &score)
	
	// Clamp to valid range
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return score, nil
}
