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
	ErrIncidentIDRequired        = errors.New("incident id is required")
	ErrIncidentIDInvalid         = errors.New("incident id must be a valid UUID")
	ErrIncidentMissing           = errors.New("incident not found")
	ErrIncidentCapabilityMissing = errors.New("incident emergency type has no configured response capabilities")
)

const (
	defaultIncidentCandidateRadiusMeters = 25000.0
	defaultIncidentCandidateLimit        = 10
)

type IncidentCandidateContextRepository interface {
	GetIncidentDiscoveryContext(
		ctx context.Context,
		incidentID string,
	) (
		emergencyType string,
		latitude float64,
		longitude float64,
		err error,
	)

	GetCapabilityRequirements(
		ctx context.Context,
		emergencyType string,
	) ([]models.IncidentCapabilityRequirement, error)
}

type IncidentCandidateDiscoveryService struct {
	contextRepository   IncidentCandidateContextRepository
	discoveryRepository OrganizationDiscoveryRepository
}

func NewIncidentCandidateDiscoveryService(
	contextRepository IncidentCandidateContextRepository,
	discoveryRepository OrganizationDiscoveryRepository,
) *IncidentCandidateDiscoveryService {
	return &IncidentCandidateDiscoveryService{
		contextRepository:   contextRepository,
		discoveryRepository: discoveryRepository,
	}
}

func (s *IncidentCandidateDiscoveryService) Discover(
	ctx context.Context,
	incidentID string,
) (*models.IncidentCandidateDiscovery, error) {
	incidentID = strings.TrimSpace(incidentID)

	if incidentID == "" {
		return nil, ErrIncidentIDRequired
	}

	if !isValidIncidentUUID(incidentID) {
		return nil, ErrIncidentIDInvalid
	}

	emergencyType, latitude, longitude, err :=
		s.contextRepository.GetIncidentDiscoveryContext(
			ctx,
			incidentID,
		)

	if err != nil {
		if errors.Is(
			err,
			repositories.ErrIncidentNotFound,
		) {
			return nil, ErrIncidentMissing
		}

		return nil, err
	}

	requirements, err :=
		s.contextRepository.GetCapabilityRequirements(
			ctx,
			emergencyType,
		)

	if err != nil {
		return nil, err
	}

	if len(requirements) == 0 {
		return nil, ErrIncidentCapabilityMissing
	}

	result := &models.IncidentCandidateDiscovery{
		IncidentID:      incidentID,
		EmergencyType:   emergencyType,
		Latitude:        latitude,
		Longitude:       longitude,
		CandidateGroups: make([]models.IncidentCandidateGroup, 0, len(requirements)),
	}

	for _, requirement := range requirements {
		organizations, err :=
			s.discoveryRepository.FindNearby(
				ctx,
				latitude,
				longitude,
				requirement.Capability,
				defaultIncidentCandidateRadiusMeters,
				defaultIncidentCandidateLimit,
			)

		if err != nil {
			return nil, err
		}

		result.CandidateGroups = append(
			result.CandidateGroups,
			models.IncidentCandidateGroup{
				Capability:       requirement.Capability,
				RequirementLevel: requirement.RequirementLevel,
				Organizations:    organizations,
			},
		)
	}

	return result, nil
}

func isValidIncidentUUID(value string) bool {
	var id pgtype.UUID

	return id.Scan(value) == nil && id.Valid
}
