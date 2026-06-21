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
	ErrInvalidClaim         = errors.New("missing or invalid claim")
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

func (a *Auth) GenerateAccessToken(userID uuid.UUID, role string) (string, error) {
	token, err := a.generateJWT(userID, role, a.cfg.AccessTTL)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *Auth) GenerateRefreshToken(userID uuid.UUID, role string) (string, error) {
	token, err := a.generateJWT(userID, role, a.cfg.RefreshTTL)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (a *Auth) ParseToken(token string) (uuid.UUID, string, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(a.cfg.Secret), nil
	})
	if err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return uuid.Nil, "", ErrInvalidToken
	}

	sub, err := a.getStringClaim(claims, "sub")
	if err != nil {
		return uuid.Nil, "", err
	}
	role, err := a.getStringClaim(claims, "role")
	if err != nil {
		return uuid.Nil, "", err
	}

	userID, err := uuid.Parse(sub)
	if err != nil {
		return uuid.Nil, "", ErrInvalidToken
	}

	return userID, role, nil
}

func (a *Auth) generateJWT(
	userID uuid.UUID,
	role string,
	ttl time.Duration,
) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": role,
		"exp":  time.Now().Add(ttl).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(a.cfg.Secret))
	if err != nil {
		return "", err
	}

	return signed, nil
}

func (a *Auth) getStringClaim(claims jwt.MapClaims, key string) (string, error) {
	val, ok := claims[key].(string)
	if !ok {
		return "", ErrInvalidClaim
	}
	return val, nil
}
