package tui

import (
	"fmt"
	"math"
	"pib/internal/repository"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("236"))

	statsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241"))

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("212"))
)

// AnalyticsModel represents the analytics view
type AnalyticsModel struct {
	statusCounts   map[string]int
	tagScores      map[string]float64
	totalCount    int64
	reviewStats    map[string]interface{}
	viewport      viewport.Model
	Ready         bool
	Cancelled     bool
}

// NewAnalyticsModel creates a new analytics model
func NewAnalyticsModel() *AnalyticsModel {
	vp := viewport.New(60, 20)
	return &AnalyticsModel{
		viewport:   vp,
		statusCounts: make(map[string]int),
		tagScores:   make(map[string]float64),
	}
}

// SetData sets the analytics data
func (m *AnalyticsModel) SetData(statusCounts map[string]int, tagScores map[string]float64, totalCount int64, reviewStats map[string]interface{}) {
	m.statusCounts = statusCounts
	m.tagScores = tagScores
	m.totalCount = totalCount
	m.reviewStats = reviewStats
	m.Ready = true
}

// Init loads the analytics data
func (m *AnalyticsModel) Init() tea.Cmd {
	return nil
}

// Update handles messages
func (m *AnalyticsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			m.Cancelled = true
			return m, nil
		case "up", "k":
			m.viewport.ScrollUp(1)
		case "down", "j":
			m.viewport.ScrollDown(1)
		}
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 6
	}
	return m, nil
}

// View renders the analytics view
func (m *AnalyticsModel) View() string {
	if !m.Ready {
		return "Loading analytics..."
	}

	var content string

	// Header
	content += headerStyle.Render(" 📊 能力分析 ") + "\n\n"

	// Progress tracking section
	content += labelStyle.Render("📈 进度追踪") + "\n"
	content += m.renderProgressBar() + "\n\n"

	// Stats summary
	content += labelStyle.Render("📋 统计概览") + "\n"
	content += m.renderStats() + "\n\n"

	// Radar chart
	if len(m.tagScores) > 0 {
		content += labelStyle.Render("🎯 能力雷达图") + "\n"
		content += m.renderRadarChart() + "\n"
	} else {
		content += labelStyle.Render("🎯 能力雷达图") + "\n"
		content += " 暂无标签数据，请先添加题目并设置标签\n"
	}

	content += "\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(" 按 [Esc/q] 返回 | [↑/k] 上翻 | [↓/j] 下翻 ")

	m.viewport.SetContent(content)
	return m.viewport.View()
}

// renderProgressBar renders a progress bar showing question status distribution
func (m *AnalyticsModel) renderProgressBar() string {
	draft := m.statusCounts["draft"]
	active := m.statusCounts["active"]
	archived := m.statusCounts["archived"]
	total := draft + active + archived

	if total == 0 {
		return "  暂无题目数据"
	}

	const barWidth = 40
	draftWidth := int(float64(draft) / float64(total) * barWidth)
	activeWidth := int(float64(active) / float64(total) * barWidth)
	archivedWidth := barWidth - draftWidth - activeWidth

	draftBar := lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(starBar(draftWidth))
	activeBar := lipgloss.NewStyle().Background(lipgloss.Color("82")).Render(starBar(activeWidth))
	archivedBar := lipgloss.NewStyle().Background(lipgloss.Color("205")).Render(starBar(archivedWidth))

	result := fmt.Sprintf("  [%s%s%s]\n", draftBar, activeBar, archivedBar)
	result += fmt.Sprintf("  New: %d  |  Active: %d  |  Mastered: %d", draft, active, archived)
	return result
}

func starBar(width int) string {
	result := ""
	for i := 0; i < width; i++ {
		result += " "
	}
	return result
}

// renderStats renders statistics
func (m *AnalyticsModel) renderStats() string {
	result := ""
	result += fmt.Sprintf("  %s: %s%d%s", 
		labelStyle.Render("总题数"),
		valueStyle.Render(""),
		int(m.totalCount),
		"\n")

	if m.reviewStats != nil {
		if total, ok := m.reviewStats["total_reviews"].(int64); ok {
			result += fmt.Sprintf("  %s: %s%d%s", 
				labelStyle.Render("复习次数"),
				valueStyle.Render(""),
				int(total),
				"\n")
		}
		if avgGrade, ok := m.reviewStats["avg_grade"].(float64); ok {
			result += fmt.Sprintf("  %s: %s%.1f/2.0%s", 
				labelStyle.Render("平均得分"),
				valueStyle.Render(""),
				avgGrade,
				"\n")
		}
	}
	return result
}

