package report

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/journey"
)

type ProjectSummary struct {
	Name       string
	Events     []journey.TimelineEvent
	FirstEvent time.Time
	LastEvent  time.Time
	Days       int
	Sessions   []Session
	TotalTime  time.Duration
}

type Session struct {
	Start    time.Time
	End      time.Time
	Duration time.Duration
	Events   int
}

type Report struct {
	StartDate time.Time
	EndDate   time.Time
	Projects  []ProjectSummary
	Total     struct {
		Events   int
		Projects int
		Days     int
		Time     time.Duration
	}
}

// Generate creates a report from journey data
func Generate(data *journey.JourneyData, sessionGap time.Duration) *Report {
	if sessionGap == 0 {
		sessionGap = 30 * time.Minute // default: 30 min gap = new session
	}

	r := &Report{
		StartDate: data.Timeline.StartTime,
		EndDate:   data.Timeline.EndTime,
	}

	// Group events by project
	byProject := make(map[string][]journey.TimelineEvent)
	for _, e := range data.Events {
		p := e.Project
		if p == "" {
			p = "(no project)"
		}
		byProject[p] = append(byProject[p], e)
	}

	// Build project summaries
	for name, events := range byProject {
		// Sort by time
		sort.Slice(events, func(i, j int) bool {
			return events[i].Timestamp.Before(events[j].Timestamp)
		})

		ps := ProjectSummary{
			Name:       name,
			Events:     events,
			FirstEvent: events[0].Timestamp,
			LastEvent:  events[len(events)-1].Timestamp,
		}

		// Calculate unique days
		days := make(map[string]bool)
		for _, e := range events {
			days[e.Timestamp.Format("2006-01-02")] = true
		}
		ps.Days = len(days)

		// Detect sessions (clusters of activity)
		ps.Sessions = detectSessions(events, sessionGap)
		for _, s := range ps.Sessions {
			ps.TotalTime += s.Duration
		}

		r.Projects = append(r.Projects, ps)
		r.Total.Events += len(events)
		r.Total.Time += ps.TotalTime
	}

	// Sort projects by total time descending
	sort.Slice(r.Projects, func(i, j int) bool {
		return r.Projects[i].TotalTime > r.Projects[j].TotalTime
	})

	r.Total.Projects = len(r.Projects)

	// Count unique days across all projects
	allDays := make(map[string]bool)
	for _, e := range data.Events {
		allDays[e.Timestamp.Format("2006-01-02")] = true
	}
	r.Total.Days = len(allDays)

	return r
}

func detectSessions(events []journey.TimelineEvent, gap time.Duration) []Session {
	if len(events) == 0 {
		return nil
	}

	var sessions []Session
	sessionStart := events[0].Timestamp
	sessionEvents := 1
	lastTime := events[0].Timestamp

	for i := 1; i < len(events); i++ {
		e := events[i]
		if e.Timestamp.Sub(lastTime) > gap {
			// End current session
			duration := lastTime.Sub(sessionStart)
			if duration < 5*time.Minute {
				duration = 5 * time.Minute // minimum session
			}
			sessions = append(sessions, Session{
				Start:    sessionStart,
				End:      lastTime,
				Duration: duration,
				Events:   sessionEvents,
			})
			// Start new session
			sessionStart = e.Timestamp
			sessionEvents = 1
		} else {
			sessionEvents++
		}
		lastTime = e.Timestamp
	}

	// Close final session
	duration := lastTime.Sub(sessionStart)
	if duration < 5*time.Minute {
		duration = 5 * time.Minute
	}
	sessions = append(sessions, Session{
		Start:    sessionStart,
		End:      lastTime,
		Duration: duration,
		Events:   sessionEvents,
	})

	return sessions
}

