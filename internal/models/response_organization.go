package models

import "time"

type ResponseOrganization struct {
	ID                 string    `json:"id"`
	Name               string    `json:"name"`
	OrganizationType   string    `json:"organization_type"`
	Latitude           float64   `json:"latitude"`
	Longitude          float64   `json:"longitude"`
	Address            *string   `json:"address,omitempty"`
	Phone              *string   `json:"phone,omitempty"`
	Email              *string   `json:"email,omitempty"`
	OnboardingSource   string    `json:"onboarding_source"`
	VerificationStatus string    `json:"verification_status"`
	IsActive           bool      `json:"is_active"`
	Capabilities       []string  `json:"capabilities"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
