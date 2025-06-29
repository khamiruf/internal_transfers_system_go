package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/lib/pq"
)

// NewTestDB creates a new test database connection
func NewTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dbURL := getEnvOrDefault("TEST_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/transfers_test?sslmode=disable")

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

// SetupTestDB sets up a test database with the schema
func SetupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	// Get the path to the migration file
	_, b, _, _ := runtime.Caller(0)
	projectRoot := filepath.Join(filepath.Dir(b), "../..")
	migrationPath := filepath.Join(projectRoot, "migrations", "001_init.sql")

	// Read and execute the migration file
	migration, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("Failed to read migration file: %v", err)
	}

	_, err = db.Exec(string(migration))
	if err != nil {
		t.Fatalf("Failed to execute migration: %v", err)
	}
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(t *testing.T, db *sql.DB) {
	t.Helper()

	tables := []string{"transactions", "accounts"}
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Errorf("Failed to clean up table %s: %v", table, err)
		}
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
