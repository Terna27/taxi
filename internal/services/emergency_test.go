package services

import (
    "context"
    "errors"
    "strings"
    "testing"
    "time"

    "taxi/internal/models"
    "taxi/internal/repositories"
)

type fakeResponseOrganizationRepository struct {
    organization *models.ResponseOrganization
    err          error

    receivedName             string
    receivedOrganizationType string
    receivedOnboardingSource string
    receivedCapabilities     []string
}

func (f *fakeResponseOrganizationRepository) Create(
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
) (*models.ResponseOrganization, error) {
    f.receivedName = name
    f.receivedOrganizationType = organizationType
    f.receivedOnboardingSource = onboardingSource
    f.receivedCapabilities = capabilities

    if f.err != nil {
        return nil, f.err
    }

    return f.organization, nil
}

func TestResponseOrganizationServiceCreateSuccess(t *testing.T) {
    latitude := 7.7318
    longitude := 8.5382

    repository := &fakeResponseOrganizationRepository{
        organization: &models.ResponseOrganization{
            ID:                 "organization-id",
            Name:               "RideRoute Test Hospital",
            OrganizationType:   "HOSPITAL",
            Latitude:           latitude,
            Longitude:          longitude,
            OnboardingSource:   "SELF_SERVICE",
            VerificationStatus: "PENDING",
            IsActive:           false,
            Capabilities: []string{
                "MEDICAL",
                "AMBULANCE",
            },
            CreatedAt: time.Now(),
            UpdatedAt: time.Now(),
        },
    }

    service := NewResponseOrganizationService(repository)

    organization, err := service.Create(
        context.Background(),
        CreateResponseOrganizationInput{
            Name:             "  RideRoute Test Hospital  ",
            OrganizationType: " hospital ",
            Latitude:         &latitude,
            Longitude:        &longitude,
            Capabilities: []string{
                " medical ",
                "ambulance",
            },
        },
    )

    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    if organization == nil {
        t.Fatal("expected organization")
    }

    if repository.receivedName != "RideRoute Test Hospital" {
        t.Fatalf(
            "expected normalized name, got %q",
            repository.receivedName,
        )
    }

    if repository.receivedOrganizationType != "HOSPITAL" {
        t.Fatalf(
            "expected HOSPITAL, got %q",
            repository.receivedOrganizationType,
        )
    }

    if repository.receivedOnboardingSource != "SELF_SERVICE" {
        t.Fatalf(
            "expected SELF_SERVICE, got %q",
            repository.receivedOnboardingSource,
        )
    }

    if len(repository.receivedCapabilities) != 2 {
        t.Fatalf(
            "expected 2 capabilities, got %d",
            len(repository.receivedCapabilities),
        )
    }

    if repository.receivedCapabilities[0] != "MEDICAL" {
        t.Fatalf(
            "expected MEDICAL, got %q",
            repository.receivedCapabilities[0],
        )
    }

    if repository.receivedCapabilities[1] != "AMBULANCE" {
        t.Fatalf(
            "expected AMBULANCE, got %q",
            repository.receivedCapabilities[1],
        )
    }
}

