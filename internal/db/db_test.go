package db_test

import (
	"testing"

	"github.com/andreypisarev/secret-santa/internal/db"
)

func TestOpenAndMigrate(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	tables := []string{"users", "groups", "memberships", "magic_links", "sessions", "messages"}
	for _, table := range tables {
		var name string
		err := database.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer database.Close()

	if err := db.Migrate(database); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if err := db.Migrate(database); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
}
