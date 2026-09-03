package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"taxi/internal/models"
	"taxi/internal/services"
)

type fakeResponseOrganizationHandlerRepository struct {
	organization *models.ResponseOrganization
	err          error
}

func (f *fakeResponseOrganizationHandlerRepository) Create(
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
	if f.err != nil {
		return nil, f.err
	}

	return f.organization, nil
}

func TestResponseOrganizationHandlerCreateSuccess(t *testing.T) {
	repository := &fakeResponseOrganizationHandlerRepository{
		organization: &models.ResponseOrganization{
			ID:                 "organization-id",
			Name:               "RideRoute Test Hospital",
			OrganizationType:   "HOSPITAL",
			Latitude:           7.7318,
			Longitude:          8.5382,
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

	service := services.NewResponseOrganizationService(repository)
	handler := NewResponseOrganizationHandler(service)

	body := strings.NewReader(
		`{
            "name":"RideRoute Test Hospital",
            "organization_type":"HOSPITAL",
            "latitude":7.7318,
            "longitude":8.5382,
            "capabilities":["MEDICAL","AMBULANCE"]
        }`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/response-organizations",
		body,
	)

	response := httptest.NewRecorder()

	handler.Create(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestResponseOrganizationHandlerRejectsOnboardingSource(t *testing.T) {
	repository := &fakeResponseOrganizationHandlerRepository{}
	service := services.NewResponseOrganizationService(repository)
	handler := NewResponseOrganizationHandler(service)

	body := strings.NewReader(
		`{
            "name":"Test Hospital",
            "organization_type":"HOSPITAL",
            "latitude":7.7318,
            "longitude":8.5382,
            "onboarding_source":"ADMIN",
            "capabilities":["MEDICAL"]
        }`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/response-organizations",
		body,
	)

	response := httptest.NewRecorder()

	handler.Create(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestResponseOrganizationHandlerRejectsUnknownField(t *testing.T) {
	repository := &fakeResponseOrganizationHandlerRepository{}
	service := services.NewResponseOrganizationService(repository)
	handler := NewResponseOrganizationHandler(service)

	body := strings.NewReader(
		`{
            "name":"Test Hospital",
            "organization_type":"HOSPITAL",
            "latitude":7.7318,
            "longitude":8.5382,
            "capabilities":["MEDICAL"],
            "unknown_field":"value"
        }`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/response-organizations",
		body,
	)

	response := httptest.NewRecorder()

	handler.Create(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestResponseOrganizationHandlerRejectsMultipleObjects(t *testing.T) {
	repository := &fakeResponseOrganizationHandlerRepository{}
	service := services.NewResponseOrganizationService(repository)
	handler := NewResponseOrganizationHandler(service)

	body := strings.NewReader(
		`{
            "name":"Test Hospital",
            "organization_type":"HOSPITAL",
            "latitude":7.7318,
            "longitude":8.5382,
            "capabilities":["MEDICAL"]
        }
        {
            "name":"Second Hospital"
        }`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/response-organizations",
		body,
	)

	response := httptest.NewRecorder()

	handler.Create(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestResponseOrganizationHandlerValidationError(t *testing.T) {
	repository := &fakeResponseOrganizationHandlerRepository{}
	service := services.NewResponseOrganizationService(repository)
	handler := NewResponseOrganizationHandler(service)

	body := strings.NewReader(
		`{
            "name":"Test Hospital",
            "organization_type":"HOSPITAL",
            "latitude":200,
            "longitude":8.5382,
            "capabilities":["MEDICAL"]
        }`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/response-organizations",
		body,
	)

	response := httptest.NewRecorder()

	handler.Create(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestResponseOrganizationHandlerInternalError(t *testing.T) {
	repository := &fakeResponseOrganizationHandlerRepository{
		err: errors.New("database failure"),
	}

	service := services.NewResponseOrganizationService(repository)
	handler := NewResponseOrganizationHandler(service)

	body := strings.NewReader(
		`{
            "name":"Test Hospital",
            "organization_type":"HOSPITAL",
            "latitude":7.7318,
            "longitude":8.5382,
            "capabilities":["MEDICAL"]
        }`,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/response-organizations",
		body,
	)

	response := httptest.NewRecorder()

	handler.Create(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			response.Code,
		)
	}
}
