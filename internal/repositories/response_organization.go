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
	ErrOrganizationTypeNotFound       = errors.New("organization type not found or inactive")
	ErrCapabilityNotFound             = errors.New("one or more capabilities were not found or are inactive")
	ErrResponseOrganizationNotFound   = errors.New("response organization not found")
	ErrOrganizationNotVerified        = errors.New("response organization must be verified before activation")
)

type ResponseOrganizationRepository struct {
	db *pgxpool.Pool
}

func NewResponseOrganizationRepository(
	db *pgxpool.Pool,
) *ResponseOrganizationRepository {
	return &ResponseOrganizationRepository{
		db: db,
	}
}

func (r *ResponseOrganizationRepository) Create(
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
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin response organization transaction: %w", err)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var organizationTypeID string

	err = tx.QueryRow(
		ctx,
		`
            SELECT id::text
            FROM response_organization_types
            WHERE code = $1
              AND is_active = TRUE
        `,
		organizationType,
	).Scan(&organizationTypeID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrganizationTypeNotFound
		}

		return nil, fmt.Errorf("find organization type: %w", err)
	}

	capabilityRows, err := tx.Query(
		ctx,
		`
            SELECT
                id::text,
                code
            FROM response_capabilities
            WHERE code = ANY($1::text[])
              AND is_active = TRUE
        `,
		capabilities,
	)
	if err != nil {
		return nil, fmt.Errorf("find response capabilities: %w", err)
	}

	capabilityIDs := make(map[string]string)

	for capabilityRows.Next() {
		var id string
		var code string

		if err := capabilityRows.Scan(&id, &code); err != nil {
			capabilityRows.Close()
			return nil, fmt.Errorf("scan response capability: %w", err)
		}

		capabilityIDs[code] = id
	}

	capabilityRows.Close()

	if err := capabilityRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate response capabilities: %w", err)
	}

	if len(capabilityIDs) != len(capabilities) {
		return nil, ErrCapabilityNotFound
	}

	var organization models.ResponseOrganization

	err = tx.QueryRow(
		ctx,
		`
            INSERT INTO response_organizations (
                organization_type_id,
                name,
                location,
                address,
                phone,
                email,
                onboarding_source
            )
            VALUES (
                $1::uuid,
                $2,
                ST_SetSRID(
                    ST_MakePoint($4, $3),
                    4326
                )::geography,
                $5,
                $6,
                $7,
                $8
            )
            RETURNING
                id::text,
                name,
                ST_Y(location::geometry),
                ST_X(location::geometry),
                address,
                phone,
                email,
                onboarding_source,
                verification_status,
                is_active,
                created_at,
                updated_at
        `,
		organizationTypeID,
		name,
		latitude,
		longitude,
		address,
		phone,
		email,
		onboardingSource,
	).Scan(
		&organization.ID,
		&organization.Name,
		&organization.Latitude,
		&organization.Longitude,
		&organization.Address,
		&organization.Phone,
		&organization.Email,
		&organization.OnboardingSource,
		&organization.VerificationStatus,
		&organization.IsActive,
		&organization.CreatedAt,
		&organization.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("create response organization: %w", err)
	}

	for _, capability := range capabilities {
		capabilityID := capabilityIDs[capability]

		_, err := tx.Exec(
			ctx,
			`
                INSERT INTO organization_capabilities (
                    organization_id,
                    capability_id
                )
                VALUES ($1::uuid, $2::uuid)
            `,
			organization.ID,
			capabilityID,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"assign capability %s: %w",
				capability,
				err,
			)
		}
	}

	organization.OrganizationType = organizationType
	organization.Capabilities = capabilities

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit response organization transaction: %w", err)
	}

	return &organization, nil
}


func (r *ResponseOrganizationRepository) UpdateVerificationStatus(
	ctx context.Context,
	organizationID string,
	status string,
) (*models.ResponseOrganizationStatus, error) {
	var organizationStatus models.ResponseOrganizationStatus

	err := r.db.QueryRow(
		ctx,
		`
			UPDATE response_organizations
			SET
				verification_status = $2,
				is_active = CASE
					WHEN $2 = 'VERIFIED' THEN is_active
					ELSE FALSE
				END,
				updated_at = NOW()
			WHERE id = $1::uuid
			RETURNING
				id::text,
				verification_status,
				is_active,
				updated_at
		`,
		organizationID,
		status,
	).Scan(
		&organizationStatus.ID,
		&organizationStatus.VerificationStatus,
		&organizationStatus.IsActive,
		&organizationStatus.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResponseOrganizationNotFound
		}

		return nil, fmt.Errorf(
			"update response organization verification status: %w",
			err,
		)
	}

	return &organizationStatus, nil
}

