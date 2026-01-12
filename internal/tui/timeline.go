package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/QRY91/uroboro/internal/journey"
)

var (
	projectColors = []lipgloss.Color{
		"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4",
		"#FECA57", "#FF9FF3", "#54A0FF", "#5F27CD",
	}

	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")).Padding(0, 1)
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4"))
	normalStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	helpStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FECA57")).MarginTop(1)
	searchStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	filterStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#4ECDC4"))
	detailStyle   = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#7D56F4")).Padding(1, 2)
)

type viewMode int

const (
	modeList viewMode = iota
	modeSearch
	modeDetail
)

type Model struct {
	data           *journey.JourneyData
	filtered       []journey.TimelineEvent
	cursor         int
	offset         int
	width, height  int
	mode           viewMode
	searchInput    textinput.Model
	searchQuery    string
	projectFilter  int // -1 = all, 0+ = index into projects
	typeFilter     int // -1 = all, 0+ = index into types
	projectColors  map[string]lipgloss.Color
	ready          bool
	groupByDay     bool
}

var eventTypes = []string{"all", "capture", "commit", "bugfix", "feature", "milestone", "learning", "decision"}

func NewModel(data *journey.JourneyData) Model {
	ti := textinput.New()
	ti.Placeholder = "search..."
	ti.CharLimit = 50

	colors := make(map[string]lipgloss.Color)
	for i, p := range data.Projects {
		colors[p.Name] = projectColors[i%len(projectColors)]
	}

	m := Model{
		data:          data,
		projectColors: colors,
		searchInput:   ti,
		projectFilter: -1,
		typeFilter:    -1,
		groupByDay:    true,
	}
	m.applyFilters()
	return m
}

func (m *Model) applyFilters() {
	m.filtered = nil
	for _, e := range m.data.Events {
		// Project filter
		if m.projectFilter >= 0 && m.projectFilter < len(m.data.Projects) {
			if e.Project != m.data.Projects[m.projectFilter].Name {
				continue
			}
		}
		// Type filter
		if m.typeFilter > 0 && m.typeFilter < len(eventTypes) {
			if e.EventType != eventTypes[m.typeFilter] {
				continue
			}
		}
		// Search filter
		if m.searchQuery != "" {
			q := strings.ToLower(m.searchQuery)
			if !strings.Contains(strings.ToLower(e.Content), q) &&
				!strings.Contains(strings.ToLower(e.Project), q) &&
				!containsTag(e.Tags, q) {
				continue
			}
		}
		m.filtered = append(m.filtered, e)
	}
	// Reset cursor if out of bounds
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	m.offset = 0
}

func containsTag(tags []string, q string) bool {
	for _, t := range tags {
		if strings.Contains(strings.ToLower(t), q) {
			return true
		}
	}
	return false
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		// Handle search mode
		if m.mode == modeSearch {
			switch msg.String() {
			case "enter":
				m.searchQuery = m.searchInput.Value()
				m.applyFilters()
				m.mode = modeList
			case "esc":
				m.searchInput.SetValue(m.searchQuery)
				m.mode = modeList
			default:
				var cmd tea.Cmd
				m.searchInput, cmd = m.searchInput.Update(msg)
				return m, cmd
			}
			return m, nil
		}

		// Handle detail mode
		if m.mode == modeDetail {
			m.mode = modeList
			return m, nil
		}

		// List mode
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			m.moveCursor(1)
		case "k", "up":
			m.moveCursor(-1)
		case "ctrl+d", "pgdown":
			m.moveCursor(10)
		case "ctrl+u", "pgup":
			m.moveCursor(-10)
		case "g", "home":
			m.cursor, m.offset = 0, 0
		case "G", "end":
			m.cursor = max(0, len(m.filtered)-1)
			m.ensureVisible()
		case "enter", " ":
			if len(m.filtered) > 0 {
				m.mode = modeDetail
			}
		case "/":
			m.mode = modeSearch
			m.searchInput.Focus()
			return m, textinput.Blink
		case "esc":
			if m.searchQuery != "" {
				m.searchQuery = ""
				m.searchInput.SetValue("")
				m.applyFilters()
			}
		case "p":
			m.projectFilter++
			if m.projectFilter >= len(m.data.Projects) {
				m.projectFilter = -1
			}
			m.applyFilters()
		case "t":
			m.typeFilter++
			if m.typeFilter >= len(eventTypes) {
				m.typeFilter = -1
			}
			m.applyFilters()
		case "d":
			m.groupByDay = !m.groupByDay
		}
	}
	return m, nil
}

