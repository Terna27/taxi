package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"taxi/internal/config"
	"taxi/internal/database"
	"taxi/internal/handlers"
	"taxi/internal/repositories"
	"taxi/internal/routes"
	"taxi/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx := context.Background()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer db.Close()

	emergencyRepository :=
		repositories.NewEmergencyRepository(db)

	emergencyService :=
		services.NewEmergencyService(
			emergencyRepository,
		)

	emergencyHandler :=
		handlers.NewEmergencyHandler(
			emergencyService,
		)

	responseOrganizationRepository :=
		repositories.NewResponseOrganizationRepository(db)

	responseOrganizationService :=
		services.NewResponseOrganizationService(
			responseOrganizationRepository,
		)

	responseOrganizationHandler :=
		handlers.NewResponseOrganizationHandler(
			responseOrganizationService,
		)

	organizationDiscoveryService :=
		services.NewOrganizationDiscoveryService(
			responseOrganizationRepository,
		)

	organizationDiscoveryHandler :=
		handlers.NewOrganizationDiscoveryHandler(
			organizationDiscoveryService,
		)

	incidentCandidateRepository :=
		repositories.NewIncidentCandidateRepository(db)

	incidentCandidateDiscoveryService :=
		services.NewIncidentCandidateDiscoveryService(
			incidentCandidateRepository,
			responseOrganizationRepository,
		)

	emergencyAssignmentRepository :=
		repositories.NewEmergencyAssignmentRepository(db)

	dispatchService :=
		services.NewDispatchService(
			incidentCandidateDiscoveryService,
			emergencyAssignmentRepository,
		)

	dispatchHandler :=
		handlers.NewDispatchHandler(
			dispatchService,
		)

	handler := routes.New(
		emergencyHandler,
		responseOrganizationHandler,
		organizationDiscoveryHandler,
		dispatchHandler,
	)

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf(
		"RideRoute server running on http://localhost:%s",
		cfg.AppPort,
	)

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatalf("server failed: %v", err)
	}
}
