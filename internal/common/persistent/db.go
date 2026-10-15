package persistent

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	db   *sql.DB
	once sync.Once
)

func InitDB() *sql.DB {
	once.Do(func() {
		dbPath := getDBPath()
		var err error
		db, err = sql.Open("sqlite", dbPath+"?_pragma=foreign_keys(1)")
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}

		if err = db.Ping(); err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		if err = runMigrations(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}

		log.Println("Database initialized successfully")
	})
	return db
}

func GetDB() *sql.DB {
	if db == nil {
		return InitDB()
	}
	return db
}

func getDBPath() string {
	if path := os.Getenv("DB_PATH"); path != "" {
		return path
	}
	return "./mugiko.db"
}

func runMigrations() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS user_auths (
			user_id TEXT PRIMARY KEY,
			role TEXT NOT NULL DEFAULT 'anonymous',
			joined_at DATETIME NOT NULL DEFAULT (datetime('now')),
			auth_type TEXT NOT NULL,
			auth_id TEXT NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT 1,
			UNIQUE(auth_type, auth_id)
		)`,
		`CREATE INDEX IF NOT EXISTS ix_user_auths_auth_type_auth_id ON user_auths(auth_type, auth_id)`,
		`CREATE TABLE IF NOT EXISTS user_profiles (
			user_id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			bio TEXT,
			email TEXT,
			FOREIGN KEY (user_id) REFERENCES user_auths(user_id) ON DELETE CASCADE
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}
	return nil
}

func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
