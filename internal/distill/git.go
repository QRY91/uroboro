package distill

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	// StyleFilterRegex matches commit messages that signal intentional style/design work.
	StyleFilterRegex = `refactor|clean.?up|simplif|extract|rename|reorganiz|restructur|` +
		`consolidat|dedup|inline|flatten|split|decompos|modular|` +
		`move.*to|pull.*out|reduce.*complex|improve.*read|` +
		`untangle|decouple|encapsulat|abstract|normalize`

	MaxFilesPerCommit = 20
	MaxDiffBytes      = 50 * 1024 // 50KB
)

type commitMeta struct {
	hash       string
	parentHash string
	date       time.Time
	message    string
}

// noiseOnlyExts are file extensions that carry no code style signal.
var noiseOnlyExts = map[string]bool{
	".md": true, ".txt": true, ".lock": true, ".sum": true,
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".svg": true,
	".ico": true,
}

// noiseOnlyNames are exact filenames that carry no code style signal.
var noiseOnlyNames = map[string]bool{
	"package-lock.json": true, "yarn.lock": true, "Pipfile.lock": true,
	"poetry.lock": true, "Cargo.lock": true,
}

// isNoiseOnlyCommit returns true when every touched file is a noise file
// (docs, lockfiles, images) — no code style signal.
func isNoiseOnlyCommit(files []string) bool {
	if len(files) == 0 {
		return false
	}
	for _, f := range files {
		base := filepath.Base(f)
		if noiseOnlyNames[base] {
			continue
		}
		ext := strings.ToLower(filepath.Ext(f))
		if noiseOnlyExts[ext] {
			continue
		}
		return false // at least one non-noise file
	}
	return true
}

type GitExtractor struct {
	RepoPath   string
	Since      *time.Time
	AllCommits bool // when true, skip StyleFilterRegex
	MaxPerRepo int  // 0 = unlimited; cap to most recent N after extraction
}

func NewGitExtractor(repoPath string, since *time.Time) *GitExtractor {
	return &GitExtractor{RepoPath: repoPath, Since: since}
}

func (e *GitExtractor) Extract() ([]GitExtract, error) {
	commits, err := e.listMatchingCommits()
	if err != nil {
		return nil, err
	}

	repo := filepath.Base(e.RepoPath)

	var results []GitExtract
	for _, c := range commits {
		// Skip root commits (no parent)
		if c.parentHash == "" {
			continue
		}

		files, stats, err := e.getDiffStats(c.hash)
		if err != nil || stats.Files > MaxFilesPerCommit {
			continue
		}

		// Skip commits that only touch noise files (docs, lockfiles, images)
		if isNoiseOnlyCommit(files) {
			continue
		}

		diff, err := e.getDiff(c.parentHash, c.hash)
		if err != nil {
			continue
		}

		results = append(results, GitExtract{
			Source:     "git",
			Repo:       repo,
			Hash:       c.hash,
			ParentHash: c.parentHash,
			Message:    c.message,
			Date:       c.date,
			Files:      files,
			Language:   detectLanguage(files),
			Diff:       diff,
			DiffStats:  stats,
		})
	}

	// Cap per-repo if requested (results are already in date-desc order from git log)
	if e.MaxPerRepo > 0 && len(results) > e.MaxPerRepo {
		results = results[:e.MaxPerRepo]
	}

	return results, nil
}

func (e *GitExtractor) gitCmd(args ...string) (string, error) {
	fullArgs := append([]string{"-C", e.RepoPath}, args...)
	out, err := exec.Command("git", fullArgs...).Output()
	return string(out), err
}

// listMatchingCommits runs git log, optionally filtered by style-signal regex.
// Format: hash|parent|authordate_iso|subject
func (e *GitExtractor) listMatchingCommits() ([]commitMeta, error) {
	args := []string{"log", "--format=%H|%P|%aI|%s", "--no-merges"}
	if !e.AllCommits {
		args = append(args, "--extended-regexp", "--grep="+StyleFilterRegex, "-i")
	}
	if e.MaxPerRepo > 0 {
		// Ask git to stop early — avoids traversing huge repos when we only want the top N.
		// We request 2× the cap to leave room for post-extraction filtering (noise, root commits).
		args = append(args, fmt.Sprintf("--max-count=%d", e.MaxPerRepo*2))
	}
	if e.Since != nil {
		args = append(args, "--since="+e.Since.Format("2006-01-02"))
	}

	out, err := e.gitCmd(args...)
	if err != nil {
		return nil, err
	}

	var commits []commitMeta
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 4)
		if len(parts) != 4 {
			continue
		}

		ts, _ := time.Parse(time.RFC3339, parts[2])

		// %P can have multiple parents (merge commits filtered by --no-merges,
		// but handle gracefully). Take first parent only.
		parent := strings.Fields(parts[1])
		parentHash := ""
		if len(parent) > 0 {
			parentHash = parent[0]
		}

		commits = append(commits, commitMeta{
			hash:       parts[0],
			parentHash: parentHash,
			date:       ts,
			message:    parts[3],
		})
	}
	return commits, nil
}

// getDiffStats runs git show --numstat to get file list and line counts.
func (e *GitExtractor) getDiffStats(hash string) ([]string, DiffStats, error) {
	out, err := e.gitCmd("show", "--numstat", "--format=", hash)
	if err != nil {
		return nil, DiffStats{}, err
	}

	var files []string
	var stats DiffStats
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		add, _ := strconv.Atoi(fields[0])
		del, _ := strconv.Atoi(fields[1])
		stats.Additions += add
		stats.Deletions += del
		stats.Files++
		files = append(files, fields[2])
	}
	return files, stats, nil
}

// getDiff gets the unified diff between parent and commit, capped at MaxDiffBytes.
func (e *GitExtractor) getDiff(parent, hash string) (string, error) {
	out, err := e.gitCmd("diff", parent, hash)
	if err != nil {
		return "", err
	}
	if len(out) > MaxDiffBytes {
		return out[:MaxDiffBytes] + "\n[truncated]", nil
	}
	return out, nil
}

// detectLanguage returns the plurality language from file extensions.
func detectLanguage(files []string) string {
	counts := make(map[string]int)
	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f))
		if lang, ok := LanguageFromExt[ext]; ok {
			counts[lang]++
		}
	}
	best := "unknown"
	bestCount := 0
	for lang, count := range counts {
		if count > bestCount {
			bestCount = count
			best = lang
		}
	}
	return best
}
