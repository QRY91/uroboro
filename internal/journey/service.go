package journey

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/database"
)

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GenerateJourney(opts Options) (*JourneyData, error) {
	dateRange := s.getDateRange(opts)

	captures, err := s.getCapturesInRange(dateRange, opts.Projects)
	if err != nil {
		return nil, fmt.Errorf("get captures: %w", err)
	}

	events := s.capturesToEvents(captures)

	// Collect git commits from CWD and any additional repos
	var allCommitEvents []TimelineEvent
	cwdCommits := s.getGitCommits(dateRange, "")
	allCommitEvents = append(allCommitEvents, s.commitsToEvents(cwdCommits, "git")...)
	for _, repo := range opts.Repos {
		repoCommits := s.getGitCommits(dateRange, repo.Path)
		name := repo.Name
		if name == "" {
			// Use directory basename
			parts := strings.Split(strings.TrimRight(repo.Path, "/"), "/")
			name = parts[len(parts)-1]
		}
		allCommitEvents = append(allCommitEvents, s.commitsToEvents(repoCommits, name)...)
	}
	allEvents := append(events, allCommitEvents...)

	sort.Slice(allEvents, func(i, j int) bool {
		return allEvents[i].Timestamp.After(allEvents[j].Timestamp)
	})

	projects := s.generateProjectSummaries(allEvents)

	return &JourneyData{
		Events: allEvents,
		Timeline: Timeline{
			StartTime:     dateRange.Start,
			EndTime:       dateRange.End,
			TotalDuration: dateRange.End.Sub(dateRange.Start).Milliseconds(),
		},
		Projects:   projects,
		Stats:      s.calcStats(allEvents, projects),
		Milestones: s.findMilestones(allEvents),
	}, nil
}

func (s *Service) getDateRange(opts Options) DateRange {
	if opts.DateRange != nil {
		return *opts.DateRange
	}
	end := time.Now()
	return DateRange{Start: end.AddDate(0, 0, -opts.Days), End: end}
}

