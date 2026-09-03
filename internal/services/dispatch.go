package services

import (
	"context"
	"errors"

	"taxi/internal/models"
	"taxi/internal/repositories"
)

var (
	ErrNoPrimaryCapability          = errors.New("incident has no primary response capability")
	ErrNoDispatchCandidate          = errors.New("no eligible dispatch candidate found")
	ErrMultiplePrimaryCapabilities  = errors.New("incident has multiple primary response capabilities")
	ErrDispatchAlreadyActive        = errors.New("an active assignment already exists for this incident capability")
	ErrDispatchCandidateUnavailable = errors.New("selected dispatch candidate is no longer eligible")
)

type DispatchCandidateDiscoverer interface {
	Discover(
		ctx context.Context,
		incidentID string,
	) (*models.IncidentCandidateDiscovery, error)
}

type DispatchAssignmentRepository interface {
	Create(
		ctx context.Context,
		incidentID string,
		organizationID string,
		capability string,
		straightLineDistanceMeters *float64,
	) (*models.EmergencyAssignment, error)
}

type DispatchService struct {
	candidateDiscovery   DispatchCandidateDiscoverer
	assignmentRepository DispatchAssignmentRepository
}

func NewDispatchService(
	candidateDiscovery DispatchCandidateDiscoverer,
	assignmentRepository DispatchAssignmentRepository,
) *DispatchService {
	return &DispatchService{
		candidateDiscovery:   candidateDiscovery,
		assignmentRepository: assignmentRepository,
	}
}

func (s *DispatchService) Dispatch(
	ctx context.Context,
	incidentID string,
) (*models.EmergencyAssignment, error) {
	discovery, err := s.candidateDiscovery.Discover(
		ctx,
		incidentID,
	)
	if err != nil {
		return nil, err
	}

	var primaryGroup *models.IncidentCandidateGroup

	for i := range discovery.CandidateGroups {
		group := &discovery.CandidateGroups[i]

		if group.RequirementLevel != "PRIMARY" {
			continue
		}

		if primaryGroup != nil {
			return nil, ErrMultiplePrimaryCapabilities
		}

		primaryGroup = group
	}

	if primaryGroup == nil {
		return nil, ErrNoPrimaryCapability
	}

	if len(primaryGroup.Organizations) == 0 {
		return nil, ErrNoDispatchCandidate
	}

	candidate := primaryGroup.Organizations[0]

	distance := candidate.DistanceMeters

	assignment, err :=
		s.assignmentRepository.Create(
			ctx,
			discovery.IncidentID,
			candidate.ID,
			primaryGroup.Capability,
			&distance,
		)

	if err != nil {
		switch {
		case errors.Is(
			err,
			repositories.ErrActiveAssignmentExists,
		):
			return nil, ErrDispatchAlreadyActive

		case errors.Is(
			err,
			repositories.ErrAssignmentOrganizationIneligible,
		):
			return nil, ErrDispatchCandidateUnavailable

		default:
			return nil, err
		}
	}

	return assignment, nil
}
