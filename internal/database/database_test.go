package database

import (
	"os"
	"testing"
)

func TestDatabaseIntegration(t *testing.T) {
	tmpFile := t.TempDir() + "/test_uroboro.sqlite"

	db, err := NewDB(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	t.Run("InsertCapture", func(t *testing.T) {
		capture, err := db.InsertCapture("Test capture content", "testproject", "tag1,tag2", "", "", nil)
		if err != nil {
			t.Fatalf("Failed to insert capture: %v", err)
		}

		if capture.ID == 0 {
			t.Error("Expected capture ID to be set")
		}
		if capture.Content != "Test capture content" {
			t.Errorf("Expected content 'Test capture content', got '%s'", capture.Content)
		}
		if capture.Project != "testproject" {
			t.Errorf("Expected project 'testproject', got '%s'", capture.Project)
		}
		if capture.Tags != "tag1,tag2" {
			t.Errorf("Expected tags 'tag1,tag2', got '%s'", capture.Tags)
		}
	})

	t.Run("GetRecentCaptures", func(t *testing.T) {
		_, err := db.InsertCapture("Recent capture 1", "project1", "", "", "", nil)
		if err != nil {
			t.Fatalf("Failed to insert: %v", err)
		}
		_, err = db.InsertCapture("Recent capture 2", "project2", "test", "", "", nil)
		if err != nil {
			t.Fatalf("Failed to insert: %v", err)
		}

		captures, err := db.GetRecentCaptures(1, "")
		if err != nil {
			t.Fatalf("Failed to get recent captures: %v", err)
		}
		if len(captures) < 2 {
			t.Errorf("Expected at least 2 captures, got %d", len(captures))
		}

		projectCaptures, err := db.GetRecentCaptures(1, "project1")
		if err != nil {
			t.Fatalf("Failed to get project captures: %v", err)
		}

		found := false
		for _, c := range projectCaptures {
			if c.Project == "project1" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected to find capture with project1")
		}
	})

	t.Run("EmptyFields", func(t *testing.T) {
		capture, err := db.InsertCapture("Content only", "", "", "", "", nil)
		if err != nil {
			t.Fatalf("Failed to insert capture with empty fields: %v", err)
		}
		if capture.Project != "" {
			t.Errorf("Expected empty project, got '%s'", capture.Project)
		}
		if capture.Tags != "" {
			t.Errorf("Expected empty tags, got '%s'", capture.Tags)
		}
	})

	t.Run("CaptureCount", func(t *testing.T) {
		count, err := db.CaptureCount()
		if err != nil {
			t.Fatalf("CaptureCount failed: %v", err)
		}
		if count < 4 {
			t.Errorf("Expected at least 4 captures, got %d", count)
		}
	})

	t.Run("QueryCaptures", func(t *testing.T) {
		captures, err := db.QueryCaptures(CaptureQuery{
			Keyword: "Recent",
			Limit:   10,
		})
		if err != nil {
			t.Fatalf("QueryCaptures failed: %v", err)
		}
		if len(captures) < 2 {
			t.Errorf("Expected at least 2 captures matching 'Recent', got %d", len(captures))
		}

		captures, err = db.QueryCaptures(CaptureQuery{
			Tags:  []string{"tag1"},
			Limit: 10,
		})
		if err != nil {
			t.Fatalf("QueryCaptures with tags failed: %v", err)
		}
		if len(captures) == 0 {
			t.Error("Expected captures with tag1")
		}
	})

	t.Run("VacuumInto", func(t *testing.T) {
		backupPath := t.TempDir() + "/backup.sqlite"
		if err := db.VacuumInto(backupPath); err != nil {
			t.Fatalf("VacuumInto failed: %v", err)
		}
		if _, err := os.Stat(backupPath); os.IsNotExist(err) {
			t.Error("Backup file was not created")
		}

		backupDB, err := NewDB(backupPath)
		if err != nil {
			t.Fatalf("Failed to open backup: %v", err)
		}
		defer backupDB.Close()

		origCount, _ := db.CaptureCount()
		backupCount, _ := backupDB.CaptureCount()
		if origCount != backupCount {
			t.Errorf("Backup has %d captures, expected %d", backupCount, origCount)
		}
	})
}

func TestConcurrentAccess(t *testing.T) {
	tmpFile := t.TempDir() + "/test_concurrent.sqlite"

	db, err := NewDB(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 5; i++ {
			if _, err := db.InsertCapture("Concurrent capture A", "projectA", "", "", "", nil); err != nil {
				t.Errorf("Concurrent insert A failed: %v", err)
			}
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 5; i++ {
			if _, err := db.InsertCapture("Concurrent capture B", "projectB", "", "", "", nil); err != nil {
				t.Errorf("Concurrent insert B failed: %v", err)
			}
		}
		done <- true
	}()

	<-done
	<-done

	captures, err := db.GetRecentCaptures(1, "")
	if err != nil {
		t.Fatalf("Failed to get captures: %v", err)
	}
	if len(captures) < 10 {
		t.Errorf("Expected at least 10 captures, got %d", len(captures))
	}
}
