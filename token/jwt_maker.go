package token

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

const minSecretKeySize = 32

type JWTTokenMaker struct {
	secretKey string
}

func NewJWTTokenMaker(secretKey string) *JWTTokenMaker {
	if len(secretKey) < minSecretKeySize {
		panic("invalid secret key size: must be at least 32 characters")
	}

	return &JWTTokenMaker{secretKey: secretKey}
}

func (maker *JWTTokenMaker) CreateToken(username string, duration int64) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	return token.SignedString([]byte(maker.secretKey))
}

func (maker *JWTTokenMaker) CreateRefreshToken(username string, duration int64) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)

	return token.SignedString([]byte(maker.secretKey))
}

func (maker *JWTTokenMaker) VerifyToken(token string) (interface{}, error) {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, ErrInvalidToken
		}

		return []byte(maker.secretKey), nil
	}

	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}

func (maker *JWTTokenMaker) VerifyRefreshToken(token string) (interface{}, error) {
	// For now, refresh tokens use the same verification logic as access tokens
	// In a more sophisticated implementation, you might want different validation rules
	payload, err := maker.VerifyToken(token)
	return payload, err
}
