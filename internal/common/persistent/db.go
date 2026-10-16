package persistent

import (
	"database/sql"
	"log"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	db   *sql.DB
	once sync.Once
)

func InitDB() *sql.DB {
	once.Do(func() {
		dbPath := getDBPath()
		var err error
		db, err = sql.Open("postgres", dbPath)
		if err != nil {
			log.Fatalf("Failed to open database: %v", err)
		}

		if err = db.Ping(); err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
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
	// Default to a local PostgreSQL connection string
	return "postgres://postgres:postgres@localhost:5432/mugiko?sslmode=disable"
}

func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
