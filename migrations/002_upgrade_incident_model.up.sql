BEGIN;

CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE emergency_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO emergency_types (code, name, description)
VALUES
    ('MEDICAL', 'Medical Emergency', 'Medical conditions requiring urgent response'),
    ('FIRE', 'Fire Emergency', 'Fire outbreaks and related hazards'),
    ('ROAD_ACCIDENT', 'Road Accident', 'Road crashes and transportation-related emergencies'),
    ('SECURITY', 'Security Emergency', 'Security threats, attacks, and violent incidents'),
    ('RESCUE', 'Rescue Emergency', 'Situations requiring rescue or extraction'),
    ('OTHER', 'Other Emergency', 'Emergency situations not covered by another category');

DROP TABLE IF EXISTS emergencies;

CREATE TABLE emergency_incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    emergency_type_id UUID NOT NULL
        REFERENCES emergency_types(id)
        ON DELETE RESTRICT,

    status VARCHAR(30) NOT NULL DEFAULT 'CREATED',

    reported_location GEOGRAPHY(POINT, 4326) NOT NULL,

    location_accuracy_meters DOUBLE PRECISION,

    description TEXT,

    idempotency_key UUID UNIQUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_incident_status
        CHECK (
            status IN (
                'CREATED',
                'SEARCHING',
                'DISPATCHED',
                'ACCEPTED',
                'EN_ROUTE',
                'ARRIVED',
                'RESOLVED',
                'CANCELLED',
                'FAILED',
                'EXPIRED'
            )
        ),

    CONSTRAINT chk_location_accuracy
        CHECK (
            location_accuracy_meters IS NULL
            OR location_accuracy_meters >= 0
        )
);

CREATE INDEX idx_emergency_incidents_status
ON emergency_incidents(status);

CREATE INDEX idx_emergency_incidents_type
ON emergency_incidents(emergency_type_id);

CREATE INDEX idx_emergency_incidents_created_at
ON emergency_incidents(created_at DESC);

CREATE INDEX idx_emergency_incidents_location
ON emergency_incidents
USING GIST (reported_location);

COMMIT;