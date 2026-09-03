package models

type IncidentCapabilityRequirement struct {
	Capability       string `json:"capability"`
	RequirementLevel string `json:"requirement_level"`
}

type IncidentCandidateGroup struct {
	Capability       string                       `json:"capability"`
	RequirementLevel string                       `json:"requirement_level"`
	Organizations    []NearbyResponseOrganization `json:"organizations"`
}

type IncidentCandidateDiscovery struct {
	IncidentID      string                   `json:"incident_id"`
	EmergencyType   string                   `json:"emergency_type"`
	Latitude        float64                  `json:"latitude"`
	Longitude       float64                  `json:"longitude"`
	CandidateGroups []IncidentCandidateGroup `json:"candidate_groups"`
}
