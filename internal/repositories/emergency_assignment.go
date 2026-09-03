package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"taxi/internal/models"
)

var (
	ErrAssignmentIncidentNotFound       = errors.New("assignment incident not found")
	ErrAssignmentOrganizationNotFound   = errors.New("assignment organization not found")
	ErrAssignmentCapabilityNotFound     = errors.New("assignment capability not found")
	ErrAssignmentOrganizationIneligible = errors.New("assignment organization is not eligible")
	ErrActiveAssignmentExists           = errors.New("active assignment already exists")
)

type EmergencyAssignmentRepository struct {
	db *pgxpool.Pool
}

func NewEmergencyAssignmentRepository(
	db *pgxpool.Pool,
) *EmergencyAssignmentRepository {
	return &EmergencyAssignmentRepository{
		db: db,
	}
}

func (r *EmergencyAssignmentRepository) Create(
	ctx context.Context,
	incidentID string,
	organizationID string,
	capability string,
	straightLineDistanceMeters *float64,
) (*models.EmergencyAssignment, error) {
	var assignment models.EmergencyAssignment

	err := r.db.QueryRow(
		ctx,
		`
			INSERT INTO emergency_assignments (
				incident_id,
				organization_id,
				capability_id,
				status,
				straight_line_distance_meters
			)
			SELECT
				ei.id,
				ro.id,
				rc.id,
				'PENDING',
				$4
			FROM emergency_incidents ei

			JOIN response_organizations ro
				ON ro.id = $2::uuid
				AND ro.verification_status = 'VERIFIED'
				AND ro.is_active = TRUE

			JOIN response_capabilities rc
				ON rc.code = $3
				AND rc.is_active = TRUE

			JOIN organization_capabilities oc
				ON oc.organization_id = ro.id
				AND oc.capability_id = rc.id

			WHERE ei.id = $1::uuid

			RETURNING
				id::text,
				incident_id::text,
				organization_id::text,
				(
					SELECT code
					FROM response_capabilities
					WHERE id = capability_id
				),
				status,
				straight_line_distance_meters,
				route_distance_meters,
				estimated_travel_seconds,
				offered_at,
				accepted_at,
				declined_at,
				en_route_at,
				arrived_at,
				completed_at,
				cancelled_at,
				expires_at,
				created_at,
				updated_at
		`,
		incidentID,
		organizationID,
		capability,
		straightLineDistanceMeters,
	).Scan(
		&assignment.ID,
		&assignment.IncidentID,
		&assignment.OrganizationID,
		&assignment.Capability,
		&assignment.Status,
		&assignment.StraightLineDistanceMeters,
		&assignment.RouteDistanceMeters,
		&assignment.EstimatedTravelSeconds,
		&assignment.OfferedAt,
		&assignment.AcceptedAt,
		&assignment.DeclinedAt,
		&assignment.EnRouteAt,
		&assignment.ArrivedAt,
		&assignment.CompletedAt,
		&assignment.CancelledAt,
		&assignment.ExpiresAt,
		&assignment.CreatedAt,
		&assignment.UpdatedAt,
	)

	if err == nil {
		return &assignment, nil
	}

	var pgErr *pgconn.PgError

	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" &&
			pgErr.ConstraintName ==
				"uq_emergency_assignments_active_incident_capability" {
			return nil, ErrActiveAssignmentExists
		}
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf(
			"create emergency assignment: %w",
			err,
		)
	}

	var incidentExists bool

	err = r.db.QueryRow(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM emergency_incidents
				WHERE id = $1::uuid
			)
		`,
		incidentID,
	).Scan(&incidentExists)

	if err != nil {
		return nil, fmt.Errorf(
			"check assignment incident: %w",
			err,
		)
	}

	if !incidentExists {
		return nil, ErrAssignmentIncidentNotFound
	}

	var organizationExists bool

	err = r.db.QueryRow(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM response_organizations
				WHERE id = $1::uuid
			)
		`,
		organizationID,
	).Scan(&organizationExists)

	if err != nil {
		return nil, fmt.Errorf(
			"check assignment organization: %w",
			err,
		)
	}

	if !organizationExists {
		return nil, ErrAssignmentOrganizationNotFound
	}

	var capabilityExists bool

	err = r.db.QueryRow(
		ctx,
		`
			SELECT EXISTS (
				SELECT 1
				FROM response_capabilities
				WHERE code = $1
				  AND is_active = TRUE
			)
		`,
		capability,
	).Scan(&capabilityExists)

	if err != nil {
		return nil, fmt.Errorf(
			"check assignment capability: %w",
			err,
		)
	}

	if !capabilityExists {
		return nil, ErrAssignmentCapabilityNotFound
	}

	return nil, ErrAssignmentOrganizationIneligible
}
