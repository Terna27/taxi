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
	ErrOrganizationIDRequired      = errors.New("organization id is required")
	ErrOrganizationIDInvalid       = errors.New("organization id must be a valid UUID")
	ErrVerificationStatusInvalid   = errors.New("verification status must be VERIFIED or REJECTED")
	ErrResponseOrganizationMissing = errors.New("response organization not found")
	ErrOrganizationMustBeVerified  = errors.New("response organization must be verified before activation")
)

type OrganizationReviewRepository interface {
	UpdateVerificationStatus(
		ctx context.Context,
		organizationID string,
		status string,
	) (*models.ResponseOrganizationStatus, error)

	UpdateActiveStatus(
		ctx context.Context,
		organizationID string,
		isActive bool,
	) (*models.ResponseOrganizationStatus, error)
}

type OrganizationReviewService struct {
	repository OrganizationReviewRepository
}

func NewOrganizationReviewService(
	repository OrganizationReviewRepository,
) *OrganizationReviewService {
	return &OrganizationReviewService{
		repository: repository,
	}
}

func (s *OrganizationReviewService) Review(
	ctx context.Context,
	organizationID string,
	status string,
) (*models.ResponseOrganizationStatus, error) {
	organizationID = strings.TrimSpace(organizationID)

	if organizationID == "" {
		return nil, ErrOrganizationIDRequired
	}

	if !isValidUUID(organizationID) {
		return nil, ErrOrganizationIDInvalid
	}

	status = strings.ToUpper(strings.TrimSpace(status))

	switch status {
	case "VERIFIED", "REJECTED":
	default:
		return nil, ErrVerificationStatusInvalid
	}

	organizationStatus, err := s.repository.UpdateVerificationStatus(
		ctx,
		organizationID,
		status,
	)

	if err != nil {
		if errors.Is(
			err,
			repositories.ErrResponseOrganizationNotFound,
		) {
			return nil, ErrResponseOrganizationMissing
		}

		return nil, err
	}

	return organizationStatus, nil
}

func (s *OrganizationReviewService) SetActive(
	ctx context.Context,
	organizationID string,
	isActive bool,
) (*models.ResponseOrganizationStatus, error) {
	organizationID = strings.TrimSpace(organizationID)

	if organizationID == "" {
		return nil, ErrOrganizationIDRequired
	}

	if !isValidUUID(organizationID) {
		return nil, ErrOrganizationIDInvalid
	}

	organizationStatus, err := s.repository.UpdateActiveStatus(
		ctx,
		organizationID,
		isActive,
	)

	if err != nil {
		switch {
		case errors.Is(
			err,
			repositories.ErrResponseOrganizationNotFound,
		):
			return nil, ErrResponseOrganizationMissing

		case errors.Is(
			err,
			repositories.ErrOrganizationNotVerified,
		):
			return nil, ErrOrganizationMustBeVerified

		default:
			return nil, err
		}
	}

	return organizationStatus, nil
}

func isValidUUID(value string) bool {
	var id pgtype.UUID

	return id.Scan(value) == nil && id.Valid
}
