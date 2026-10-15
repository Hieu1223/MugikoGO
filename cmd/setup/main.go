package main

import (
	"fmt"

	"mugiko/internal/common/persistent"
)

func main() {
	// Initialize database
	persistent.InitDB()
	defer persistent.CloseDB()

	fmt.Println("Database setup complete!")
	fmt.Println("Tables created: user_auths, user_profiles")
	
	// Create admin user if not exists
	createAdminIfNotExists()
	
	fmt.Println("Setup completed successfully!")
}

func createAdminIfNotExists() {
	// Admin user creation would go here
	// For now, just print info
	fmt.Println("Admin user creation skipped - implement as needed")
}