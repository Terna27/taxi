package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"taxi/internal/services"
)

type OrganizationDiscoveryHandler struct {
	service *services.OrganizationDiscoveryService
}

func NewOrganizationDiscoveryHandler(
	service *services.OrganizationDiscoveryService,
) *OrganizationDiscoveryHandler {
	return &OrganizationDiscoveryHandler{
		service: service,
	}
}

func (h *OrganizationDiscoveryHandler) FindNearby(
	w http.ResponseWriter,
	r *http.Request,
) {
	query := r.URL.Query()

	latitude, err := parseRequiredFloatQuery(
		query.Get("latitude"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "latitude must be a valid number",
		})
		return
	}

	longitude, err := parseRequiredFloatQuery(
		query.Get("longitude"),
	)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "longitude must be a valid number",
		})
		return
	}

	var radiusMeters *float64

	if value := query.Get("radius_meters"); value != "" {
		parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "radius_meters must be a valid number",
			})
			return
		}

		radiusMeters = &parsed
	}

	var limit *int

	if value := query.Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": "limit must be a valid integer",
			})
			return
		}

		limit = &parsed
	}

	organizations, err := h.service.FindNearby(
		r.Context(),
		services.FindNearbyOrganizationsInput{
			Latitude:     latitude,
			Longitude:    longitude,
			Capability:   query.Get("capability"),
			RadiusMeters: radiusMeters,
			Limit:        limit,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrDiscoveryLatitudeRequired),
			errors.Is(err, services.ErrDiscoveryLongitudeRequired),
			errors.Is(err, services.ErrDiscoveryLatitudeInvalid),
			errors.Is(err, services.ErrDiscoveryLongitudeInvalid),
			errors.Is(err, services.ErrDiscoveryCapabilityRequired),
			errors.Is(err, services.ErrDiscoveryRadiusInvalid),
			errors.Is(err, services.ErrDiscoveryRadiusTooLarge),
			errors.Is(err, services.ErrDiscoveryLimitInvalid):

			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})

		default:
			writeJSON(
				w,
				http.StatusInternalServerError,
				map[string]string{
					"error": "failed to discover response organizations",
				},
			)
		}

		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"count":         len(organizations),
		"organizations": organizations,
	})
}

func parseRequiredFloatQuery(
	value string,
) (*float64, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}