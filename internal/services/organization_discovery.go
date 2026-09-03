package services

import (
	"context"
	"errors"
	"strings"

	"taxi/internal/models"
)

var (
	ErrDiscoveryLatitudeRequired   = errors.New("latitude is required")
	ErrDiscoveryLongitudeRequired  = errors.New("longitude is required")
	ErrDiscoveryLatitudeInvalid    = errors.New("latitude must be between -90 and 90")
	ErrDiscoveryLongitudeInvalid   = errors.New("longitude must be between -180 and 180")
	ErrDiscoveryCapabilityRequired = errors.New("capability is required")
	ErrDiscoveryRadiusInvalid      = errors.New("radius_meters must be greater than 0")
	ErrDiscoveryRadiusTooLarge     = errors.New("radius_meters must not exceed 100000")
	ErrDiscoveryLimitInvalid       = errors.New("limit must be between 1 and 50")
)

const (
	defaultDiscoveryRadiusMeters = 10000.0
	maxDiscoveryRadiusMeters     = 100000.0

	defaultDiscoveryLimit = 10
	maxDiscoveryLimit     = 50
)

type OrganizationDiscoveryRepository interface {
	FindNearby(
		ctx context.Context,
		latitude float64,
		longitude float64,
		capability string,
		radiusMeters float64,
		limit int,
	) ([]models.NearbyResponseOrganization, error)
}

type FindNearbyOrganizationsInput struct {
	Latitude     *float64
	Longitude    *float64
	Capability   string
	RadiusMeters *float64
	Limit        *int
}

type OrganizationDiscoveryService struct {
	repository OrganizationDiscoveryRepository
}

func NewOrganizationDiscoveryService(
	repository OrganizationDiscoveryRepository,
) *OrganizationDiscoveryService {
	return &OrganizationDiscoveryService{
		repository: repository,
	}
}

func (s *OrganizationDiscoveryService) FindNearby(
	ctx context.Context,
	input FindNearbyOrganizationsInput,
) ([]models.NearbyResponseOrganization, error) {
	if input.Latitude == nil {
		return nil, ErrDiscoveryLatitudeRequired
	}

	if input.Longitude == nil {
		return nil, ErrDiscoveryLongitudeRequired
	}

	if *input.Latitude < -90 || *input.Latitude > 90 {
		return nil, ErrDiscoveryLatitudeInvalid
	}

	if *input.Longitude < -180 || *input.Longitude > 180 {
		return nil, ErrDiscoveryLongitudeInvalid
	}

	capability := strings.ToUpper(
		strings.TrimSpace(input.Capability),
	)

	if capability == "" {
		return nil, ErrDiscoveryCapabilityRequired
	}

	radiusMeters := defaultDiscoveryRadiusMeters

	if input.RadiusMeters != nil {
		radiusMeters = *input.RadiusMeters
	}

	if radiusMeters <= 0 {
		return nil, ErrDiscoveryRadiusInvalid
	}

	if radiusMeters > maxDiscoveryRadiusMeters {
		return nil, ErrDiscoveryRadiusTooLarge
	}

	limit := defaultDiscoveryLimit

	if input.Limit != nil {
		limit = *input.Limit
	}

	if limit < 1 || limit > maxDiscoveryLimit {
		return nil, ErrDiscoveryLimitInvalid
	}

	return s.repository.FindNearby(
		ctx,
		*input.Latitude,
		*input.Longitude,
		capability,
		radiusMeters,
		limit,
	)
}