// FormatMarkdown outputs a markdown report
func (r *Report) FormatMarkdown() string {
	var sb strings.Builder

	sb.WriteString("# Work Report\n\n")
	sb.WriteString(fmt.Sprintf("**Period:** %s to %s\n\n",
		r.StartDate.Format("Jan 2, 2006"),
		r.EndDate.Format("Jan 2, 2006")))

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Time:** %s\n", formatDuration(r.Total.Time)))
	sb.WriteString(fmt.Sprintf("- **Projects:** %d\n", r.Total.Projects))
	sb.WriteString(fmt.Sprintf("- **Events:** %d\n", r.Total.Events))
	sb.WriteString(fmt.Sprintf("- **Active Days:** %d\n\n", r.Total.Days))

	// By project
	sb.WriteString("## By Project\n\n")
	sb.WriteString("| Project | Time | Events | Days | Sessions |\n")
	sb.WriteString("|---------|------|--------|------|----------|\n")
	for _, p := range r.Projects {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %d |\n",
			p.Name,
			formatDuration(p.TotalTime),
			len(p.Events),
			p.Days,
			len(p.Sessions)))
	}
	sb.WriteString("\n")

	// Detailed breakdown
	sb.WriteString("## Session Details\n\n")
	for _, p := range r.Projects {
		sb.WriteString(fmt.Sprintf("### %s\n\n", p.Name))
		for _, s := range p.Sessions {
			sb.WriteString(fmt.Sprintf("- **%s** %s-%s (%s, %d events)\n",
				s.Start.Format("Jan 2"),
				s.Start.Format("15:04"),
				s.End.Format("15:04"),
				formatDuration(s.Duration),
				s.Events))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatPlain outputs a simple text report
func (r *Report) FormatPlain() string {
	var sb strings.Builder

	sb.WriteString("WORK REPORT\n")
	sb.WriteString(strings.Repeat("=", 50) + "\n\n")
	sb.WriteString(fmt.Sprintf("Period: %s to %s\n\n",
		r.StartDate.Format("Jan 2, 2006"),
		r.EndDate.Format("Jan 2, 2006")))

	sb.WriteString(fmt.Sprintf("Total Time:   %s\n", formatDuration(r.Total.Time)))
	sb.WriteString(fmt.Sprintf("Projects:     %d\n", r.Total.Projects))
	sb.WriteString(fmt.Sprintf("Events:       %d\n", r.Total.Events))
	sb.WriteString(fmt.Sprintf("Active Days:  %d\n\n", r.Total.Days))

	sb.WriteString("BY PROJECT\n")
	sb.WriteString(strings.Repeat("-", 50) + "\n")
	for _, p := range r.Projects {
		sb.WriteString(fmt.Sprintf("%-20s %8s  %3d events  %2d days\n",
			truncate(p.Name, 20),
			formatDuration(p.TotalTime),
			len(p.Events),
			p.Days))
	}
	sb.WriteString("\n")

	sb.WriteString("SESSION LOG\n")
	sb.WriteString(strings.Repeat("-", 50) + "\n")
	for _, p := range r.Projects {
		sb.WriteString(fmt.Sprintf("\n[%s]\n", p.Name))
		for _, s := range p.Sessions {
			sb.WriteString(fmt.Sprintf("  %s  %s-%s  %s\n",
				s.Start.Format("Jan 02"),
				s.Start.Format("15:04"),
				s.End.Format("15:04"),
				formatDuration(s.Duration)))
		}
	}

	return sb.String()
}

// FormatCSV outputs CSV for spreadsheets
func (r *Report) FormatCSV() string {
	var sb strings.Builder
	sb.WriteString("Project,Date,Start,End,Duration (min),Events\n")
	for _, p := range r.Projects {
		for _, s := range p.Sessions {
			sb.WriteString(fmt.Sprintf("%s,%s,%s,%s,%.0f,%d\n",
				p.Name,
				s.Start.Format("2006-01-02"),
				s.Start.Format("15:04"),
				s.End.Format("15:04"),
				s.Duration.Minutes(),
				s.Events))
		}
	}
	return sb.String()
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
