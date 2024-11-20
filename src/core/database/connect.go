package database

import (
	"fmt"
	"log"
	"Backend/src/core/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var DB *gorm.DB

func ConnectDB() {
	// Fetch configuration values from environment or config files
	host := config.Config("DB_HOST")
	port := config.Config("DB_PORT")
	user := config.Config("DB_USER")
	password := config.Config("DB_PASSWORD")
	dbname := config.Config("DB_NAME")

	// Build the connection string
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	// Connect to the database with a custom config
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		// Disable automatic statement caching by setting this flag
		PrepareStmt: false, // This disables prepared statement caching in GORM

		// Disable automatic schema reflection in GORM (optional but may help in some cases)
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "", // Custom table prefix (if you have any)
			SingularTable: false, // Disable singular table names
		},
	})
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	fmt.Println("Database successfully connected!")
}
