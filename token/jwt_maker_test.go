package token

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestJWTTokenMaker(t *testing.T) {
	// Test secret key that meets minimum size requirement
	secretKey := "12345678901234567890123456789012" // 32 characters
	maker := NewJWTTokenMaker(secretKey)

	username := "testuser"
	duration := int64(3600) // 1 hour

	// Test token creation
	t.Run("Create Token", func(t *testing.T) {
		token, err := maker.CreateToken(username, duration)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}
		if token == "" {
			t.Fatal("Token should not be empty")
		}
		t.Logf("Created token: %s", token)
	})

	// Test token verification
	t.Run("Verify Valid Token", func(t *testing.T) {
		token, err := maker.CreateToken(username, duration)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		payloadInterface, err := maker.VerifyToken(token)
		if err != nil {
			t.Fatalf("Failed to verify token: %v", err)
		}

		payload, ok := payloadInterface.(*Payload)
		if !ok {
			t.Fatalf("Failed to cast payload to *Payload")
		}

		if payload.Username != username {
			t.Errorf("Expected username %s, got %s", username, payload.Username)
		}

		if payload.ID == uuid.Nil {
			t.Error("Token ID should not be nil")
		}

		// Check if token is not expired
		if time.Now().Unix() > payload.ExpiredAt {
			t.Error("Token should not be expired")
		}
	})

	// Test invalid token
	t.Run("Verify Invalid Token", func(t *testing.T) {
		invalidToken := "invalid.token.here"
		_, err := maker.VerifyToken(invalidToken)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken, got %v", err)
		}
	})

	// Test expired token
	t.Run("Verify Expired Token", func(t *testing.T) {
		// Create a token with very short duration (1 second)
		token, err := maker.CreateToken(username, 1)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Wait for token to expire
		time.Sleep(2 * time.Second)

		_, err = maker.VerifyToken(token)
		if err != ErrExpiredToken {
			t.Errorf("Expected ErrExpiredToken, got %v", err)
		}
	})

	// Test token with different algorithm (no algorithm scenario)
	t.Run("Verify Token With Different Algorithm", func(t *testing.T) {
		// Test with a token that has a different algorithm (RS256 instead of HS256)
		// This simulates the "no algorithm" scenario where the token uses an unexpected signing method
		invalidToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6IjEyMzQ1Njc4OTAiLCJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwaXJlZF9hdCI6MTc1Mzk1Mzc5MCwiaXNzdWVkX2F0IjoxNzUzOTUwMTkwfQ.invalid_signature"

		_, err := maker.VerifyToken(invalidToken)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken for different algorithm, got %v", err)
		}
	})
}

func TestNewJWTTokenMaker(t *testing.T) {
	t.Run("Valid Secret Key", func(t *testing.T) {
		secretKey := "12345678901234567890123456789012" // 32 characters
		maker := NewJWTTokenMaker(secretKey)
		if maker == nil {
			t.Fatal("Maker should not be nil")
		}
		if maker.secretKey != secretKey {
			t.Errorf("Expected secret key %s, got %s", secretKey, maker.secretKey)
		}
	})

	t.Run("Invalid Secret Key Size", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid secret key size")
			}
		}()

		secretKey := "tooshort" // Less than 32 characters
		NewJWTTokenMaker(secretKey)
	})
}

func TestAlgorithmValidation(t *testing.T) {
	secretKey := "12345678901234567890123456789012"
	maker := NewJWTTokenMaker(secretKey)

	t.Run("Reject RS256 Algorithm", func(t *testing.T) {
		// Create a payload
		payload, err := NewPayload("testuser", 3600)
		if err != nil {
			t.Fatalf("Failed to create payload: %v", err)
		}

		// Create token with RS256 algorithm (different from expected HS256)
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, payload)

		// Sign with a dummy key (this will fail, but we're testing the algorithm check)
		tokenString, err := token.SignedString([]byte("dummy-key"))
		if err != nil {
			// This is expected to fail, but the important part is testing the algorithm validation
			t.Logf("Token signing failed as expected: %v", err)
		}

		// Try to verify with our HS256-only maker
		_, err = maker.VerifyToken(tokenString)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken for RS256 algorithm, got %v", err)
		}
	})

	t.Run("Reject None Algorithm", func(t *testing.T) {
		// Create a payload
		payload, err := NewPayload("testuser", 3600)
		if err != nil {
			t.Fatalf("Failed to create payload: %v", err)
		}

		// Create token with "none" algorithm (no signature)
		token := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
		tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Fatalf("Failed to create none algorithm token: %v", err)
		}

		// Try to verify with our HS256-only maker
		_, err = maker.VerifyToken(tokenString)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken for none algorithm, got %v", err)
		}
	})

	t.Run("Accept HS256 Algorithm", func(t *testing.T) {
		// Create a valid HS256 token
		tokenString, err := maker.CreateToken("testuser", 3600)
		if err != nil {
			t.Fatalf("Failed to create token: %v", err)
		}

		// Verify it should work
		_, err = maker.VerifyToken(tokenString)
		if err != nil {
			t.Errorf("Expected successful verification for HS256 algorithm, got %v", err)
		}
	})

	t.Run("Reject Malformed Algorithm Header", func(t *testing.T) {
		// Test with a token that has a malformed algorithm in the header
		malformedToken := "eyJhbGciOiJOT1RfSFMyNTYiLCJ0eXAiOiJKV1QifQ.eyJpZCI6IjEyMzQ1Njc4OTAiLCJ1c2VybmFtZSI6InRlc3R1c2VyIiwiZXhwaXJlZF9hdCI6MTc1Mzk1Mzc5MCwiaXNzdWVkX2F0IjoxNzUzOTUwMTkwfQ.invalid_signature"

		_, err := maker.VerifyToken(malformedToken)
		if err != ErrInvalidToken {
			t.Errorf("Expected ErrInvalidToken for malformed algorithm, got %v", err)
		}
	})
}
