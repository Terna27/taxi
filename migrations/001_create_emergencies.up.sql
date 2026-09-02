CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS emergencies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,

    status VARCHAR(30) NOT NULL DEFAULT 'requested',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_emergencies_status
ON emergencies(status);

CREATE INDEX IF NOT EXISTS idx_emergencies_created_at
ON emergencies(created_at DESC);