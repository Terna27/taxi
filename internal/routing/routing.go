package routing

import (
	"context"

	"taxi/internal/models"
)

type RouteEstimate struct {
	OrganizationID    string
	DistanceMeters    int
	TravelTimeSeconds int
}

type Service interface {
	EstimateRoutes(
		ctx context.Context,
		originLatitude float64,
		originLongitude float64,
		organizations []models.NearbyResponseOrganization,
	) ([]RouteEstimate, error)
}
