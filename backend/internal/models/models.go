package models

import "time"

type Device struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Type     string     `json:"type"`
	Zone     string     `json:"zone"`
	Status   string     `json:"status"` // "active" | "inactive"
	LastSeen *time.Time `json:"last_seen"`
}

type SensorReading struct {
	ID            int64     `json:"id"`
	DeviceID      string    `json:"device_id"`
	SoilMoisture  float64   `json:"soil_moisture"`
	Temperature   float64   `json:"temperature"`
	Humidity      float64   `json:"humidity"`
	CreatedAt     time.Time `json:"created_at"`
}

// SensorPayload is the raw MQTT/ingest payload shape from ESP32.
type SensorPayload struct {
	DeviceID     string    `json:"device_id"`
	SoilMoisture float64   `json:"soil_moisture"`
	Temperature  float64   `json:"temperature"`
	Humidity     float64   `json:"humidity"`
	Timestamp    time.Time `json:"timestamp"`
}

type LeafImage struct {
	ID         int64     `json:"id"`
	DeviceID   string    `json:"device_id"`
	ImageURL   string    `json:"image_url"`
	CVResult   string    `json:"cv_result"`
	Confidence float64   `json:"confidence"`
	CreatedAt  time.Time `json:"created_at"`
}

type AgentDecision struct {
	ID               int64      `json:"id"`
	DeviceID         string     `json:"device_id"`
	Action           string     `json:"action"` // "water_on" | "water_off" | "no_action" | "alert"
	Reason           string     `json:"reason"`
	Confidence       float64    `json:"confidence"`
	NeedHumanConfirm bool       `json:"need_human_confirm"`
	ConfirmedBy      *string    `json:"confirmed_by"`
	ConfirmedAt      *time.Time `json:"confirmed_at"`
	Status           string     `json:"status"`         // "pending" | "confirmed" | "rejected" | "auto_executed"
	TriggerSource    string     `json:"trigger_source"` // "threshold" | "delta" | "cv" | "schedule" | "manual"
	CreatedAt        time.Time  `json:"created_at"`
}

// LLMDecision is the parsed JSON response expected from the LLM agent call.
type LLMDecision struct {
	Action           string  `json:"action"`
	Reason           string  `json:"reason"`
	NeedHumanConfirm bool    `json:"need_human_confirm"`
	Confidence       float64 `json:"confidence"`
}
