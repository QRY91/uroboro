package status

import (
	"testing"

	"github.com/QRY91/uroboro/internal/database"
)

func TestShowFromDB(t *testing.T) {
	dbPath := t.TempDir() + "/test.sqlite"

	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	db.InsertCapture("Test capture 1", "project1", "tag1", "", nil)
	db.InsertCapture("Test capture 2", "project2", "tag2", "", nil)
	db.Close()

	svc := NewService()
	err = svc.Show(7, dbPath, "")
	if err != nil {
		t.Errorf("Show failed: %v", err)
	}
}

func TestShowFromDBEmpty(t *testing.T) {
	dbPath := t.TempDir() + "/test.sqlite"

	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	db.Close()

	svc := NewService()
	err = svc.Show(7, dbPath, "")
	if err != nil {
		t.Errorf("Show with empty DB should not error: %v", err)
	}
}

func TestShowWithProjectFilter(t *testing.T) {
	dbPath := t.TempDir() + "/test.sqlite"

	db, err := database.NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create DB: %v", err)
	}
	db.InsertCapture("Project A work", "projectA", "", "", nil)
	db.InsertCapture("Project B work", "projectB", "", "", nil)
	db.Close()

	svc := NewService()
	err = svc.Show(7, dbPath, "projectA")
	if err != nil {
		t.Errorf("Show with project filter failed: %v", err)
	}
}
