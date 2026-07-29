package service

import (
	"os"
	"testing"

	"vortexuipro/internal/database"
)

// TestMain initialises a single shared in-memory SQLite database for the
// entire service test suite. This avoids concurrent writes to the global
// database.DB variable when the race detector is active.
func TestMain(m *testing.M) {
	if err := database.InitDB(database.Config{
		Type:     "sqlite",
		DSN:      ":memory:",
		LogLevel: "silent",
	}); err != nil {
		panic("failed to init test database: " + err.Error())
	}
	os.Exit(m.Run())
}
