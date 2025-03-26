package models

type Device struct {
    DeviceID   string `json:"deviceId"`
    AccountID  string `json:"accountId,omitempty"`
}

type HealthRecord struct {
    DeviceID string `json:"deviceId"`
    SPO2     int    `json:"spo2"`
    Pulse    int    `json:"pulse"`
}

type FilteredRecords struct {
    Records []HealthRecord `json:"records"`
    Average struct {
        SPO2  float64 `json:"spo2"`
        Pulse float64 `json:"pulse"`
    } `json:"average"`
}