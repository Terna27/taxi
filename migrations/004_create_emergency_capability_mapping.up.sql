BEGIN;

CREATE TABLE emergency_type_capabilities (
    emergency_type_id UUID NOT NULL
        REFERENCES emergency_types(id)
        ON DELETE CASCADE,

    capability_id UUID NOT NULL
        REFERENCES response_capabilities(id)
        ON DELETE RESTRICT,

    requirement_level VARCHAR(20) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (
        emergency_type_id,
        capability_id
    ),

    CONSTRAINT chk_emergency_capability_requirement_level
        CHECK (
            requirement_level IN (
                'PRIMARY',
                'SUPPORTING'
            )
        )
);

CREATE INDEX idx_emergency_type_capabilities_capability
ON emergency_type_capabilities(capability_id);

INSERT INTO emergency_type_capabilities (
    emergency_type_id,
    capability_id,
    requirement_level
)
SELECT
    et.id,
    rc.id,
    mapping.requirement_level
FROM (
    VALUES
        ('MEDICAL',       'MEDICAL',   'PRIMARY'),
        ('MEDICAL',       'AMBULANCE', 'SUPPORTING'),

        ('FIRE',          'FIRE',      'PRIMARY'),
        ('FIRE',          'MEDICAL',   'SUPPORTING'),
        ('FIRE',          'AMBULANCE', 'SUPPORTING'),

        ('ROAD_ACCIDENT', 'AMBULANCE', 'PRIMARY'),
        ('ROAD_ACCIDENT', 'MEDICAL',   'SUPPORTING'),
        ('ROAD_ACCIDENT', 'RESCUE',    'SUPPORTING'),

        ('SECURITY',      'SECURITY',  'PRIMARY'),

        ('RESCUE',        'RESCUE',    'PRIMARY'),
        ('RESCUE',        'MEDICAL',   'SUPPORTING'),
        ('RESCUE',        'AMBULANCE', 'SUPPORTING')
) AS mapping(
    emergency_type_code,
    capability_code,
    requirement_level
)
JOIN emergency_types et
    ON et.code = mapping.emergency_type_code
JOIN response_capabilities rc
    ON rc.code = mapping.capability_code;

COMMIT;