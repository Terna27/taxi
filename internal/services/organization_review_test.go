package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"taxi/internal/models"
	"taxi/internal/repositories"
)

type fakeOrganizationReviewRepository struct {
	status *models.ResponseOrganizationStatus
	err    error

	receivedOrganizationID string
	receivedStatus         string
	receivedActive         bool
}

func (f *fakeOrganizationReviewRepository) UpdateVerificationStatus(
	ctx context.Context,
	organizationID string,
	status string,
) (*models.ResponseOrganizationStatus, error) {
	f.receivedOrganizationID = organizationID
	f.receivedStatus = status

	if f.err != nil {
		return nil, f.err
	}

	return f.status, nil
}

func (f *fakeOrganizationReviewRepository) UpdateActiveStatus(
	ctx context.Context,
	organizationID string,
	isActive bool,
) (*models.ResponseOrganizationStatus, error) {
	f.receivedOrganizationID = organizationID
	f.receivedActive = isActive

	if f.err != nil {
		return nil, f.err
	}

	return f.status, nil
}

func TestOrganizationReviewServiceVerify(t *testing.T) {
	organizationID := "6717b804-32b4-43f6-b310-acee0250e053"

	repository := &fakeOrganizationReviewRepository{
		status: &models.ResponseOrganizationStatus{
			ID:                 organizationID,
			VerificationStatus: "VERIFIED",
			IsActive:           false,
			UpdatedAt:          time.Now(),
		},
	}

	service := NewOrganizationReviewService(repository)

	result, err := service.Review(
		context.Background(),
		organizationID,
		" verified ",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.VerificationStatus != "VERIFIED" {
		t.Fatalf(
			"expected VERIFIED, got %q",
			result.VerificationStatus,
		)
	}

	if repository.receivedStatus != "VERIFIED" {
		t.Fatalf(
			"expected repository status VERIFIED, got %q",
			repository.receivedStatus,
		)
	}
}

func TestOrganizationReviewServiceReject(t *testing.T) {
	organizationID := "6717b804-32b4-43f6-b310-acee0250e053"

	repository := &fakeOrganizationReviewRepository{
		status: &models.ResponseOrganizationStatus{
			ID:                 organizationID,
			VerificationStatus: "REJECTED",
			IsActive:           false,
			UpdatedAt:          time.Now(),
		},
	}

	service := NewOrganizationReviewService(repository)

	result, err := service.Review(
		context.Background(),
		organizationID,
		"REJECTED",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.VerificationStatus != "REJECTED" {
		t.Fatalf(
			"expected REJECTED, got %q",
			result.VerificationStatus,
		)
	}
}

func TestOrganizationReviewServiceInvalidVerificationStatus(
	t *testing.T,
) {
	organizationID := "6717b804-32b4-43f6-b310-acee0250e053"

	repository := &fakeOrganizationReviewRepository{}
	service := NewOrganizationReviewService(repository)

	_, err := service.Review(
		context.Background(),
		organizationID,
		"PENDING",
	)

	if !errors.Is(err, ErrVerificationStatusInvalid) {
		t.Fatalf(
			"expected %v, got %v",
			ErrVerificationStatusInvalid,
			err,
		)
	}
}

func TestOrganizationReviewServiceInvalidOrganizationID(
	t *testing.T,
) {
	repository := &fakeOrganizationReviewRepository{}
	service := NewOrganizationReviewService(repository)

	_, err := service.Review(
		context.Background(),
		"not-a-uuid",
		"VERIFIED",
	)

	if !errors.Is(err, ErrOrganizationIDInvalid) {
		t.Fatalf(
			"expected %v, got %v",
			ErrOrganizationIDInvalid,
			err,
		)
	}
}

func TestOrganizationReviewServiceOrganizationNotFound(
	t *testing.T,
) {
	organizationID := "6717b804-32b4-43f6-b310-acee0250e053"

	repository := &fakeOrganizationReviewRepository{
		err: repositories.ErrResponseOrganizationNotFound,
	}

	service := NewOrganizationReviewService(repository)

	_, err := service.Review(
		context.Background(),
		organizationID,
		"VERIFIED",
	)

	if !errors.Is(err, ErrResponseOrganizationMissing) {
		t.Fatalf(
			"expected %v, got %v",
			ErrResponseOrganizationMissing,
			err,
		)
	}
}

func TestOrganizationReviewServiceActivate(t *testing.T) {
	organizationID := "6717b804-32b4-43f6-b310-acee0250e053"

	repository := &fakeOrganizationReviewRepository{
		status: &models.ResponseOrganizationStatus{
			ID:                 organizationID,
			VerificationStatus: "VERIFIED",
			IsActive:           true,
			UpdatedAt:          time.Now(),
		},
	}

	service := NewOrganizationReviewService(repository)

	result, err := service.SetActive(
		context.Background(),
		organizationID,
		true,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !result.IsActive {
		t.Fatal("expected organization to be active")
	}

	if !repository.receivedActive {
		t.Fatal("expected repository to receive active=true")
	}
}

func TestOrganizationReviewServiceCannotActivateUnverified(
	t *testing.T,
) {
	organizationID := "6717b804-32b4-43f6-b310-acee0250e053"

	repository := &fakeOrganizationReviewRepository{
		err: repositories.ErrOrganizationNotVerified,
	}

	service := NewOrganizationReviewService(repository)

	_, err := service.SetActive(
		context.Background(),
		organizationID,
		true,
	)

	if !errors.Is(err, ErrOrganizationMustBeVerified) {
		t.Fatalf(
			"expected %v, got %v",
			ErrOrganizationMustBeVerified,
			err,
		)
	}
}

func TestOrganizationReviewServiceDeactivate(t *testing.T) {
	organizationID := "6717b804-32b4-43f6-b310-acee0250e053"

	repository := &fakeOrganizationReviewRepository{
		status: &models.ResponseOrganizationStatus{
			ID:                 organizationID,
			VerificationStatus: "VERIFIED",
			IsActive:           false,
			UpdatedAt:          time.Now(),
		},
	}

	service := NewOrganizationReviewService(repository)

	result, err := service.SetActive(
		context.Background(),
		organizationID,
		false,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.IsActive {
		t.Fatal("expected organization to be inactive")
	}
}