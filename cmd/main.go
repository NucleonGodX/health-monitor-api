package main

import (
    "log"
    "net/http"

    "github.com/gorilla/mux"
    
    "health-tracking-api/database"
    "health-tracking-api/handlers"
)

func main() {
    // Initialize database connection
    database.InitDB()
    defer database.CloseDB()

    // Create router
    r := mux.NewRouter()

    // Define routes
    r.HandleFunc("/api/data", handlers.ReceiveData).Methods("POST")
    r.HandleFunc("/api/devices/register", handlers.RegisterDevice).Methods("POST")
    r.HandleFunc("/api/data/last", handlers.GetLastRecord).Methods("GET")
    r.HandleFunc("/api/data/recent", handlers.GetRecentRecords).Methods("GET")
    r.HandleFunc("/api/data/filter", handlers.GetFilteredRecords).Methods("GET")

    // Start server
    log.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}