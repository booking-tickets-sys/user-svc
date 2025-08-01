package token

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPasetoTokenMaker(t *testing.T) {
	maker := NewPasetoMaker("12345678901234567890123456789012") // 32 bytes for AES-256

	username := "testuser"
	duration := int64(2)

	t.Run("Create_Token", func(t *testing.T) {
		token, err := maker.CreateToken(username, duration)
		require.NoError(t, err)
		require.NotEmpty(t, token)
		t.Logf("Created PASETO token: %s", token)
	})

	t.Run("Verify_Valid_Token", func(t *testing.T) {
		token, err := maker.CreateToken(username, duration)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		payloadInterface, err := maker.VerifyToken(token)
		require.NoError(t, err)
		require.NotNil(t, payloadInterface)

		payload, ok := payloadInterface.(*Payload)
		require.True(t, ok, "Failed to cast payload to *Payload")
		require.Equal(t, username, payload.Username)
		require.NotEmpty(t, payload.ID)
		require.Greater(t, payload.ExpiredAt, payload.IssuedAt)
	})

	t.Run("Verify_Invalid_Token", func(t *testing.T) {
		invalidToken := "invalid.token.here"
		payload, err := maker.VerifyToken(invalidToken)
		require.Error(t, err)
		require.Nil(t, payload)
	})

	t.Run("Verify_Expired_Token", func(t *testing.T) {
		// Create a token with very short duration
		token, err := maker.CreateToken(username, 1)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Wait for token to expire
		time.Sleep(2 * time.Second)

		payload, err := maker.VerifyToken(token)
		require.Error(t, err)
		require.Nil(t, payload)
	})

	t.Run("Verify_Token_With_Different_Key", func(t *testing.T) {
		token, err := maker.CreateToken(username, duration)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		// Create a different maker with different key
		differentMaker := NewPasetoMaker("differentkey1234567890123456789012")
		payload, err := differentMaker.VerifyToken(token)
		require.Error(t, err)
		require.Nil(t, payload)
	})
}

func TestNewPasetoMaker(t *testing.T) {
	t.Run("Valid_Secret_Key", func(t *testing.T) {
		secretKey := "12345678901234567890123456789012" // 32 bytes
		maker := NewPasetoMaker(secretKey)
		require.NotNil(t, maker)
		require.Equal(t, []byte(secretKey), maker.secretKey)
	})

	t.Run("Short_Secret_Key", func(t *testing.T) {
		// PASETO v2 requires 32 bytes for AES-256, but the library will handle this
		secretKey := "shortkey"
		maker := NewPasetoMaker(secretKey)
		require.NotNil(t, maker)
		require.Equal(t, []byte(secretKey), maker.secretKey)
	})
}

func TestPasetoTokenPayload(t *testing.T) {
	maker := NewPasetoMaker("12345678901234567890123456789012")
	username := "testuser"
	duration := int64(3600) // 1 hour

	t.Run("Token_Payload_Fields", func(t *testing.T) {
		token, err := maker.CreateToken(username, duration)
		require.NoError(t, err)

		payloadInterface, err := maker.VerifyToken(token)
		require.NoError(t, err)
		require.NotNil(t, payloadInterface)

		payload, ok := payloadInterface.(*Payload)
		require.True(t, ok, "Failed to cast payload to *Payload")

		// Check all payload fields
		require.Equal(t, username, payload.Username)
		require.NotEqual(t, uuid.Nil, payload.ID)
		require.Greater(t, payload.ExpiredAt, payload.IssuedAt)
		require.Greater(t, payload.ExpiredAt, time.Now().Unix())
		require.LessOrEqual(t, payload.IssuedAt, time.Now().Unix())
	})

	t.Run("Token_Expiration_Time", func(t *testing.T) {
		startTime := time.Now()
		token, err := maker.CreateToken(username, duration)
		require.NoError(t, err)

		payloadInterface, err := maker.VerifyToken(token)
		require.NoError(t, err)

		payload, ok := payloadInterface.(*Payload)
		require.True(t, ok, "Failed to cast payload to *Payload")

		// Check that expiration is approximately duration seconds from creation
		expectedExpiration := startTime.Add(time.Duration(duration) * time.Second).Unix()
		actualExpiration := payload.ExpiredAt

		// Allow for small time differences (within 5 seconds)
		require.InDelta(t, expectedExpiration, actualExpiration, 5)
	})
}

func TestPasetoTokenSecurity(t *testing.T) {
	maker := NewPasetoMaker("12345678901234567890123456789012")
	username := "testuser"
	duration := int64(3600)

	t.Run("Tokens_Are_Unique", func(t *testing.T) {
		token1, err := maker.CreateToken(username, duration)
		require.NoError(t, err)

		token2, err := maker.CreateToken(username, duration)
		require.NoError(t, err)

		// Tokens should be different due to unique IDs
		require.NotEqual(t, token1, token2)

		// Both tokens should be valid
		payload1Interface, err := maker.VerifyToken(token1)
		require.NoError(t, err)
		require.NotNil(t, payload1Interface)

		payload1, ok := payload1Interface.(*Payload)
		require.True(t, ok, "Failed to cast payload1 to *Payload")

		payload2Interface, err := maker.VerifyToken(token2)
		require.NoError(t, err)
		require.NotNil(t, payload2Interface)

		payload2, ok := payload2Interface.(*Payload)
		require.True(t, ok, "Failed to cast payload2 to *Payload")

		// IDs should be different
		require.NotEqual(t, payload1.ID, payload2.ID)
	})

	t.Run("Token_Format_Validation", func(t *testing.T) {
		token, err := maker.CreateToken(username, duration)
		require.NoError(t, err)

		// PASETO v2 local tokens should start with "v2.local."
		require.Contains(t, token, "v2.local.")
	})
}
