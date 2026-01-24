package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/furan917/go-solar-system/internal/constants"
	"github.com/furan917/go-solar-system/internal/models"
)

const (
	MaxResponseSize = 10 * 1024 * 1024
	MaxBodiesCount  = 10000

	userAgent = "go-solar-system/1.0 (+https://github.com/furan917/go-solar-system)"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

func NewClient() *Client {
	// Use environment variable first; fallback to baked-in constant
	apiKey := ""
	if envKey := os.Getenv("SOLAR_SYSTEM_API_KEY"); envKey != "" {
		apiKey = envKey
	} else {
		// This is a light obscuring of a key used for ID only, its simply to avoid repo scrapers, nothing secure here
		key, err := constants.UnobscureKey(constants.SolarSystemAPIKeyObfuscated)
		if err == nil {
			apiKey = key
		}
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: constants.DefaultTimeout,
		},
		baseURL: constants.SolarSystemAPIBase,
		apiKey:  apiKey,
	}
}

/*
------------------------------------------------
Internal HTTP helper
------------------------------------------------
*/

func (c *Client) doGET(path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")

	// Include API key as Bearer token
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("failed to close response body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	reader := io.LimitReader(resp.Body, MaxResponseSize)
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

/*
------------------------------------------------
Public API methods
------------------------------------------------
*/

func (c *Client) GetAllBodies() ([]models.CelestialBody, error) {
	body, err := c.doGET("/bodies")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch bodies: %w", err)
	}

	// Sanitize the response: replace literal newlines in JSON string values with spaces
	// This handles cases where the API returns unescaped newlines (e.g., in discoveredBy fields)
	body = sanitizeJSONResponse(body)

	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bodies response: %w", err)
	}

	if err := validateAPIResponse(apiResponse); err != nil {
		return nil, fmt.Errorf("invalid API response: %w", err)
	}

	// Normalize and filter bodies
	validBodies := make([]models.CelestialBody, 0, len(apiResponse.Bodies))
	for _, body := range apiResponse.Bodies {
		if err := validateCelestialBody(body); err == nil {
			// Normalize: ensure EnglishName is populated (fallback to Name, then ID)
			if body.EnglishName == "" {
				if body.Name != "" {
					body.EnglishName = body.Name
				} else if body.ID != "" {
					body.EnglishName = body.ID
				}
			}
			validBodies = append(validBodies, body)
		}
	}

	if len(validBodies) == 0 {
		return nil, fmt.Errorf("no valid celestial bodies in API response")
	}

	return validBodies, nil
}

// sanitizeJSONResponse replaces literal control characters (\n, \r, \t) in JSON string values
// with spaces. This is necessary because the Solar System API returns malformed JSON with
// unescaped newlines in fields like discoveredBy (e.g., "William Cranch Bond,\nGeorge Phillips Bond").
func sanitizeJSONResponse(jsonBytes []byte) []byte {
	result := make([]byte, 0, len(jsonBytes))
	inString := false
	escapeNext := false

	for i := 0; i < len(jsonBytes); i++ {
		b := jsonBytes[i]

		if escapeNext {
			result = append(result, b)
			escapeNext = false
			continue
		}

		if b == '\\' {
			escapeNext = true
			result = append(result, b)
			continue
		}

		if b == '"' {
			inString = !inString
			result = append(result, b)
			continue
		}

		if inString && (b == '\n' || b == '\r' || b == '\t') {
			// Replace control characters with spaces when inside a string
			result = append(result, ' ')
			continue
		}

		result = append(result, b)
	}

	return result
}

func (c *Client) GetBody(id string) (*models.CelestialBody, error) {
	if id == "" {
		return nil, fmt.Errorf("body id is empty")
	}

	path := fmt.Sprintf("/bodies/%s", url.PathEscape(id))
	body, err := c.doGET(path)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch body %s: %w", id, err)
	}

	var celestialBody models.CelestialBody
	if err := json.Unmarshal(body, &celestialBody); err != nil {
		return nil, fmt.Errorf("failed to unmarshal body %s: %w", id, err)
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

	planets := make([]models.CelestialBody, 0)
	for _, body := range bodies {
		if body.IsPlanet {
			planets = append(planets, body)
		}
	}

	return planets, nil
}

func (c *Client) GetBodiesWithFilter(filter string) ([]models.CelestialBody, error) {
	if filter == "" {
		return nil, fmt.Errorf("filter is empty")
	}

	path := fmt.Sprintf("/bodies?filter[]=%s", url.QueryEscape(filter))
	body, err := c.doGET(path)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch filtered bodies: %w", err)
	}

	var apiResponse models.APIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal filtered response: %w", err)
	}

	if err := validateAPIResponse(apiResponse); err != nil {
		return nil, fmt.Errorf("invalid filtered API response: %w", err)
	}

	return apiResponse.Bodies, nil
}

func (c *Client) GetMoonData(moonID string) (*models.CelestialBody, error) {
	if moonID == "" {
		return nil, fmt.Errorf("moon ID is empty")
	}

	return c.GetBody(moonID)
}

/*
------------------------------------------------
Validation
------------------------------------------------
*/

func validateAPIResponse(response models.APIResponse) error {
	if response.Bodies == nil {
		return fmt.Errorf("API response missing bodies field")
	}

	if len(response.Bodies) > MaxBodiesCount {
		return fmt.Errorf(
			"API response contains too many celestial bodies: %d (max: %d)",
			len(response.Bodies),
			MaxBodiesCount,
		)
	}

	// Individual body validation happens during filtering in GetAllBodies()
	return nil
}

func validateCelestialBody(body models.CelestialBody) error {
	// Must have at least one name identifier
	if body.EnglishName == "" && body.Name == "" && body.ID == "" {
		return fmt.Errorf("celestial body missing all name identifiers")
	}

	displayName := body.EnglishName
	if displayName == "" {
		displayName = body.Name
	}
	if displayName == "" {
		displayName = body.ID
	}

	if body.MeanRadius < 0 {
		return fmt.Errorf("celestial body %s has negative mean radius", displayName)
	}

	if body.SemimajorAxis < 0 {
		return fmt.Errorf("celestial body %s has negative semimajor axis", displayName)
	}

	if body.Density < 0 {
		return fmt.Errorf("celestial body %s has negative density", displayName)
	}

	if body.Gravity < 0 {
		return fmt.Errorf("celestial body %s has negative gravity", displayName)
	}

	if body.Eccentricity < 0 || body.Eccentricity > 1 {
		return fmt.Errorf(
			"celestial body %s has unrealistic eccentricity: %.6f",
			displayName,
			body.Eccentricity,
		)
	}

	return nil
}
