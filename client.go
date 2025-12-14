package keyclaim

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	defaultBaseURLB64 = "aHR0cHM6Ly9rZXljbGFpbS5vcmc=" // https://keyclaim.org
	defaultTimeout    = 30 * time.Second
	defaultTTL        = 30
)

// ResponseMethod represents the method for generating a response
type ResponseMethod string

const (
	ResponseMethodEcho  ResponseMethod = "echo"
	ResponseMethodHMAC  ResponseMethod = "hmac"
	ResponseMethodHash  ResponseMethod = "hash"
	ResponseMethodCustom ResponseMethod = "custom"
)

// Config holds the configuration for KeyClaimClient
type Config struct {
	APIKey string
	Secret string // Optional, defaults to API key
}

// KeyClaimClient is the main client for interacting with the KeyClaim API
type KeyClaimClient struct {
	apiKey  string
	baseURL string
	secret  string
	client  *http.Client
}

// NewClient creates a new KeyClaimClient with the given API key
func NewClient(apiKey string) (*KeyClaimClient, error) {
	return NewClientWithSecret(apiKey, apiKey)
}

// NewClientWithSecret creates a new KeyClaimClient with API key and custom secret
func NewClientWithSecret(apiKey, secret string) (*KeyClaimClient, error) {
	return NewClientWithConfig(Config{
		APIKey: apiKey,
		Secret: secret,
	})
}

// NewClientWithConfig creates a new KeyClaimClient with a Config struct
func NewClientWithConfig(config Config) (*KeyClaimClient, error) {
	if config.APIKey == "" || !hasPrefix(config.APIKey, "kc_") {
		return nil, fmt.Errorf("invalid API key format. API key must start with \"kc_\"")
	}

	// Decode default base URL from base64
	baseURLBytes, err := base64.StdEncoding.DecodeString(defaultBaseURLB64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base URL: %w", err)
	}
	baseURL := string(baseURLBytes)

	secret := config.Secret
	if secret == "" {
		secret = config.APIKey
	}

	return &KeyClaimClient{
		apiKey:  config.APIKey,
		baseURL: baseURL,
		secret:  secret,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}, nil
}

// CreateChallengeOptions holds options for creating a challenge
type CreateChallengeOptions struct {
	TTL int `json:"ttl,omitempty"`
}

// CreateChallengeResponse represents the response from creating a challenge
type CreateChallengeResponse struct {
	Challenge string `json:"challenge"`
	ExpiresIn int    `json:"expires_in"`
	Encrypted *bool  `json:"encrypted,omitempty"`
}

