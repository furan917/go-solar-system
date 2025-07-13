package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/furan917/go-solar-system/internal/models"
)

const (
	MaxResponseSize = 10 * 1024 * 1024
	MaxBodiesCount  = 10000
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: constants.DefaultTimeout,
		},
		baseURL: constants.SolarSystemAPIBase,
	}
}

func (c *Client) GetAllBodies() ([]models.CelestialBody, error) {
	targetUrl := fmt.Sprintf("%s/bodies", c.baseURL)

	resp, err := c.httpClient.Get(targetUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bodies: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body: %v\n", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	limitedReader := io.LimitReader(resp.Body, MaxResponseSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if err := validateAPIResponse(apiResponse); err != nil {
		return nil, fmt.Errorf("invalid API response: %w", err)
	}

	return apiResponse.Bodies, nil
}

func (c *Client) GetBody(id string) (*models.CelestialBody, error) {
	targetUrl := fmt.Sprintf("%s/bodies/%s", c.baseURL, url.QueryEscape(id))

	resp, err := c.httpClient.Get(targetUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch body %s: %w", id, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body for %s: %v\n", id, err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d for body %s", resp.StatusCode, id)
	}

	limitedReader := io.LimitReader(resp.Body, MaxResponseSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var celestialBody models.CelestialBody
	if err := json.Unmarshal(body, &celestialBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if err := validateCelestialBody(celestialBody); err != nil {
		return nil, fmt.Errorf("invalid celestial body data for %s: %w", id, err)
	}

	return &celestialBody, nil
}

func (c *Client) GetPlanets() ([]models.CelestialBody, error) {
	bodies, err := c.GetAllBodies()
	if err != nil {
		return nil, err
	}

	var planets []models.CelestialBody
	for _, body := range bodies {
		if body.IsPlanet {
			planets = append(planets, body)
		}
	}

	return planets, nil
}

func (c *Client) GetBodiesWithFilter(filter string) ([]models.CelestialBody, error) {
	targetUrl := fmt.Sprintf("%s/bodies?filter[]=%s", c.baseURL, url.QueryEscape(filter))

	resp, err := c.httpClient.Get(targetUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch filtered bodies: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("Error closing response body for filter %s: %v\n", filter, err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	limitedReader := io.LimitReader(resp.Body, MaxResponseSize)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if err := validateAPIResponse(apiResponse); err != nil {
		return nil, fmt.Errorf("invalid filtered API response: %w", err)
	}

	return apiResponse.Bodies, nil
}

// GetMoonData attempts to fetch detailed moon data from the API
func (c *Client) GetMoonData(moonID string) (*models.CelestialBody, error) {
	if moonID == "" {
		return nil, fmt.Errorf("moon ID is empty")
	}

	body, err := c.GetBody(moonID)
	if err != nil {
		return nil, err
	}

	if err := validateCelestialBody(*body); err != nil {
		return nil, fmt.Errorf("invalid celestial body data for %s: %w", moonID, err)
	}

	return body, nil
}

// validateAPIResponse validates the structure and content of API responses
func validateAPIResponse(response models.APIResponse) error {
	if len(response.Bodies) == 0 {
		return fmt.Errorf("API response contains no celestial bodies")
	}

	if len(response.Bodies) > MaxBodiesCount {
		return fmt.Errorf("API response contains too many celestial bodies: %d (max: %d)", len(response.Bodies), MaxBodiesCount)
	}

	for i, body := range response.Bodies {
		if err := validateCelestialBody(body); err != nil {
			return fmt.Errorf("invalid celestial body at index %d: %w", i, err)
		}
	}

	return nil
}

// validateCelestialBody validates individual celestial body data
func validateCelestialBody(body models.CelestialBody) error {
	if body.EnglishName == "" {
		return fmt.Errorf("celestial body missing English name")
	}

	if body.MeanRadius < 0 {
		return fmt.Errorf("celestial body %s has negative radius: %.2f", body.EnglishName, body.MeanRadius)
	}

	if body.SemimajorAxis < 0 {
		return fmt.Errorf("celestial body %s has negative semimajor axis: %.2f", body.EnglishName, body.SemimajorAxis)
	}

	if body.Density < 0 {
		return fmt.Errorf("celestial body %s has negative density: %.2f", body.EnglishName, body.Density)
	}

	if body.Gravity < 0 {
		return fmt.Errorf("celestial body %s has negative gravity: %.2f", body.EnglishName, body.Gravity)
	}

	if body.Eccentricity < 0 || body.Eccentricity > 1 {
		return fmt.Errorf("celestial body %s has unrealistic eccentricity: %.6f", body.EnglishName, body.Eccentricity)
	}

	return nil
}