func (m *Model) moveCursor(delta int) {
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}
	m.ensureVisible()
}

func (m *Model) ensureVisible() {
	visibleLines := m.height - 4
	if m.cursor < m.offset {
		m.offset = m.cursor
	} else if m.cursor >= m.offset+visibleLines {
		m.offset = m.cursor - visibleLines + 1
	}
}

func (m Model) View() string {
	if !m.ready {
		return "Loading..."
	}

	switch m.mode {
	case modeDetail:
		return m.viewDetail()
	case modeSearch:
		return m.viewWithSearch()
	default:
		return m.viewList()
	}
}

func (m Model) viewList() string {
	// Title
	title := m.buildTitle()

	// Events
	content := m.renderEvents()

	// Help
	help := m.buildHelp()

	return fmt.Sprintf("%s\n%s\n%s", title, content, help)
}

func (m Model) viewWithSearch() string {
	title := titleStyle.Render(" uroboro timeline ")
	search := fmt.Sprintf("\n %s %s\n", searchStyle.Render("/"), m.searchInput.View())
	content := m.renderEvents()
	help := helpStyle.Render("enter: search | esc: cancel")
	return fmt.Sprintf("%s%s%s\n%s", title, search, content, help)
}

func (m Model) buildTitle() string {
	parts := []string{fmt.Sprintf(" uroboro | %d events ", len(m.filtered))}

	if m.projectFilter >= 0 && m.projectFilter < len(m.data.Projects) {
		p := m.data.Projects[m.projectFilter]
		parts = append(parts, filterStyle.Render(fmt.Sprintf("[%s]", p.Name)))
	}
	if m.typeFilter > 0 && m.typeFilter < len(eventTypes) {
		parts = append(parts, filterStyle.Render(fmt.Sprintf("<%s>", eventTypes[m.typeFilter])))
	}
	if m.searchQuery != "" {
		parts = append(parts, searchStyle.Render(fmt.Sprintf("/%s/", m.searchQuery)))
	}

	return titleStyle.Render(strings.Join(parts, " "))
}

func (m Model) buildHelp() string {
	var parts []string
	parts = append(parts, "j/k:nav")
	parts = append(parts, "/:search")
	parts = append(parts, "p:project")
	parts = append(parts, "t:type")
	parts = append(parts, "d:group")
	parts = append(parts, "enter:detail")
	parts = append(parts, "q:quit")
	return helpStyle.Render(strings.Join(parts, " | "))
}

func (m Model) renderEvents() string {
	if len(m.filtered) == 0 {
		return dimStyle.Render("\n  No events match filters.\n")
	}

	visibleLines := m.height - 4
	var sb strings.Builder
	var lastDay string
	lineCount := 0

	for i, event := range m.filtered {
		if i < m.offset {
			continue
		}
		if lineCount >= visibleLines {
			break
		}

		// Day header
		if m.groupByDay {
			day := event.Timestamp.Format("Mon, Jan 2 2006")
			if day != lastDay {
				if lineCount > 0 {
					sb.WriteString("\n")
					lineCount++
				}
				sb.WriteString(headerStyle.Render(fmt.Sprintf("── %s ──", day)))
				sb.WriteString("\n")
				lineCount++
				lastDay = day
				if lineCount >= visibleLines {
					break
				}
			}
		}

		sb.WriteString(m.renderEvent(i, event))
		sb.WriteString("\n")
		lineCount++
	}

	return sb.String()
}

