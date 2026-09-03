BEGIN;

CREATE UNIQUE INDEX uq_emergency_assignments_active_incident_capability
ON emergency_assignments (
    incident_id,
    capability_id
)
WHERE status IN (
    'PENDING',
    'OFFERED',
    'ACCEPTED',
    'EN_ROUTE',
    'ARRIVED'
);

COMMIT;