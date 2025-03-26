package database

import (
    "database/sql"
    "fmt"
    "log"
    "os"

    _ "github.com/lib/pq"
    "github.com/joho/godotenv"
)

var DB *sql.DB

func InitDB() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    connStr := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=require",
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_HOST"),
        os.Getenv("DB_NAME"),
    )

    db, err := sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal(err)
    }

    err = db.Ping()
    if err != nil {
        log.Fatal(err)
    }

    DB = db
    log.Println("Connected to PostgreSQL database")
}

func CloseDB() {
    if DB != nil {
        DB.Close()
    }
}