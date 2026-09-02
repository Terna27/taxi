BEGIN;

CREATE TABLE response_organization_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO response_organization_types (
    code,
    name,
    description
)
VALUES
    (
        'HOSPITAL',
        'Hospital',
        'Hospitals and medical facilities providing emergency healthcare'
    ),
    (
        'CLINIC',
        'Clinic',
        'Clinics and smaller medical facilities capable of emergency response'
    ),
    (
        'AMBULANCE_SERVICE',
        'Ambulance Service',
        'Organizations primarily providing ambulance and pre-hospital response'
    ),
    (
        'FIRE_SERVICE',
        'Fire Service',
        'Organizations providing fire suppression and related emergency response'
    ),
    (
        'SECURITY_AGENCY',
        'Security Agency',
        'Organizations providing security and emergency protection services'
    ),
    (
        'RESCUE_SERVICE',
        'Rescue Service',
        'Organizations providing rescue and extraction services'
    ),
    (
        'MULTI_SERVICE',
        'Multi-Service Emergency Organization',
        'Organizations capable of providing multiple emergency response services'
    ),
    (
        'OTHER',
        'Other',
        'Other verified emergency response organizations'
    );

CREATE TABLE response_capabilities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,

    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO response_capabilities (
    code,
    name,
    description
)
VALUES
    (
        'MEDICAL',
        'Medical Response',
        'Emergency medical assessment and treatment'
    ),
    (
        'AMBULANCE',
        'Ambulance Response',
        'Emergency patient transport and pre-hospital response'
    ),
    (
        'FIRE',
        'Fire Response',
        'Fire suppression and fire-related emergency response'
    ),
    (
        'SECURITY',
        'Security Response',
        'Security intervention and protection response'
    ),
    (
        'RESCUE',
        'Rescue Response',
        'Rescue, extraction, and recovery operations'
    );

CREATE TABLE response_organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    organization_type_id UUID NOT NULL
        REFERENCES response_organization_types(id)
        ON DELETE RESTRICT,

    name VARCHAR(200) NOT NULL,

    location GEOGRAPHY(POINT, 4326) NOT NULL,

    address TEXT,

    phone VARCHAR(50),

    email VARCHAR(255),

    onboarding_source VARCHAR(30) NOT NULL DEFAULT 'ADMIN',

    verification_status VARCHAR(30) NOT NULL DEFAULT 'PENDING',

    is_active BOOLEAN NOT NULL DEFAULT FALSE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_response_organization_onboarding_source
        CHECK (
            onboarding_source IN (
                'ADMIN',
                'SELF_SERVICE',
                'PARTNER_IMPORT'
            )
        ),

    CONSTRAINT chk_response_organization_verification_status
        CHECK (
            verification_status IN (
                'PENDING',
                'VERIFIED',
                'REJECTED'
            )
        ),

    CONSTRAINT chk_response_organization_email
        CHECK (
            email IS NULL
            OR POSITION('@' IN email) > 1
        )
);

CREATE TABLE organization_capabilities (
    organization_id UUID NOT NULL
        REFERENCES response_organizations(id)
        ON DELETE CASCADE,

    capability_id UUID NOT NULL
        REFERENCES response_capabilities(id)
        ON DELETE RESTRICT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (
        organization_id,
        capability_id
    )
);

CREATE INDEX idx_response_organizations_type
ON response_organizations(organization_type_id);

CREATE INDEX idx_response_organizations_verification
ON response_organizations(verification_status);

CREATE INDEX idx_response_organizations_active
ON response_organizations(is_active);

CREATE INDEX idx_response_organizations_location
ON response_organizations
USING GIST (location);

CREATE INDEX idx_organization_capabilities_capability
ON organization_capabilities(capability_id);

COMMIT;