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
		driver := getDBDriver()
		var err error
		db, err = sql.Open(driver, dbPath)
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
	path := os.Getenv("DB_PATH")
	if path == "" {
		panic("Env must contain DB_PATH")
	}
	return path
}

func getDBDriver() string {
	path := os.Getenv("DB_DRIVER")
	if path == "" {
		panic("Env must contain DB_DRIVER")
	}
	return path
}

func CloseDB() error {
	if db != nil {
		return db.Close()
	}
	return nil
}
