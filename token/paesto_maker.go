package token

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/o1egl/paseto"
)

type PasetoMaker struct {
	secretKey []byte
}

func NewPasetoMaker(secretKey string) *PasetoMaker {
	return &PasetoMaker{secretKey: []byte(secretKey)}
}

func (maker *PasetoMaker) CreateToken(username string, duration int64) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	// Create a PASETO v2 JSON token
	token := &paseto.JSONToken{
		Jti:        payload.ID.String(),
		Subject:    payload.Username,
		Expiration: time.Unix(payload.ExpiredAt, 0),
		IssuedAt:   time.Unix(payload.IssuedAt, 0),
		NotBefore:  time.Unix(payload.IssuedAt, 0),
	}

	// Create V2 instance and encrypt the token
	v2 := paseto.NewV2()
	signedToken, err := v2.Encrypt(maker.secretKey, token, nil)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt token: %w", err)
	}

	return signedToken, nil
}

func (maker *PasetoMaker) CreateRefreshToken(username string, duration int64) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	// Create a PASETO v2 JSON token
	token := &paseto.JSONToken{
		Jti:        payload.ID.String(),
		Subject:    payload.Username,
		Expiration: time.Unix(payload.ExpiredAt, 0),
		IssuedAt:   time.Unix(payload.IssuedAt, 0),
		NotBefore:  time.Unix(payload.IssuedAt, 0),
	}

	// Create V2 instance and encrypt the token
	v2 := paseto.NewV2()
	signedToken, err := v2.Encrypt(maker.secretKey, token, nil)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt token: %w", err)
	}

	return signedToken, nil
}

func (maker *PasetoMaker) VerifyToken(tokenString string) (interface{}, error) {
	// Create V2 instance and decrypt the token
	v2 := paseto.NewV2()
	var token paseto.JSONToken

	err := v2.Decrypt(tokenString, maker.secretKey, &token, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Create payload from token claims
	payload := &Payload{
		Username:  token.Subject,
		ExpiredAt: token.Expiration.Unix(),
		IssuedAt:  token.IssuedAt.Unix(),
	}

	// Parse JTI as UUID
	if token.Jti != "" {
		id, err := uuid.Parse(token.Jti)
		if err != nil {
			return nil, fmt.Errorf("invalid token id: %w", err)
		}
		payload.ID = id
	}

	// Validate the payload
	if err := payload.Valid(); err != nil {
		return nil, err
	}

	return payload, nil
}

func (maker *PasetoMaker) VerifyRefreshToken(tokenString string) (interface{}, error) {
	// For now, refresh tokens use the same verification logic as access tokens
	// In a more sophisticated implementation, you might want different validation rules
	payload, err := maker.VerifyToken(tokenString)
	return payload, err
}
