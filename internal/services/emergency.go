package services

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"taxi/internal/models"
	"taxi/internal/repositories"
)

var (
	ErrEmergencyTypeRequired   = errors.New("emergency_type is required")
	ErrEmergencyTypeNotFound   = errors.New("emergency type not found or inactive")
	ErrLatitudeRequired        = errors.New("latitude is required")
	ErrLongitudeRequired       = errors.New("longitude is required")
	ErrInvalidLatitude         = errors.New("latitude must be between -90 and 90")
	ErrInvalidLongitude        = errors.New("longitude must be between -180 and 180")
	ErrInvalidLocationAccuracy = errors.New("location_accuracy_meters cannot be negative")
	ErrDescriptionTooLong      = errors.New("description cannot exceed 2000 characters")
	ErrIdempotencyKeyRequired  = errors.New("idempotency_key is required")
	ErrInvalidIdempotencyKey   = errors.New("idempotency_key must be a valid UUID")
	ErrDuplicateIncident       = errors.New("incident with this idempotency key already exists")
)

type EmergencyRepository interface {
	Create(
		ctx context.Context,
		emergencyType string,
		latitude float64,
		longitude float64,
		locationAccuracyMeters *float64,
		description *string,
		idempotencyKey string,
	) (*models.EmergencyIncident, error)
}

type CreateEmergencyInput struct {
	EmergencyType          string
	Latitude               *float64
	Longitude              *float64
	LocationAccuracyMeters *float64
	Description            *string
	IdempotencyKey         string
}

type EmergencyService struct {
	repository EmergencyRepository
}

func NewEmergencyService(
	repository EmergencyRepository,
) *EmergencyService {
	return &EmergencyService{
		repository: repository,
	}
}

func (s *EmergencyService) Create(
	ctx context.Context,
	input CreateEmergencyInput,
) (*models.EmergencyIncident, error) {
	input.EmergencyType = strings.ToUpper(
		strings.TrimSpace(input.EmergencyType),
	)

	if input.EmergencyType == "" {
		return nil, ErrEmergencyTypeRequired
	}

	if input.Latitude == nil {
		return nil, ErrLatitudeRequired
	}

	if input.Longitude == nil {
		return nil, ErrLongitudeRequired
	}

	if *input.Latitude < -90 || *input.Latitude > 90 {
		return nil, ErrInvalidLatitude
	}

	if *input.Longitude < -180 || *input.Longitude > 180 {
		return nil, ErrInvalidLongitude
	}

	if input.LocationAccuracyMeters != nil &&
		*input.LocationAccuracyMeters < 0 {
		return nil, ErrInvalidLocationAccuracy
	}

	if input.Description != nil {
		trimmedDescription := strings.TrimSpace(*input.Description)

		if len(trimmedDescription) > 2000 {
			return nil, ErrDescriptionTooLong
		}

		if trimmedDescription == "" {
			input.Description = nil
		} else {
			input.Description = &trimmedDescription
		}
	}

	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)

	if input.IdempotencyKey == "" {
		return nil, ErrIdempotencyKeyRequired
	}

	var uuid pgtype.UUID
	if err := uuid.Scan(input.IdempotencyKey); err != nil || !uuid.Valid {
		return nil, ErrInvalidIdempotencyKey
	}

	incident, err := s.repository.Create(
		ctx,
		input.EmergencyType,
		*input.Latitude,
		*input.Longitude,
		input.LocationAccuracyMeters,
		input.Description,
		input.IdempotencyKey,
	)

	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrEmergencyTypeNotFound):
			return nil, ErrEmergencyTypeNotFound

		case errors.Is(err, repositories.ErrDuplicateIdempotencyKey):
			return nil, ErrDuplicateIncident

		default:
			return nil, err
		}
	}

	return incident, nil
}
