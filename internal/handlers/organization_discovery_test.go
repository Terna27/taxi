package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"taxi/internal/models"
	"taxi/internal/services"
)

type fakeOrganizationDiscoveryHandlerRepository struct {
	organizations []models.NearbyResponseOrganization
	err           error
}

func (f *fakeOrganizationDiscoveryHandlerRepository) FindNearby(
	ctx context.Context,
	latitude float64,
	longitude float64,
	capability string,
	radiusMeters float64,
	limit int,
) ([]models.NearbyResponseOrganization, error) {
	if f.err != nil {
		return nil, f.err
	}

	return f.organizations, nil
}

func TestOrganizationDiscoveryHandlerFindNearbySuccess(t *testing.T) {
	repository := &fakeOrganizationDiscoveryHandlerRepository{
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

	service := services.NewOrganizationDiscoveryService(repository)
	handler := NewOrganizationDiscoveryHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/response-organizations/nearby?latitude=7.7318&longitude=8.5382&capability=MEDICAL&radius_meters=10000&limit=10",
		nil,
	)

	response := httptest.NewRecorder()

	handler.FindNearby(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestOrganizationDiscoveryHandlerMalformedLatitude(t *testing.T) {
	repository := &fakeOrganizationDiscoveryHandlerRepository{}
	service := services.NewOrganizationDiscoveryService(repository)
	handler := NewOrganizationDiscoveryHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/response-organizations/nearby?latitude=wrong&longitude=8.5382&capability=MEDICAL",
		nil,
	)

	response := httptest.NewRecorder()

	handler.FindNearby(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestOrganizationDiscoveryHandlerMalformedLongitude(t *testing.T) {
	repository := &fakeOrganizationDiscoveryHandlerRepository{}
	service := services.NewOrganizationDiscoveryService(repository)
	handler := NewOrganizationDiscoveryHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/response-organizations/nearby?latitude=7.7318&longitude=wrong&capability=MEDICAL",
		nil,
	)

	response := httptest.NewRecorder()

	handler.FindNearby(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestOrganizationDiscoveryHandlerMalformedRadius(t *testing.T) {
	repository := &fakeOrganizationDiscoveryHandlerRepository{}
	service := services.NewOrganizationDiscoveryService(repository)
	handler := NewOrganizationDiscoveryHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/response-organizations/nearby?latitude=7.7318&longitude=8.5382&capability=MEDICAL&radius_meters=wrong",
		nil,
	)

	response := httptest.NewRecorder()

	handler.FindNearby(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestOrganizationDiscoveryHandlerMalformedLimit(t *testing.T) {
	repository := &fakeOrganizationDiscoveryHandlerRepository{}
	service := services.NewOrganizationDiscoveryService(repository)
	handler := NewOrganizationDiscoveryHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/response-organizations/nearby?latitude=7.7318&longitude=8.5382&capability=MEDICAL&limit=wrong",
		nil,
	)

	response := httptest.NewRecorder()

	handler.FindNearby(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestOrganizationDiscoveryHandlerServiceValidation(t *testing.T) {
	repository := &fakeOrganizationDiscoveryHandlerRepository{}
	service := services.NewOrganizationDiscoveryService(repository)
	handler := NewOrganizationDiscoveryHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/response-organizations/nearby?latitude=7.7318&longitude=8.5382&capability=MEDICAL&radius_meters=500000",
		nil,
	)

	response := httptest.NewRecorder()

	handler.FindNearby(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}

func TestOrganizationDiscoveryHandlerRepositoryError(t *testing.T) {
	repository := &fakeOrganizationDiscoveryHandlerRepository{
		err: errors.New("database failure"),
	}

	service := services.NewOrganizationDiscoveryService(repository)
	handler := NewOrganizationDiscoveryHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/response-organizations/nearby?latitude=7.7318&longitude=8.5382&capability=MEDICAL",
		nil,
	)

	response := httptest.NewRecorder()

	handler.FindNearby(response, request)

	if response.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusInternalServerError,
			response.Code,
		)
	}
}

func TestOrganizationDiscoveryHandlerEmptyResults(t *testing.T) {
	repository := &fakeOrganizationDiscoveryHandlerRepository{
		organizations: []models.NearbyResponseOrganization{},
	}

	service := services.NewOrganizationDiscoveryService(repository)
	handler := NewOrganizationDiscoveryHandler(service)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/response-organizations/nearby?latitude=7.7318&longitude=8.5382&capability=FIRE",
		nil,
	)

	response := httptest.NewRecorder()

	handler.FindNearby(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			response.Code,
		)
	}

	expected := `"count":0`

	if !containsResponseText(
		response.Body.String(),
		expected,
	) {
		t.Fatalf(
			"expected response to contain %s, got %s",
			expected,
			response.Body.String(),
		)
	}
}

func containsResponseText(value string, expected string) bool {
	for i := 0; i+len(expected) <= len(value); i++ {
		if value[i:i+len(expected)] == expected {
			return true
		}
	}

	return false
} 