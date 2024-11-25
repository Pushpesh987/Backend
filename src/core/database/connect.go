package database

import (
	"fmt"
	"log"
	"time"
	"strconv"

	"Backend/src/core/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() {
	// Fetch configuration values from environment or config files
	host := config.Config("DB_HOST")
	port := config.Config("DB_PORT")
	user := config.Config("DB_USER")
	password := config.Config("DB_PASSWORD")
	dbname := config.Config("DB_NAME")

	// Convert port from string to uint, similar to the reference code
	portNum, err := strconv.ParseUint(port, 10, 32)
	if err != nil {
		log.Fatalf("Failed to parse database port: %v", err)
	}

	// Build the connection string
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, portNum, user, password, dbname,
	)

	// Connect to the database with custom configuration
	DB, err = gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // Bypass prepared statement caching issues
	}), &gorm.Config{
		// Disable prepared statement caching in GORM
		PrepareStmt: false,

		// Custom naming strategy (optional)
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "",    // Custom table prefix (if required)
			SingularTable: false, // Disable singular table names
		},

		// Logger settings
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	// Setting up connection pooling
	sqlDB, err := DB.DB()  // Use the same 'err' variable
	if err != nil {
		log.Fatalf("Error setting up database connection pool: %v", err)
	}

	// Pool settings: Adjust these values based on your app's expected workload
	sqlDB.SetMaxIdleConns(10)               // Allow up to 10 idle connections
	sqlDB.SetMaxOpenConns(50)               // Allow up to 50 open connections
	sqlDB.SetConnMaxLifetime(15 * time.Minute) // Limit connection reuse to 15 minutes
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)  // Limit idle connection time to 5 minutes

	// Debugging: Confirm database connection
	fmt.Println("Database successfully connected!")
}