// CreateChallenge creates a new challenge
func (c *KeyClaimClient) CreateChallenge(ttl int) (*CreateChallengeResponse, error) {
	if ttl == 0 {
		ttl = defaultTTL
	}

	reqBody := map[string]interface{}{
		"ttl": ttl,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/challenge/create", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create challenge: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp, "Failed to create challenge")
	}

	var challengeResp CreateChallengeResponse
	if err := json.NewDecoder(resp.Body).Decode(&challengeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &challengeResp, nil
}

// GenerateResponse generates a response from a challenge using the specified method
func (c *KeyClaimClient) GenerateResponse(challenge string, method ResponseMethod, customData interface{}) (string, error) {
	switch method {
	case ResponseMethodEcho:
		return challenge, nil

	case ResponseMethodHMAC:
		h := hmac.New(sha256.New, []byte(c.secret))
		h.Write([]byte(challenge))
		return hex.EncodeToString(h.Sum(nil)), nil

	case ResponseMethodHash:
		hash := sha256.Sum256([]byte(challenge + c.secret))
		return hex.EncodeToString(hash[:]), nil

	case ResponseMethodCustom:
		if customData == nil {
			return "", fmt.Errorf("custom data is required for custom method")
		}

		var data string
		switch v := customData.(type) {
		case string:
			data = challenge + ":" + v
		default:
			jsonData, err := json.Marshal(customData)
			if err != nil {
				return "", fmt.Errorf("failed to marshal custom data: %w", err)
			}
			data = challenge + ":" + string(jsonData)
		}

		hash := sha256.Sum256([]byte(data))
		return hex.EncodeToString(hash[:]), nil

	default:
		return "", fmt.Errorf("unknown response method: %s", method)
	}
}

// ValidateChallengeOptions holds options for validating a challenge
type ValidateChallengeOptions struct {
	Challenge         string `json:"challenge"`
	Response          string `json:"response"`
	DecryptedChallenge *string `json:"decryptedChallenge,omitempty"`
}

// ValidateChallengeResponse represents the response from validating a challenge
type ValidateChallengeResponse struct {
	Valid    *bool  `json:"valid"`
	Signature *string `json:"signature,omitempty"`
	Quota    *Quota `json:"quota,omitempty"`
	Error    *string `json:"error,omitempty"`
}

// Quota represents quota information
type Quota struct {
	Used      int         `json:"used"`
	Remaining int         `json:"remaining"`
	Quota     interface{} `json:"quota"` // Can be int or "unlimited"
}

// ValidateChallenge validates a challenge-response pair
func (c *KeyClaimClient) ValidateChallenge(challenge, response string, decryptedChallenge *string) (*ValidateChallengeResponse, error) {
	reqBody := ValidateChallengeOptions{
		Challenge: challenge,
		Response:  response,
	}

	if decryptedChallenge != nil {
		reqBody.DecryptedChallenge = decryptedChallenge
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/api/challenge/validate", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to validate challenge: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var validationResp ValidateChallengeResponse
	if err := json.Unmarshal(bodyBytes, &validationResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// If the API returns a validation response (even if invalid), return it
	if resp.StatusCode == http.StatusBadRequest || resp.StatusCode == http.StatusUnprocessableEntity {
		if validationResp.Valid != nil {
			return &validationResp, nil
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponseFromBody(bodyBytes, resp.StatusCode, "Failed to validate challenge")
	}

	return &validationResp, nil
}

// Validate completes the full flow: create challenge, generate response, and validate
func (c *KeyClaimClient) Validate(method ResponseMethod, ttl int, customData interface{}) (*ValidateChallengeResponse, error) {
	// Create challenge
	challenge, err := c.CreateChallenge(ttl)
	if err != nil {
		return nil, err
	}

	// Generate response
	response, err := c.GenerateResponse(challenge.Challenge, method, customData)
	if err != nil {
		return nil, err
	}

	// Validate
	return c.ValidateChallenge(challenge.Challenge, response, nil)
}

// IsValid checks if a validation response is valid
func (v *ValidateChallengeResponse) IsValid() bool {
	return v.Valid != nil && *v.Valid
}

// KeyClaimError represents an error from the KeyClaim API
type KeyClaimError struct {
	Message    string
	Code       string
	StatusCode int
}

func (e *KeyClaimError) Error() string {
	return e.Message
}

func (c *KeyClaimClient) handleErrorResponse(resp *http.Response, defaultMessage string) error {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &KeyClaimError{
			Message:    defaultMessage,
			StatusCode: resp.StatusCode,
		}
	}
	return c.handleErrorResponseFromBody(bodyBytes, resp.StatusCode, defaultMessage)
}

func (c *KeyClaimClient) handleErrorResponseFromBody(bodyBytes []byte, statusCode int, defaultMessage string) error {
	var errorData map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &errorData); err != nil {
		return &KeyClaimError{
			Message:    defaultMessage,
			StatusCode: statusCode,
		}
	}

	var errorMessage string
	var errorCode string

	if err, ok := errorData["error"].(string); ok {
		errorMessage = err
		errorCode = err
	} else if msg, ok := errorData["message"].(string); ok {
		errorMessage = msg
	} else {
		errorMessage = defaultMessage
	}

	return &KeyClaimError{
		Message:    errorMessage,
		Code:       errorCode,
		StatusCode: statusCode,
	}
}

// Helper function to check prefix (for Go 1.20 compatibility)
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

