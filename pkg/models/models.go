package models

import "time"

type Device struct {
    ID          string    `json:"id"`
    DeviceID    string    `json:"deviceId"`
    AccountID   string    `json:"accountId"`
    CreatedAt   time.Time `json:"createdAt"`
    LastActive  time.Time `json:"lastActive"`
}

type HealthRecord struct {
    DeviceID     string    `json:"deviceId"`
    SPO2         int       `json:"spo2"`
    Pulse        int       `json:"pulse"`
    Temperature  float64   `json:"temperature"`
    RecordedAt   time.Time `json:"recordedAt"`
}

type FilteredRecords struct {
    Records []HealthRecord `json:"records"`
    Average struct {
        SPO2        float64 `json:"spo2"`
        Pulse       float64 `json:"pulse"`
        Temperature float64 `json:"temperature"`
    } `json:"average"`
}