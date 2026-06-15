package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mykytakuzminov/sprintly-api/internal/config"
)

var (
	ErrInvalidToken         = errors.New("invalid token")
	ErrInvalidSigningMethod = errors.New("unexpected signing method")
)

type Auth struct {
	cfg *config.JWTConfig
}

func NewAuth(cfg *config.JWTConfig) *Auth {
	return &Auth{cfg: cfg}
}

func (a *Auth) RefreshTTL() time.Duration {
	return a.cfg.RefreshTTL
}

func (a *Auth) GenerateAccessToken(userID uuid.UUID) (string, error) {
	token, err := a.generateJWT(userID, a.cfg.AccessTTL)
	if err != nil {
		return "", err
	}

	return token, err
}

func (a *Auth) GenerateRefreshToken(userID uuid.UUID) (string, error) {

	token, err := a.generateJWT(userID, a.cfg.RefreshTTL)
	if err != nil {
		return "", err
	}

	return token, err
}

func (a *Auth) ParseToken(token string) (uuid.UUID, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(a.cfg.Secret), nil
	})
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}

func (a *Auth) generateJWT(userID uuid.UUID, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(a.cfg.Secret))
	if err != nil {
		return "", err
	}

	return signed, nil
}
