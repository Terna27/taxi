package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"taxi/internal/models"
	"taxi/internal/repositories"
	"taxi/internal/services"
)

type fakeHandlerEmergencyRepository struct {
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

func (f *fakeHandlerEmergencyRepository) Create(
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

func TestEmergencyHandlerCreateSuccess(t *testing.T) {
	repository := &fakeHandlerEmergencyRepository{
		createFn: func(
			ctx context.Context,
			emergencyType string,
			latitude float64,
			longitude float64,
			locationAccuracyMeters *float64,
			description *string,
			idempotencyKey string,
		) (*models.EmergencyIncident, error) {
			return &models.EmergencyIncident{
				ID:                     "incident-1",
				EmergencyType:          emergencyType,
				Status:                 "CREATED",
				Latitude:               latitude,
				Longitude:              longitude,
				LocationAccuracyMeters: locationAccuracyMeters,
				Description:            description,
				IdempotencyKey:         idempotencyKey,
				CreatedAt:              time.Now(),
				UpdatedAt:              time.Now(),
			}, nil
		},
	}

	service := services.NewEmergencyService(repository)
	handler := NewEmergencyHandler(service)

	body := `{
        "emergency_type": "MEDICAL",
        "latitude": 7.7318,
        "longitude": 8.5382,
        "location_accuracy_meters": 8.5,
        "description": "Medical emergency",
        "idempotency_key": "550e8400-e29b-41d4-a716-446655440001"
    }`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/incidents",
		strings.NewReader(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusCreated,
			recorder.Code,
			recorder.Body.String(),
		)
	}

	var response struct {
		Message  string                   `json:"message"`
		Incident models.EmergencyIncident `json:"incident"`
	}

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf(
			"decode response: %v",
			err,
		)
	}

	if response.Incident.Status != "CREATED" {
		t.Fatalf(
			"expected status CREATED, got %s",
			response.Incident.Status,
		)
	}

	if response.Incident.EmergencyType != "MEDICAL" {
		t.Fatalf(
			"expected MEDICAL, got %s",
			response.Incident.EmergencyType,
		)
	}
}

func TestEmergencyHandlerCreateInvalidLatitude(t *testing.T) {
	repository := &fakeHandlerEmergencyRepository{
		createFn: func(
			ctx context.Context,
			emergencyType string,
			latitude float64,
			longitude float64,
			locationAccuracyMeters *float64,
			description *string,
			idempotencyKey string,
		) (*models.EmergencyIncident, error) {
			t.Fatal("repository should not be called")
			return nil, nil
		},
	}

	service := services.NewEmergencyService(repository)
	handler := NewEmergencyHandler(service)

	body := `{
        "emergency_type": "MEDICAL",
        "latitude": 200,
        "longitude": 8.5382,
        "idempotency_key": "550e8400-e29b-41d4-a716-446655440001"
    }`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/incidents",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestEmergencyHandlerCreateUnknownType(t *testing.T) {
	repository := &fakeHandlerEmergencyRepository{
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

	service := services.NewEmergencyService(repository)
	handler := NewEmergencyHandler(service)

	body := `{
        "emergency_type": "ALIEN_INVASION",
        "latitude": 7.7318,
        "longitude": 8.5382,
        "idempotency_key": "550e8400-e29b-41d4-a716-446655440001"
    }`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/incidents",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusBadRequest,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestEmergencyHandlerCreateDuplicateIncident(t *testing.T) {
	repository := &fakeHandlerEmergencyRepository{
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

	service := services.NewEmergencyService(repository)
	handler := NewEmergencyHandler(service)

	body := `{
        "emergency_type": "MEDICAL",
        "latitude": 7.7318,
        "longitude": 8.5382,
        "idempotency_key": "550e8400-e29b-41d4-a716-446655440001"
    }`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/incidents",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusConflict {
		t.Fatalf(
			"expected status %d, got %d; body=%s",
			http.StatusConflict,
			recorder.Code,
			recorder.Body.String(),
		)
	}
}

func TestEmergencyHandlerRejectsUnknownJSONField(t *testing.T) {
	repository := &fakeHandlerEmergencyRepository{
		createFn: func(
			ctx context.Context,
			emergencyType string,
			latitude float64,
			longitude float64,
			locationAccuracyMeters *float64,
			description *string,
			idempotencyKey string,
		) (*models.EmergencyIncident, error) {
			t.Fatal("repository should not be called")
			return nil, nil
		},
	}

	service := services.NewEmergencyService(repository)
	handler := NewEmergencyHandler(service)

	body := `{
        "emergency_type": "MEDICAL",
        "latitude": 7.7318,
        "longitude": 8.5382,
        "unexpected_field": "bad",
        "idempotency_key": "550e8400-e29b-41d4-a716-446655440001"
    }`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/incidents",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}

func TestEmergencyHandlerRejectsMultipleJSONObjects(t *testing.T) {
	repository := &fakeHandlerEmergencyRepository{
		createFn: func(
			ctx context.Context,
			emergencyType string,
			latitude float64,
			longitude float64,
			locationAccuracyMeters *float64,
			description *string,
			idempotencyKey string,
		) (*models.EmergencyIncident, error) {
			t.Fatal("repository should not be called")
			return nil, nil
		},
	}

	service := services.NewEmergencyService(repository)
	handler := NewEmergencyHandler(service)

	body := `{
        "emergency_type": "MEDICAL",
        "latitude": 7.7318,
        "longitude": 8.5382,
        "idempotency_key": "550e8400-e29b-41d4-a716-446655440001"
    }
    {
        "extra": "object"
    }`

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/incidents",
		strings.NewReader(body),
	)

	recorder := httptest.NewRecorder()

	handler.Create(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			recorder.Code,
		)
	}
}
