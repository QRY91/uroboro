package database

import (
	"os"
	"testing"
)

func TestDatabaseIntegration(t *testing.T) {
	// Create temporary database
	tmpFile := "/tmp/test_uroboro.sqlite"
	defer os.Remove(tmpFile)

	db, err := NewDB(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	t.Run("InsertCapture", func(t *testing.T) {
		capture, err := db.InsertCapture("Test capture content", "testproject", "tag1,tag2")
		if err != nil {
			t.Fatalf("Failed to insert capture: %v", err)
		}

		if capture.ID == 0 {
			t.Error("Expected capture ID to be set")
		}

		if capture.Content != "Test capture content" {
			t.Errorf("Expected content 'Test capture content', got '%s'", capture.Content)
		}

		if !capture.Project.Valid || capture.Project.String != "testproject" {
			t.Errorf("Expected project 'testproject', got %v", capture.Project)
		}

		if !capture.Tags.Valid || capture.Tags.String != "tag1,tag2" {
			t.Errorf("Expected tags 'tag1,tag2', got %v", capture.Tags)
		}
	})

	t.Run("GetRecentCaptures", func(t *testing.T) {
		// Insert test captures
		_, err := db.InsertCapture("Recent capture 1", "project1", "")
		if err != nil {
			t.Fatalf("Failed to insert test capture: %v", err)
		}

		_, err = db.InsertCapture("Recent capture 2", "project2", "test")
		if err != nil {
			t.Fatalf("Failed to insert test capture: %v", err)
		}

		// Get recent captures
		captures, err := db.GetRecentCaptures(1, "")
		if err != nil {
			t.Fatalf("Failed to get recent captures: %v", err)
		}

		if len(captures) < 2 {
			t.Errorf("Expected at least 2 captures, got %d", len(captures))
		}

		// Test project filtering
		projectCaptures, err := db.GetRecentCaptures(1, "project1")
		if err != nil {
			t.Fatalf("Failed to get project captures: %v", err)
		}

		found := false
		for _, capture := range projectCaptures {
			if capture.Project.Valid && capture.Project.String == "project1" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Expected to find capture with project1")
		}
	})

	t.Run("InsertPublication", func(t *testing.T) {
		sourceCaptureIDs := []int64{1, 2}
		pub, err := db.InsertPublication(
			"Test Publication",
			"Publication content",
			"markdown",
			"blog",
			"testproject",
			"/tmp/test.md",
			sourceCaptureIDs,
		)

		if err != nil {
			t.Fatalf("Failed to insert publication: %v", err)
		}

		if pub.ID == 0 {
			t.Error("Expected publication ID to be set")
		}

		if pub.Title != "Test Publication" {
			t.Errorf("Expected title 'Test Publication', got '%s'", pub.Title)
		}

		if pub.Type != "blog" {
			t.Errorf("Expected type 'blog', got '%s'", pub.Type)
		}
	})

	t.Run("NullHandling", func(t *testing.T) {
		// Test with empty/null values
		capture, err := db.InsertCapture("Content only", "", "")
		if err != nil {
			t.Fatalf("Failed to insert capture with nulls: %v", err)
		}

		if capture.Project.Valid {
			t.Error("Expected project to be null/invalid")
		}

		if capture.Tags.Valid {
			t.Error("Expected tags to be null/invalid")
		}

		// Verify we can read it back
		captures, err := db.GetRecentCaptures(1, "")
		if err != nil {
			t.Fatalf("Failed to get captures: %v", err)
		}

		found := false
		for _, c := range captures {
			if c.Content == "Content only" {
				found = true
				if c.Project.Valid && c.Project.String != "" {
					t.Error("Expected project to remain null")
				}
				break
			}
		}

		if !found {
			t.Error("Failed to find the null-field capture")
		}
	})
}

func TestSchemaCreation(t *testing.T) {
	tmpFile := "/tmp/test_schema.sqlite"
	defer os.Remove(tmpFile)

	db, err := NewDB(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify tables exist
	tables := []string{"captures", "publications", "projects", "tool_messages", "usage_stats", "schema_migrations"}

	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check table %s: %v", table, err)
		}

		if count != 1 {
			t.Errorf("Expected table %s to exist", table)
		}
	}

	// Verify migration record
	var version int
	err = db.QueryRow("SELECT version FROM schema_migrations WHERE version = 1").Scan(&version)
	if err != nil {
		t.Fatalf("Failed to find migration record: %v", err)
	}

	if version != 1 {
		t.Errorf("Expected migration version 1, got %d", version)
	}
}

func TestConcurrentAccess(t *testing.T) {
	tmpFile := "/tmp/test_concurrent.sqlite"
	defer os.Remove(tmpFile)

	db, err := NewDB(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test concurrent writes
	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 5; i++ {
			_, err := db.InsertCapture("Concurrent capture A", "projectA", "")
			if err != nil {
				t.Errorf("Concurrent insert A failed: %v", err)
			}
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 5; i++ {
			_, err := db.InsertCapture("Concurrent capture B", "projectB", "")
			if err != nil {
				t.Errorf("Concurrent insert B failed: %v", err)
			}
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Verify all captures were inserted
	captures, err := db.GetRecentCaptures(1, "")
	if err != nil {
		t.Fatalf("Failed to get captures after concurrent test: %v", err)
	}

	if len(captures) < 10 {
		t.Errorf("Expected at least 10 captures from concurrent test, got %d", len(captures))
	}
}
