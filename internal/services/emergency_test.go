package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"taxi/internal/models"
	"taxi/internal/repositories"
)

type fakeEmergencyRepository struct {
	createFn func(
		ctx context.Context,
		emergencyType string,
		latitude float64,
		longitude float64,
		locationAccuracyMeters *float64,
		description *string,
		idempotencyKey string,
	) (*models.EmergencyIncident, error)
}

func (f *fakeEmergencyRepository) Create(
	ctx context.Context,
	emergencyType string,
	latitude float64,
	longitude float64,
	locationAccuracyMeters *float64,
	description *string,
	idempotencyKey string,
) (*models.EmergencyIncident, error) {
	return f.createFn(
		ctx,
		emergencyType,
		latitude,
		longitude,
		locationAccuracyMeters,
		description,
		idempotencyKey,
	)
}

func TestEmergencyServiceCreateSuccess(t *testing.T) {
	latitude := 7.7318
	longitude := 8.5382
	accuracy := 8.5
	description := "Medical emergency"

	repository := &fakeEmergencyRepository{
		createFn: func(
			ctx context.Context,
			emergencyType string,
			receivedLatitude float64,
			receivedLongitude float64,
			locationAccuracyMeters *float64,
			receivedDescription *string,
			idempotencyKey string,
		) (*models.EmergencyIncident, error) {
			if emergencyType != "MEDICAL" {
				t.Fatalf(
					"expected emergency type MEDICAL, got %s",
					emergencyType,
				)
			}

			if receivedLatitude != latitude {
				t.Fatalf(
					"expected latitude %f, got %f",
					latitude,
					receivedLatitude,
				)
			}

			if receivedLongitude != longitude {
				t.Fatalf(
					"expected longitude %f, got %f",
					longitude,
					receivedLongitude,
				)
			}

			return &models.EmergencyIncident{
				ID:                     "incident-1",
				EmergencyType:          emergencyType,
				Status:                 "CREATED",
				Latitude:               receivedLatitude,
				Longitude:              receivedLongitude,
				LocationAccuracyMeters: locationAccuracyMeters,
				Description:            receivedDescription,
				IdempotencyKey:         idempotencyKey,
				CreatedAt:              time.Now(),
				UpdatedAt:              time.Now(),
			}, nil
		},
	}

	service := NewEmergencyService(repository)

	incident, err := service.Create(
		context.Background(),
		CreateEmergencyInput{
			EmergencyType:          " medical ",
			Latitude:               &latitude,
			Longitude:              &longitude,
			LocationAccuracyMeters: &accuracy,
			Description:            &description,
			IdempotencyKey:         "550e8400-e29b-41d4-a716-446655440001",
		},
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if incident.Status != "CREATED" {
		t.Fatalf(
			"expected status CREATED, got %s",
			incident.Status,
		)
	}

	if incident.EmergencyType != "MEDICAL" {
		t.Fatalf(
			"expected emergency type MEDICAL, got %s",
			incident.EmergencyType,
		)
	}
}

func TestEmergencyServiceCreateValidation(t *testing.T) {
	validLatitude := 7.7318
	validLongitude := 8.5382
	negativeAccuracy := -1.0
	invalidLatitude := 91.0
	invalidLongitude := 181.0
	longDescription := makeString(2001)

	repository := &fakeEmergencyRepository{
		createFn: func(
			ctx context.Context,
			emergencyType string,
			latitude float64,
			longitude float64,
			locationAccuracyMeters *float64,
			description *string,
			idempotencyKey string,
		) (*models.EmergencyIncident, error) {
			t.Fatal("repository should not be called for validation errors")
			return nil, nil
		},
	}

	service := NewEmergencyService(repository)

	tests := []struct {
		name     string
		input    CreateEmergencyInput
		expected error
	}{
		{
			name: "missing emergency type",
			input: CreateEmergencyInput{
				Latitude:       &validLatitude,
				Longitude:      &validLongitude,
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			},
			expected: ErrEmergencyTypeRequired,
		},
		{
			name: "missing latitude",
			input: CreateEmergencyInput{
				EmergencyType:  "MEDICAL",
				Longitude:      &validLongitude,
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			},
			expected: ErrLatitudeRequired,
		},
		{
			name: "missing longitude",
			input: CreateEmergencyInput{
				EmergencyType:  "MEDICAL",
				Latitude:       &validLatitude,
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			},
			expected: ErrLongitudeRequired,
		},
		{
			name: "invalid latitude",
			input: CreateEmergencyInput{
				EmergencyType:  "MEDICAL",
				Latitude:       &invalidLatitude,
				Longitude:      &validLongitude,
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			},
			expected: ErrInvalidLatitude,
		},
		{
			name: "invalid longitude",
			input: CreateEmergencyInput{
				EmergencyType:  "MEDICAL",
				Latitude:       &validLatitude,
				Longitude:      &invalidLongitude,
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			},
			expected: ErrInvalidLongitude,
		},
		{
			name: "negative accuracy",
			input: CreateEmergencyInput{
				EmergencyType:          "MEDICAL",
				Latitude:               &validLatitude,
				Longitude:              &validLongitude,
				LocationAccuracyMeters: &negativeAccuracy,
				IdempotencyKey:         "550e8400-e29b-41d4-a716-446655440001",
			},
			expected: ErrInvalidLocationAccuracy,
		},
		{
			name: "description too long",
			input: CreateEmergencyInput{
				EmergencyType:  "MEDICAL",
				Latitude:       &validLatitude,
				Longitude:      &validLongitude,
				Description:    &longDescription,
				IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
			},
			expected: ErrDescriptionTooLong,
		},
		{
			name: "missing idempotency key",
			input: CreateEmergencyInput{
				EmergencyType: "MEDICAL",
				Latitude:      &validLatitude,
				Longitude:     &validLongitude,
			},
			expected: ErrIdempotencyKeyRequired,
		},
		{
			name: "invalid idempotency key",
			input: CreateEmergencyInput{
				EmergencyType:  "MEDICAL",
				Latitude:       &validLatitude,
				Longitude:      &validLongitude,
				IdempotencyKey: "not-a-uuid",
			},
			expected: ErrInvalidIdempotencyKey,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.Create(
				context.Background(),
				test.input,
			)

			if !errors.Is(err, test.expected) {
				t.Fatalf(
					"expected error %v, got %v",
					test.expected,
					err,
				)
			}
		})
	}
}

