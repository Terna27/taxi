package models

import "time"

type ResponseOrganizationStatus struct {
	ID                 string    `json:"id"`
	VerificationStatus string    `json:"verification_status"`
	IsActive           bool      `json:"is_active"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type NearbyResponseOrganization struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	OrganizationType string   `json:"organization_type"`
	Latitude         float64  `json:"latitude"`
	Longitude        float64  `json:"longitude"`
	Address          *string  `json:"address,omitempty"`
	Phone            *string  `json:"phone,omitempty"`
	DistanceMeters   float64  `json:"distance_meters"`
	Capabilities     []string `json:"capabilities"`
}