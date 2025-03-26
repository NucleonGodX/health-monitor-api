package database

import (
    "database/sql"
    "log"
    "os"
    "fmt"
    _ "github.com/lib/pq"
    "github.com/joho/godotenv"
)

var DB *sql.DB

func InitDB() error {
    // Get absolute path to the root directory
    rootDir, err := os.Getwd()
    if err != nil {
        log.Fatal("Failed to get working directory:", err)
    }
    envPath := rootDir + "/configs/.env"
    err = godotenv.Load(envPath)
    if err != nil {
        log.Println("Error loading .env file from", envPath, ", using system environment")
    }

    connectionString := os.Getenv("DB_CONNECTION_STRING")
    if connectionString == "" {
        log.Fatal("DB_CONNECTION_STRING is not set")
    }

    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return fmt.Errorf("error opening database: %v", err)
    }

    err = db.Ping()
    if err != nil {
        return fmt.Errorf("error connecting to the database: %v", err)
    }

    DB = db
    log.Println("Successfully connected to the database")

    return nil
}


func CloseDB() {
    if DB != nil {
        DB.Close()
    }
}