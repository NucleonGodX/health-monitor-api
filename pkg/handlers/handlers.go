package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/NucleonGodX/health-monitor-api/internal/database"
    "github.com/NucleonGodX/health-monitor-api/pkg/models"
)

// CreateDevice handles device registration
func CreateDevice(w http.ResponseWriter, r *http.Request) {
    var deviceReq struct {
        DeviceID    string `json:"deviceId"`
        AccountID   string `json:"accountId"`
    }

    if err := json.NewDecoder(r.Body).Decode(&deviceReq); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    _, err := database.DB.Exec(
        `INSERT INTO devices (device_id, account_id) 
         VALUES ($1, $2) 
         ON CONFLICT (device_id) DO UPDATE 
         SET account_id = EXCLUDED.account_id`,
        deviceReq.DeviceID, deviceReq.AccountID,
    )
    if err != nil {
        http.Error(w, "Failed to register device", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}

// GetDeviceByID retrieves device details
func GetDeviceByID(w http.ResponseWriter, r *http.Request) {
    deviceID := r.URL.Query().Get("deviceId")
    if deviceID == "" {
        http.Error(w, "Device ID is required", http.StatusBadRequest)
        return
    }

    var device models.Device
    err := database.DB.QueryRow(
        `SELECT id, device_id, account_id, created_at, last_active 
         FROM devices 
         WHERE device_id = $1`,
        deviceID,
    ).Scan(&device.ID, &device.DeviceID, &device.AccountID, &device.CreatedAt, &device.LastActive)

    if err != nil {
        http.Error(w, "Device not found", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(device)
}

// ListDevices retrieves all devices for an account
func ListDevices(w http.ResponseWriter, r *http.Request) {
    accountID := r.URL.Query().Get("accountId")
    if accountID == "" {
        http.Error(w, "Account ID is required", http.StatusBadRequest)
        return
    }

    rows, err := database.DB.Query(
        `SELECT id, device_id, account_id, created_at, last_active 
         FROM devices 
         WHERE account_id = $1 
         ORDER BY last_active DESC`,
        accountID,
    )
    if err != nil {
        http.Error(w, "Failed to retrieve devices", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var devices []models.Device
    for rows.Next() {
        var device models.Device
        if err := rows.Scan(&device.ID, &device.DeviceID, &device.AccountID, &device.CreatedAt, &device.LastActive); err != nil {
            http.Error(w, "Error scanning devices", http.StatusInternalServerError)
            return
        }
        devices = append(devices, device)
    }

    json.NewEncoder(w).Encode(devices)
}

// AddHealthRecord handles adding a new health record
func AddHealthRecord(w http.ResponseWriter, r *http.Request) {
    var record models.HealthRecord
    if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate record data
    if record.DeviceID == "" || record.SPO2 < 0 || record.SPO2 > 100 || 
       record.Pulse < 0 || record.Pulse > 300 || 
       record.Temperature < 95.0 || record.Temperature > 106.0 {
        http.Error(w, "Invalid health record data", http.StatusBadRequest)
        return
    }

    _, err := database.DB.Exec(
        `INSERT INTO health_records (device_id, spo2, pulse, temperature) 
         VALUES ($1, $2, $3, $4)`,
        record.DeviceID, record.SPO2, record.Pulse, record.Temperature,
    )
    if err != nil {
        http.Error(w, "Failed to store health record", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}

// GetLatestHealthRecord retrieves the most recent health record for a device
func GetLatestHealthRecord(w http.ResponseWriter, r *http.Request) {
    deviceID := r.URL.Query().Get("deviceId")
    if deviceID == "" {
        http.Error(w, "Device ID is required", http.StatusBadRequest)
        return
    }

    var record models.HealthRecord
    err := database.DB.QueryRow(
        `SELECT device_id, spo2, pulse, temperature, recorded_at 
         FROM health_records 
         WHERE device_id = $1 
         ORDER BY recorded_at DESC 
         LIMIT 1`,
        deviceID,
    ).Scan(&record.DeviceID, &record.SPO2, &record.Pulse, &record.Temperature, &record.RecordedAt)

    if err != nil {
        http.Error(w, "No records found", http.StatusNotFound)
        return
    }

    json.NewEncoder(w).Encode(record)
}

// GetHealthRecords retrieves health records with filtering options
func GetHealthRecords(w http.ResponseWriter, r *http.Request) {
    deviceID := r.URL.Query().Get("deviceId")
    filter := r.URL.Query().Get("filter")
    if deviceID == "" {
        http.Error(w, "Device ID is required", http.StatusBadRequest)
        return
    }

    var query string
    var args []interface{}
    args = append(args, deviceID)

    switch filter {
    case "7d":
        query = `
            SELECT device_id, spo2, pulse, temperature, recorded_at 
            FROM health_records 
            WHERE device_id = $1 AND recorded_at >= NOW() - INTERVAL '7 days' 
            ORDER BY recorded_at DESC
        `
    case "30d":
        query = `
            SELECT device_id, spo2, pulse, temperature, recorded_at 
            FROM health_records 
            WHERE device_id = $1 AND recorded_at >= NOW() - INTERVAL '30 days' 
            ORDER BY recorded_at DESC
        `
    case "6m":
        query = `
            SELECT device_id, spo2, pulse, temperature, recorded_at 
            FROM health_records 
            WHERE device_id = $1 AND recorded_at >= NOW() - INTERVAL '6 months' 
            ORDER BY recorded_at DESC
        `
    case "1y":
        query = `
            SELECT device_id, spo2, pulse, temperature, recorded_at 
            FROM health_records 
            WHERE device_id = $1 AND recorded_at >= NOW() - INTERVAL '1 year' 
            ORDER BY recorded_at DESC
        `
    case "all":
        query = `
            SELECT device_id, spo2, pulse, temperature, recorded_at 
            FROM health_records 
            WHERE device_id = $1 
            ORDER BY recorded_at DESC
        `
    default:
        http.Error(w, "Invalid filter", http.StatusBadRequest)
        return
    }

    rows, err := database.DB.Query(query, args...)
    if err != nil {
        http.Error(w, "Failed to retrieve records", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var records []models.HealthRecord
    var totalSPO2, totalPulse, totalTemp float64
    var count int

    for rows.Next() {
        var record models.HealthRecord
        if err := rows.Scan(&record.DeviceID, &record.SPO2, &record.Pulse, &record.Temperature, &record.RecordedAt); err != nil {
            http.Error(w, "Error scanning records", http.StatusInternalServerError)
            return
        }
        records = append(records, record)
        
        totalSPO2 += float64(record.SPO2)
        totalPulse += float64(record.Pulse)
        totalTemp += record.Temperature
        count++
    }

    var result struct {
        Records []models.HealthRecord `json:"records"`
        Average struct {
            SPO2        float64 `json:"spo2"`
            Pulse       float64 `json:"pulse"`
            Temperature float64 `json:"temperature"`
        } `json:"average"`
    }

    result.Records = records
    if count > 0 {
        result.Average.SPO2 = totalSPO2 / float64(count)
        result.Average.Pulse = totalPulse / float64(count)
        result.Average.Temperature = totalTemp / float64(count)
    }

    json.NewEncoder(w).Encode(result)
}