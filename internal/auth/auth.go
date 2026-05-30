package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/mykytakuzminov/task-manager-api/internal/config"
)

type Auth struct {
	cfg *config.JWTConfig
}

func NewAuth(cfg *config.JWTConfig) *Auth {
	return &Auth{cfg: cfg}
}

func (a *Auth) GenerateAccessToken(userID uuid.UUID) (string, error) {
	accessTTL, err := time.ParseDuration(a.cfg.AccessTTL)
	if err != nil {
		return "", err
	}

	token, err := a.generateJWT(userID, accessTTL)
	if err != nil {
		return "", err
	}

	return token, err
}

func (a *Auth) GenerateRefreshToken(userID uuid.UUID) (string, error) {
	refreshTTL, err := time.ParseDuration(a.cfg.RefreshTTL)
	if err != nil {
		return "", err
	}

	token, err := a.generateJWT(userID, refreshTTL)
	if err != nil {
		return "", err
	}

	return token, err
}

func (a *Auth) ParseToken(tokenString string) (uuid.UUID, error) {
	parsedToken, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(a.cfg.Secret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	userID, err := uuid.Parse(claims["sub"].(string))
	if err != nil {
		return uuid.Nil, errors.New("invalid user id in token")
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