func (r *ResponseOrganizationRepository) UpdateActiveStatus(
	ctx context.Context,
	organizationID string,
	isActive bool,
) (*models.ResponseOrganizationStatus, error) {
	var organizationStatus models.ResponseOrganizationStatus

	if isActive {
		err := r.db.QueryRow(
			ctx,
			`
				UPDATE response_organizations
				SET
					is_active = TRUE,
					updated_at = NOW()
				WHERE id = $1::uuid
				  AND verification_status = 'VERIFIED'
				RETURNING
					id::text,
					verification_status,
					is_active,
					updated_at
			`,
			organizationID,
		).Scan(
			&organizationStatus.ID,
			&organizationStatus.VerificationStatus,
			&organizationStatus.IsActive,
			&organizationStatus.UpdatedAt,
		)

		if err == nil {
			return &organizationStatus, nil
		}

		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf(
				"activate response organization: %w",
				err,
			)
		}

		var verificationStatus string

		err = r.db.QueryRow(
			ctx,
			`
				SELECT verification_status
				FROM response_organizations
				WHERE id = $1::uuid
			`,
			organizationID,
		).Scan(&verificationStatus)

		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrResponseOrganizationNotFound
			}

			return nil, fmt.Errorf(
				"check response organization verification status: %w",
				err,
			)
		}

		return nil, ErrOrganizationNotVerified
	}

	err := r.db.QueryRow(
		ctx,
		`
			UPDATE response_organizations
			SET
				is_active = FALSE,
				updated_at = NOW()
			WHERE id = $1::uuid
			RETURNING
				id::text,
				verification_status,
				is_active,
				updated_at
		`,
		organizationID,
	).Scan(
		&organizationStatus.ID,
		&organizationStatus.VerificationStatus,
		&organizationStatus.IsActive,
		&organizationStatus.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrResponseOrganizationNotFound
		}

		return nil, fmt.Errorf(
			"deactivate response organization: %w",
			err,
		)
	}

	return &organizationStatus, nil
}

func (r *ResponseOrganizationRepository) FindNearby(
	ctx context.Context,
	latitude float64,
	longitude float64,
	capability string,
	radiusMeters float64,
	limit int,
) ([]models.NearbyResponseOrganization, error) {
	rows, err := r.db.Query(
		ctx,
		`
			SELECT
				ro.id::text,
				ro.name,
				rot.code,
				ST_Y(ro.location::geometry) AS latitude,
				ST_X(ro.location::geometry) AS longitude,
				ro.address,
				ro.phone,

				ST_Distance(
					ro.location,
					ST_SetSRID(
						ST_MakePoint($2, $1),
						4326
					)::geography
				) AS distance_meters,

				COALESCE(
					array_agg(
						DISTINCT all_rc.code
						ORDER BY all_rc.code
					) FILTER (
						WHERE all_rc.code IS NOT NULL
					),
					ARRAY[]::varchar[]
				) AS capabilities

			FROM response_organizations ro

			JOIN response_organization_types rot
				ON rot.id = ro.organization_type_id
				AND rot.is_active = TRUE

			JOIN organization_capabilities required_oc
				ON required_oc.organization_id = ro.id

			JOIN response_capabilities required_rc
				ON required_rc.id = required_oc.capability_id
				AND required_rc.code = $3
				AND required_rc.is_active = TRUE

			LEFT JOIN organization_capabilities all_oc
				ON all_oc.organization_id = ro.id

			LEFT JOIN response_capabilities all_rc
				ON all_rc.id = all_oc.capability_id
				AND all_rc.is_active = TRUE

			WHERE ro.verification_status = 'VERIFIED'
			  AND ro.is_active = TRUE

			  AND ST_DWithin(
					ro.location,
					ST_SetSRID(
						ST_MakePoint($2, $1),
						4326
					)::geography,
					$4
			  )

			GROUP BY
				ro.id,
				ro.name,
				rot.code,
				ro.location,
				ro.address,
				ro.phone

			ORDER BY distance_meters ASC

			LIMIT $5
		`,
		latitude,
		longitude,
		capability,
		radiusMeters,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf(
			"find nearby response organizations: %w",
			err,
		)
	}
	defer rows.Close()

	organizations := make(
		[]models.NearbyResponseOrganization,
		0,
	)

	for rows.Next() {
		var organization models.NearbyResponseOrganization

		if err := rows.Scan(
			&organization.ID,
			&organization.Name,
			&organization.OrganizationType,
			&organization.Latitude,
			&organization.Longitude,
			&organization.Address,
			&organization.Phone,
			&organization.DistanceMeters,
			&organization.Capabilities,
		); err != nil {
			return nil, fmt.Errorf(
				"scan nearby response organization: %w",
				err,
			)
		}

		organizations = append(
			organizations,
			organization,
		)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate nearby response organizations: %w",
			err,
		)
	}

	return organizations, nil
}