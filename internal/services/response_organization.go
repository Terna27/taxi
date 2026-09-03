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
	ErrCapabilitiesRequired          = errors.New("at least one capability is required")
	ErrCapabilityNotFound            = errors.New("one or more capabilities were not found or are inactive")
	ErrDuplicateCapability           = errors.New("capabilities must not contain duplicates")
	ErrOrganizationEmailInvalid      = errors.New("email is invalid")
	ErrOrganizationNameTooLong       = errors.New("name must not exceed 200 characters")
	ErrOrganizationAddressTooLong    = errors.New("address must not exceed 1000 characters")
	ErrOrganizationPhoneTooLong      = errors.New("phone must not exceed 50 characters")
	ErrOrganizationEmailTooLong      = errors.New("email must not exceed 255 characters")
	ErrTooManyCapabilities           = errors.New("too many capabilities")
)

const (
	selfServiceOnboardingSource  = "SELF_SERVICE"
	maxOrganizationNameLength    = 200
	maxOrganizationAddressLength = 1000
	maxOrganizationPhoneLength   = 50
	maxOrganizationEmailLength   = 255
	maxOrganizationCapabilities  = 20
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

	if len(input.Name) > maxOrganizationNameLength {
		return nil, ErrOrganizationNameTooLong
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

	if len(input.Capabilities) == 0 {
		return nil, ErrCapabilitiesRequired
	}

	if len(input.Capabilities) > maxOrganizationCapabilities {
		return nil, ErrTooManyCapabilities
	}

	normalizedCapabilities := make([]string, 0, len(input.Capabilities))
	seenCapabilities := make(map[string]struct{}, len(input.Capabilities))

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

	input.Address = cleanOptionalString(input.Address)
	input.Phone = cleanOptionalString(input.Phone)
	input.Email = cleanOptionalString(input.Email)

	if input.Address != nil &&
		len(*input.Address) > maxOrganizationAddressLength {
		return nil, ErrOrganizationAddressTooLong
	}

	if input.Phone != nil &&
		len(*input.Phone) > maxOrganizationPhoneLength {
		return nil, ErrOrganizationPhoneTooLong
	}

	if input.Email != nil {
		if len(*input.Email) > maxOrganizationEmailLength {
			return nil, ErrOrganizationEmailTooLong
		}

		if !isValidSimpleEmail(*input.Email) {
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
		selfServiceOnboardingSource,
		normalizedCapabilities,
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

func isValidSimpleEmail(value string) bool {
	if strings.ContainsAny(value, " \t\r\n") {
		return false
	}

	at := strings.IndexByte(value, '@')

	if at <= 0 || at == len(value)-1 {
		return false
	}

	if strings.IndexByte(value[at+1:], '@') >= 0 {
		return false
	}

	domain := value[at+1:]

	return strings.Contains(domain, ".")
}