func TestEmergencyServiceCreateUnknownEmergencyType(t *testing.T) {
	latitude := 7.7318
	longitude := 8.5382

	repository := &fakeEmergencyRepository{
		createFn: func(
			ctx context.Context,
			emergencyType string,
			latitude float64,
			longitude float64,
			locationAccuracyMeters *float64,
			description *string,
			idempotencyKey string,
		) (*models.EmergencyIncident, error) {
			return nil, repositories.ErrEmergencyTypeNotFound
		},
	}

	service := NewEmergencyService(repository)

	_, err := service.Create(
		context.Background(),
		CreateEmergencyInput{
			EmergencyType:  "UNKNOWN",
			Latitude:       &latitude,
			Longitude:      &longitude,
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
		},
	)

	if !errors.Is(err, ErrEmergencyTypeNotFound) {
		t.Fatalf(
			"expected ErrEmergencyTypeNotFound, got %v",
			err,
		)
	}
}

func TestEmergencyServiceCreateDuplicateIncident(t *testing.T) {
	latitude := 7.7318
	longitude := 8.5382

	repository := &fakeEmergencyRepository{
		createFn: func(
			ctx context.Context,
			emergencyType string,
			latitude float64,
			longitude float64,
			locationAccuracyMeters *float64,
			description *string,
			idempotencyKey string,
		) (*models.EmergencyIncident, error) {
			return nil, repositories.ErrDuplicateIdempotencyKey
		},
	}

	service := NewEmergencyService(repository)

	_, err := service.Create(
		context.Background(),
		CreateEmergencyInput{
			EmergencyType:  "MEDICAL",
			Latitude:       &latitude,
			Longitude:      &longitude,
			IdempotencyKey: "550e8400-e29b-41d4-a716-446655440001",
		},
	)

	if !errors.Is(err, ErrDuplicateIncident) {
		t.Fatalf(
			"expected ErrDuplicateIncident, got %v",
			err,
		)
	}
}

func makeString(length int) string {
	value := make([]byte, length)

	for i := range value {
		value[i] = 'a'
	}

	return string(value)
}
