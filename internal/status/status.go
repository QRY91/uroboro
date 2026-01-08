package status

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/QRY91/uroboro/internal/common"
	"github.com/QRY91/uroboro/internal/database"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Show(days int, dbPath, project string) error {
	fmt.Println("🐍 uroboro status")

	if dbPath != "" {
		return s.showFromDB(days, dbPath, project)
	}
	return s.showFromFiles(days)
}

func (s *Service) showFromDB(days int, dbPath, project string) error {
	db, err := database.NewDB(dbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	captures, err := db.GetRecentCaptures(days, project)
	if err != nil {
		return err
	}

	fmt.Printf("Recent activity (%d days): %d items\n\n", days, len(captures))
	fmt.Printf("📝 Recent Captures (last %d days):\n", days)

	if len(captures) == 0 {
		fmt.Println("  No recent captures found")
		return nil
	}

	for i, c := range captures {
		if i >= 10 {
			break
		}
		content := c.Content
		if len(content) > 80 {
			content = content[:80] + "..."
		}
		if c.Project != "" {
			fmt.Printf("  📄 [%s] %s\n", c.Project, content)
		} else {
			fmt.Printf("  📄 %s\n", content)
		}
	}

	return nil
}

func (s *Service) showFromFiles(days int) error {
	dataDir := common.GetDataDir()
	cutoff := time.Now().AddDate(0, 0, -days)

	entries, err := os.ReadDir(dataDir)
	if err != nil {
		fmt.Printf("Recent activity (%d days): 0 items\n\n", days)
		fmt.Printf("📝 Recent Captures (last %d days):\n", days)
		fmt.Println("  No recent captures found")
		return nil
	}

	var count int
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			if info, err := e.Info(); err == nil && info.ModTime().After(cutoff) {
				count++
			}
		}
	}

	fmt.Printf("Recent activity (%d days): %d items\n\n", days, count)
	fmt.Printf("📝 Recent Captures (last %d days):\n", days)

	shown := 0
	for _, e := range entries {
		if shown >= 10 {
			break
		}
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		info, err := e.Info()
		if err != nil || !info.ModTime().After(cutoff) {
			continue
		}

		content, err := os.ReadFile(filepath.Join(dataDir, e.Name()))
		if err != nil {
			continue
		}

		for _, line := range strings.Split(string(content), "\n") {
			if shown >= 10 {
				break
			}
			// Skip headers and metadata
			if strings.HasPrefix(line, "##") || strings.HasPrefix(line, "Project:") ||
				strings.HasPrefix(line, "Tags:") || strings.TrimSpace(line) == "" {
				continue
			}
			text := strings.TrimSpace(line)
			if len(text) > 80 {
				text = text[:80] + "..."
			}
			if text != "" {
				fmt.Printf("  📄 %s\n", text)
				shown++
			}
		}
	}

	if shown == 0 {
		fmt.Println("  No recent captures found")
	}

	return nil
}
