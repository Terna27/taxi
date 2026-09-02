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
	ErrOrganizationTypeNotFound = errors.New("organization type not found or inactive")
	ErrCapabilityNotFound       = errors.New("one or more capabilities were not found or are inactive")
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
