package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taxi/internal/models"
)

var (
	ErrIncidentNotFound = errors.New("incident not found")
)

type IncidentCandidateRepository struct {
	db *pgxpool.Pool
}

func NewIncidentCandidateRepository(
	db *pgxpool.Pool,
) *IncidentCandidateRepository {
	return &IncidentCandidateRepository{
		db: db,
	}
}

func (r *IncidentCandidateRepository) GetIncidentDiscoveryContext(
	ctx context.Context,
	incidentID string,
) (
	string,
	float64,
	float64,
	error,
) {
	var (
		emergencyType string
		latitude      float64
		longitude     float64
	)

	err := r.db.QueryRow(
		ctx,
		`
			SELECT
				et.code,
				ST_Y(ei.reported_location::geometry) AS latitude,
				ST_X(ei.reported_location::geometry) AS longitude
			FROM emergency_incidents ei
			JOIN emergency_types et
				ON et.id = ei.emergency_type_id
			WHERE ei.id = $1::uuid
		`,
		incidentID,
	).Scan(
		&emergencyType,
		&latitude,
		&longitude,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, 0, ErrIncidentNotFound
		}

		return "", 0, 0, fmt.Errorf(
			"get incident discovery context: %w",
			err,
		)
	}

	return emergencyType, latitude, longitude, nil
}

func (r *IncidentCandidateRepository) GetCapabilityRequirements(
	ctx context.Context,
	emergencyType string,
) ([]models.IncidentCapabilityRequirement, error) {
	rows, err := r.db.Query(
		ctx,
		`
			SELECT
				rc.code,
				etc.requirement_level
			FROM emergency_type_capabilities etc
			JOIN emergency_types et
				ON et.id = etc.emergency_type_id
			JOIN response_capabilities rc
				ON rc.id = etc.capability_id
			WHERE et.code = $1
			  AND et.is_active = TRUE
			  AND rc.is_active = TRUE
			ORDER BY
				CASE etc.requirement_level
					WHEN 'PRIMARY' THEN 1
					ELSE 2
				END,
				rc.code
		`,
		emergencyType,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"get emergency capability requirements: %w",
			err,
		)
	}
	defer rows.Close()

	requirements := make(
		[]models.IncidentCapabilityRequirement,
		0,
	)

	for rows.Next() {
		var requirement models.IncidentCapabilityRequirement

		if err := rows.Scan(
			&requirement.Capability,
			&requirement.RequirementLevel,
		); err != nil {
			return nil, fmt.Errorf(
				"scan emergency capability requirement: %w",
				err,
			)
		}

		requirements = append(
			requirements,
			requirement,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate emergency capability requirements: %w",
			err,
		)
	}

	return requirements, nil
}