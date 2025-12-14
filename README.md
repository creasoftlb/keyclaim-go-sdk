# KeyClaim Go SDK

Official KeyClaim SDK for Go - MITM protection and challenge validation

## Installation

```bash
go get github.com/creasoftlb/keyclaim-go-sdk
```

## Usage

### Basic Example

```go
package main

import (
    "fmt"
    "github.com/creasoftlb/keyclaim-go-sdk"
)

func main() {
    // Create client
    client, err := keyclaim.NewClient("kc_your_api_key")
    if err != nil {
        panic(err)
    }

    // Create challenge
    challenge, err := client.CreateChallenge(30)
    if err != nil {
        panic(err)
    }

    // Generate response using HMAC
    response, err := client.GenerateResponse(challenge.Challenge, keyclaim.ResponseMethodHMAC, nil)
    if err != nil {
        panic(err)
    }

    // Validate challenge
    result, err := client.ValidateChallenge(challenge.Challenge, response, nil)
    if err != nil {
        panic(err)
    }

    if result.IsValid() {
        fmt.Println("Validation successful!")
    }
}
```

### Complete Flow

```go
// Complete flow: create, generate, and validate in one call
result, err := client.Validate(keyclaim.ResponseMethodHMAC, 30, nil)
if err != nil {
    panic(err)
}

if result.IsValid() {
    fmt.Println("Validation successful!")
    fmt.Printf("Quota remaining: %d\n", result.Quota.Remaining)
}
```

### Response Methods

```go
// Echo (testing only)
echoResponse, _ := client.GenerateResponse(challenge, keyclaim.ResponseMethodEcho, nil)

// HMAC (recommended)
hmacResponse, _ := client.GenerateResponse(challenge, keyclaim.ResponseMethodHMAC, nil)

// Hash
hashResponse, _ := client.GenerateResponse(challenge, keyclaim.ResponseMethodHash, nil)

// Custom with string data
customResponse, _ := client.GenerateResponse(
    challenge,
    keyclaim.ResponseMethodCustom,
    "custom-data",
)

// Custom with object data
customData := map[string]interface{}{
    "userId":    "123",
    "timestamp": time.Now().Unix(),
}
customObjectResponse, _ := client.GenerateResponse(
    challenge,
    keyclaim.ResponseMethodCustom,
    customData,
)
```

### Error Handling

```go
challenge, err := client.CreateChallenge(30)
if err != nil {
    if keyclaimErr, ok := err.(*keyclaim.KeyClaimError); ok {
        fmt.Printf("Error: %s\n", keyclaimErr.Message)
        fmt.Printf("Code: %s\n", keyclaimErr.Code)
        fmt.Printf("Status: %d\n", keyclaimErr.StatusCode)
    } else {
        fmt.Printf("Error: %v\n", err)
    }
    return
}
```

### Using Config

```go
config := keyclaim.Config{
    APIKey: "kc_your_api_key",
    Secret: "custom-secret", // Optional, defaults to API key
}

client, err := keyclaim.NewClientWithConfig(config)
if err != nil {
    panic(err)
}
```

## API Reference

### KeyClaimClient

#### Constructor Functions

```go
// With API key only (secret defaults to API key)
client, err := keyclaim.NewClient("kc_your_api_key")

// With API key and custom secret
client, err := keyclaim.NewClientWithSecret("kc_your_api_key", "custom-secret")

// With config object
config := keyclaim.Config{
    APIKey: "kc_your_api_key",
    Secret: "custom-secret",
}
client, err := keyclaim.NewClientWithConfig(config)
```

#### Methods

- `CreateChallenge(ttl int) (*CreateChallengeResponse, error)` - Create a new challenge
- `GenerateResponse(challenge string, method ResponseMethod, customData interface{}) (string, error)` - Generate response
- `ValidateChallenge(challenge, response string, decryptedChallenge *string) (*ValidateChallengeResponse, error)` - Validate challenge
- `Validate(method ResponseMethod, ttl int, customData interface{}) (*ValidateChallengeResponse, error)` - Complete flow

### ResponseMethod Constants

- `keyclaim.ResponseMethodEcho` - Echo the challenge (testing only)
- `keyclaim.ResponseMethodHMAC` - HMAC-SHA256 (recommended)
- `keyclaim.ResponseMethodHash` - SHA-256 hash
- `keyclaim.ResponseMethodCustom` - Custom hash with data

### Types

- `CreateChallengeResponse` - Challenge creation response
- `ValidateChallengeResponse` - Validation response
- `Quota` - Quota information
- `KeyClaimError` - Custom error type
- `Config` - Client configuration

## Requirements

- Go 1.21 or higher

## Testing

```bash
go test ./...
```

## License

MIT License - see [LICENSE](LICENSE) file for details

## Support

For support, email support@keyclaim.org or visit https://keyclaim.org