func TestResponseOrganizationServiceCreateValidation(t *testing.T) {
    validLatitude := 7.7318
    validLongitude := 8.5382
    invalidLatitude := 91.0
    invalidLongitude := 181.0

    tests := []struct {
        name     string
        input    CreateResponseOrganizationInput
        expected error
    }{
        {
            name: "missing name",
            input: CreateResponseOrganizationInput{
                OrganizationType: "HOSPITAL",
                Latitude:         &validLatitude,
                Longitude:        &validLongitude,
                Capabilities:     []string{"MEDICAL"},
            },
            expected: ErrOrganizationNameRequired,
        },
        {
            name: "missing organization type",
            input: CreateResponseOrganizationInput{
                Name:         "Hospital",
                Latitude:     &validLatitude,
                Longitude:    &validLongitude,
                Capabilities: []string{"MEDICAL"},
            },
            expected: ErrOrganizationTypeRequired,
        },
        {
            name: "missing latitude",
            input: CreateResponseOrganizationInput{
                Name:             "Hospital",
                OrganizationType: "HOSPITAL",
                Longitude:        &validLongitude,
                Capabilities:     []string{"MEDICAL"},
            },
            expected: ErrOrganizationLatitudeRequired,
        },
        {
            name: "missing longitude",
            input: CreateResponseOrganizationInput{
                Name:             "Hospital",
                OrganizationType: "HOSPITAL",
                Latitude:         &validLatitude,
                Capabilities:     []string{"MEDICAL"},
            },
            expected: ErrOrganizationLongitudeRequired,
        },
        {
            name: "invalid latitude",
            input: CreateResponseOrganizationInput{
                Name:             "Hospital",
                OrganizationType: "HOSPITAL",
                Latitude:         &invalidLatitude,
                Longitude:        &validLongitude,
                Capabilities:     []string{"MEDICAL"},
            },
            expected: ErrOrganizationLatitudeInvalid,
        },
        {
            name: "invalid longitude",
            input: CreateResponseOrganizationInput{
                Name:             "Hospital",
                OrganizationType: "HOSPITAL",
                Latitude:         &validLatitude,
                Longitude:        &invalidLongitude,
                Capabilities:     []string{"MEDICAL"},
            },
            expected: ErrOrganizationLongitudeInvalid,
        },
        {
            name: "missing capabilities",
            input: CreateResponseOrganizationInput{
                Name:             "Hospital",
                OrganizationType: "HOSPITAL",
                Latitude:         &validLatitude,
                Longitude:        &validLongitude,
            },
            expected: ErrCapabilitiesRequired,
        },
        {
            name: "duplicate capability",
            input: CreateResponseOrganizationInput{
                Name:             "Hospital",
                OrganizationType: "HOSPITAL",
                Latitude:         &validLatitude,
                Longitude:        &validLongitude,
                Capabilities: []string{
                    "MEDICAL",
                    "medical",
                },
            },
            expected: ErrDuplicateCapability,
        },
        {
            name: "invalid email",
            input: CreateResponseOrganizationInput{
                Name:             "Hospital",
                OrganizationType: "HOSPITAL",
                Latitude:         &validLatitude,
                Longitude:        &validLongitude,
                Email:            testString("invalid-email"),
                Capabilities:     []string{"MEDICAL"},
            },
            expected: ErrOrganizationEmailInvalid,
        },
        {
            name: "name too long",
            input: CreateResponseOrganizationInput{
                Name:             strings.Repeat("a", 201),
                OrganizationType: "HOSPITAL",
                Latitude:         &validLatitude,
                Longitude:        &validLongitude,
                Capabilities:     []string{"MEDICAL"},
            },
            expected: ErrOrganizationNameTooLong,
        },
    }

    for _, test := range tests {
        t.Run(test.name, func(t *testing.T) {
            repository := &fakeResponseOrganizationRepository{}
            service := NewResponseOrganizationService(repository)

            _, err := service.Create(
                context.Background(),
                test.input,
            )

            if !errors.Is(err, test.expected) {
                t.Fatalf(
                    "expected %v, got %v",
                    test.expected,
                    err,
                )
            }
        })
    }
}

func TestResponseOrganizationServiceCreateUnknownType(t *testing.T) {
    latitude := 7.7318
    longitude := 8.5382

    repository := &fakeResponseOrganizationRepository{
        err: repositories.ErrOrganizationTypeNotFound,
    }

    service := NewResponseOrganizationService(repository)

    _, err := service.Create(
        context.Background(),
        CreateResponseOrganizationInput{
            Name:             "Unknown Organization",
            OrganizationType: "UNKNOWN",
            Latitude:         &latitude,
            Longitude:        &longitude,
            Capabilities:     []string{"MEDICAL"},
        },
    )

    if !errors.Is(err, ErrOrganizationTypeNotFound) {
        t.Fatalf(
            "expected %v, got %v",
            ErrOrganizationTypeNotFound,
            err,
        )
    }
}

func TestResponseOrganizationServiceCreateUnknownCapability(t *testing.T) {
    latitude := 7.7318
    longitude := 8.5382

    repository := &fakeResponseOrganizationRepository{
        err: repositories.ErrCapabilityNotFound,
    }

    service := NewResponseOrganizationService(repository)

    _, err := service.Create(
        context.Background(),
        CreateResponseOrganizationInput{
            Name:             "Hospital",
            OrganizationType: "HOSPITAL",
            Latitude:         &latitude,
            Longitude:        &longitude,
            Capabilities:     []string{"UNKNOWN"},
        },
    )

    if !errors.Is(err, ErrCapabilityNotFound) {
        t.Fatalf(
            "expected %v, got %v",
            ErrCapabilityNotFound,
            err,
        )
    }
}

func testString(value string) *string {
    return &value
}