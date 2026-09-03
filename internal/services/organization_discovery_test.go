package services

import (
	"context"
	"errors"
	"testing"

	"taxi/internal/models"
)

type fakeOrganizationDiscoveryRepository struct {
	organizations []models.NearbyResponseOrganization
	err           error

	receivedLatitude     float64
	receivedLongitude    float64
	receivedCapability   string
	receivedRadiusMeters float64
	receivedLimit        int
}

func (f *fakeOrganizationDiscoveryRepository) FindNearby(
	ctx context.Context,
	latitude float64,
	longitude float64,
	capability string,
	radiusMeters float64,
	limit int,
) ([]models.NearbyResponseOrganization, error) {
	f.receivedLatitude = latitude
	f.receivedLongitude = longitude
	f.receivedCapability = capability
	f.receivedRadiusMeters = radiusMeters
	f.receivedLimit = limit

	if f.err != nil {
		return nil, f.err
	}

	return f.organizations, nil
}

func TestOrganizationDiscoveryServiceFindNearbySuccess(t *testing.T) {
	latitude := 7.7318
	longitude := 8.5382
	radius := 15000.0
	limit := 5

	repository := &fakeOrganizationDiscoveryRepository{
		organizations: []models.NearbyResponseOrganization{
			{
				ID:               "organization-id",
				Name:             "RideRoute Test Hospital",
				OrganizationType: "HOSPITAL",
				Latitude:         7.7318,
				Longitude:        8.5382,
				DistanceMeters:   0,
				Capabilities: []string{
					"AMBULANCE",
					"MEDICAL",
				},
			},
		},
	}

	service := NewOrganizationDiscoveryService(repository)

	organizations, err := service.FindNearby(
		context.Background(),
		FindNearbyOrganizationsInput{
			Latitude:     &latitude,
			Longitude:    &longitude,
			Capability:   " medical ",
			RadiusMeters: &radius,
			Limit:        &limit,
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(organizations) != 1 {
		t.Fatalf(
			"expected 1 organization, got %d",
			len(organizations),
		)
	}

	if repository.receivedCapability != "MEDICAL" {
		t.Fatalf(
			"expected MEDICAL, got %q",
			repository.receivedCapability,
		)
	}

	if repository.receivedRadiusMeters != 15000 {
		t.Fatalf(
			"expected radius 15000, got %f",
			repository.receivedRadiusMeters,
		)
	}

	if repository.receivedLimit != 5 {
		t.Fatalf(
			"expected limit 5, got %d",
			repository.receivedLimit,
		)
	}
}

func TestOrganizationDiscoveryServiceDefaults(t *testing.T) {
	latitude := 7.7318
	longitude := 8.5382

	repository := &fakeOrganizationDiscoveryRepository{
		organizations: []models.NearbyResponseOrganization{},
	}

	service := NewOrganizationDiscoveryService(repository)

	_, err := service.FindNearby(
		context.Background(),
		FindNearbyOrganizationsInput{
			Latitude:   &latitude,
			Longitude:  &longitude,
			Capability: "MEDICAL",
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if repository.receivedRadiusMeters != defaultDiscoveryRadiusMeters {
		t.Fatalf(
			"expected default radius %f, got %f",
			defaultDiscoveryRadiusMeters,
			repository.receivedRadiusMeters,
		)
	}

	if repository.receivedLimit != defaultDiscoveryLimit {
		t.Fatalf(
			"expected default limit %d, got %d",
			defaultDiscoveryLimit,
			repository.receivedLimit,
		)
	}
}

func TestOrganizationDiscoveryServiceValidation(t *testing.T) {
	validLatitude := 7.7318
	validLongitude := 8.5382

	invalidLatitude := 91.0
	invalidLongitude := 181.0
	zeroRadius := 0.0
	negativeRadius := -1.0
	tooLargeRadius := 100001.0
	zeroLimit := 0
	tooLargeLimit := 51

	tests := []struct {
		name     string
		input    FindNearbyOrganizationsInput
		expected error
	}{
		{
			name: "missing latitude",
			input: FindNearbyOrganizationsInput{
				Longitude:  &validLongitude,
				Capability: "MEDICAL",
			},
			expected: ErrDiscoveryLatitudeRequired,
		},
		{
			name: "missing longitude",
			input: FindNearbyOrganizationsInput{
				Latitude:   &validLatitude,
				Capability: "MEDICAL",
			},
			expected: ErrDiscoveryLongitudeRequired,
		},
		{
			name: "invalid latitude",
			input: FindNearbyOrganizationsInput{
				Latitude:   &invalidLatitude,
				Longitude:  &validLongitude,
				Capability: "MEDICAL",
			},
			expected: ErrDiscoveryLatitudeInvalid,
		},
		{
			name: "invalid longitude",
			input: FindNearbyOrganizationsInput{
				Latitude:   &validLatitude,
				Longitude:  &invalidLongitude,
				Capability: "MEDICAL",
			},
			expected: ErrDiscoveryLongitudeInvalid,
		},
		{
			name: "missing capability",
			input: FindNearbyOrganizationsInput{
				Latitude:  &validLatitude,
				Longitude: &validLongitude,
			},
			expected: ErrDiscoveryCapabilityRequired,
		},
		{
			name: "zero radius",
			input: FindNearbyOrganizationsInput{
				Latitude:     &validLatitude,
				Longitude:    &validLongitude,
				Capability:   "MEDICAL",
				RadiusMeters: &zeroRadius,
			},
			expected: ErrDiscoveryRadiusInvalid,
		},
		{
			name: "negative radius",
			input: FindNearbyOrganizationsInput{
				Latitude:     &validLatitude,
				Longitude:    &validLongitude,
				Capability:   "MEDICAL",
				RadiusMeters: &negativeRadius,
			},
			expected: ErrDiscoveryRadiusInvalid,
		},
		{
			name: "radius too large",
			input: FindNearbyOrganizationsInput{
				Latitude:     &validLatitude,
				Longitude:    &validLongitude,
				Capability:   "MEDICAL",
				RadiusMeters: &tooLargeRadius,
			},
			expected: ErrDiscoveryRadiusTooLarge,
		},
		{
			name: "zero limit",
			input: FindNearbyOrganizationsInput{
				Latitude:   &validLatitude,
				Longitude:  &validLongitude,
				Capability: "MEDICAL",
				Limit:      &zeroLimit,
			},
			expected: ErrDiscoveryLimitInvalid,
		},
		{
			name: "limit too large",
			input: FindNearbyOrganizationsInput{
				Latitude:   &validLatitude,
				Longitude:  &validLongitude,
				Capability: "MEDICAL",
				Limit:      &tooLargeLimit,
			},
			expected: ErrDiscoveryLimitInvalid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &fakeOrganizationDiscoveryRepository{}
			service := NewOrganizationDiscoveryService(repository)

			_, err := service.FindNearby(
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

func TestOrganizationDiscoveryServiceRepositoryError(t *testing.T) {
	latitude := 7.7318
	longitude := 8.5382

	expectedError := errors.New("database failure")

	repository := &fakeOrganizationDiscoveryRepository{
		err: expectedError,
	}

	service := NewOrganizationDiscoveryService(repository)

	_, err := service.FindNearby(
		context.Background(),
		FindNearbyOrganizationsInput{
			Latitude:   &latitude,
			Longitude:  &longitude,
			Capability: "MEDICAL",
		},
	)

	if !errors.Is(err, expectedError) {
		t.Fatalf(
			"expected repository error, got %v",
			err,
		)
	}
}

func TestOrganizationDiscoveryServiceEmptyResults(t *testing.T) {
	latitude := 7.7318
	longitude := 8.5382

	repository := &fakeOrganizationDiscoveryRepository{
		organizations: []models.NearbyResponseOrganization{},
	}

	service := NewOrganizationDiscoveryService(repository)

	organizations, err := service.FindNearby(
		context.Background(),
		FindNearbyOrganizationsInput{
			Latitude:   &latitude,
			Longitude:  &longitude,
			Capability: "FIRE",
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if organizations == nil {
		t.Fatal("expected empty slice, got nil")
	}

	if len(organizations) != 0 {
		t.Fatalf(
			"expected 0 organizations, got %d",
			len(organizations),
		)
	}
}