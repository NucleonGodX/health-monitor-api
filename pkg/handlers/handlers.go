package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "time"

    "health-tracking-api/database"
    "health-tracking-api/models"
)

func ReceiveData(w http.ResponseWriter, r *http.Request) {
    var record models.HealthRecord
    err := json.NewDecoder(r.Body).Decode(&record)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    _, err = database.DB.Exec(
        "INSERT INTO health_records (device_id, spo2, pulse) VALUES ($1, $2, $3)",
        record.DeviceID, record.SPO2, record.Pulse,
    )
    if err != nil {
        http.Error(w, "Failed to store data", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
}

func RegisterDevice(w http.ResponseWriter, r *http.Request) {
    var device models.Device
    err := json.NewDecoder(r.Body).Decode(&device)
    if err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    _, err = database.DB.Exec(
        "INSERT INTO devices (device_id, account_id) VALUES ($1, $2) ON CONFLICT (device_id) DO NOTHING",
        device.DeviceID, device.AccountID,
    )
    if err != nil {
        http.Error(w, "Failed to register device", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}

func GetLastRecord(w http.ResponseWriter, r *http.Request) {
    deviceID := r.URL.Query().Get("deviceId")
    if deviceID == "" {
        http.Error(w, "Device ID is required", http.StatusBadRequest)
        return
    }

    var record models.HealthRecord
    err := database.DB.QueryRow(
        "SELECT spo2, pulse FROM health_records WHERE device_id = $1 ORDER BY created_at DESC LIMIT 1",
        deviceID,
    ).Scan(&record.SPO2, &record.Pulse)

    if err != nil {
        http.Error(w, "No records found", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(record)
}

func GetRecentRecords(w http.ResponseWriter, r *http.Request) {
    deviceID := r.URL.Query().Get("deviceId")
    if deviceID == "" {
        http.Error(w, "Device ID is required", http.StatusBadRequest)
        return
    }

    rows, err := database.DB.Query(
        "SELECT spo2, pulse FROM health_records WHERE device_id = $1 ORDER BY created_at DESC LIMIT 5",
        deviceID,
    )
    if err != nil {
        http.Error(w, "Failed to retrieve records", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var records []models.HealthRecord
    for rows.Next() {
        var record models.HealthRecord
        if err := rows.Scan(&record.SPO2, &record.Pulse); err != nil {
            http.Error(w, "Error scanning records", http.StatusInternalServerError)
            return
        }
        records = append(records, record)
    }

    json.NewEncoder(w).Encode(records)
}

func GetFilteredRecords(w http.ResponseWriter, r *http.Request) {
    deviceID := r.URL.Query().Get("deviceId")
    filter := r.URL.Query().Get("filter")
    if deviceID == "" || filter == "" {
        http.Error(w, "Device ID and filter are required", http.StatusBadRequest)
        return
    }

    var query string
    var days int
    switch filter {
    case "7d": days = 7
    case "30d": days = 30
    case "6m": days = 180
    case "1y": days = 365
    case "all": days = 0
    default:
        http.Error(w, "Invalid filter", http.StatusBadRequest)
        return
    }

    var result models.FilteredRecords
    if days > 0 {
        query = `
            SELECT spo2, pulse FROM health_records 
            WHERE device_id = $1 AND created_at >= NOW() - INTERVAL '1 days' * $2
            ORDER BY created_at DESC
        `
        rows, err := database.DB.Query(query, deviceID, days)
        if err != nil {
            http.Error(w, "Failed to retrieve records", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var totalSPO2, totalPulse float64
        var count int
        for rows.Next() {
            var record models.HealthRecord
            if err := rows.Scan(&record.SPO2, &record.Pulse); err != nil {
                http.Error(w, "Error scanning records", http.StatusInternalServerError)
                return
            }
            result.Records = append(result.Records, record)
            totalSPO2 += float64(record.SPO2)
            totalPulse += float64(record.Pulse)
            count++
        }

        if count > 0 {
            result.Average.SPO2 = totalSPO2 / float64(count)
            result.Average.Pulse = totalPulse / float64(count)
        }
    } else {
        // For 'all' filter, retrieve all records
        query = "SELECT spo2, pulse FROM health_records WHERE device_id = $1 ORDER BY created_at DESC"
        rows, err := database.DB.Query(query, deviceID)
        if err != nil {
            http.Error(w, "Failed to retrieve records", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        var totalSPO2, totalPulse float64
        var count int
        for rows.Next() {
            var record models.HealthRecord
            if err := rows.Scan(&record.SPO2, &record.Pulse); err != nil {
                http.Error(w, "Error scanning records", http.StatusInternalServerError)
                return
            }
            result.Records = append(result.Records, record)
            totalSPO2 += float64(record.SPO2)
            totalPulse += float64(record.Pulse)
            count++
        }

        if count > 0 {
            result.Average.SPO2 = totalSPO2 / float64(count)
            result.Average.Pulse = totalPulse / float64(count)
        }
    }

    json.NewEncoder(w).Encode(result)
}