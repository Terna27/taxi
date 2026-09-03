package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"taxi/internal/services"
)

const maxResponseOrganizationRequestBodyBytes = 64 * 1024

type ResponseOrganizationHandler struct {
	service *services.ResponseOrganizationService
}

type createResponseOrganizationRequest struct {
	Name             string   `json:"name"`
	OrganizationType string   `json:"organization_type"`
	Latitude         *float64 `json:"latitude"`
	Longitude        *float64 `json:"longitude"`
	Address          *string  `json:"address"`
	Phone            *string  `json:"phone"`
	Email            *string  `json:"email"`
	Capabilities     []string `json:"capabilities"`
}

func NewResponseOrganizationHandler(
	service *services.ResponseOrganizationService,
) *ResponseOrganizationHandler {
	return &ResponseOrganizationHandler{
		service: service,
	}
}

func (h *ResponseOrganizationHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		maxResponseOrganizationRequestBodyBytes,
	)

	defer r.Body.Close()

	var request createResponseOrganizationRequest

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

	organization, err := h.service.Create(
		r.Context(),
		services.CreateResponseOrganizationInput{
			Name:             request.Name,
			OrganizationType: request.OrganizationType,
			Latitude:         request.Latitude,
			Longitude:        request.Longitude,
			Address:          request.Address,
			Phone:            request.Phone,
			Email:            request.Email,
			Capabilities:     request.Capabilities,
		},
	)

	if err != nil {
		switch {
		case errors.Is(err, services.ErrOrganizationNameRequired),
			errors.Is(err, services.ErrOrganizationNameTooLong),
			errors.Is(err, services.ErrOrganizationTypeRequired),
			errors.Is(err, services.ErrOrganizationTypeNotFound),
			errors.Is(err, services.ErrOrganizationLatitudeRequired),
			errors.Is(err, services.ErrOrganizationLongitudeRequired),
			errors.Is(err, services.ErrOrganizationLatitudeInvalid),
			errors.Is(err, services.ErrOrganizationLongitudeInvalid),
			errors.Is(err, services.ErrCapabilitiesRequired),
			errors.Is(err, services.ErrTooManyCapabilities),
			errors.Is(err, services.ErrCapabilityNotFound),
			errors.Is(err, services.ErrDuplicateCapability),
			errors.Is(err, services.ErrOrganizationAddressTooLong),
			errors.Is(err, services.ErrOrganizationPhoneTooLong),
			errors.Is(err, services.ErrOrganizationEmailTooLong),
			errors.Is(err, services.ErrOrganizationEmailInvalid):

			writeJSON(w, http.StatusBadRequest, map[string]string{
				"error": err.Error(),
			})

		default:
			writeJSON(w, http.StatusInternalServerError, map[string]string{
				"error": "failed to create response organization",
			})
		}

		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message":      "response organization created",
		"organization": organization,
	})
}
