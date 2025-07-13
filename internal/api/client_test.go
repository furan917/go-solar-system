package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/furan917/go-solar-system/internal/models"
)

func TestClient_GetAllBodies(t *testing.T) {
	mockResponse := models.APIResponse{
		Bodies: []models.CelestialBody{
			{
				ID:          "terre",
				Name:        "Terre",
				EnglishName: "Earth",
				IsPlanet:    true,
				MeanRadius:  6371.0,
				Mass: models.Mass{
					MassValue:    5.97237,
					MassExponent: 24,
				},
			},
			{
				ID:          "mars",
				Name:        "Mars",
				EnglishName: "Mars",
				IsPlanet:    true,
				MeanRadius:  3389.5,
				Mass: models.Mass{
					MassValue:    6.4171,
					MassExponent: 23,
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bodies" {
			t.Errorf("Expected path /bodies, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(mockResponse)
		if err != nil {
			return
		}
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	bodies, err := client.GetAllBodies()
	if err != nil {
		t.Fatalf("GetAllBodies() error = %v", err)
	}

	if len(bodies) != 2 {
		t.Errorf("Expected 2 bodies, got %d", len(bodies))
	}

	if bodies[0].EnglishName != "Earth" {
		t.Errorf("Expected first body to be Earth, got %s", bodies[0].EnglishName)
	}

	if bodies[1].EnglishName != "Mars" {
		t.Errorf("Expected second body to be Mars, got %s", bodies[1].EnglishName)
	}
}

func TestClient_GetBody(t *testing.T) {
	mockBody := models.CelestialBody{
		ID:          "terre",
		Name:        "Terre",
		EnglishName: "Earth",
		IsPlanet:    true,
		MeanRadius:  6371.0,
		Mass: models.Mass{
			MassValue:    5.97237,
			MassExponent: 24,
		},
		Moons: []models.Moon{
			{
				ID:          "lune",
				Name:        "Lune",
				EnglishName: "Moon",
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/bodies/terre" {
			t.Errorf("Expected path /bodies/terre, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(mockBody)
		if err != nil {
			return
		}
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	body, err := client.GetBody("terre")
	if err != nil {
		t.Fatalf("GetBody() error = %v", err)
	}

	if body.EnglishName != "Earth" {
		t.Errorf("Expected body to be Earth, got %s", body.EnglishName)
	}

	if len(body.Moons) != 1 {
		t.Errorf("Expected 1 moon, got %d", len(body.Moons))
	}

	if body.Moons[0].EnglishName != "Moon" {
		t.Errorf("Expected moon to be Moon, got %s", body.Moons[0].EnglishName)
	}
}

func TestClient_GetPlanets(t *testing.T) {
	mockResponse := models.APIResponse{
		Bodies: []models.CelestialBody{
			{
				ID:          "terre",
				Name:        "Terre",
				EnglishName: "Earth",
				IsPlanet:    true,
			},
			{
				ID:          "lune",
				Name:        "Lune",
				EnglishName: "Moon",
				IsPlanet:    false,
			},
			{
				ID:          "mars",
				Name:        "Mars",
				EnglishName: "Mars",
				IsPlanet:    true,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(mockResponse)
		if err != nil {
			return
		}
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	planets, err := client.GetPlanets()
	if err != nil {
		t.Fatalf("GetPlanets() error = %v", err)
	}

	if len(planets) != 2 {
		t.Errorf("Expected 2 planets, got %d", len(planets))
	}

	planetNames := make([]string, len(planets))
	for i, planet := range planets {
		planetNames[i] = planet.EnglishName
	}

	expectedNames := []string{"Earth", "Mars"}
	for i, expected := range expectedNames {
		if planetNames[i] != expected {
			t.Errorf("Expected planet %d to be %s, got %s", i, expected, planetNames[i])
		}
	}
}

func TestClient_GetBody_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	_, err := client.GetBody("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent body, got nil")
	}
}

func TestClient_GetBodiesWithFilter(t *testing.T) {
	mockResponse := models.APIResponse{
		Bodies: []models.CelestialBody{
			{
				ID:          "terre",
				Name:        "Terre",
				EnglishName: "Earth",
				IsPlanet:    true,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "filter[]=isPlanet%2Ceq%2Ctrue" {
			t.Errorf("Expected query parameter filter[], got %s", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(mockResponse)
		if err != nil {
			return
		}
	}))
	defer server.Close()

	client := NewClient()
	client.baseURL = server.URL

	bodies, err := client.GetBodiesWithFilter("isPlanet,eq,true")
	if err != nil {
		t.Fatalf("GetBodiesWithFilter() error = %v", err)
	}

	if len(bodies) != 1 {
		t.Errorf("Expected 1 body, got %d", len(bodies))
	}

	if bodies[0].EnglishName != "Earth" {
		t.Errorf("Expected body to be Earth, got %s", bodies[0].EnglishName)
	}
}
