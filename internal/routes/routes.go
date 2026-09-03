package routes

import (
	"encoding/json"
	"net/http"

	"taxi/internal/handlers"
)

func New(
	emergencyHandler *handlers.EmergencyHandler,
	responseOrganizationHandler *handlers.ResponseOrganizationHandler,
	organizationDiscoveryHandler *handlers.OrganizationDiscoveryHandler,
	dispatchHandler *handlers.DispatchHandler,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(
		"GET /health",
		healthHandler,
	)

	mux.HandleFunc(
		"POST /api/v1/incidents",
		emergencyHandler.Create,
	)

	mux.HandleFunc(
		"POST /api/v1/response-organizations",
		responseOrganizationHandler.Create,
	)

	mux.HandleFunc(
		"GET /api/v1/response-organizations/nearby",
		organizationDiscoveryHandler.FindNearby,
	)

	mux.HandleFunc(
		"POST /api/v1/incidents/{incidentID}/dispatch",
		dispatchHandler.Dispatch,
	)

	return mux
}

func healthHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(
		map[string]string{
			"status":  "ok",
			"service": "rideroute",
		},
	)
}