func (m Model) renderEvent(idx int, event journey.TimelineEvent) string {
	timeStr := event.Timestamp.Format("15:04")

	// Project
	var project string
	if event.Project != "" {
		color := m.projectColors[event.Project]
		project = lipgloss.NewStyle().Foreground(color).Bold(true).Render(fmt.Sprintf("[%s]", truncate(event.Project, 12)))
	} else {
		project = dimStyle.Render("[--]")
	}

	// Type icon
	icon := m.typeIcon(event.EventType)

	// Content
	maxLen := m.width - 35
	content := truncate(event.Content, max(20, maxLen))

	line := fmt.Sprintf(" %s %s %s %s", dimStyle.Render(timeStr), project, icon, content)

	if idx == m.cursor {
		return selectedStyle.Render(line)
	}
	return line
}

func (m Model) typeIcon(t string) string {
	icons := map[string]struct{ icon string; color lipgloss.Color }{
		"commit":    {"git", "#626262"},
		"milestone": {"***", "#FECA57"},
		"learning":  {"lrn", "#54A0FF"},
		"decision":  {"dec", "#FF9FF3"},
		"bugfix":    {"fix", "#FF6B6B"},
		"feature":   {"fea", "#4ECDC4"},
	}
	if ic, ok := icons[t]; ok {
		return lipgloss.NewStyle().Foreground(ic.color).Render(ic.icon)
	}
	return dimStyle.Render("cap")
}

func (m Model) viewDetail() string {
	if m.cursor >= len(m.filtered) {
		return ""
	}
	e := m.filtered[m.cursor]

	var sb strings.Builder
	sb.WriteString(titleStyle.Render(" Event Details "))
	sb.WriteString("\n\n")

	sb.WriteString(dimStyle.Render("Time: "))
	sb.WriteString(e.Timestamp.Format(time.RFC1123))
	sb.WriteString("\n\n")

	sb.WriteString(dimStyle.Render("Project: "))
	if c, ok := m.projectColors[e.Project]; ok {
		sb.WriteString(lipgloss.NewStyle().Foreground(c).Bold(true).Render(e.Project))
	} else {
		sb.WriteString(e.Project)
	}
	sb.WriteString("\n\n")

	sb.WriteString(dimStyle.Render("Type: "))
	sb.WriteString(e.EventType)
	sb.WriteString("\n\n")

	if len(e.Tags) > 0 {
		sb.WriteString(dimStyle.Render("Tags: "))
		sb.WriteString(strings.Join(e.Tags, ", "))
		sb.WriteString("\n\n")
	}

	sb.WriteString(dimStyle.Render("Content:\n"))
	sb.WriteString(wordWrap(e.Content, m.width-10))
	sb.WriteString("\n\n")

	if e.GitHash != "" {
		sb.WriteString(dimStyle.Render("Commit: "))
		sb.WriteString(e.GitHash[:min(8, len(e.GitHash))])
		sb.WriteString("\n\n")
	}

	sb.WriteString(helpStyle.Render("\nPress any key to close"))

	return detailStyle.Width(m.width - 4).Render(sb.String())
}

func truncate(s string, n int) string {
	if n <= 0 || len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func wordWrap(s string, width int) string {
	if width <= 0 {
		return s
	}
	var result strings.Builder
	for _, line := range strings.Split(s, "\n") {
		for len(line) > width {
			result.WriteString(line[:width])
			result.WriteString("\n")
			line = line[width:]
		}
		result.WriteString(line)
		result.WriteString("\n")
	}
	return strings.TrimSuffix(result.String(), "\n")
}

func Run(data *journey.JourneyData) error {
	p := tea.NewProgram(NewModel(data), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
