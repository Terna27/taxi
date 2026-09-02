package services

import (
	"context"
	"errors"
	"strings"

	"taxi/internal/models"
	"taxi/internal/repositories"
)

var (
	ErrOrganizationNameRequired      = errors.New("name is required")
	ErrOrganizationTypeRequired      = errors.New("organization_type is required")
	ErrOrganizationTypeNotFound      = errors.New("organization type not found or inactive")
	ErrOrganizationLatitudeRequired  = errors.New("latitude is required")
	ErrOrganizationLongitudeRequired = errors.New("longitude is required")
	ErrOrganizationLatitudeInvalid   = errors.New("latitude must be between -90 and 90")
	ErrOrganizationLongitudeInvalid  = errors.New("longitude must be between -180 and 180")
	ErrOnboardingSourceInvalid       = errors.New("invalid onboarding_source")
	ErrCapabilitiesRequired          = errors.New("at least one capability is required")
	ErrCapabilityNotFound            = errors.New("one or more capabilities were not found or are inactive")
	ErrDuplicateCapability           = errors.New("capabilities must not contain duplicates")
	ErrOrganizationEmailInvalid      = errors.New("email is invalid")
)

type ResponseOrganizationRepository interface {
	Create(
		ctx context.Context,
		name string,
		organizationType string,
		latitude float64,
		longitude float64,
		address *string,
		phone *string,
		email *string,
		onboardingSource string,
		capabilities []string,
	) (*models.ResponseOrganization, error)
}

type CreateResponseOrganizationInput struct {
	Name             string
	OrganizationType string
	Latitude         *float64
	Longitude        *float64
	Address          *string
	Phone            *string
	Email            *string
	OnboardingSource string
	Capabilities     []string
}

type ResponseOrganizationService struct {
	repository ResponseOrganizationRepository
}

func NewResponseOrganizationService(
	repository ResponseOrganizationRepository,
) *ResponseOrganizationService {
	return &ResponseOrganizationService{
		repository: repository,
	}
}

func (s *ResponseOrganizationService) Create(
	ctx context.Context,
	input CreateResponseOrganizationInput,
) (*models.ResponseOrganization, error) {
	input.Name = strings.TrimSpace(input.Name)

	if input.Name == "" {
		return nil, ErrOrganizationNameRequired
	}

	input.OrganizationType = strings.ToUpper(
		strings.TrimSpace(input.OrganizationType),
	)

	if input.OrganizationType == "" {
		return nil, ErrOrganizationTypeRequired
	}

	if input.Latitude == nil {
		return nil, ErrOrganizationLatitudeRequired
	}

	if input.Longitude == nil {
		return nil, ErrOrganizationLongitudeRequired
	}

	if *input.Latitude < -90 || *input.Latitude > 90 {
		return nil, ErrOrganizationLatitudeInvalid
	}

	if *input.Longitude < -180 || *input.Longitude > 180 {
		return nil, ErrOrganizationLongitudeInvalid
	}

	input.OnboardingSource = strings.ToUpper(
		strings.TrimSpace(input.OnboardingSource),
	)

	if input.OnboardingSource == "" {
		input.OnboardingSource = "ADMIN"
	}

	switch input.OnboardingSource {
	case "ADMIN", "SELF_SERVICE", "PARTNER_IMPORT":
	default:
		return nil, ErrOnboardingSourceInvalid
	}

	if len(input.Capabilities) == 0 {
		return nil, ErrCapabilitiesRequired
	}

	normalizedCapabilities := make([]string, 0, len(input.Capabilities))
	seenCapabilities := make(map[string]struct{})

	for _, capability := range input.Capabilities {
		capability = strings.ToUpper(
			strings.TrimSpace(capability),
		)

		if capability == "" {
			return nil, ErrCapabilityNotFound
		}

		if _, exists := seenCapabilities[capability]; exists {
			return nil, ErrDuplicateCapability
		}

		seenCapabilities[capability] = struct{}{}
		normalizedCapabilities = append(
			normalizedCapabilities,
			capability,
		)
	}

	input.Capabilities = normalizedCapabilities

	input.Address = cleanOptionalString(input.Address)
	input.Phone = cleanOptionalString(input.Phone)
	input.Email = cleanOptionalString(input.Email)

	if input.Email != nil {
		if !strings.Contains(*input.Email, "@") {
			return nil, ErrOrganizationEmailInvalid
		}
	}

	organization, err := s.repository.Create(
		ctx,
		input.Name,
		input.OrganizationType,
		*input.Latitude,
		*input.Longitude,
		input.Address,
		input.Phone,
		input.Email,
		input.OnboardingSource,
		input.Capabilities,
	)

	if err != nil {
		switch {
		case errors.Is(err, repositories.ErrOrganizationTypeNotFound):
			return nil, ErrOrganizationTypeNotFound

		case errors.Is(err, repositories.ErrCapabilityNotFound):
			return nil, ErrCapabilityNotFound

		default:
			return nil, err
		}
	}

	return organization, nil
}

func cleanOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	cleaned := strings.TrimSpace(*value)

	if cleaned == "" {
		return nil
	}

	return &cleaned
}
