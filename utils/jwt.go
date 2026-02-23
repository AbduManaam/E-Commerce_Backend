package utils

import (
	"backend/utils/logging"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// ------------------------- ACCESS TOKEN -------------------------

func GenerateAccessToken(userID uint, role, accessSecret string, accessExpiry time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessExpiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(accessSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}

	logging.LogDebug("access token generated", "user_id", userID, "role", role)
	return signed, nil
}

func ValidateAccessToken(tokenStr, accessSecret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(accessSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			logging.LogDebug("access token expired", "error", err)
			return nil, errors.New("token expired")
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			logging.LogWarn("access token signature invalid", "error", err)
			return nil, errors.New("invalid token signature")
		}
		logging.LogDebug("access token parse error", "error", err)
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		logging.LogWarn("access token claims invalid")
		return nil, errors.New("invalid token")
	}

	logging.LogDebug("access token valid", "user_id", claims.UserID, "role", claims.Role)
	return claims, nil
}

// ------------------------- REFRESH TOKEN -------------------------

func GenerateRefreshToken(userID uint, role, refreshSecret string, refreshExpiry time.Duration) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(refreshExpiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(refreshSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	logging.LogDebug("refresh token generated", "user_id", userID, "role", role)
	return signed, nil
}

func ValidateRefreshToken(tokenStr, refreshSecret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(refreshSecret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			logging.LogDebug("refresh token expired", "error", err)
			return nil, errors.New("token expired")
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			logging.LogWarn("refresh token signature invalid", "error", err)
			return nil, errors.New("invalid token signature")
		}
		logging.LogDebug("refresh token parse error", "error", err)
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		logging.LogWarn("refresh token claims invalid")
		return nil, errors.New("invalid token")
	}

	logging.LogDebug("refresh token valid", "user_id", claims.UserID, "role", claims.Role)
	return claims, nil
}
