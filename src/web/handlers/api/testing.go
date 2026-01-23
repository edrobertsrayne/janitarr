package api

import (
	"testing"

	"github.com/edrobertsrayne/janitarr/src/database"
)

// testDB creates a new in-memory database for testing.
func testDB(t *testing.T) *database.DB {
	t.Helper()
	db, err := database.New(":memory:", t.TempDir()+"/key")
	if err != nil {
		t.Fatalf("creating test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
