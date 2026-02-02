package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint `json:"user_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}



func GenerateAccessToken(userID uint,role,Accesssecret string,Accessexpiry time.Duration) (string,error){
	claims:= Claims{
		UserID:userID,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(Accessexpiry)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
		
	}
	token:= jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
		return token.SignedString([]byte(Accesssecret))
}

func ValidateAccessToken(tokenStr, Accesssecret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _,ok:= token.Method.(*jwt.SigningMethodHMAC);!ok{
	     	return nil,errors.New("unexpected signing method")
	}
			return []byte(Accesssecret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}



func GenerateRefreshToken(userID uint,role,Refreshsecret string,Refreshexpiry time.Duration) (string,error){
	claims:= Claims{
		UserID:userID,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(Refreshexpiry)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
		
	}
	token:= jwt.NewWithClaims(jwt.SigningMethodHS256,claims)
		return token.SignedString([]byte(Refreshsecret))
}

func ValidateRefreshToken(tokenStr, Refreshsecret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			if _,ok:= token.Method.(*jwt.SigningMethodHMAC);!ok{
	    	return nil,errors.New("unexpected signing method")
	}
			return []byte(Refreshsecret), nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}