func (s *Service) getCapturesInRange(dr DateRange, projects []string) ([]database.Capture, error) {
	q := database.CaptureQuery{Since: &dr.Start}

	if len(projects) == 1 {
		q.Project = projects[0]
	}

	captures, err := s.db.QueryCaptures(q)
	if err != nil {
		return nil, err
	}

	// Filter by end date and multiple projects if needed
	var result []database.Capture
	for _, c := range captures {
		if c.Timestamp.After(dr.End) {
			continue
		}
		if len(projects) > 1 {
			found := false
			for _, p := range projects {
				if c.Project == p {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		result = append(result, c)
	}
	return result, nil
}

func (s *Service) getGitCommits(dr DateRange, repoPath string) []GitCommit {
	var cmd *exec.Cmd
	if repoPath != "" {
		cmd = exec.Command("git", "-C", repoPath, "log",
			"--pretty=format:%H|%s|%at|%an",
			fmt.Sprintf("--since=%s", dr.Start.Format("2006-01-02")),
			fmt.Sprintf("--until=%s", dr.End.Format("2006-01-02")))
	} else {
		cmd = exec.Command("git", "log",
			"--pretty=format:%H|%s|%at|%an",
			fmt.Sprintf("--since=%s", dr.Start.Format("2006-01-02")),
			fmt.Sprintf("--until=%s", dr.End.Format("2006-01-02")))
	}

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var commits []GitCommit
	for _, line := range strings.Split(string(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) >= 4 {
			var unix int64
			fmt.Sscanf(parts[2], "%d", &unix)
			ts := time.Unix(unix, 0)
			commits = append(commits, GitCommit{
				Hash: parts[0], Message: parts[1], Timestamp: ts, Author: parts[3],
			})
		}
	}
	return commits
}

func (s *Service) capturesToEvents(captures []database.Capture) []TimelineEvent {
	var events []TimelineEvent
	for _, c := range captures {
		var tags []string
		if c.Tags != "" {
			for _, t := range strings.Split(c.Tags, ",") {
				tags = append(tags, strings.TrimSpace(t))
			}
		}
		events = append(events, TimelineEvent{
			Timestamp:  c.Timestamp,
			Content:    c.Content,
			Project:    c.Project,
			Branch:     c.Branch,
			Tags:       tags,
			EventType:  s.determineEventType(c.Content, c.Tags),
			Importance: s.calcImportance(c.Content, c.Tags),
		})
	}
	return events
}

func (s *Service) commitsToEvents(commits []GitCommit, projectName string) []TimelineEvent {
	var events []TimelineEvent
	for _, c := range commits {
		events = append(events, TimelineEvent{
			Timestamp:  c.Timestamp,
			Content:    c.Message,
			Project:    projectName,
			Tags:       []string{"git", "commit"},
			EventType:  EventTypeCommit,
			Importance: s.calcCommitImportance(c.Message),
			GitHash:    c.Hash,
		})
	}
	return events
}

func (s *Service) generateProjectSummaries(events []TimelineEvent) []ProjectSummary {
	pm := make(map[string]*ProjectSummary)
	colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FECA57", "#FF9FF3", "#54A0FF", "#5F27CD"}
	ci := 0

	for _, e := range events {
		if e.Project == "" {
			continue
		}
		if _, ok := pm[e.Project]; !ok {
			pm[e.Project] = &ProjectSummary{
				Name: e.Project, Color: colors[ci%len(colors)],
				StartDate: e.Timestamp, LastActive: e.Timestamp,
			}
			ci++
		}
		p := pm[e.Project]
		p.EventCount++
		if e.Timestamp.Before(p.StartDate) {
			p.StartDate = e.Timestamp
		}
		if e.Timestamp.After(p.LastActive) {
			p.LastActive = e.Timestamp
		}
	}

	var result []ProjectSummary
	for _, p := range pm {
		result = append(result, *p)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].EventCount > result[j].EventCount
	})
	return result
}

func (s *Service) calcStats(events []TimelineEvent, projects []ProjectSummary) JourneyStats {
	stats := JourneyStats{
		TotalEvents:  len(events),
		ProjectCount: len(projects),
	}
	var totalImp int
	for _, e := range events {
		totalImp += e.Importance
		if e.EventType == EventTypeMilestone {
			stats.MilestoneCount++
		}
		if e.EventType == EventTypeLearning || hasTag(e.Tags, "learning") {
			stats.LearningMoments++
		}
	}
	if len(events) > 0 {
		stats.ProductivityScore = float64(totalImp) / float64(len(events))
	}
	return stats
}

func (s *Service) findMilestones(events []TimelineEvent) []TimelineEvent {
	var ms []TimelineEvent
	for _, e := range events {
		if e.EventType == EventTypeMilestone || e.Importance >= ImportanceHigh {
			ms = append(ms, e)
		}
	}
	return ms
}

func (s *Service) determineEventType(content, tags string) string {
	lc, lt := strings.ToLower(content), strings.ToLower(tags)
	switch {
	case strings.Contains(lt, "milestone") || strings.Contains(lc, "milestone"):
		return EventTypeMilestone
	case strings.Contains(lt, "learning") || strings.Contains(lc, "learned"):
		return EventTypeLearning
	case strings.Contains(lt, "decision") || strings.Contains(lc, "decided"):
		return EventTypeDecision
	case strings.Contains(lt, "bug") || strings.Contains(lc, "fixed"):
		return EventTypeBugfix
	case strings.Contains(lt, "feature") || strings.Contains(lc, "implemented"):
		return EventTypeFeature
	case strings.Contains(lt, "blocker"):
		return EventTypeBlocker
	case strings.Contains(lt, "question"):
		return EventTypeQuestion
	default:
		return EventTypeCapture
	}
}

func (s *Service) calcImportance(content, tags string) int {
	lc, lt := strings.ToLower(content), strings.ToLower(tags)
	switch {
	case strings.Contains(lt, "critical") || strings.Contains(lc, "critical"):
		return ImportanceCritical
	case strings.Contains(lt, "milestone") || strings.Contains(lt, "important"):
		return ImportanceHigh
	case strings.Contains(lt, "decision") || strings.Contains(lt, "integration"):
		return ImportanceMedium
	default:
		return ImportanceLow
	}
}

func (s *Service) calcCommitImportance(msg string) int {
	m := strings.ToLower(msg)
	switch {
	case strings.Contains(m, "feat"):
		return ImportanceHigh
	case strings.Contains(m, "fix") || strings.Contains(m, "refactor"):
		return ImportanceMedium
	default:
		return ImportanceLow
	}
}

func hasTag(tags []string, target string) bool {
	for _, t := range tags {
		if strings.ToLower(t) == target {
			return true
		}
	}
	return false
}

type GitCommit struct {
	Hash, Message, Author string
	Timestamp             time.Time
}
