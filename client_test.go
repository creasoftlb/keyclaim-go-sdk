package keyclaim

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	client, err := NewClient("kc_test123456789012345678901234567890123456789012345678901234567890")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if client == nil {
		t.Fatal("Expected client to be created")
	}
}

func TestNewClient_InvalidAPIKey(t *testing.T) {
	_, err := NewClient("invalid-key")
	if err == nil {
		t.Fatal("Expected error for invalid API key")
	}
}

func TestNewClient_EmptyAPIKey(t *testing.T) {
	_, err := NewClient("")
	if err == nil {
		t.Fatal("Expected error for empty API key")
	}
}

func TestCreateChallenge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/challenge/create" {
			t.Errorf("Expected path /api/challenge/create, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		response := CreateChallengeResponse{
			Challenge: "test-challenge-123",
			ExpiresIn: 30,
			Encrypted: boolPtr(false),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, _ := NewClient("kc_test123456789012345678901234567890123456789012345678901234567890")
	client.baseURL = server.URL

	challenge, err := client.CreateChallenge(30)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if challenge.Challenge != "test-challenge-123" {
		t.Errorf("Expected challenge 'test-challenge-123', got %s", challenge.Challenge)
	}
}

func TestGenerateResponse_Echo(t *testing.T) {
	client, _ := NewClient("kc_test123456789012345678901234567890123456789012345678901234567890")
	
	challenge := "test-challenge"
	response, err := client.GenerateResponse(challenge, ResponseMethodEcho, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if response != challenge {
		t.Errorf("Expected %s, got %s", challenge, response)
	}
}

func TestGenerateResponse_HMAC(t *testing.T) {
	client, _ := NewClientWithSecret("kc_test123456789012345678901234567890123456789012345678901234567890", "test-secret")
	
	challenge := "test-challenge"
	response, err := client.GenerateResponse(challenge, ResponseMethodHMAC, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(response) != 64 {
		t.Errorf("Expected HMAC length 64, got %d", len(response))
	}
}

func TestGenerateResponse_Hash(t *testing.T) {
	client, _ := NewClientWithSecret("kc_test123456789012345678901234567890123456789012345678901234567890", "test-secret")
	
	challenge := "test-challenge"
	response, err := client.GenerateResponse(challenge, ResponseMethodHash, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(response) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(response))
	}
}

func TestGenerateResponse_Custom(t *testing.T) {
	client, _ := NewClient("kc_test123456789012345678901234567890123456789012345678901234567890")
	
	challenge := "test-challenge"
	customData := "custom-string"
	response, err := client.GenerateResponse(challenge, ResponseMethodCustom, customData)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if len(response) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(response))
	}
}

func TestGenerateResponse_Custom_NoData(t *testing.T) {
	client, _ := NewClient("kc_test123456789012345678901234567890123456789012345678901234567890")
	
	_, err := client.GenerateResponse("test-challenge", ResponseMethodCustom, nil)
	if err == nil {
		t.Fatal("Expected error for custom method without data")
	}
}

func TestValidateChallenge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ValidateChallengeResponse{
			Valid: boolPtr(true),
			Quota: &Quota{
				Used:      10,
				Remaining: 90,
				Quota:     100,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, _ := NewClient("kc_test123456789012345678901234567890123456789012345678901234567890")
	client.baseURL = server.URL

	result, err := client.ValidateChallenge("test-challenge", "test-response", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !result.IsValid() {
		t.Error("Expected validation to be valid")
	}
}

func TestValidateChallenge_Invalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		response := ValidateChallengeResponse{
			Valid: boolPtr(false),
			Error:  stringPtr("Invalid response"),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, _ := NewClient("kc_test123456789012345678901234567890123456789012345678901234567890")
	client.baseURL = server.URL

	result, err := client.ValidateChallenge("test-challenge", "test-response", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result.IsValid() {
		t.Error("Expected validation to be invalid")
	}
}

func TestValidate(t *testing.T) {
	createServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := CreateChallengeResponse{
			Challenge: "test-challenge-123",
			ExpiresIn: 30,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer createServer.Close()

	validateServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := ValidateChallengeResponse{
			Valid: boolPtr(true),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer validateServer.Close()

	client, _ := NewClient("kc_test123456789012345678901234567890123456789012345678901234567890")
	client.baseURL = createServer.URL

	// Override baseURL for validation (in real usage, both would use the same baseURL)
	originalBaseURL := client.baseURL
	client.baseURL = validateServer.URL

	result, err := client.Validate(ResponseMethodHMAC, 30, nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if !result.IsValid() {
		t.Error("Expected validation to be valid")
	}

	client.baseURL = originalBaseURL
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}