// renderRadarChart renders an ASCII hexagonal radar chart
func (m *AnalyticsModel) renderRadarChart() string {
	// Get top 6 tags
	type tagScore struct {
		name  string
		score float64
	}
	var tags []tagScore
	for name, score := range m.tagScores {
		tags = append(tags, tagScore{name: name, score: score})
	}
	if len(tags) > 6 {
		tags = tags[:6]
	}
	if len(tags) == 0 {
		return "  暂无数据"
	}

	// Use exactly 6 dimensions for hexagon (pad if needed)
	dimensions := []string{"技能1", "技能2", "技能3", "技能4", "技能5", "技能6"}
	scores := []float64{0, 0, 0, 0, 0, 0}
	for i, t := range tags {
		dimensions[i] = t.name
		scores[i] = t.score
	}

	// Chart configuration
	const (
		chartWidth  = 50
		chartHeight = 18
		centerX     = 25
		centerY     = 10
		maxRadius   = 8
	)

	// Create grid
	grid := make([][]rune, chartHeight)
	for i := range grid {
		grid[i] = make([]rune, chartWidth)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Draw hexagonal grid lines
	drawHexagon(grid, centerX, centerY, maxRadius, '-')
	drawHexagon(grid, centerX, centerY, maxRadius*2/3, '.')
	drawHexagon(grid, centerX, centerY, maxRadius*1/3, '.')

	// Draw axis lines and data
	numPoints := len(dimensions)
	for i := 0; i < numPoints; i++ {
		angle := float64(i)*2*math.Pi/float64(numPoints) - math.Pi/2
		x := centerX + int(float64(maxRadius)*math.Cos(angle))
		y := centerY + int(float64(maxRadius)*math.Sin(angle))
		
		// Draw axis line
		drawLine(grid, centerX, centerY, x, y, '|')
		
		// Draw data point
		radius := float64(maxRadius) * scores[i] / 100
		dataX := centerX + int(radius*math.Cos(angle))
		dataY := centerY + int(radius*math.Sin(angle))
		if dataY >= 0 && dataY < chartHeight && dataX >= 0 && dataX < chartWidth {
			grid[dataY][dataX] = '●'
		}
	}

	// Draw connecting lines between data points
	for i := 0; i < numPoints; i++ {
		next := (i + 1) % numPoints
		angle1 := float64(i)*2*math.Pi/float64(numPoints) - math.Pi/2
		angle2 := float64(next)*2*math.Pi/float64(numPoints) - math.Pi/2
		radius1 := float64(maxRadius) * scores[i] / 100
		radius2 := float64(maxRadius) * scores[next] / 100
		x1 := centerX + int(radius1*math.Cos(angle1))
		y1 := centerY + int(radius1*math.Sin(angle1))
		x2 := centerX + int(radius2*math.Cos(angle2))
		y2 := centerY + int(radius2*math.Sin(angle2))
		drawLine(grid, x1, y1, x2, y2, '•')
	}

	// Convert to string
	result := "  "
	for y := 0; y < chartHeight; y++ {
		for x := 0; x < chartWidth; x++ {
			char := grid[y][x]
			if char == 0 {
				char = ' '
			}
			result += string(char)
		}
		result += "\n  "
	}

	// Add labels
	result += "\n         "
	labelY := centerY + maxRadius + 1
	if labelY < chartHeight {
		for i, dim := range dimensions {
			angle := float64(i)*2*math.Pi/float64(numPoints) - math.Pi/2
			labelX := centerX + int(float64(maxRadius+2)*math.Cos(angle)) - len(dim)/2
			if labelX >= 0 && labelX < chartWidth && labelY >= 0 && labelY < chartHeight {
				result += dim + "  "
				break
			}
		}
	}

	return result
}

func drawHexagon(grid [][]rune, cx, cy, radius int, char rune) {
	numPoints := 6
	for i := 0; i < numPoints; i++ {
		angle1 := float64(i) * 2 * math.Pi / 6
		angle2 := float64(i+1) * 2 * math.Pi / 6
		x1 := cx + int(float64(radius)*math.Cos(angle1))
		y1 := cy + int(float64(radius)*math.Sin(angle1))
		x2 := cx + int(float64(radius)*math.Cos(angle2))
		y2 := cy + int(float64(radius)*math.Sin(angle2))
		drawLine(grid, x1, y1, x2, y2, char)
	}
}

func drawLine(grid [][]rune, x1, y1, x2, y2 int, char rune) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := -1
	if x1 < x2 {
		sx = 1
	}
	sy := -1
	if y1 < y2 {
		sy = 1
	}
	err := dx - dy

	for {
		if y1 >= 0 && y1 < len(grid) && x1 >= 0 && x1 < len(grid[0]) {
			if grid[y1][x1] == ' ' {
				grid[y1][x1] = char
			}
		}
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// LoadAnalytics loads analytics data from repository
func LoadAnalytics(db *repository.SQLiteDB) (map[string]int, map[string]float64, int64, map[string]interface{}, error) {
	statusCounts, err := db.GetQuestionCountByStatus()
	if err != nil {
		return nil, nil, 0, nil, err
	}

	tagScores, err := db.GetTagScores()
	if err != nil {
		return nil, nil, 0, nil, err
	}

	totalCount, err := db.GetTotalQuestionCount()
	if err != nil {
		return nil, nil, 0, nil, err
	}

	reviewStats, err := db.GetReviewStats()
	if err != nil {
		return nil, nil, 0, nil, err
	}

	return statusCounts, tagScores, totalCount, reviewStats, nil
}
