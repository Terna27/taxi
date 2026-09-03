package models

import "time"

type EmergencyAssignment struct {
	ID                         string     `json:"id"`
	IncidentID                 string     `json:"incident_id"`
	OrganizationID             string     `json:"organization_id"`
	Capability                 string     `json:"capability"`
	Status                     string     `json:"status"`
	StraightLineDistanceMeters *float64   `json:"straight_line_distance_meters,omitempty"`
	RouteDistanceMeters        *float64   `json:"route_distance_meters,omitempty"`
	EstimatedTravelSeconds     *int       `json:"estimated_travel_seconds,omitempty"`
	OfferedAt                  *time.Time `json:"offered_at,omitempty"`
	AcceptedAt                 *time.Time `json:"accepted_at,omitempty"`
	DeclinedAt                 *time.Time `json:"declined_at,omitempty"`
	EnRouteAt                  *time.Time `json:"en_route_at,omitempty"`
	ArrivedAt                  *time.Time `json:"arrived_at,omitempty"`
	CompletedAt                *time.Time `json:"completed_at,omitempty"`
	CancelledAt                *time.Time `json:"cancelled_at,omitempty"`
	ExpiresAt                  *time.Time `json:"expires_at,omitempty"`
	CreatedAt                  time.Time  `json:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at"`
}