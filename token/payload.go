package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	ExpiredAt int64     `json:"expired_at"`
	IssuedAt  int64     `json:"issued_at"`
}

func NewPayload(username string, duration int64) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now().Unix(),
		ExpiredAt: time.Now().Add(time.Duration(duration) * time.Second).Unix(),
	}

	return payload, nil
}

func (payload *Payload) Valid() error {
	if time.Now().Unix() > payload.ExpiredAt {
		return jwt.ErrTokenExpired
	}

	if payload.ID == uuid.Nil {
		return jwt.ErrTokenInvalidId
	}

	if payload.Username == "" {
		return jwt.ErrTokenRequiredClaimMissing
	}

	return nil
}

func (payload *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(payload.ExpiredAt, 0)), nil
}

func (payload *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (payload *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(payload.IssuedAt, 0)), nil
}

func (payload *Payload) GetIssuer() (string, error) {
	return "", nil
}

func (payload *Payload) GetSubject() (string, error) {
	return "", nil
}

func (payload *Payload) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}
