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
	ErrEmergencyTypeNotFound   = errors.New("emergency type not found or inactive")
	ErrDuplicateIdempotencyKey = errors.New("incident with this idempotency key already exists")
)

type EmergencyRepository struct {
	db *pgxpool.Pool
}

func NewEmergencyRepository(db *pgxpool.Pool) *EmergencyRepository {
	return &EmergencyRepository{
		db: db,
	}
}

func (r *EmergencyRepository) Create(
	ctx context.Context,
	emergencyType string,
	latitude float64,
	longitude float64,
	locationAccuracyMeters *float64,
	description *string,
	idempotencyKey string,
) (*models.EmergencyIncident, error) {
	query := `
        WITH selected_type AS (
            SELECT id
            FROM emergency_types
            WHERE code = $1
              AND is_active = TRUE
        ),
        inserted_incident AS (
            INSERT INTO emergency_incidents (
                emergency_type_id,
                reported_location,
                location_accuracy_meters,
                description,
                idempotency_key
            )
            SELECT
                selected_type.id,
                ST_SetSRID(
                    ST_MakePoint($3, $2),
                    4326
                )::geography,
                $4,
                $5,
                $6::uuid
            FROM selected_type
            RETURNING
                id,
                emergency_type_id,
                status,
                reported_location,
                location_accuracy_meters,
                description,
                idempotency_key,
                created_at,
                updated_at
        )
        SELECT
            incident.id::text,
            emergency_type.code,
            incident.status,
            ST_Y(incident.reported_location::geometry),
            ST_X(incident.reported_location::geometry),
            incident.location_accuracy_meters,
            incident.description,
            incident.idempotency_key::text,
            incident.created_at,
            incident.updated_at
        FROM inserted_incident AS incident
        JOIN emergency_types AS emergency_type
            ON emergency_type.id = incident.emergency_type_id
    `

	var incident models.EmergencyIncident

	err := r.db.QueryRow(
		ctx,
		query,
		emergencyType,
		latitude,
		longitude,
		locationAccuracyMeters,
		description,
		idempotencyKey,
	).Scan(
		&incident.ID,
		&incident.EmergencyType,
		&incident.Status,
		&incident.Latitude,
		&incident.Longitude,
		&incident.LocationAccuracyMeters,
		&incident.Description,
		&incident.IdempotencyKey,
		&incident.CreatedAt,
		&incident.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrEmergencyTypeNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicateIdempotencyKey
		}

		return nil, fmt.Errorf("create emergency incident: %w", err)
	}

	return &incident, nil
}
