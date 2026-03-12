package agent

import (
	"fmt"
)

// ScoutResult represents the AI Scout result
type ScoutResult struct {
	Beginner  string // 入门级
	Expert    string // 专家级
	BigTech   string // 大厂面试官版
}

// GenerateAnswers generates 3 types of reference answers for a question
func (c *LLMClient) GenerateAnswers(question string, tags []string) (*ScoutResult, error) {
	// Build context from tags
	tagsStr := ""
	if len(tags) > 0 {
		tagsStr = fmt.Sprintf("相关技术标签: %s", joinTags(tags))
	}

	// Generate answers in parallel using separate prompts
	beginnerChan := make(chan string, 1)
	expertChan := make(chan string, 1)
	bigTechChan := make(chan string, 1)

	// Generate beginner level answer
	go func() {
		msg := []Message{
			{Role: "system", Content: "你是一位技术面试辅导老师，善于用通俗易懂的语言讲解技术概念。"},
			{Role: "user", Content: fmt.Sprintf("题目: %s\n%s\n\n请用入门级的角度回答这个问题，适合初学者理解。要求：1) 解释基本概念 2) 给出简单的代码示例 3) 结论清晰。", question, tagsStr)},
		}
		resp, err := c.Chat(msg)
		if err != nil {
			beginnerChan <- ""
		} else {
			beginnerChan <- resp
		}
	}()

	// Generate expert level answer
	go func() {
		msg := []Message{
			{Role: "system", Content: "你是一位资深技术专家，对技术有深入的理解和丰富的实战经验。"},
			{Role: "user", Content: fmt.Sprintf("题目: %s\n%s\n\n请用专家级的角度回答这个问题，展示深度技术理解。要求：1) 深入分析原理 2) 结合实际场景 3) 包含最佳实践和常见坑。", question, tagsStr)},
		}
		resp, err := c.Chat(msg)
		if err != nil {
			expertChan <- ""
		} else {
			expertChan <- resp
		}
	}()

	// Generate big tech interviewer version
	go func() {
		msg := []Message{
			{Role: "system", Content: "你是一位大厂面试官（BAT/字节/美团等），负责技术岗位面试评估。"},
			{Role: "user", Content: fmt.Sprintf("题目: %s\n%s\n\n请从大厂面试官的角度回答这个问题，模拟真实面试场景。要求：1) 先让候选人分析问题 2) 给出考察点 3) 给出标准答案要点 4) 可适当追问。", question, tagsStr)},
		}
		resp, err := c.Chat(msg)
		if err != nil {
			bigTechChan <- ""
		} else {
			bigTechChan <- resp
		}
	}()

	// Wait for all results
	beginner := <-beginnerChan
	expert := <-expertChan
	bigTech := <-bigTechChan

	// Check if at least one succeeded
	if beginner == "" && expert == "" && bigTech == "" {
		return nil, fmt.Errorf("failed to generate all answers")
	}

	return &ScoutResult{
		Beginner: beginner,
		Expert:   expert,
		BigTech:  bigTech,
	}, nil
}

func joinTags(tags []string) string {
	result := ""
	for i, tag := range tags {
		if i > 0 {
			result += ", "
		}
		result += tag
	}
	return result
}
