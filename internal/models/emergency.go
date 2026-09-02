package models

import "time"

type EmergencyIncident struct {
	ID                     string    `json:"id"`
	EmergencyType          string    `json:"emergency_type"`
	Status                 string    `json:"status"`
	Latitude               float64   `json:"latitude"`
	Longitude              float64   `json:"longitude"`
	LocationAccuracyMeters *float64  `json:"location_accuracy_meters,omitempty"`
	Description            *string   `json:"description,omitempty"`
	IdempotencyKey         string    `json:"idempotency_key"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}
