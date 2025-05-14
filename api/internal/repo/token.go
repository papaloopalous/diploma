package repo

import (
	"api/internal/messages"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type MyClaims struct {
	SessionID uuid.UUID `json:"sessionID"`
	jwt.RegisteredClaims
}

type TokenData struct {
	key []byte
}

type TokenRepo interface {
	GetData() (token []byte)
	SetData(token string)
	GenerateJWT(sessionID uuid.UUID) (string, error)
	ParseJWT(tokenString string) (*MyClaims, error)
}

var _ TokenRepo = &TokenData{}

func NewTokenRepo() *TokenData {
	return &TokenData{
		key: make([]byte, 0),
	}
}

func (p *TokenData) GetData() (token []byte) {
	res := p.key
	return res
}

func (p *TokenData) SetData(token string) {
	p.key = []byte(token)
}

func (p *TokenData) GenerateJWT(sessionID uuid.UUID) (string, error) {
	claims := MyClaims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	res, err := token.SignedString(p.GetData())
	if err != nil {
		return res, errors.New(messages.ErrGenToken)
	}

	return res, nil
}

func (p *TokenData) ParseJWT(tokenString string) (*MyClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &MyClaims{}, func(token *jwt.Token) (interface{}, error) {
		return p.GetData(), nil
	})

	if err != nil {
		return nil, errors.New(messages.ErrParseToken)
	}

	if claims, ok := token.Claims.(*MyClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New(messages.ErrBadToken)
}
