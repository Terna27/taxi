package routing

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"taxi/internal/models"
)

const googleRouteMatrixURL =
	"https://routes.googleapis.com/distanceMatrix/v2:computeRouteMatrix"

var (
	ErrGoogleAPIKeyMissing = errors.New("google routes api key is required")
	ErrNoRouteResults       = errors.New("no route estimates returned")
)

type GoogleService struct {
	apiKey     string
	httpClient *http.Client
}

func NewGoogleService(
	apiKey string,
) (*GoogleService, error) {
	apiKey = strings.TrimSpace(apiKey)

	if apiKey == "" {
		return nil, ErrGoogleAPIKeyMissing
	}

	return &GoogleService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 8 * time.Second,
		},
	}, nil
}

type googleLatLng struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type googleLocation struct {
	LatLng googleLatLng `json:"latLng"`
}

type googleWaypoint struct {
	Location googleLocation `json:"location"`
}

type googleRouteMatrixOrigin struct {
	Waypoint googleWaypoint `json:"waypoint"`
}

type googleRouteMatrixDestination struct {
	Waypoint googleWaypoint `json:"waypoint"`
}

type googleRouteMatrixRequest struct {
	Origins           []googleRouteMatrixOrigin      `json:"origins"`
	Destinations      []googleRouteMatrixDestination `json:"destinations"`
	TravelMode        string                         `json:"travelMode"`
	RoutingPreference string                         `json:"routingPreference"`
}

type googleRouteMatrixStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type googleRouteMatrixElement struct {
	OriginIndex      int                     `json:"originIndex"`
	DestinationIndex int                     `json:"destinationIndex"`
	Status           googleRouteMatrixStatus `json:"status"`
	Condition        string                  `json:"condition"`
	DistanceMeters   int                     `json:"distanceMeters"`
	Duration         string                  `json:"duration"`
}

func (s *GoogleService) EstimateRoutes(
	ctx context.Context,
	originLatitude float64,
	originLongitude float64,
	organizations []models.NearbyResponseOrganization,
) ([]RouteEstimate, error) {
	if len(organizations) == 0 {
		return nil, ErrNoRouteResults
	}

	requestPayload := googleRouteMatrixRequest{
		Origins: []googleRouteMatrixOrigin{
			{
				Waypoint: googleWaypoint{
					Location: googleLocation{
						LatLng: googleLatLng{
							Latitude:  originLatitude,
							Longitude: originLongitude,
						},
					},
				},
			},
		},
		Destinations: make(
			[]googleRouteMatrixDestination,
			0,
			len(organizations),
		),
		TravelMode:        "DRIVE",
		RoutingPreference: "TRAFFIC_AWARE",
	}

	for _, organization := range organizations {
		requestPayload.Destinations = append(
			requestPayload.Destinations,
			googleRouteMatrixDestination{
				Waypoint: googleWaypoint{
					Location: googleLocation{
						LatLng: googleLatLng{
							Latitude:  organization.Latitude,
							Longitude: organization.Longitude,
						},
					},
				},
			},
		)
	}

	body, err := json.Marshal(requestPayload)
	if err != nil {
		return nil, fmt.Errorf(
			"marshal google route matrix request: %w",
			err,
		)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		googleRouteMatrixURL,
		bytes.NewReader(body),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"create google route matrix request: %w",
			err,
		)
	}

	req.Header.Set(
		"Content-Type",
		"application/json",
	)

	req.Header.Set(
		"X-Goog-Api-Key",
		s.apiKey,
	)

	req.Header.Set(
		"X-Goog-FieldMask",
		"originIndex,destinationIndex,status,condition,distanceMeters,duration",
	)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"google route matrix request failed: %w",
			err,
		)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(
		io.LimitReader(resp.Body, 1024*1024),
	)
	if err != nil {
		return nil, fmt.Errorf(
			"read google route matrix response: %w",
			err,
		)
	}

	if resp.StatusCode < 200 ||
		resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"google routes returned status %d: %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	var elements []googleRouteMatrixElement

	if err := json.Unmarshal(
		responseBody,
		&elements,
	); err != nil {
		return nil, fmt.Errorf(
			"decode google route matrix response: %w",
			err,
		)
	}

	estimates := make(
		[]RouteEstimate,
		0,
		len(elements),
	)

	for _, element := range elements {
		if element.Status.Code != 0 {
			continue
		}

		if element.Condition != "" &&
			element.Condition != "ROUTE_EXISTS" {
			continue
		}

		if element.DestinationIndex < 0 ||
			element.DestinationIndex >= len(organizations) {
			continue
		}

		durationSeconds, err :=
			parseGoogleDurationSeconds(
				element.Duration,
			)

		if err != nil {
			continue
		}

		organization :=
			organizations[element.DestinationIndex]

		estimates = append(
			estimates,
			RouteEstimate{
				OrganizationID: organization.ID,
				DistanceMeters: element.DistanceMeters,
				TravelTimeSeconds: durationSeconds,
			},
		)
	}

	if len(estimates) == 0 {
		return nil, ErrNoRouteResults
	}

	return estimates, nil
}

func parseGoogleDurationSeconds(
	value string,
) (int, error) {
	value = strings.TrimSpace(value)

	if !strings.HasSuffix(value, "s") {
		return 0, fmt.Errorf(
			"invalid google duration: %s",
			value,
		)
	}

	value = strings.TrimSuffix(value, "s")

	seconds, err := strconv.ParseFloat(
		value,
		64,
	)
	if err != nil {
		return 0, fmt.Errorf(
			"parse google duration: %w",
			err,
		)
	}

	return int(seconds + 0.5), nil
}