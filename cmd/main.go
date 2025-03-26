package main

import (
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"
    
    "github.com/NucleonGodX/health-monitor-api/internal/database"
    "github.com/NucleonGodX/health-monitor-api/pkg/handlers"
)

func main() {

    err := database.InitDB()
    if err != nil {
        log.Fatalf("Database initialization failed: %v", err)
    }
    defer database.CloseDB()
    r := mux.NewRouter()

    r.HandleFunc("/api/devices", handlers.CreateDevice).Methods("POST")
    r.HandleFunc("/api/devices", handlers.GetDeviceByID).Methods("GET")
    r.HandleFunc("/api/devices/list", handlers.ListDevices).Methods("GET")

    r.HandleFunc("/api/health", handlers.AddHealthRecord).Methods("POST")
    r.HandleFunc("/api/health/latest", handlers.GetLatestHealthRecord).Methods("GET")
    r.HandleFunc("/api/health/records", handlers.GetHealthRecords).Methods("GET")

    port := os.Getenv("SERVER_PORT")
    if port == "" {
        port = "8080"
    }
    serverAddr := fmt.Sprintf(":%s", port)
    log.Printf("Server starting on %s", serverAddr)
    log.Fatal(http.ListenAndServe(serverAddr, r))
}