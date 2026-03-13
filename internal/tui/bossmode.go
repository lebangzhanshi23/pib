package tui

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var (
	bossLogStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("249"))
)

// BossModeModel displays a fake log stream to hide the app
type BossModeModel struct {
	logs       []string
	lineIndex  int
	isActive   bool
	fakeProcs  []string
}

// NewBossModeModel creates a new boss mode model
func NewBossModeModel() *BossModeModel {
	// Initialize with some fake log entries
	logs := generateFakeLogs(50)
	procs := generateFakeProcesses()
	
	return &BossModeModel{
		logs:      logs,
		lineIndex: 0,
		isActive:  true,
		fakeProcs: procs,
	}
}

// generateFakeLogs creates fake nginx access logs
func generateFakeLogs(count int) []string {
	ips := []string{
		"192.168.1.100", "192.168.1.101", "10.0.0.45", "172.16.0.23",
		"192.168.1.55", "10.0.0.12", "172.16.0.8", "192.168.1.200",
	}
	paths := []string{
		"/", "/api/status", "/static/css/main.css", "/api/users",
		"/health", "/metrics", "/api/v2/data", "/static/js/app.js",
	}
	statuses := []int{200, 200, 200, 304, 200, 201, 200, 200}
	methods := []string{"GET", "GET", "GET", "POST", "GET", "GET", "POST", "GET"}
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		"curl/7.68.0",
		"nginx/1.18.0",
		"Prometheus/2.30.0",
	}
	
	logs := make([]string, count)
	for i := 0; i < count; i++ {
		ip := ips[rand.Intn(len(ips))]
		path := paths[rand.Intn(len(paths))]
		status := statuses[rand.Intn(len(statuses))]
		method := methods[rand.Intn(len(methods))]
		ua := userAgents[rand.Intn(len(userAgents))]
		size := rand.Intn(50000) + 100
		
		timestamp := time.Now().Add(-time.Duration(rand.Intn(3600)) * time.Second).Format("14/Jan/2006:15:04:05 -0700")
		log := fmt.Sprintf(`%s - - [%s] "%s %s HTTP/1.1" %d %d "-" "%s"`, 
			ip, timestamp, method, path, status, size, ua)
		logs[i] = log
	}
	
	return logs
}

// generateFakeProcesses creates fake process list
func generateFakeProcesses() []string {
	procs := []string{
		"    1 ?        S      0:00 /sbin/init",
		"  123 ?        S      0:02 /usr/sbin/sshd -D",
		"  456 ?        Ss     0:00 nginx: master process /usr/sbin/nginx",
		"  457 ?        S      0:00 nginx: worker process",
		"  458 ?        S      0:00 nginx: worker process",
		"  459 ?        S      0:00 nginx: worker process",
		"  460 ?        S      0:00 nginx: worker process",
		"  789 ?        Ss     0:01 /usr/bin/containerd",
		"  890 ?        S      0:00 /usr/bin/docker-proxy",
		" 1001 ?        Ss     0:00 /usr/local/bin/systemd-monitor",
		" 1102 ?        S      0:00 [kworker/0:1]",
		" 1203 ?        S      0:00 [kworker/1:2]",
		" 1304 ?        Ss     0:00 /usr/sbin/cron",
		" 1405 ?        Ss     0:00 /usr/sbin/rsyslogd",
		" 1506 ?        Ss     0:00 /usr/sbin/gunicorn --workers 4",
	}
	return procs
}

// Init initializes the model
func (m *BossModeModel) Init() tea.Cmd {
	// Continuously add new log entries
	return tickBossMode()
}

// tickBossMode creates a tick for adding new logs
func tickBossMode() tea.Cmd {
	return tea.Tick(800*time.Millisecond, func(t time.Time) tea.Msg {
		return bossModeTick{}
	})
}

type bossModeTick struct{}

// Update handles messages
func (m *BossModeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case bossModeTick:
		// Add a new fake log entry
		newLog := generateFakeLogs(1)[0]
		m.logs = append(m.logs, newLog)
		// Keep only last 100 lines
		if len(m.logs) > 100 {
			m.logs = m.logs[len(m.logs)-100:]
		}
		// Return tick for next update
		return m, tickBossMode()
	}
	return m, nil
}

// View renders the fake log stream
func (m *BossModeModel) View() string {
	s := ""
	
	// Header mimicking a terminal
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Background(lipgloss.Color("235")).
		Padding(0, 1).
		Render("root@server:~# tail -f /var/log/nginx/access.log")
	s += header + "\n\n"
	
	// Fake process info
	s += lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("top - 14:32:01 up 45 days,  3:22,  1 user,  load average: 0.15, 0.10, 0.08\n")
	s += lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Tasks: 127 total,   1 running, 126 sleeping,   0 stopped,   0 zombie\n")
	s += lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("%Cpu(s):  2.3 us,  1.2 sy,  0.0 ni, 96.2 id,  0.0 wa,  0.0 hi,  0.3 si,  0.0 st\n")
	s += "\n"
	
	// Process list header
	s += lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Background(lipgloss.Color("235")).
		Render(" PID USER      PR  NI    VIRT    RES    SHR S  %CPU  %MEM     TIME+ COMMAND  ")
	s += "\n"
	
	// Show fake processes (including nginx to blend in)
	for _, proc := range m.fakeProcs {
		s += bossLogStyle.Render(proc) + "\n"
	}
	
	s += "\n"
	
	// Log entries header
	s += lipgloss.NewStyle().
		Foreground(lipgloss.Color("250")).
		Background(lipgloss.Color("235")).
		Render("==> /var/log/nginx/access.log <==")
	s += "\n"
	
	// Show last few log lines
	startIdx := 0
	if len(m.logs) > 15 {
		startIdx = len(m.logs) - 15
	}
	for i := startIdx; i < len(m.logs); i++ {
		// Color GET requests green, POST blue, errors red
		log := m.logs[i]
		var coloredLog string
		switch {
		case contains(log, `"GET`):
			coloredLog = lipgloss.NewStyle().Foreground(lipgloss.Color("82")).Render(log)
		case contains(log, `"POST`):
			coloredLog = lipgloss.NewStyle().Foreground(lipgloss.Color("75")).Render(log)
		case contains(log, " 5") || contains(log, " 4"):
			coloredLog = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(log)
		default:
			coloredLog = bossLogStyle.Render(log)
		}
		s += coloredLog + "\n"
	}
	
	// Help at bottom
	s += "\n" + lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Render("Press ESC or Ctrl+B to exit Boss Mode")
	
	return s
}

// contains is a simple string contains check
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
