package handlers

import (
	"errors"
	"net/http"

	"taxi/internal/services"
)

type DispatchHandler struct {
	service *services.DispatchService
}

func NewDispatchHandler(
	service *services.DispatchService,
) *DispatchHandler {
	return &DispatchHandler{
		service: service,
	}
}

func (h *DispatchHandler) Dispatch(
	w http.ResponseWriter,
	r *http.Request,
) {
	incidentID := r.PathValue("incidentID")

	assignment, err := h.service.Dispatch(
		r.Context(),
		incidentID,
	)

	if err != nil {
		switch {
		case errors.Is(
			err,
			services.ErrIncidentIDRequired,
		),
			errors.Is(
				err,
				services.ErrIncidentIDInvalid,
			):

			writeJSON(
				w,
				http.StatusBadRequest,
				map[string]string{
					"error": err.Error(),
				},
			)

		case errors.Is(
			err,
			services.ErrIncidentMissing,
		):

			writeJSON(
				w,
				http.StatusNotFound,
				map[string]string{
					"error": err.Error(),
				},
			)

		case errors.Is(
			err,
			services.ErrIncidentCapabilityMissing,
		),
			errors.Is(
				err,
				services.ErrNoPrimaryCapability,
			),
			errors.Is(
				err,
				services.ErrMultiplePrimaryCapabilities,
			):

			writeJSON(
				w,
				http.StatusUnprocessableEntity,
				map[string]string{
					"error": err.Error(),
				},
			)

		case errors.Is(
			err,
			services.ErrNoDispatchCandidate,
		),
			errors.Is(
				err,
				services.ErrDispatchCandidateUnavailable,
			):

			writeJSON(
				w,
				http.StatusConflict,
				map[string]string{
					"error": err.Error(),
				},
			)

		case errors.Is(
			err,
			services.ErrDispatchAlreadyActive,
		):

			writeJSON(
				w,
				http.StatusConflict,
				map[string]string{
					"error": err.Error(),
				},
			)

		default:
			writeJSON(
				w,
				http.StatusInternalServerError,
				map[string]string{
					"error": "failed to dispatch incident",
				},
			)
		}

		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		map[string]any{
			"assignment": assignment,
		},
	)
}
