BEGIN;

CREATE TABLE emergency_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    incident_id UUID NOT NULL
        REFERENCES emergency_incidents(id)
        ON DELETE CASCADE,

    organization_id UUID NOT NULL
        REFERENCES response_organizations(id)
        ON DELETE RESTRICT,

    capability_id UUID NOT NULL
        REFERENCES response_capabilities(id)
        ON DELETE RESTRICT,

    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',

    straight_line_distance_meters DOUBLE PRECISION,

    route_distance_meters DOUBLE PRECISION,

    estimated_travel_seconds INTEGER,

    offered_at TIMESTAMPTZ,

    accepted_at TIMESTAMPTZ,

    declined_at TIMESTAMPTZ,

    en_route_at TIMESTAMPTZ,

    arrived_at TIMESTAMPTZ,

    completed_at TIMESTAMPTZ,

    cancelled_at TIMESTAMPTZ,

    expires_at TIMESTAMPTZ,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_emergency_assignment_status
        CHECK (
            status IN (
                'PENDING',
                'OFFERED',
                'ACCEPTED',
                'DECLINED',
                'EXPIRED',
                'CANCELLED',
                'EN_ROUTE',
                'ARRIVED',
                'COMPLETED'
            )
        ),

    CONSTRAINT chk_assignment_straight_line_distance
        CHECK (
            straight_line_distance_meters IS NULL
            OR straight_line_distance_meters >= 0
        ),

    CONSTRAINT chk_assignment_route_distance
        CHECK (
            route_distance_meters IS NULL
            OR route_distance_meters >= 0
        ),

    CONSTRAINT chk_assignment_travel_seconds
        CHECK (
            estimated_travel_seconds IS NULL
            OR estimated_travel_seconds >= 0
        )
);

CREATE INDEX idx_emergency_assignments_incident
ON emergency_assignments(incident_id);

CREATE INDEX idx_emergency_assignments_organization
ON emergency_assignments(organization_id);

CREATE INDEX idx_emergency_assignments_capability
ON emergency_assignments(capability_id);

CREATE INDEX idx_emergency_assignments_status
ON emergency_assignments(status);

CREATE INDEX idx_emergency_assignments_incident_status
ON emergency_assignments(
    incident_id,
    status
);

CREATE INDEX idx_emergency_assignments_organization_status
ON emergency_assignments(
    organization_id,
    status
);

COMMIT;
