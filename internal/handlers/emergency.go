package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"taxi/internal/services"
)

const maxIncidentRequestBodyBytes = 64 * 1024

type EmergencyHandler struct {
	service *services.EmergencyService
}

type createEmergencyRequest struct {
	EmergencyType          string   `json:"emergency_type"`
	Latitude               *float64 `json:"latitude"`
	Longitude              *float64 `json:"longitude"`
	LocationAccuracyMeters *float64 `json:"location_accuracy_meters"`
	Description            *string  `json:"description"`
	IdempotencyKey         string   `json:"idempotency_key"`
}

func NewEmergencyHandler(
	service *services.EmergencyService,
) *EmergencyHandler {
	return &EmergencyHandler{
		service: service,
	}
}

func (h *EmergencyHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxIncidentRequestBodyBytes,
	)

	defer r.Body.Close()

	var request createEmergencyRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "invalid request body",
		})
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "request body must contain exactly one JSON object",
		})
		return
	}

	incident, err := h.service.Create(
		r.Context(),
		services.CreateEmergencyInput{
			EmergencyType:          request.EmergencyType,
			Latitude:               request.Latitude,
			Longitude:              request.Longitude,
			LocationAccuracyMeters: request.LocationAccuracyMeters,
			Description:            request.Description,
			IdempotencyKey:         request.IdempotencyKey,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrEmergencyTypeRequired),
			errors.Is(err, services.ErrEmergencyTypeNotFound),
			errors.Is(err, services.ErrLatitudeRequired),
			errors.Is(err, services.ErrLongitudeRequired),
			errors.Is(err, services.ErrInvalidLatitude),
			errors.Is(err, services.ErrInvalidLongitude),
			errors.Is(err, services.ErrInvalidLocationAccuracy),
			errors.Is(err, services.ErrDescriptionTooLong),
			errors.Is(err, services.ErrIdempotencyKeyRequired),
			errors.Is(err, services.ErrInvalidIdempotencyKey):

			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})

		case errors.Is(err, services.ErrDuplicateIncident):
			writeJSON(w, http.StatusConflict, map[string]string{
				"error": err.Error(),
			})

		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to create emergency incident",
			})
		}

		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":  "emergency incident created",
		"incident": incident,
	})
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}
