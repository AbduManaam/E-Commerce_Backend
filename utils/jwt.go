package utils

import (
	"errors"
	"fmt"
	"log"
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

	log.Printf("Generated access token: userID=%d role=%s expiresAt=%v", userID, role, claims.ExpiresAt.Time)
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
		// jwt/v5 returns specific errors
		if errors.Is(err, jwt.ErrTokenExpired) {
			log.Printf("Access token expired: %v", err)
			return nil, errors.New("token expired")
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			log.Printf("Access token signature invalid: %v", err)
			return nil, errors.New("invalid token signature")
		}
		log.Printf("Access token parse error: %v", err)
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		log.Printf("Access token claims invalid or token not valid")
		return nil, errors.New("invalid token")
	}

	log.Printf("Access token valid: userID=%d role=%s expiresAt=%v", claims.UserID, claims.Role, claims.ExpiresAt.Time)
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

	log.Printf("Generated refresh token: userID=%d role=%s expiresAt=%v", userID, role, claims.ExpiresAt.Time)
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
			log.Printf("Refresh token expired: %v", err)
			return nil, errors.New("token expired")
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			log.Printf("Refresh token signature invalid: %v", err)
			return nil, errors.New("invalid token signature")
		}
		log.Printf("Refresh token parse error: %v", err)
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		log.Printf("Refresh token claims invalid or token not valid")
		return nil, errors.New("invalid token")
	}

	log.Printf("Refresh token valid: userID=%d role=%s expiresAt=%v", claims.UserID, claims.Role, claims.ExpiresAt.Time)
	return claims, nil
}